package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cdzombak/lychee-meta-tool/backend/config"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
	driver string
}

func Connect(cfg *config.Config) (*DB, error) {
	var driverName string
	switch cfg.Database.Type {
	case "mysql":
		driverName = "mysql"
	case "postgres":
		driverName = "postgres"
	case "sqlite":
		driverName = "sqlite3"
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}

	db, err := sql.Open(driverName, cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return &DB{
		DB:     db,
		driver: cfg.Database.Type,
	}, nil
}

func (db *DB) Driver() string {
	return db.driver
}

func (db *DB) Health() error {
	return db.Ping()
}