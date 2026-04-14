package store

import "database/sql"

type Storage struct {
	Teachers interface {
		GetByUsername(string) (*Teacher, error)
	}
	Students interface {
		Create(*Student) error
		GetByUsername(string) (*Student, error)
		GetByCourseId(int) ([]*Student, error)
		ChangePassword(int, string) error
		CompareHashAndPassword(string, string) bool
	}

	Courses interface {
		Create(*Course) error
		GetById(int) (*Course, error)
		GetByJoinCode(string) (*Course, error)
		GetByTeacherId(int) ([]*Course, error)
		Update(*Course) error
	}

	Assignments interface {
		Create(*Assignment) error
		GetWithGradeByStudentId(int) ([]*AssignmentWithGrade, error)
		GetWithSubmissionByAssignmentAndStudentIds(int, int) (*AssignmentSubmission, error)

		GetById(int) (*Assignment, error)
		GetByCourseId(int) ([]*Assignment, error)
		Update(*Assignment) error
	}

	Submissions interface {
		Create(int, int, string) error
		GetByAssignmentAndStudentIds(int, int) (*Submission, error)
		GetNextPendingSubmission() (*Submission, error)
		Update(*Submission) error
		GetAllByCourseId(int) ([]*SubmissionExport, error)
	}
}

func New(db *sql.DB) Storage {
	return Storage{
		Teachers:    &TeacherStore{db},
		Students:    &StudentStore{db},
		Assignments: &AssignmentStore{db},
		Courses:     &CourseStore{db},
		Submissions: &SubmissionStore{db},
	}
}
