package storagetesting

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/goosemigrator"
)

// Open connects to the DB and creates a template database from migrations if it doesn't exist.
// For each test, it creates a new database from the template and returns a connection to it.
// For more information, see https://github.com/peterldowns/pgtestdb.
func Open(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Fatal("please provide database URL via DATABASE_URL environment variable")
	}

	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal("get abs path", err)
	}

	// need to pass the root directory to the migrator, so it is able to read the migrations folder
	migrator := goosemigrator.New("migrations", goosemigrator.WithFS(os.DirFS(root)))
	cfg, err := ConfigFromURL("postgres", dbURL)
	if err != nil {
		t.Fatalf("config from URL: %s", err)
	}
	db := pgtestdb.New(t, cfg, migrator)

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatal("close DB", err)
		}
	})

	return db
}

// ConfigFromURL is a helper function to create a [pgtestdb.Config] from a connection string
// like "postgres://bob:secret@1.2.3.4:5432/mydb?sslmode=verify-full".
func ConfigFromURL(driverName, connString string) (pgtestdb.Config, error) {
	cfg, err := parseURL(connString)
	if err != nil {
		return pgtestdb.Config{}, err
	}

	cfg.DriverName = driverName

	return cfg, nil
}

func parseURL(connString string) (pgtestdb.Config, error) {
	connURL, err := url.Parse(connString)
	if err != nil {
		return pgtestdb.Config{}, err
	}

	if connURL.Scheme != "postgres" && connURL.Scheme != "postgresql" {
		return pgtestdb.Config{}, fmt.Errorf("%w: %s", errInvalidProtocol, connURL.Scheme)
	}

	cfg := pgtestdb.Config{
		Host:    connURL.Hostname(),
		Port:    connURL.Port(),
		Options: connURL.RawQuery,
	}

	if len(connURL.Path) > 1 {
		cfg.Database = connURL.Path[1:]
	}

	if connURL.User != nil {
		cfg.User = connURL.User.Username()
		cfg.Password, _ = connURL.User.Password()
	}

	return cfg, nil
}

var errInvalidProtocol = errors.New("invalid connection protocol")
