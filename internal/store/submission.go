package store

import (
	"database/sql"
)

type Submission struct {
	Id           int
	StudentId    int
	AssignmentId int
	Code         string
	Grade        int
	SubmittedOn  int64
	Comments     sql.NullString
	Status       string
	GradedOn     sql.NullInt64
}

type SubmissionExport struct {
	StudentNumber  string
	AssignmentName string
	Grade          sql.NullInt16
}

type SubmissionStore struct {
	db *sql.DB
}

func (s *SubmissionStore) Create(assignmentId, studentId int, code string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if a submission already exists for this student and assignment
	var submissionId int
	err = tx.QueryRow(
		`SELECT id 
    FROM submission 
    WHERE student_id = $1 AND assignment_id = $2`,
		studentId, assignmentId).Scan(&submissionId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if submissionId > 0 {
		// Update existing submission
		_, err = tx.Exec(
			`UPDATE submission
      SET code=$1, status='grading', comments='Working on it...', submitted_on = EXTRACT(EPOCH FROM now())
      WHERE id = $2`, code, submissionId)
		if err != nil {
			return err
		}
	} else {
		// Insert new submission
		_, err = tx.Exec(
			`INSERT INTO submission (student_id, assignment_id, code, grade, comments, status)
      VALUES ($1, $2, $3, 0, 'Working on it...', 'grading')`,
			studentId,
			assignmentId,
			code)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}

func (s *SubmissionStore) GetNextPendingSubmission() (*Submission, error) {
	submission := &Submission{}
	err := s.db.QueryRow(
		`SELECT id, student_id, assignment_id, code, grade, comments, status, submitted_on, graded_on
    FROM submission
    WHERE status='grading'
    ORDER BY submitted_on ASC
    LIMIT 1`).Scan(
		&submission.Id,
		&submission.StudentId,
		&submission.AssignmentId,
		&submission.Code,
		&submission.Grade,
		&submission.Comments,
		&submission.Status,
		&submission.SubmittedOn,
		&submission.GradedOn,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return submission, nil
}

func (s *SubmissionStore) GetByAssignmentAndStudentIds(assignmentId, studentId int) (*Submission, error) {
	submission := &Submission{}

	err := s.db.QueryRow(`SELECT 
		id, 
    student_id,
    assignment_id,
    code,
    grade,
    submitted_on,
    comments,
    status,
    graded_on
	FROM submission
	WHERE student_id = $1 AND assignment_id = $2`, studentId, assignmentId).Scan(
		&submission.Id,
		&submission.StudentId,
		&submission.AssignmentId,
		&submission.Code,
		&submission.Grade,
		&submission.SubmittedOn,
		&submission.Comments,
		&submission.Status,
		&submission.GradedOn)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return submission, nil
}

func (s *SubmissionStore) Update(submission *Submission) error {
	_, err := s.db.Exec(`UPDATE submission
  SET grade = $1, comments = $2, status = $3, graded_on = EXTRACT(EPOCH FROM now())
  WHERE id=$4`, submission.Grade, submission.Comments, submission.Status, submission.Id)

	return err
}

func (s *SubmissionStore) GetAllByCourseId(courseId int) ([]*SubmissionExport, error) {
	rows, err := s.db.Query(`SELECT student.student_number, assignment.name, submission.grade
  FROM assignment
  JOIN student on student.course_id = assignment.course_id
  LEFT JOIN submission
    ON submission.student_id = student.id 
    AND submission.assignment_id = assignment.id
  WHERE assignment.course_id = $1
  ORDER BY student.student_number`, courseId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	submissionExport := []*SubmissionExport{}
	for rows.Next() {
		export := &SubmissionExport{}
		err := rows.Scan(&export.StudentNumber, &export.AssignmentName, &export.Grade)
		if err != nil {
			return nil, err
		}
		submissionExport = append(submissionExport, export)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return submissionExport, nil
}
