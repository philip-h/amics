package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DbConfig struct {
	User string
	Password string
	Host string
	DbName string
	Params	 string
}

func (cfg *DbConfig) ToSqlite3Addr() string {
	return fmt.Sprintf("./%s.db", cfg.DbName)
}

func New(dbCfg *DbConfig) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbCfg.ToSqlite3Addr())
	if err != nil {
		return nil, err
	}
	return db, nil
}
