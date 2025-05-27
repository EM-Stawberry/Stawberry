package database

import (
	"log"

	"go.uber.org/zap"

	"github.com/EM-Stawberry/Stawberry/config"

	// Import pgx driver to enable database connection via database/sql
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func InitDB(cfg *config.DBConfig) (*sqlx.DB, func()) {
	db, err := sqlx.Connect("pgx", cfg.GetDBConnString())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	closer := func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}

	return db, closer
}

func SeedDatabase(cfg *config.Config, db *sqlx.DB, log *zap.Logger) {
	if cfg.Environment == config.EnvDev || cfg.Environment == config.EnvTest {
		log.Info("Seeding database with test data")
		_, err := sqlx.LoadFile(db, "migrations/seed_data/seed_data.sql")
		if err != nil {
			log.Error("Failed to load seed data SQL", zap.Error(err))
		}
		log.Info("Database seeded with test data")
	}
}
