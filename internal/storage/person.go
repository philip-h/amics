package storage

import (
	"context"
	"database/sql"
	"fmt"

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

func (s *PersonStore) Create(ctx context.Context, student *Person, courseId int) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(student.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("db(person.create) %w", err)
	}
	student.Password = string(hashedPassword)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db(person.create) %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
  INSERT INTO person (id, first_name, username, password) 
  VALUES ($1, $2, $3, $4)`,
		student.Id,
		student.FirstName,
		student.Username,
		student.Password)

	if err != nil {
		return fmt.Errorf("db(person.create) %w", err)
	}

	_, err = tx.ExecContext(ctx, `
  INSERT INTO student_course (student_id, course_id)
  VALUES ($1, $2)`,
		student.Id, courseId,
	)
	if err != nil {
		return fmt.Errorf("db(person.create) %w", err)
	}

	return tx.Commit()
}

func (s *PersonStore) GetById(ctx context.Context, id int) (*Person, error) {
	person := &Person{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, first_name, username, password, role
    FROM person
    WHERE id = $1`, id,
	).Scan(&person.Id, &person.FirstName, &person.Username, &person.Password, &person.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db(person.getById): %w", err)
	}
	return person, nil
}

func (s *PersonStore) GetByUsername(ctx context.Context, username string) (*Person, error) {
	student := &Person{}
	err := s.db.QueryRowContext(ctx, `SELECT id, first_name, username, password, role
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
		return nil, fmt.Errorf("db(person.getByUsername): %w", err)
	}
	return student, nil
}

func (s *PersonStore) GetByCourseId(ctx context.Context, courseId int) ([]*Person, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, first_name, username, password, role
	FROM person
  JOIN student_course ON student_course.student_id = person.id
	WHERE student_course.course_id = $1`, courseId)

	if err != nil {
		return nil, fmt.Errorf("db(person.getByCourseId): %w", err)
	}
	defer rows.Close()

	people := []*Person{}
	for rows.Next() {
		person := &Person{}

		err := rows.Scan(&person.Id, &person.FirstName, &person.Username, &person.Password, &person.Role)
		if err != nil {
			return nil, fmt.Errorf("db(person.getByCourseId): %w", err)
		}
		people = append(people, person)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("db(person.getByCourseId): %w", err)
	}
	return people, nil
}

func (s *PersonStore) CompareHashAndPassword(hash, pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}

func (s *PersonStore) ChangePassword(ctx context.Context, studentId int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("db(person.changePassword) %w", err)
	}
	_, err = s.db.ExecContext(ctx, `UPDATE person
  SET password=$1
  WHERE id=$2`, string(hashedPassword), studentId)

	if err != nil {
		return fmt.Errorf("db(person.changePassword) %w", err)
	}
	return nil
}
