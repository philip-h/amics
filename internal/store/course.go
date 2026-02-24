package store

import "database/sql"

type Course struct {
	Id        int
	Year      int
	Semester  string
	Name      string
	TeacherId int
}

type CourseStore struct {
	db *sql.DB
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
