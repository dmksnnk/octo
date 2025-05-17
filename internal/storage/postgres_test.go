package storage_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/dmksnnk/octo/internal"
	"github.com/dmksnnk/octo/internal/service"
	"github.com/dmksnnk/octo/internal/storage"
	"github.com/dmksnnk/octo/internal/storage/queries"
	"github.com/dmksnnk/octo/internal/storage/storagetesting"
)

func TestProduct(t *testing.T) {
	db := storagetesting.Open(t)
	products := []queries.Product{
		storagetesting.NewProduct(t, db),
		storagetesting.NewProduct(t, db),
		// deleted product
		storagetesting.NewProduct(t, db, func(p *queries.InsertProductParams) {
			p.DeletedAt = sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			}
		}),
	}
	// price for each product
	prices := make([]queries.Price, len(products))
	for i, product := range products {
		prices[i] = storagetesting.NewPrice(t, db, product.ID)
	}

	t.Run("get all products", func(t *testing.T) {
		gotProducts, err := storage.NewPostgres(db).Products(context.TODO(), internal.CapabilityRequestNone)
		if err != nil {
			t.Fatalf("get products: %v", err)
		}
		wantProductsLen := len(products) - 1 // one product is deleted
		if len(gotProducts) != wantProductsLen {
			t.Fatalf("want %d products, got %d", wantProductsLen, len(gotProducts))
		}
		for i, gotProduct := range gotProducts {
			assertProductEqual(t, products[i], gotProduct.(internal.ProductBase))
		}
	})

	t.Run("get all products with prices", func(t *testing.T) {
		gotProducts, err := storage.NewPostgres(db).Products(context.TODO(), internal.CapabilityRequestPrice)
		if err != nil {
			t.Fatalf("get products: %v", err)
		}
		wantProductsLen := len(products) - 1 // one product is deleted
		if len(gotProducts) != wantProductsLen {
			t.Fatalf("want %d products, got %d", wantProductsLen, len(gotProducts))
		}
		for i, gotProduct := range gotProducts {
			want := queries.ProductWithPriceRow{
				Product: products[i],
				Price:   prices[i],
			}
			assertProductWithPriceEqual(t, want, gotProduct.(internal.ProductWithPrice))
		}
	})

	t.Run("get product by ID", func(t *testing.T) {
		product := products[0]
		gotProduct, err := storage.NewPostgres(db).Product(context.TODO(), int(product.ID), internal.CapabilityRequestNone)
		if err != nil {
			t.Fatalf("get product: %v", err)
		}
		assertProductEqual(t, product, gotProduct.(internal.ProductBase))
	})

	t.Run("get product by ID not found", func(t *testing.T) {
		_, err := storage.NewPostgres(db).Product(context.TODO(), -1, internal.CapabilityRequestNone)
		if !errors.Is(err, service.ErrNotFound) {
			t.Errorf("want error %v, got %v", service.ErrNotFound, err)
		}
	})
}

