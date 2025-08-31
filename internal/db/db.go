package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func openDb(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("db successfully pinged")

	return db, nil
}

func ConnectDb(dsn string) (*sql.DB, error) {
	db, err := openDb(dsn)
	if err != nil {
		return nil, err
	}
	return db, nil

}
