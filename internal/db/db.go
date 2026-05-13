package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type DbConfig struct {
	ConnStr string
}

func (cfg *DbConfig) String() string {
	return cfg.ConnStr
}

func New(dbCfg *DbConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbCfg.String())
	if err != nil {
		return nil, err
	}
	return db, nil
}
