package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dmksnnk/octo/internal"
	"github.com/dmksnnk/octo/internal/auth"
)

type API struct {
	service Service
}

type Service interface {
	Products(ctx context.Context, capability internal.CapabilityRequest) ([]internal.Product, error)
	Product(ctx context.Context, id int, capability internal.CapabilityRequest) (internal.Product, error)
	Availability(ctx context.Context, productID int, localDate time.Time, capability internal.CapabilityRequest) (internal.Availability, error)
	Availabilities(ctx context.Context, productID int, localDateStart, localDateEnd time.Time, capability internal.CapabilityRequest) ([]internal.Availability, error)
	// CreateBooking creates a booking for the given product and availability.
	// Return internal.ErrNotAvailable if the product is not available.
	CreateBooking(ctx context.Context, params internal.CreateBookingRequest) (int, error)
	// ConfirmBooking confirms a booking for the given product and availability and generates tickets.
	ConfirmBooking(ctx context.Context, id, userID int) error
	Booking(ctx context.Context, id, userID int, capability internal.CapabilityRequest) (internal.Booking, error)
}

func NewAPI(service Service) API {
	return API{
		service: service,
	}
}

func (a API) Products(w http.ResponseWriter, r *http.Request) {
	var capability CapabilityRequest
	if err := capability.UnmarshalHTTP(r); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest, err.Error())
		return
	}

	products, err := a.service.Products(r.Context(), internal.CapabilityRequest(capability))
	if err != nil {
		writeError(w, "failed to get products", http.StatusInternalServerError, err.Error())
		return
	}

	_ = writeJSON(w, http.StatusOK, products)
}

func (a API) Product(w http.ResponseWriter, r *http.Request) {
	var capability CapabilityRequest
	if err := capability.UnmarshalHTTP(r); err != nil {
		writeError(w, "failed to decode capability", http.StatusBadRequest, err.Error())
		return
	}

	var id IDPathValue
	if err := id.UnmarshalHTTP(r); err != nil {
		writeError(w, "invalid product ID", http.StatusBadRequest, err.Error())
		return
	}

	product, err := a.service.Product(r.Context(), int(id), internal.CapabilityRequest(capability))
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			writeError(w, "product not found", http.StatusNotFound)
			return
		}

		writeError(w, "failed to get product", http.StatusInternalServerError, err.Error())
		return
	}

	_ = writeJSON(w, http.StatusOK, product)
}

func (a API) Availability(w http.ResponseWriter, r *http.Request) {
	var capability CapabilityRequest
	if err := capability.UnmarshalHTTP(r); err != nil {
		writeError(w, "failed to decode capability", http.StatusBadRequest, err.Error())
		return
	}

	var availabilityReq AvailabilityRequest
	if err := availabilityReq.UnmarshalHTTP(r); err != nil {
		writeError(w, "failed to decode availability request", http.StatusBadRequest, err.Error())
		return
	}

	if !time.Time(availabilityReq.LocalDate).IsZero() { // request for a single availability
		availabilities, err := a.singleAvailability(r.Context(), availabilityReq, capability)
		if err != nil {
			writeError(w, "failed to get availability", http.StatusInternalServerError, err.Error())
			return
		}
		_ = writeJSON(w, http.StatusOK, availabilities)
		return
	}

	// request for a range of availabilities
	availabilities, err := a.service.Availabilities(
		r.Context(),
		int(availabilityReq.ProductID),
		time.Time(availabilityReq.LocalDateStart),
		time.Time(availabilityReq.LocalDateEnd),
		internal.CapabilityRequest(capability),
	)
	if err != nil {
		writeError(w, "failed to get availability", http.StatusInternalServerError, err.Error())
		return
	}

	_ = writeJSON(w, http.StatusOK, availabilities)
}

func (a API) singleAvailability(ctx context.Context, req AvailabilityRequest, capability CapabilityRequest) ([]internal.Availability, error) {
	availability, err := a.service.Availability(
		ctx,
		int(req.ProductID),
		time.Time(req.LocalDate),
		internal.CapabilityRequest(capability),
	)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			return []internal.Availability{}, nil // return empty slice
		}

		return nil, err
	}

	return []internal.Availability{availability}, nil
}

func (a API) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var capability CapabilityRequest
	if err := capability.UnmarshalHTTP(r); err != nil {
		writeError(w, "failed to decode capability", http.StatusBadRequest, err.Error())
		return
	}
	var bookingReq BookingRequest
	if err := bookingReq.UnmarshalHTTP(r); err != nil {
		writeError(w, "failed to decode booking request", http.StatusBadRequest, err.Error())
		return
	}

	user, _ := auth.ContextUser(r.Context())
	params := internal.CreateBookingRequest{
		ProductID:      int(bookingReq.ProductID),
		AvailabilityID: int(bookingReq.AvailabilityID),
		Units:          int(bookingReq.Units),
		UserID:         user.ID,
	}
	id, err := a.service.CreateBooking(r.Context(), params)
	if err != nil {
		if errors.Is(err, internal.ErrNotAvailable) {
			writeError(w, "not available", http.StatusConflict)
			return
		}
		writeError(w, "failed to create booking", http.StatusInternalServerError, err.Error())
		return
	}

	a.booking(r.Context(), w, id, user.ID, internal.CapabilityRequest(capability))
}

func (a API) ConfirmBooking(w http.ResponseWriter, r *http.Request) {
	var capability CapabilityRequest
	if err := capability.UnmarshalHTTP(r); err != nil {
		writeError(w, "failed to decode capability", http.StatusBadRequest, err.Error())
		return
	}

	var id IDPathValue
	if err := id.UnmarshalHTTP(r); err != nil {
		writeError(w, "invalid booking ID", http.StatusBadRequest, err.Error())
		return
	}

	user, _ := auth.ContextUser(r.Context())
	if err := a.service.ConfirmBooking(r.Context(), int(id), user.ID); err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			writeError(w, "booking not found", http.StatusNotFound)
			return
		}

		writeError(w, "failed to confirm booking", http.StatusInternalServerError, err.Error())
		return
	}

	a.booking(r.Context(), w, int(id), user.ID, internal.CapabilityRequest(capability))
}

func (a API) Booking(w http.ResponseWriter, r *http.Request) {
	var capability CapabilityRequest
	if err := capability.UnmarshalHTTP(r); err != nil {
		writeError(w, "failed to decode capability", http.StatusBadRequest, err.Error())
		return
	}

	var id IDPathValue
	if err := id.UnmarshalHTTP(r); err != nil {
		writeError(w, "invalid booking ID", http.StatusBadRequest, err.Error())
		return
	}

	user, _ := auth.ContextUser(r.Context())
	a.booking(r.Context(), w, int(id), user.ID, internal.CapabilityRequest(capability))
}

func (a API) booking(ctx context.Context, w http.ResponseWriter, id, userID int, capability internal.CapabilityRequest) {
	booking, err := a.service.Booking(ctx, int(id), userID, internal.CapabilityRequest(capability))
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			writeError(w, "booking not found", http.StatusNotFound)
			return
		}

		writeError(w, "failed to get booking", http.StatusInternalServerError, err.Error())
		return
	}

	_ = writeJSON(w, http.StatusOK, booking)
}

func writeJSON(w http.ResponseWriter, status int, resp any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(resp)
}