func TestAvailability(t *testing.T) {
	db := storagetesting.Open(t)
	products := []queries.Product{
		storagetesting.NewProduct(t, db),
		storagetesting.NewProduct(t, db),
	}
	// price for each product
	prices := make([]queries.Price, len(products))
	for i, product := range products {
		prices[i] = storagetesting.NewPrice(t, db, product.ID)
	}
	dateSoldOut := time.Now().AddDate(0, 0, -1)
	dateStart := time.Now().AddDate(0, 0, 1)
	dateEnd := dateStart.AddDate(0, 0, 2)
	// availability for each product
	availabilities := []queries.Availability{
		// sold out, not in range
		storagetesting.NewAvailability(t, db, products[0].ID, func(iap *queries.InsertAvailabilityParams) {
			iap.Vacancies = 0
			iap.LocalDate = dateSoldOut
		}),
		// in range
		storagetesting.NewAvailability(t, db, products[1].ID, func(iap *queries.InsertAvailabilityParams) {
			iap.LocalDate = dateStart
		}),
		storagetesting.NewAvailability(t, db, products[1].ID, func(iap *queries.InsertAvailabilityParams) {
			iap.LocalDate = dateEnd
		}),
	}

	t.Run("get single availability", func(t *testing.T) {
		product := products[0]
		availability := availabilities[0]
		wantAvailability := internal.AvailabilityBase{
			ID:        strconv.Itoa(int(availability.ID)),
			LocalDate: internal.Date(availability.LocalDate),
			Status:    internal.AvailabilityStatusSoldOut,
			Vacancies: int(availability.Vacancies),
			Available: false,
		}
		gotAvailability, err := storage.NewPostgres(db).Availability(context.TODO(), int(product.ID), availability.LocalDate, internal.CapabilityRequestNone)
		if err != nil {
			t.Fatalf("get availability: %v", err)
		}
		if !reflect.DeepEqual(wantAvailability, gotAvailability) {
			t.Errorf("want availability %v, got %v", wantAvailability, gotAvailability)
		}
	})

	t.Run("get single availability with price", func(t *testing.T) {
		product := products[0]
		availability := availabilities[0]
		price := prices[0]
		wantAvailability := internal.AvailabilityWithPrice{
			AvailabilityBase: internal.AvailabilityBase{
				ID:        strconv.Itoa(int(availability.ID)),
				LocalDate: internal.Date(availability.LocalDate),
				Status:    internal.AvailabilityStatusSoldOut,
				Vacancies: int(availability.Vacancies),
				Available: false,
			},
			CapabilityPrice: internal.CapabilityPrice{
				Price:    int(price.Price),
				Currency: price.Currency,
			},
		}
		gotAvailability, err := storage.NewPostgres(db).Availability(context.TODO(), int(product.ID), availability.LocalDate, internal.CapabilityRequestPrice)
		if err != nil {
			t.Fatalf("get availability: %v", err)
		}
		if !reflect.DeepEqual(wantAvailability, gotAvailability) {
			t.Errorf("want availability %v, got %v", wantAvailability, gotAvailability)
		}
	})

	t.Run("get availability not found", func(t *testing.T) {
		_, err := storage.NewPostgres(db).Availability(context.TODO(), -1, time.Now(), internal.CapabilityRequestNone)
		if !errors.Is(err, service.ErrNotFound) {
			t.Errorf("want error %v, got %v", service.ErrNotFound, err)
		}
	})

	t.Run("get availability range", func(t *testing.T) {
		product := products[1]
		wantAvailabilities := []internal.AvailabilityBase{
			{
				ID:        strconv.Itoa(int(availabilities[1].ID)),
				LocalDate: internal.Date(availabilities[1].LocalDate),
				Status:    internal.AvailabilityStatusAvailable,
				Vacancies: int(availabilities[1].Vacancies),
				Available: true,
			},
			{
				ID:        strconv.Itoa(int(availabilities[2].ID)),
				LocalDate: internal.Date(availabilities[2].LocalDate),
				Status:    internal.AvailabilityStatusAvailable,
				Vacancies: int(availabilities[2].Vacancies),
				Available: true,
			},
		}

		gotAvailabilities, err := storage.NewPostgres(db).Availabilities(context.TODO(), int(product.ID), dateStart, dateEnd, internal.CapabilityRequestNone)
		if err != nil {
			t.Fatalf("get availabilities: %v", err)
		}
		if len(gotAvailabilities) != len(wantAvailabilities) {
			t.Fatalf("want %d availabilities, got %d", len(wantAvailabilities), len(gotAvailabilities))
		}

		for i, gotAvailability := range gotAvailabilities {
			if gotAvailability != wantAvailabilities[i] {
				t.Errorf("want availability %v, got %v", wantAvailabilities[i], gotAvailability)
			}
		}
	})
}

