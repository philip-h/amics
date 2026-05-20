package storage

import "database/sql"

type Storage struct {
	Teachers interface {
		GetByUsername(string) (*Teacher, error)
	}

	People interface {
		Create(*Person, int) error
		GetById(int) (*Person, error)
		GetByUsername(string) (*Person, error)
		GetByCourseId(int) ([]*Person, error)
		ChangePassword(int, string) error
		CompareHashAndPassword(string, string) bool
	}

	Courses interface {
		Create(*Course, int) error
		GetById(int) (*Course, error)
		GetByJoinCode(string) (*Course, error)
		GetByTeacherId(int) ([]*Course, error)
    Update(*Course) error
	}

	Assignments interface {
		Create(*Assignment) error
		GetWithGrade(int) ([]*AssignmentWithGrade, error)
		GetWithSubmissionByAssignmentAndStudentIds(int, int) (*AssignmentSubmission, error)

		GetById(int) (*Assignment, error)
		GetByCourseId(int) ([]*Assignment, error)
		Update(*Assignment) error
		UpdateAll(int, [][]string) error
	}

	Submissions interface {
		Create(int, int, string) error
		GetByAssignmentAndStudentIds(int, int) (*Submission, error)
		GetByAssignmentId(int) ([]*Submission, error)
		GetNextPendingSubmission() (*Submission, error)
		Update(*Submission) error
		GetAllByCourseId(int) ([]*SubmissionExport, error)
	}

	Sessions interface {
		Create(int, bool) (string, error)
		GetById(string) (*Session, error)
		Delete(string) error
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
