package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type PostgreDB struct {
	PDB       *sql.DB
	DbTimeout time.Duration
}

func (pDB *PostgreDB) Connection() *sql.DB {
	return pDB.PDB
}

type TIdGmailPassword struct {
	Id int `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (pDB *PostgreDB) GetIdEmailPasswordUser(user TIdGmailPassword) (*TIdGmailPassword, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pDB.DbTimeout)
	defer cancel()

	query := `
	select id, email, password from users where email = $1 AND password = $2
	`
	row := pDB.PDB.QueryRowContext(ctx, query, user.Email, user.Password)

	u := TIdGmailPassword{}

	err := row.Scan(
		&u.Id,
		&u.Email,
		&u.Password,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, errors.New("wrong credentials")
		default:
			return nil, errors.New("something went wrong")
		}
	}
	return &u, nil
}
