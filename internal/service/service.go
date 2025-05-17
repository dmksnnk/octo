package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dmksnnk/octo/internal"
)

type DB interface {
	// Products returns all products.
	Products(ctx context.Context, capability internal.CapabilityRequest) ([]internal.Product, error)
	// Product returns a product by id.
	// It returns ErrNotFound if the product is not found.
	Product(ctx context.Context, id int, capability internal.CapabilityRequest) (internal.Product, error)
	// Availability returns availability for a product on a given date.
	// It returns ErrNotFound if the product is not found.
	Availability(ctx context.Context, productID int, localDate time.Time, capability internal.CapabilityRequest) (internal.Availability, error)
	// Availabilities returns availabilities for a product in a given date range.
	Availabilities(ctx context.Context, productID int, localDateStart, localDateEnd time.Time, capability internal.CapabilityRequest) ([]internal.Availability, error)
	// CreateBooking creates a booking for a product.
	// It returns ErrNotAvailable if the product is not available for booking.
	CreateBooking(ctx context.Context, params CreateBookingParams) (int, error)
	// ConfirmBooking confirms a booking.
	// It returns ErrNotFound if the booking is not found.
	ConfirmBooking(ctx context.Context, id int, userID int) error
	// Booking returns a booking by id.
	// It returns ErrNotFound if the booking is not found.
	Booking(ctx context.Context, id int, userID int, capability internal.CapabilityRequest) (internal.Booking, error)
}

type CreateBookingParams struct {
	ProductID      int
	AvailabilityID int
	Units          int
	UserID         int
}

var (
	// ErrNotFound is returned by database when a resource is not found.
	ErrNotFound = fmt.Errorf("not found")
	// ErrNotAvailable is returned by database when a product is not available for booking.
	ErrNotAvailable = fmt.Errorf("not available")
)

type Service struct {
	db DB
}

func NewService(db DB) Service {
	return Service{
		db: db,
	}
}

func (s Service) Products(ctx context.Context, capability internal.CapabilityRequest) ([]internal.Product, error) {
	products, err := s.db.Products(ctx, capability)
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}

	return products, nil
}

func (s Service) Product(ctx context.Context, id int, capability internal.CapabilityRequest) (internal.Product, error) {
	product, err := s.db.Product(ctx, id, capability)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, internal.ErrNotFound
		}

		return nil, fmt.Errorf("get product: %w", err)
	}

	return product, nil
}

func (s Service) Availability(ctx context.Context, productID int, localDate time.Time, capability internal.CapabilityRequest) (internal.Availability, error) {
	availability, err := s.db.Availability(ctx, productID, localDate, capability)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, internal.ErrNotFound
		}
		return nil, fmt.Errorf("get availability: %w", err)
	}

	return availability, nil
}

func (s Service) Availabilities(ctx context.Context, productID int, localDateStart, localDateEnd time.Time, capability internal.CapabilityRequest) ([]internal.Availability, error) {
	availabilities, err := s.db.Availabilities(ctx, productID, localDateStart, localDateEnd, capability)
	if err != nil {
		return nil, fmt.Errorf("get availabilities: %w", err)
	}

	return availabilities, nil
}

func (s Service) CreateBooking(ctx context.Context, req internal.CreateBookingRequest) (int, error) {
	params := CreateBookingParams{
		ProductID:      req.ProductID,
		AvailabilityID: req.AvailabilityID,
		Units:          req.Units,
		UserID:         req.UserID,
	}
	id, err := s.db.CreateBooking(ctx, params)
	if err != nil {
		if errors.Is(err, ErrNotAvailable) {
			return 0, internal.ErrNotAvailable
		}

		return 0, fmt.Errorf("create booking: %w", err)
	}

	return id, nil
}

func (s Service) ConfirmBooking(ctx context.Context, id, userID int) error {
	if err := s.db.ConfirmBooking(ctx, id, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return internal.ErrNotFound
		}
		return fmt.Errorf("confirm booking: %w", err)
	}

	return nil
}

func (s Service) Booking(ctx context.Context, id, userID int, capability internal.CapabilityRequest) (internal.Booking, error) {
	booking, err := s.db.Booking(ctx, id, userID, capability)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, internal.ErrNotFound
		}
		return nil, fmt.Errorf("get booking: %w", err)
	}

	return booking, nil
}
