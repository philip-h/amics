package store

import "database/sql"

type Storage struct {
	Teachers interface {
		GetByUsername(string) (*Teacher, error)
	}
	Students interface {
		Create(*Student) error
		GetByUsername(string) (*Student, error)
	}

	Courses interface {
		GetByTeacherId(int) ([]*Course, error)
	}

	Assignments interface {
		GetWithGradeByStudentId(int) ([]*AssignmentWithGrade, error)
		GetWithSubmissionByAssignmentAndStudentIds(int, int) (*AssignmentSubmission, error)
		Submit(int, int, *PyFile) error

		GetById(int) (*Assignment, error)
		GetByCourseId(int) ([]*Assignment, error)
	}
}

func New(db *sql.DB) Storage {
	return Storage{
		Teachers:    &TeacherStore{db},
		Students:    &StudentStore{db},
		Assignments: &AssignmentStore{db},
		Courses:     &CourseStore{db},
	}
}
