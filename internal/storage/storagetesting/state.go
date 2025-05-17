package storagetesting

import (
	"context"
	"database/sql"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/dmksnnk/octo/internal/storage/queries"
)

func NewUser(t *testing.T, db *sql.DB, ops ...func(*queries.InsertUserParams)) queries.User {
	t.Helper()

	p := queries.InsertUserParams{
		Email: gofakeit.Email(),
		ApiKey: sql.NullString{
			String: gofakeit.Word(),
			Valid:  true,
		},
	}
	for _, op := range ops {
		op(&p)
	}

	user, err := queries.New(db).InsertUser(context.TODO(), p)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	return user
}

func NewProduct(t *testing.T, db *sql.DB, ops ...func(*queries.InsertProductParams)) queries.Product {
	t.Helper()

	p := queries.InsertProductParams{
		Name:     gofakeit.ProductName(),
		Capacity: int32(gofakeit.IntRange(1, 100)),
	}
	for _, op := range ops {
		op(&p)
	}

	product, err := queries.New(db).InsertProduct(context.TODO(), p)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}

	return product
}

func NewPrice(t *testing.T, db *sql.DB, productID int32, ops ...func(*queries.InsertPriceParams)) queries.Price {
	t.Helper()

	p := queries.InsertPriceParams{
		ProductID: productID,
		Price:     int32(gofakeit.IntRange(1, 1000)),
		Currency:  gofakeit.CurrencyShort(),
	}
	for _, op := range ops {
		op(&p)
	}

	price, err := queries.New(db).InsertPrice(context.TODO(), p)
	if err != nil {
		t.Fatalf("insert price: %v", err)
	}

	return price
}

func NewAvailability(t *testing.T, db *sql.DB, productID int32, ops ...func(*queries.InsertAvailabilityParams)) queries.Availability {
	t.Helper()

	p := queries.InsertAvailabilityParams{
		ProductID: productID,
		LocalDate: gofakeit.Date(),
		Vacancies: int32(gofakeit.IntRange(1, 1000)),
	}
	for _, op := range ops {
		op(&p)
	}

	availability, err := queries.New(db).InsertAvailability(context.TODO(), p)
	if err != nil {
		t.Fatalf("insert availability: %v", err)
	}

	return availability
}
