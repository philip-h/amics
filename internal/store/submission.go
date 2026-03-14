package store

import (
	"database/sql"
	"log"
)

type Submission struct {
	Id           int
	StudentId    int
	AssignmentId int
	Code         string
	Grade        int
	SubmittedOn  int64
	Comments     sql.NullString
	Status       sql.NullString
	GradedOn     sql.NullInt64
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
	var existingSubmissionId int
	err = tx.QueryRow(
		`SELECT id 
    FROM submission 
    WHERE student_id = $1 AND assignment_id = $2`,
		studentId, assignmentId).Scan(&existingSubmissionId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if existingSubmissionId > 0 {
		// Update existing submission
		_, err = tx.Exec(
			`UPDATE submission
      SET code=$1, status='grading', comments='Working on it...', submitted_on = EXTRACT(EPOCH FROM now())
      WHERE id = $2`, code, existingSubmissionId)
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

func (s *SubmissionStore) Update(submission *Submission) error {
	log.Printf("Updating subission %d with status %v", submission.Id, submission.Status)
	_, err := s.db.Exec(`UPDATE submission
  SET grade = $1, comments = $2, status = $3, graded_on = EXTRACT(EPOCH FROM now())
  WHERE id=$4`, submission.Grade, submission.Comments, submission.Status, submission.Id)

	return err
}
