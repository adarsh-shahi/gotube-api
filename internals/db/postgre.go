package db

import "database/sql"

type PostgreDB struct{
	PDB *sql.DB
}

func (pDB *PostgreDB) Connection() *sql.DB{
	return pDB.PDB
}
