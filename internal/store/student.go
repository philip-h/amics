package store

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type Student struct {
	Id            int    `json:"id" db:"id"`
	StudentNumber string `json:"student_number" db:"student_number"`
	Username      string `json:"username" db:"username"`
	Password      string `json:"password" db:"password"`
	CourseId      int    `json:"course_id" db:"course_id"`
}

type StudentStore struct {
	db *sql.DB
}

func (s *StudentStore) Create(student *Student) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(student.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	student.Password = string(hashedPassword)
	_, err = s.db.Exec("INSERT INTO student (student_number, username, password, course_id) VALUES ($1, $2, $3, $4)", student.StudentNumber, student.Username, student.Password, student.CourseId)
	return err
}

func (s *StudentStore) GetByUsername(username string) (*Student, error) {
	student := &Student{}
	err := s.db.QueryRow("SELECT id, student_number, username, password, course_id FROM student WHERE username = $1", username).Scan(&student.Id, &student.StudentNumber, &student.Username, &student.Password, &student.CourseId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return student, nil
}
