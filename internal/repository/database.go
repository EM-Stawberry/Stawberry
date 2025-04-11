package repository

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/zuzaaa-dev/stawberry/config"
)

func InitDB(cfg *config.Config) *sqlx.DB {
	db, err := sqlx.Connect("pgx", cfg.GetDBConnString())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	return db
}
