package store

import "database/sql"

type Teacher struct {
	Id             int    `json:"id" db:"id"`
	EmployeeNumber string `json:"employee_number" db:"employee_number"`
	Username       string `json:"username" db:"username"`
	Password       string `json:"password" db:"password"`
}

type TeacherStore struct {
	db *sql.DB
}

func (s *TeacherStore) GetByUsername(username string) (*Teacher, error) {
	teacher := &Teacher{}
	err := s.db.QueryRow("SELECT id, employee_number, username, password FROM teacher WHERE username = $1", username).Scan(&teacher.Id, &teacher.EmployeeNumber, &teacher.Username, &teacher.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return teacher, nil
}