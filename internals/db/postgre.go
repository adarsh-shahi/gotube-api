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

type TGmailPassword struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (pDB *PostgreDB) GetEmailPasswordUser(user TGmailPassword) (*TGmailPassword, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pDB.DbTimeout)
	defer cancel()

	query := `
	select email, password from users where email = $1 AND password = $2
	`
	row := pDB.PDB.QueryRowContext(ctx, query, user.Email, user.Password)

	u := TGmailPassword{}

	err := row.Scan(
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
