package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type DbConfig struct {
	User     string
	Password string
	Host     string
	DbName   string
	Params   string
}

func (cfg *DbConfig) ToSqlite3Addr() string {
	return fmt.Sprintf("./%s.db", cfg.DbName)
}

func (cfg *DbConfig) ToPGAddr() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.DbName,
		cfg.Params,
	)
}

func New(dbCfg *DbConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbCfg.ToPGAddr())
	if err != nil {
		return nil, err
	}
	return db, nil
}
