package store

import "database/sql"

type Course struct {
	Id        int
	Year      int
	Semester  int
	Name      string
	JoinCode  string
	TeacherId int
}

type CourseStore struct {
	db *sql.DB
}

func (s *CourseStore) Create(course *Course) error {
	_, err := s.db.Exec("INSERT INTO course (year, semester, name, join_code, teacher_id) VALUES ($1, $2, $3, $4, $5)",
		course.Year, course.Semester, course.Name, course.JoinCode, course.TeacherId)
	return err
}

func (s *CourseStore) GetById(courseId int) (*Course, error) {
	course := &Course{}

	err := s.db.QueryRow(`SELECT 
		id, 
		year,
		semester,
		name,
    join_code
	FROM course
	WHERE id = $1`, courseId).Scan(
		&course.Id,
		&course.Year,
		&course.Semester,
		&course.Name,
		&course.JoinCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return course, nil
}

func (s *CourseStore) GetByJoinCode(joinCode string) (*Course, error) {
	course := &Course{}

	err := s.db.QueryRow(`SELECT 
		id, 
		year,
		semester,
		name,
    join_code
	FROM course
	WHERE join_code = $1`, joinCode).Scan(
		&course.Id,
		&course.Year,
		&course.Semester,
		&course.Name,
		&course.JoinCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return course, nil
}

func (s *CourseStore) GetByTeacherId(teacherId int) ([]*Course, error) {
	rows, err := s.db.Query("SELECT id, year, semester, name, join_code, teacher_id FROM course WHERE teacher_id = $1;", teacherId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := []*Course{}
	for rows.Next() {
		course := &Course{}
		err := rows.Scan(&course.Id, &course.Year, &course.Semester, &course.Name, &course.JoinCode, &course.TeacherId)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	return courses, nil
}

func (s *CourseStore) Update(course *Course) error {
	_, err := s.db.Exec(`UPDATE course
  SET year=$1, semester=$2, name=$3, join_code=$4
  WHERE id=$5`,
		course.Year, course.Semester, course.Name, course.JoinCode, course.Id)

	return err
}
