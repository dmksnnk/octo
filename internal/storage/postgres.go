package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dmksnnk/octo/internal"
	"github.com/dmksnnk/octo/internal/auth"
	"github.com/dmksnnk/octo/internal/platform"
	"github.com/dmksnnk/octo/internal/service"
	"github.com/dmksnnk/octo/internal/storage/queries"
)

type Postgres struct {
	db *sql.DB
}

var (
	_ service.DB = (*Postgres)(nil)
	_ auth.DB    = (*Postgres)(nil)
)

func NewPostgres(db *sql.DB) Postgres {
	return Postgres{
		db: db,
	}
}

func (p Postgres) UserByAPIKey(ctx context.Context, key string) (internal.User, error) {
	user, err := queries.New(p.db).UserByAPIKey(ctx, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return internal.User{}, auth.ErrNotFound
		}
		return internal.User{}, fmt.Errorf("get user by api key: %w", err)
	}

	return internal.User{
		ID:    int(user.ID),
		Email: user.Email,
	}, nil
}

func (p Postgres) Products(ctx context.Context, capability internal.CapabilityRequest) ([]internal.Product, error) {
	switch capability {
	case internal.CapabilityRequestPrice:
		productsWithPrices, err := queries.New(p.db).ProductsWithPrices(ctx)
		if err != nil {
			return nil, fmt.Errorf("get products with prices: %w", err)
		}

		return mapp(
			productsWithPrices,
			func(pp queries.ProductsWithPricesRow) internal.Product {
				return toProductWithPrice(pp.Product, pp.Price)
			},
		), nil
	default:
		products, err := queries.New(p.db).Products(ctx)
		if err != nil {
			return nil, fmt.Errorf("get products: %w", err)
		}
		return mapp(
			products,
			func(p queries.Product) internal.Product {
				return toProduct(p)
			},
		), nil
	}
}

func (p Postgres) Product(ctx context.Context, id int, capability internal.CapabilityRequest) (internal.Product, error) {
	switch capability {
	case internal.CapabilityRequestPrice:
		productWithPrice, err := queries.New(p.db).ProductWithPrice(ctx, int32(id))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, service.ErrNotFound
			}
			return nil, fmt.Errorf("get product with price: %w", err)
		}

		return toProductWithPrice(productWithPrice.Product, productWithPrice.Price), nil
	default:
		product, err := queries.New(p.db).Product(ctx, int32(id))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, service.ErrNotFound
			}
			return internal.ProductBase{}, fmt.Errorf("get product: %w", err)
		}

		return toProduct(product), nil
	}
}

func (p Postgres) Availability(ctx context.Context, productID int, localDate time.Time, capability internal.CapabilityRequest) (internal.Availability, error) {
	switch capability {
	case internal.CapabilityRequestPrice:
		params := queries.AvalabilityWithPriceParams{
			ProductID: int32(productID),
			LocalDate: localDate,
		}
		availability, err := queries.New(p.db).AvalabilityWithPrice(ctx, params)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, service.ErrNotFound
			}
			return nil, err
		}

		return toAvailabilityWithPrice(availability.Availability, availability.Price), nil
	default:
		params := queries.AvalabilityParams{
			ProductID: int32(productID),
			LocalDate: localDate,
		}
		availability, err := queries.New(p.db).Avalability(ctx, params)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, service.ErrNotFound
			}
			return nil, err
		}

		return toAvailability(availability), nil
	}
}

