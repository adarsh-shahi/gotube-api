package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	UType    string `json:"utype"`
	Role     string `json:"role"`
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
	u.UType = user.UType
	u.Role = user.Role
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, errors.New("user not found")
		default:
			return nil, errors.New("something went wrong")
		}
	}
	return &u, nil
}

type AddUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UType    string `json:"utype"`
}

func (pDB *PostgreDB) AddUser(user AddUser) error {
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

func (pDB *PostgreDB) AddInvites(sender, receiver int, role string) error {
	query := fmt.Sprintf("insert into invites(sender, receiver, role) values(%d, %d, '%s')", sender, receiver, role)
	_, err := pDB.PDB.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}
	return nil
}

func (pDB *PostgreDB) UpdateInviteRole(sender, receiver int, role string) error {
	query := fmt.Sprintf("update invites set role = '%s' where sender = %d AND receiver = %d", role, sender, receiver)
	_, err := pDB.PDB.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}
	return nil
}

func (pDB *PostgreDB) DeleteInvite(sender, receiver int) error {
	query := fmt.Sprintf("delete from invites where sender = %d AND receiver = %d", sender, receiver)
	_, err := pDB.PDB.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}
	return nil
}

func (pDB *PostgreDB) DeleteInviteAndAddToTeam(sender, receiver int, role string) error {
	ctx := context.Background()
	tx, err := pDB.PDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	queryToDelete := fmt.Sprintf(
		"delete from invites where sender = %d and receiver = %d and role = '%s'",
		sender,
		receiver,
		role,
	)
	result, err := tx.ExecContext(ctx, queryToDelete)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows != 1 {
		tx.Rollback()
		return errors.New("unable to find invite")
	}

	queryToAdd := fmt.Sprintf("insert into teams(owner, member, role) values(%d, %d, '%s')", sender, receiver, role)
	result, err = tx.ExecContext(ctx, queryToAdd)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

type Invites struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (pDB *PostgreDB) GetAllInvites(id int, utype string) ([]Invites, error) {
	var list []Invites
	query := ""

	if utype == "owner" {
		// get all invites sent by the owner to his users (team)
		query = fmt.Sprintf("select users.email, invites.role from invites inner join users on invites.receiver = users.id where invites.sender = %d", id)
	} else if utype == "user" {
		// get all invitations sent by different owners to from specific user
		query = fmt.Sprintf("select owners.email, invites.role from invites inner join owners on invites.sender = owners.id where invites.receiver = %d", id)
	}

	rows, err := pDB.PDB.QueryContext(context.Background(), query)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		invite := Invites{}
		rows.Scan(&invite.Email, &invite.Role)
		list = append(list, invite)
	}
	fmt.Println(list)
	return list, nil
}
