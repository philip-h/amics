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
	Assignments interface {
		GetById(int, string) (*AssignmentSubmission, error)
		GetByUsername(string) ([]*AssignmentWithGrade, error)
		Submit(int, string, *PyFile) error
	}
}

func New(db *sql.DB) Storage {
	return Storage{
		Teachers:    &TeacherStore{db},
		Students:    &StudentStore{db},
		Assignments: &AssignmentStore{db},
	}
}
