package main

import (
	"database/sql"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func (app *appConfig) connectDB() (*sql.DB, error) {
	connection, err := sql.Open("pgx", app.dbConnectionCreds)
	if err != nil {
		return nil, err
	}
	err = connection.Ping()
	if err != nil {
		return nil, err
	}
	return connection, nil
}
