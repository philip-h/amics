package store

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id int `json:"id" db:"id"`
	StudentNumber string `json:"student_number" db:"student_number"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	_, err = s.db.Exec("INSERT INTO users (student_number, username, password) VALUES ($1, $2, $3)", user.StudentNumber, user.Username, user.Password)
	return err
}

func (s *UserStore) GetByUsername(username string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow("SELECT id, student_number, username, password FROM users WHERE username = $1", username).Scan(&user.Id, &user.StudentNumber, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}