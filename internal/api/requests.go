package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dmksnnk/octo/internal"
)

// AvailabilityRequest represents a request to check product availability.
type AvailabilityRequest struct {
	ProductID      IntString     `json:"productId"`
	LocalDate      internal.Date `json:"localDate"`
	LocalDateStart internal.Date `json:"localDateStart"`
	LocalDateEnd   internal.Date `json:"localDateEnd"`
}

func (a *AvailabilityRequest) UnmarshalHTTP(r *http.Request) error {
	return json.NewDecoder(r.Body).Decode(&a)
}

// IntString is a custom type that marshals and unmarshals an integer as a string.
type IntString int

func (i *IntString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	id, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	*i = IntString(id)
	return nil
}

func (i IntString) MarshalJSON() ([]byte, error) {
	s := strconv.Itoa(int(i))
	return json.Marshal(s)
}

// BookingRequest represents a request to create a booking.
type BookingRequest struct {
	ProductID      IntString `json:"productId"`
	AvailabilityID IntString `json:"availabilityId"`
	Units          int       `json:"units"`
}

func (b *BookingRequest) UnmarshalHTTP(r *http.Request) error {
	return json.NewDecoder(r.Body).Decode(&b)
}

type CapabilityRequest internal.CapabilityRequest

func (c *CapabilityRequest) UnmarshalHTTP(r *http.Request) error {
	text := r.Header.Get("Capability")
	if err := c.UnmarshalText([]byte(text)); err != nil {
		return err
	}

	return nil
}

func (c *CapabilityRequest) UnmarshalText(text []byte) error {
	switch string(text) {
	case "price":
		*c = CapabilityRequest(internal.CapabilityRequestPrice)
	default:
		*c = CapabilityRequest(internal.CapabilityRequestNone)
	}
	return nil
}

type IDPathValue int

func (i *IDPathValue) UnmarshalHTTP(r *http.Request) error {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}

	*i = IDPathValue(id)
	return nil
}
