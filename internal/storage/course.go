package storage

import (
	"database/sql"
	"fmt"
)

type Course struct {
	Id         int
	CourseCode string
	Section    int
	Name       string
	Year       int
	Semester   int
	JoinCode   string
}

type CourseStore struct {
	db *sql.DB
}

func (s *CourseStore) Create(course *Course, teacherId int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("db(course.create): %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRow(`
  INSERT INTO course (course_code, section, name, year, semester, join_code)
  VALUES ($1, $2, $3, $4)
  RETURNING id`,
		course.CourseCode,
		course.Section,
		course.Name,
		course.Year,
		course.Semester,
		course.JoinCode,
	).Scan(&course.Id)

	if err != nil {
		return fmt.Errorf("db(course.create): %w", err)
	}

	_, err = tx.Exec(`
  INSERT INTO teacher_course
  VALUES ($1, $2)
  ON CONFLICT (teacher_id, course_id)
  DO NOTHING`,
		teacherId,
		course.Id,
	)
	if err != nil {
		return fmt.Errorf("db(course.create): %w", err)
	}

	return tx.Commit()
}

func (s *CourseStore) GetById(courseId int) (*Course, error) {
	course := &Course{}

	err := s.db.QueryRow(`SELECT 
		id, 
		course_code,
		section,
		name,
		year,
		semester,
		join_code
	FROM course
	WHERE id = $1`, courseId).Scan(
		&course.Id,
		&course.CourseCode,
		&course.Section,
		&course.Name,
		&course.Year,
		&course.Semester,
		&course.JoinCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db(course.getById): %w", err)
	}
	return course, nil
}

func (s *CourseStore) GetByJoinCode(joinCode string) (*Course, error) {
	course := &Course{}

	err := s.db.QueryRow(`SELECT 
		id, 
		course_code,
		section,
		name,
		year,
		semester,
		join_code
	FROM course
	WHERE join_code = $1`, joinCode).Scan(
		&course.Id,
		&course.CourseCode,
		&course.Section,
		&course.Name,
		&course.Year,
		&course.Semester,
		&course.JoinCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db(course.getByJoinCode): %w", err)
	}
	return course, nil
}

func (s *CourseStore) GetByTeacherId(teacherId int) ([]*Course, error) {
	rows, err := s.db.Query(`SELECT 
	id, 
	course_code,
	section,
	name,
	year, 
	semester, 
	join_code
  FROM course 
  JOIN teacher_course tc ON course.id = tc.course_id
  WHERE tc.teacher_id = $1;`, teacherId)
	if err != nil {
		return nil, fmt.Errorf("db(course.getByTeacherId): %w", err)
	}
	defer rows.Close()

	courses := []*Course{}
	for rows.Next() {
		course := &Course{}
		err := rows.Scan(
			&course.Id,
			&course.CourseCode,
			&course.Section,
			&course.Name,
			&course.Year,
			&course.Semester,
			&course.JoinCode,
		)
		if err != nil {
			return nil, fmt.Errorf("db(course.getByTeacherId): %w", err)
		}
		courses = append(courses, course)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db(course.getByTeacherId): %w", err)
	}

	return courses, nil
}

func (s *CourseStore) GetByStudentId(studentId int) ([]*Course, error) {
	rows, err := s.db.Query(`SELECT 
		id, 
		course_code,
		section,
		name,
		year, 
		semester,
		join_code
  FROM course 
  JOIN student_course sc ON course.id = sc.course_id
  WHERE sc.student_id = $1;`, studentId)
	if err != nil {
		return nil, fmt.Errorf("db(course.getByStudentId): %w", err)
	}
	defer rows.Close()

	courses := []*Course{}
	for rows.Next() {
		course := &Course{}
		err := rows.Scan(
			&course.Id,
			&course.CourseCode,
			&course.Section,
			&course.Name,
			&course.Year,
			&course.Semester,
			&course.JoinCode,
		)
		if err != nil {
			return nil, fmt.Errorf("db(course.getByStudentId): %w", err)
		}
		courses = append(courses, course)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db(course.getByStudentId): %w", err)
	}

	return courses, nil
}

func (s *CourseStore) Update(course *Course) error {
	_, err := s.db.Exec(`UPDATE course
  SET 
	course_code=$1,
	section=$2,
	name=$3, 
	year=$4, 
	semester=$5, 
	join_code=$6
  WHERE id=$7`,
		course.CourseCode, course.Section, course.Name, course.Year, course.Semester, course.JoinCode, course.Id)
	if err != nil {
		return fmt.Errorf("db(course.update): %w", err)
	}

	return nil
}