func (p Postgres) Availabilities(ctx context.Context, productID int, localDateStart, localDateEnd time.Time, capability internal.CapabilityRequest) ([]internal.Availability, error) {
	switch capability {
	case internal.CapabilityRequestPrice:
		params := queries.AvalabilityWithPriceRangeParams{
			ProductID:      int32(productID),
			LocalDateStart: localDateStart,
			LocalDateEnd:   localDateEnd,
		}
		availabilities, err := queries.New(p.db).AvalabilityWithPriceRange(ctx, params)
		if err != nil {
			return nil, err
		}

		return mapp(
			availabilities,
			func(a queries.AvalabilityWithPriceRangeRow) internal.Availability {
				return toAvailabilityWithPrice(a.Availability, a.Price)
			},
		), nil
	default:
		params := queries.AvalabilityRangeParams{
			ProductID:      int32(productID),
			LocalDateStart: localDateStart,
			LocalDateEnd:   localDateEnd,
		}
		availabilities, err := queries.New(p.db).AvalabilityRange(ctx, params)
		if err != nil {
			return nil, err
		}

		return mapp(
			availabilities,
			func(a queries.Availability) internal.Availability {
				return toAvailability(a)
			},
		), nil
	}
}

func (p Postgres) CreateBooking(ctx context.Context, params service.CreateBookingParams) (int, error) {
	q := queries.CreateBookingParams{
		ProductID:      int32(params.ProductID),
		AvailabilityID: int32(params.AvailabilityID),
		Units:          int32(params.Units),
		UserID:         int32(params.UserID),
	}
	id, err := queries.New(p.db).CreateBooking(ctx, q)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, service.ErrNotAvailable
		}
		return 0, fmt.Errorf("create booking: %w", err)
	}

	return int(id), nil
}

func (p Postgres) ConfirmBooking(ctx context.Context, id int, userID int) error {
	params := queries.BookingForUpdateParams{
		ID:     int64(id),
		UserID: int32(userID),
	}
	return p.withTx(ctx, func(tx *sql.Tx) error {
		qrs := queries.New(tx)
		bookingWithUnits, err := qrs.BookingForUpdate(ctx, params)
		if err != nil {
			return fmt.Errorf("get booking for update: %w", err)
		}
		if len(bookingWithUnits) == 0 {
			return service.ErrNotFound
		}

		if bookingWithUnits[0].Status == internal.BookingStatusConfirmed { // already confirmed
			return nil
		}

		if err := qrs.ConfirmBooking(ctx, int64(id)); err != nil {
			return fmt.Errorf("confirm booking: %w", err)
		}

		for _, row := range bookingWithUnits {
			if err := p.createTicket(ctx, tx, row.UnitID); err != nil {
				return fmt.Errorf("create ticket: %w", err)
			}
		}

		return nil
	})
}

func (p Postgres) createTicket(ctx context.Context, tx *sql.Tx, unitID int64) error {
	ticket := platform.RandString()
	params := queries.SetUnitTicketParams{
		ID: unitID,
		Ticket: sql.NullString{
			String: ticket,
			Valid:  true,
		},
	}
	if err := queries.New(tx).SetUnitTicket(ctx, params); err != nil {
		return fmt.Errorf("set unit ticket: %w", err)
	}

	return nil
}

func (p Postgres) Booking(ctx context.Context, id int, userID int, capability internal.CapabilityRequest) (internal.Booking, error) {
	switch capability {
	case internal.CapabilityRequestPrice:
		params := queries.BookingWithPriceParams{
			ID:     int64(id),
			UserID: int32(userID),
		}
		bookingsWithPrice, err := queries.New(p.db).BookingWithPrice(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("get booking with price: %w", err)
		}

		if len(bookingsWithPrice) == 0 {
			return nil, service.ErrNotFound
		}

		return toBookingsWithPrice(bookingsWithPrice)[0], nil
	default:
		params := queries.BookingParams{
			ID:     int64(id),
			UserID: int32(userID),
		}
		bookings, err := queries.New(p.db).Booking(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("get booking: %w", err)
		}

		if len(bookings) == 0 {
			return nil, service.ErrNotFound
		}

		return toBookings(bookings)[0], nil
	}
}

func (p Postgres) withTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin database transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("rollback database transaction after %w: %w", err, rollbackErr)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit database transaction: %w", err)
	}

	return nil
}

func toProductWithPrice(product queries.Product, price queries.Price) internal.ProductWithPrice {
	return internal.ProductWithPrice{
		ProductBase:     toProduct(product),
		CapabilityPrice: toPrice(price),
	}
}

