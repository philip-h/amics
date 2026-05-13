package storage

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type Person struct {
	Id        int
	FirstName string
	Username  string
	Password  string
	Role      string
}

type PersonStore struct {
	db *sql.DB
}

func (s *PersonStore) Create(student *Person, courseId int) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(student.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	student.Password = string(hashedPassword)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var id int64
	err = tx.QueryRow(`
  INSERT INTO person (student_number, first_name, username, password) 
  VALUES ($1, $2, $3, $4)
  RETURNING id`,
		student.Id,
		student.FirstName,
		student.Username,
		student.Password).Scan(&id)
	if err != nil {
		return err
	}

	student.Id = int(id)

	_, err = tx.Exec(`
  INSERT INTO student_course (student_id, couse_id)
  VALUES ($1, $2)`,
		id, courseId,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PersonStore) GetById(id int) (*Person, error) {
	person := &Person{}
	err := s.db.QueryRow(
		`SELECT id, first_name, username, password, role
    FROM person
    WHERE id = $1`, id,
	).Scan(&person.Id, &person.FirstName, &person.Username, &person.Password, &person.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return person, nil
}

func (s *PersonStore) GetByUsername(username string) (*Person, error) {
	student := &Person{}
	err := s.db.QueryRow(`SELECT id, first_name, username, password, role
  FROM person
  WHERE username = $1`, username).Scan(
		&student.Id,
		&student.FirstName,
		&student.Username,
		&student.Password,
		&student.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return student, nil
}

func (s *PersonStore) GetByCourseId(courseId int) ([]*Person, error) {
	rows, err := s.db.Query(`SELECT id, first_name, username, password, role
	FROM person
  JOIN student_course ON student_course.student_id = person.id
	WHERE student_course.course_id = $1`, courseId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	people := []*Person{}
	for rows.Next() {
		person := &Person{}

		err := rows.Scan(&person.Id, &person.FirstName, &person.Username, &person.Password, &person.Role)
		if err != nil {
			return nil, err
		}
		people = append(people, person)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return people, nil
}

func (s *PersonStore) CompareHashAndPassword(hash, pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}

func (s *PersonStore) ChangePassword(studentId int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE person
  SET password=$1
  WHERE id=$2`, string(hashedPassword), studentId)

	return err
}
