package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

const dateFormat = "2006-01-02"

// User represents an application user.
type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func (u User) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("id", u.ID),
		slog.String("email", u.Email),
	)
}

// ProductBase is a product without any additional capabilities.
type ProductBase struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

func (p ProductBase) IsProduct() {}

// ProductWithPrice is a product with price capability.
type ProductWithPrice struct {
	ProductBase
	CapabilityPrice
}

// Product represents a product in the system.
type Product interface {
	IsProduct()
}

// AvailabilityBase is avalability without any additional capabilities.
type AvailabilityBase struct {
	ID        string             `json:"id"`
	LocalDate Date               `json:"localDate"`
	Status    AvailabilityStatus `json:"status"`
	Vacancies int                `json:"vacancies"`
	Available bool               `json:"available"`
}

func (a AvailabilityBase) IsAvailability() {}

type AvailabilityStatus string

const (
	AvailabilityStatusAvailable AvailabilityStatus = "AVAILABLE"
	AvailabilityStatusSoldOut   AvailabilityStatus = "SOLD_OUT"
)

// AvailabilityWithPrice is an availability with price capability.
type AvailabilityWithPrice struct {
	AvailabilityBase
	CapabilityPrice
}

// Availability represents the availability of a product.
type Availability interface {
	IsAvailability()
}

// BookingBase is a booking without any additional capabilities.
type BookingBase struct {
	ID             string        `json:"id"`
	Status         BookingStatus `json:"status"`
	ProductID      string        `json:"productId"`
	AvailabilityID string        `json:"availabilityId"`
	Units          []Unit        `json:"units"`
}

func (b BookingBase) IsBooking() {}

// BookingWithPrice is a booking with price capability.
type BookingWithPrice struct {
	BookingBase
	CapabilityPrice
}

// Booking represents a booking in the system.
type Booking interface {
	IsBooking()
}

type BookingStatus string

const (
	BookingStatusReserved  BookingStatus = "RESERVED"
	BookingStatusConfirmed BookingStatus = "CONFIRMED"
)

// Scan implements the [sql.Scanner] interface.
func (b *BookingStatus) Scan(value any) error {
	switch v := value.(type) {
	case string:
		*b = BookingStatus(v)
		return nil
	case []byte:
		*b = BookingStatus(string(v))
		return nil
	default:
		return fmt.Errorf("cannot convert %T to BookingStatus", value)
	}
}

type CapabilityRequest int

const (
	CapabilityRequestNone CapabilityRequest = iota
	CapabilityRequestPrice
)

// Capability represents additional capabilities for objects.
type Capability interface {
	IsCapability()
}

// UnitBase is a unit without any additional capabilities.
type UnitBase struct {
	ID     string  `json:"id"`
	Ticket *string `json:"ticket"`
}

func (u UnitBase) IsUnit() {}

// UnitWithPrice is a unit with price capability.
type UnitWithPrice struct {
	UnitBase
	CapabilityPrice
}

// Unit represents a unit in a booking.
type Unit interface {
	IsUnit()
}

// CapabilityNone has no additional capabilities.
type CapabilityNone struct{}

func (c CapabilityNone) IsCapability() {}

// CapabilityPrice adds a price capability for objects.
type CapabilityPrice struct {
	Price    int    `json:"price"`
	Currency string `json:"currency"`
}

func (c CapabilityPrice) IsCapability() {}

type CreateBookingRequest struct {
	ProductID      int
	AvailabilityID int
	Units          int
	UserID         int
}

// Date is custom type for handling date JSON serialization and deserialization.
type Date time.Time

func (d Date) MarshalJSON() ([]byte, error) {
	formattedDate := time.Time(d).Format(dateFormat)
	return json.Marshal(formattedDate)
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var dateString string
	if err := json.Unmarshal(data, &dateString); err != nil {
		return err
	}

	parsedDate, err := time.Parse(dateFormat, dateString)
	if err != nil {
		return err
	}

	*d = Date(parsedDate)
	return nil
}
