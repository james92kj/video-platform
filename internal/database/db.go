package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func New(connectionString string) (*DB, error) {

	conn, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to ping database: %w", err)
	}

	// set the connection pool
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	db.conn.Close()
	return nil
}

func (db *DB) GetConn() *sql.DB {
	return db.conn
}