func TestBooking(t *testing.T) {
	db := storagetesting.Open(t)
	user := storagetesting.NewUser(t, db)
	product := storagetesting.NewProduct(t, db)
	price := storagetesting.NewPrice(t, db, product.ID)

	t.Run("booking not found", func(t *testing.T) {
		_, err := storage.NewPostgres(db).Booking(context.TODO(), -1, int(user.ID), internal.CapabilityRequestNone)
		if !errors.Is(err, service.ErrNotFound) {
			t.Errorf("want error %v, got %v", service.ErrNotFound, err)
		}
	})

	t.Run("create booking", func(t *testing.T) {
		pg := storage.NewPostgres(db)
		availability := storagetesting.NewAvailability(t, db, product.ID)

		params := service.CreateBookingParams{
			ProductID:      int(product.ID),
			AvailabilityID: int(availability.ID),
			Units:          3,
			UserID:         int(user.ID),
		}
		id, err := pg.CreateBooking(context.TODO(), params)
		if err != nil {
			t.Fatalf("create booking: %v", err)
		}

		wantBookingWithPrice := internal.BookingWithPrice{
			BookingBase: internal.BookingBase{
				ProductID:      strconv.Itoa(int(product.ID)),
				AvailabilityID: strconv.Itoa(int(availability.ID)),
				Status:         internal.BookingStatusReserved,
				Units: []internal.Unit{
					internal.UnitWithPrice{
						CapabilityPrice: internal.CapabilityPrice{
							Price:    int(price.Price),
							Currency: price.Currency,
						},
					},
					internal.UnitWithPrice{
						CapabilityPrice: internal.CapabilityPrice{
							Price:    int(price.Price),
							Currency: price.Currency,
						},
					},
					internal.UnitWithPrice{
						CapabilityPrice: internal.CapabilityPrice{
							Price:    int(price.Price),
							Currency: price.Currency,
						},
					},
				},
			},
			CapabilityPrice: internal.CapabilityPrice{
				Price:    int(price.Price) * 3,
				Currency: price.Currency,
			},
		}
		gotBookingWithPrice, err := pg.Booking(context.TODO(), id, int(user.ID), internal.CapabilityRequestPrice)
		if err != nil {
			t.Fatalf("get booking: %v", err)
		}
		assertBookingWithPrice(t, wantBookingWithPrice, gotBookingWithPrice.(internal.BookingWithPrice))

		// confirm booking
		wantConfirmedBooking := internal.BookingBase{
			ProductID:      strconv.Itoa(int(product.ID)),
			AvailabilityID: strconv.Itoa(int(availability.ID)),
			Status:         internal.BookingStatusConfirmed,
			Units:          make([]internal.Unit, 3),
		}
		if err := pg.ConfirmBooking(context.TODO(), id, int(user.ID)); err != nil {
			t.Fatalf("confirm booking: %v", err)
		}
		confirmedBooking, err := pg.Booking(context.TODO(), id, int(user.ID), internal.CapabilityRequestNone)
		if err != nil {
			t.Fatalf("get booking: %v", err)
		}
		for _, unit := range confirmedBooking.(internal.BookingBase).Units {
			if unit.(internal.UnitBase).Ticket == nil {
				t.Errorf("want ticket, got empty")
			}
		}
		assertBooking(t, wantConfirmedBooking, confirmedBooking.(internal.BookingBase))

		// confirm booking again, should be no error
		err = pg.ConfirmBooking(context.TODO(), id, int(user.ID))
		if err != nil {
			t.Fatalf("confirm booking again: %v", err)
		}
	})

	t.Run("create booking out of vacancies", func(t *testing.T) {
		availability := storagetesting.NewAvailability(t, db, product.ID, func(iap *queries.InsertAvailabilityParams) {
			iap.Vacancies = 2
		})
		pg := storage.NewPostgres(db)
		params := service.CreateBookingParams{
			ProductID:      int(product.ID),
			AvailabilityID: int(availability.ID),
			Units:          3,
			UserID:         int(user.ID),
		}

		_, err := pg.CreateBooking(context.TODO(), params)
		if !errors.Is(err, service.ErrNotAvailable) {
			t.Errorf("want error %v, got %v", service.ErrNotAvailable, err)
		}
	})

	t.Run("confirm booking not found", func(t *testing.T) {
		err := storage.NewPostgres(db).ConfirmBooking(context.TODO(), -1, int(user.ID))
		if !errors.Is(err, service.ErrNotFound) {
			t.Errorf("want error %v, got %v", service.ErrNotFound, err)
		}
	})
}

func assertProductEqual(t *testing.T, want queries.Product, got internal.ProductBase) {
	t.Helper()

	wantID := strconv.Itoa(int(want.ID))
	if wantID != got.ID {
		t.Errorf("want product ID %s, got %s", wantID, got.ID)
	}
	if want.Name != got.Name {
		t.Errorf("want product name %s, got %s", want.Name, got.Name)
	}
}

func assertProductWithPriceEqual(t *testing.T, want queries.ProductWithPriceRow, got internal.ProductWithPrice) {
	t.Helper()

	assertProductEqual(t, want.Product, got.ProductBase)
	assertPriceEqual(t, want.Price, got.CapabilityPrice)
}

func assertPriceEqual(t *testing.T, want queries.Price, got internal.CapabilityPrice) {
	t.Helper()

	if want.Price != int32(got.Price) {
		t.Errorf("want price %d, got %d", want.Price, got.Price)
	}
	if want.Currency != got.Currency {
		t.Errorf("want currency %s, got %s", want.Currency, got.Currency)
	}
}

func assertBooking(t *testing.T, want, got internal.BookingBase) {
	t.Helper()

	if want.AvailabilityID != got.AvailabilityID {
		t.Errorf("want availability ID %s, got %s", want.AvailabilityID, got.AvailabilityID)
	}
	if want.Status != got.Status {
		t.Errorf("want booking status %s, got %s", want.Status, got.Status)
	}
	if len(want.Units) != len(got.Units) {
		t.Errorf("want %d units, got %d", len(want.Units), len(got.Units))
	}
}

func assertBookingWithPrice(t *testing.T, want, got internal.BookingWithPrice) {
	t.Helper()

	assertBooking(t, want.BookingBase, got.BookingBase)
	assertPriceEqual(
		t,
		queries.Price{
			Price:    int32(want.Price),
			Currency: want.Currency,
		},
		got.CapabilityPrice,
	)
	for i, unit := range want.Units {
		want := unit.(internal.UnitWithPrice)
		assertPriceEqual(
			t,
			queries.Price{
				Price:    int32(want.Price),
				Currency: want.Currency,
			},
			got.Units[i].(internal.UnitWithPrice).CapabilityPrice,
		)
	}
}
