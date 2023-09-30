package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PostgreDB struct {
	PDB       *sql.DB
	DbTimeout time.Duration
}

func (pDB *PostgreDB) Connection() *sql.DB {
	return pDB.PDB
}

type TIdEmailPassword struct {
	Id int `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	UType string `json:"utype"`
}

func (pDB *PostgreDB) GetIdEmailPasswordUser(user TIdEmailPassword) (*TIdEmailPassword, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pDB.DbTimeout)
	defer cancel()

	var table string
	if user.UType == "owner" {
		table = "owners"
	} else {
		table = "users"
	}

	query := fmt.Sprintf("select id, email, password from %s where email = '%s'", table, user.Email)
	row := pDB.PDB.QueryRowContext(ctx, query)

	u := TIdEmailPassword{}

	err := row.Scan(
		&u.Id,
		&u.Email,
		&u.Password,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, errors.New("user not found please create a account first.")
		default:
			return nil, errors.New("something went wrong")
		}
	}
	return &u, nil
}



type AddUser struct {
	Email string `json:"email"`
	Password string `json:"password"`
	UType string `json:"utype"`
}

func (pDB *PostgreDB) AddUser(user AddUser) (error) {
	var table string
	if user.UType == "owner" {
		table = "owners"
	} else {
		table = "users"
	}
	query := fmt.Sprintf("insert into %s(email, password) values('%s', '%s')", table, user.Email, user.Password)
	fmt.Println(query)
	result, err := pDB.PDB.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}
