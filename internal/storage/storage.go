package storage

import (
	"context"
	"database/sql"
)

type Storage struct {
	Teachers interface {
		GetByUsername(context.Context, string) (*Teacher, error)
	}

	People interface {
		Create(context.Context, *Person, int) error
		GetById(context.Context, int) (*Person, error)
		GetByUsername(context.Context, string) (*Person, error)
		GetByCourseId(context.Context, int) ([]*Person, error)
		ChangePassword(context.Context, int, string) error
		CompareHashAndPassword(string, string) bool
	}

	Courses interface {
		Create(context.Context, *Course, int) error
		GetById(context.Context, int) (*Course, error)
		GetByJoinCode(context.Context, string) (*Course, error)
		GetByStudentId(context.Context, int) ([]*Course, error)
		GetByTeacherId(context.Context, int) ([]*Course, error)
		Update(context.Context, *Course) error
	}

	Assignments interface {
		Create(context.Context, *Assignment) error
		GetWithGrade(context.Context, int, int) ([]*AssignmentWithGrade, error)
		GetWithSubmissionByAssignmentAndStudentIds(context.Context, int, int) (*AssignmentSubmission, error)
		GetAllWithSubmissionByCourseAndStudentIds(context.Context, int, int) ([]*AssignmentSubmission, error)

		GetById(context.Context, int) (*Assignment, error)
		GetByCourseId(context.Context, int) ([]*Assignment, error)
		Update(context.Context, *Assignment) error

		GetUnitNamesByCourseId(context.Context, int) ([]string, error)
	}

	Submissions interface {
		Create(context.Context, int, int, string) error
		GetByAssignmentAndStudentIds(context.Context, int, int) (*Submission, error)
		ListByAssignmentId(context.Context, int) ([]*Submission, error)
		GetNextPendingSubmission(context.Context) (*Submission, error)
		Update(context.Context, *Submission) error
		UpdateAll(context.Context, []*SubmissionImport) error
		ListByCourseId(context.Context, int) ([]*SubmissionExport, error)
	}

	Sessions interface {
		Create(context.Context, int, bool) (string, error)
		GetById(context.Context, string) (*Session, error)
		Delete(context.Context, string) error
	}
}

func New(db *sql.DB) *Storage {
	return &Storage{
		Teachers:    &TeacherStore{db},
		People:      &PersonStore{db},
		Assignments: &AssignmentStore{db},
		Courses:     &CourseStore{db},
		Submissions: &SubmissionStore{db},
		Sessions:    &SessionStore{db},
	}
}
