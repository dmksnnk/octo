package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/dmksnnk/octo/internal/platform"
	"github.com/dmksnnk/octo/internal/storage/queries"
	_ "github.com/lib/pq"
)

type config struct {
	Products         int
	Users            int
	AvailabilityDays int
	DatabaseURL      string
}

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := parseConfig()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		slog.Error("failed to begin transaction", "error", err)
		os.Exit(1)
	}

	if err := generate(rootCtx, tx, cfg); err != nil {
		slog.Error("failed to generate data", "error", err)
		if err := tx.Rollback(); err != nil {
			slog.Error("failed to rollback transaction", "error", err)
		}
		os.Exit(1)
	}
	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit transaction", "error", err)
		os.Exit(1)
	}

	slog.Info("data generation completed successfully")
}

func parseConfig() config {
	var cfg config
	flag.StringVar(&cfg.DatabaseURL, "database-url", "", "Database connection URL (required)")
	flag.IntVar(&cfg.Products, "products", 10, "Number of products to generate")
	flag.IntVar(&cfg.Users, "users", 2, "Number of users to generate")
	flag.IntVar(&cfg.AvailabilityDays, "availability-days", 365, "Number of days to generate availability for")
	flag.Parse()
	return cfg
}

func generate(ctx context.Context, tx *sql.Tx, cfg config) error {
	qrs := queries.New(tx)
	for range cfg.Products {
		p, err := qrs.InsertProduct(ctx, queries.InsertProductParams{
			Name:     gofakeit.ProductName(),
			Capacity: int32(gofakeit.Number(1, 1000)),
		})
		if err != nil {
			return fmt.Errorf("insert product: %w", err)
		}

		now := time.Now()
		for i := range cfg.AvailabilityDays {
			_, err = qrs.InsertAvailability(ctx, queries.InsertAvailabilityParams{
				ProductID: p.ID,
				LocalDate: now.AddDate(0, 0, i),
				Vacancies: int32(gofakeit.Number(1, 1000)),
			})
			if err != nil {
				return fmt.Errorf("insert availability: %w", err)
			}
		}

		_, err = qrs.InsertPrice(ctx, queries.InsertPriceParams{
			ProductID: p.ID,
			Price:     int32(gofakeit.Price(10, 1000)),
			Currency:  gofakeit.CurrencyShort(),
		})
		if err != nil {
			return fmt.Errorf("insert price: %w", err)
		}
	}

	for range cfg.Users {
		_, err := qrs.InsertUser(ctx, queries.InsertUserParams{
			Email: gofakeit.Email(),
			ApiKey: sql.NullString{
				String: platform.RandString(),
				Valid:  true,
			},
		})
		if err != nil {
			return fmt.Errorf("insert user: %w", err)
		}
	}

	return nil
}
