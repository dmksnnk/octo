package storagetesting_test

import (
	"testing"

	"github.com/dmksnnk/octo/internal/storage/storagetesting"
	"github.com/google/go-cmp/cmp"
	"github.com/peterldowns/pgtestdb"
)

func TestUnitConfigFromURL(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		url := "postgres://bob:secret@1.2.3.4:5432/mydb?sslmode=verify-full"
		want := pgtestdb.Config{
			DriverName: "postgres",
			Host:       "1.2.3.4",
			Port:       "5432",
			User:       "bob",
			Password:   "secret",
			Database:   "mydb",
			Options:    "sslmode=verify-full",
		}

		cfg, err := storagetesting.ConfigFromURL("postgres", url)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(want, cfg); diff != "" {
			t.Errorf("config mismatch (-want +got):\n%s", diff)
		}

		if cfg.URL() != url {
			t.Fatalf("unexpected URL, want %s, got %s", url, cfg.URL())
		}
	})

	t.Run("bad protocol", func(t *testing.T) {
		url := "http://example.com"
		_, err := storagetesting.ConfigFromURL("pgx", url)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
