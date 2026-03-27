package store

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type Student struct {
	Id            int
	StudentNumber string
	Username      string
	Password      string
	CourseId      int
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
  var id int64
	err = s.db.QueryRow("INSERT INTO student (student_number, username, password, course_id) VALUES ($1, $2, $3, $4) RETURNING id", student.StudentNumber, student.Username, student.Password, student.CourseId).Scan(&id)
	if err != nil {
		return err
	}

  // Put the id back into the student
	student.Id = int(id)

	return nil
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

func (s *StudentStore) GetByCourseId(courseId int) ([]*Student, error) {
	rows, err := s.db.Query(`SELECT 
		id,
		student_number,
		username,
		password,
		course_id
	FROM student
	WHERE course_id = $1`, courseId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := []*Student{}
	for rows.Next() {
		student := &Student{}

		err := rows.Scan(&student.Id, &student.StudentNumber, &student.Username, &student.Password, &student.CourseId)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return students, nil
}

func (s *StudentStore) CompareHashAndPassword(hash, pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}

func (s *StudentStore) ChangePassword(studentId int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE student
  SET password=$1
  WHERE id=$2`, string(hashedPassword), studentId)

	return err
}
