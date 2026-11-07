package db

import (
	"embed"
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type migrationLogger struct{}

func (migrationLogger) Printf(format string, args ...interface{}) {
	log.Printf("migrations: "+format, args...)
}

func (migrationLogger) Verbose() bool { return false }

// RunMigrations applies pending migrations before the API accepts requests.
func RunMigrations(databaseURL string) error {
	log.Print("migrations: checking for pending migrations")
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		_ = source.Close()
		return err
	}
	defer m.Close()
	m.Log = migrationLogger{}
	if err := m.Up(); errors.Is(err, migrate.ErrNoChange) {
		log.Print("migrations: database is already up to date")
		return nil
	} else if err != nil {
		return err
	}
	log.Print("migrations: completed successfully")
	return nil
}
