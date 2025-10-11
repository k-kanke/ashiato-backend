package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DBClient struct {
	DB *sql.DB
}

func NewDBClient(dsn string) (*DBClient, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 必要に応じて接続プール設定

	return &DBClient{DB: db}, nil
}
