package store

import "database/sql"

type Course struct {
	Id        int
	Year      int
	Semester  int
	Name      string
	TeacherId int
}

type CourseStore struct {
	db *sql.DB
}

func (s *CourseStore) Create(course *Course) error {
	_, err := s.db.Exec("INSERT INTO course (year, semester, name, teacher_id) VALUES ($1, $2, $3, $4)", course.Year, course.Semester, course.Name, course.TeacherId)
	return err
}

func (s *CourseStore) GetById(courseId int) (*Course, error) {
	course := &Course{}

	err := s.db.QueryRow(`SELECT 
		id, 
		year,
		semester,
		name
	FROM course
	WHERE id = $1`, courseId).Scan(
		&course.Id,
		&course.Year,
		&course.Semester,
		&course.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return course, nil
}

func (s *CourseStore) GetByTeacherId(teacherId int) ([]*Course, error) {
	rows, err := s.db.Query("SELECT id, year, semester, name, teacher_id FROM course WHERE teacher_id = ?", teacherId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := []*Course{}
	for rows.Next() {
		course := &Course{}
		err := rows.Scan(&course.Id, &course.Year, &course.Semester, &course.Name, &course.TeacherId)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	return courses, nil
}

func (s *CourseStore) Update(course *Course) error {
	_, err := s.db.Exec(`UPDATE course
  SET year=?, semester=?, name=?
  WHERE id=?`,
		course.Year, course.Semester, course.Name, course.Id)

	return err
}
