package migrator

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

// RunMigrations applies database migrations using *sqlx.DB.
func RunMigrations(db *sqlx.DB, migrationsDir string) {

	// Set the database dialect (PostgreSQL)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	// Apply migrations
	if err := goose.Up(db.DB, migrationsDir); err != nil {
		log.Fatal(err)
	}

	log.Println("Migrations applied successfully!")
}