func toProduct(p queries.Product) internal.ProductBase {
	return internal.ProductBase{
		ID:       strconv.Itoa(int(p.ID)),
		Name:     p.Name,
		Capacity: int(p.Capacity),
	}
}

func toAvailability(a queries.Availability) internal.AvailabilityBase {
	return internal.AvailabilityBase{
		ID:        strconv.Itoa(int(a.ID)),
		LocalDate: internal.Date(a.LocalDate),
		Status:    toAvailabilityStatus(int(a.Vacancies)),
		Vacancies: int(a.Vacancies),
		Available: a.Vacancies > 0,
	}
}

func toAvailabilityStatus(vacancies int) internal.AvailabilityStatus {
	if vacancies > 0 {
		return internal.AvailabilityStatusAvailable
	}
	return internal.AvailabilityStatusSoldOut
}

func toAvailabilityWithPrice(a queries.Availability, p queries.Price) internal.AvailabilityWithPrice {
	return internal.AvailabilityWithPrice{
		AvailabilityBase: toAvailability(a),
		CapabilityPrice:  toPrice(p),
	}
}

func toBookings(rows []queries.BookingRow) []internal.BookingBase {
	bookings := make(map[int64]*internal.BookingBase)
	for _, row := range rows {
		if _, ok := bookings[row.Booking.ID]; !ok {
			b := toBooking(row.Booking)
			bookings[row.Booking.ID] = &b
		}

		unit := toUnit(row.Unit)
		units := bookings[row.Booking.ID].Units
		bookings[row.Booking.ID].Units = append(units, unit)
	}

	result := make([]internal.BookingBase, 0, len(bookings))
	for _, booking := range bookings {
		result = append(result, *booking)
	}

	return result
}

func toBookingsWithPrice(rows []queries.BookingWithPriceRow) []internal.BookingWithPrice {
	bookings := make(map[int64]*internal.BookingWithPrice)
	for _, row := range rows {
		if _, ok := bookings[row.Booking.ID]; !ok {
			b := toBooking(row.Booking)
			bookings[row.Booking.ID] = &internal.BookingWithPrice{
				BookingBase: b,
			}
		}

		price := bookings[row.Booking.ID].CapabilityPrice
		price.Price += int(row.Price.Price)
		price.Currency = row.Price.Currency
		bookings[row.Booking.ID].CapabilityPrice = price

		unit := toUnitWithPrice(row.Unit, row.Price)
		units := bookings[row.Booking.ID].Units
		bookings[row.Booking.ID].Units = append(units, unit)
	}

	result := make([]internal.BookingWithPrice, 0, len(bookings))
	for _, booking := range bookings {
		result = append(result, *booking)
	}

	return result
}

func toBooking(b queries.Booking) internal.BookingBase {
	return internal.BookingBase{
		ID:             strconv.Itoa(int(b.ID)),
		ProductID:      strconv.Itoa(int(b.ProductID)),
		AvailabilityID: strconv.Itoa(int(b.AvailabilityID)),
		Status:         b.Status,
		Units:          []internal.Unit{},
	}
}

func toUnit(u queries.Unit) internal.UnitBase {
	return internal.UnitBase{
		ID:     strconv.Itoa(int(u.ID)),
		Ticket: nullStringToPtr(u.Ticket),
	}
}

func toUnitWithPrice(u queries.Unit, p queries.Price) internal.UnitWithPrice {
	return internal.UnitWithPrice{
		UnitBase:        toUnit(u),
		CapabilityPrice: toPrice(p),
	}
}

func toPrice(p queries.Price) internal.CapabilityPrice {
	return internal.CapabilityPrice{
		Price:    int(p.Price),
		Currency: p.Currency,
	}
}

func nullStringToPtr(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func mapp[T any, V any](slice []T, f func(T) V) []V {
	mapped := make([]V, len(slice))
	for i, v := range slice {
		mapped[i] = f(v)
	}

	return mapped
}
