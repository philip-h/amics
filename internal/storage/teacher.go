package storage

import (
	"context"
	"database/sql"
)

type Teacher struct {
	Id             int
	EmployeeNumber string
	Username       string
	Password       string
}

type TeacherStore struct {
	db *sql.DB
}

func (s *TeacherStore) GetByUsername(ctx context.Context, username string) (*Teacher, error) {
	teacher := &Teacher{}
	err := s.db.QueryRowContext(ctx, "SELECT id, employee_number, username, password FROM teacher WHERE username = $1", username).Scan(&teacher.Id, &teacher.EmployeeNumber, &teacher.Username, &teacher.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return teacher, nil
}
