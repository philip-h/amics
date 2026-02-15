package store

import "database/sql"

type Storage struct {
	Users interface {
		Create(*User) error
		GetByUsername(string) (*User, error)
	}
}

func New(db *sql.DB) Storage {
	return Storage{
		Users: &UserStore{db},
	}
}