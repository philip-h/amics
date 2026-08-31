package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Submission struct {
	Id           int
	StudentId    int
	AssignmentId int
	Code         string
	Grade        int
	SubmittedOn  time.Time
	Comments     sql.NullString
	Status       string
	GradedOn     sql.NullTime
}

type SubmissionExport struct {
	StudentNumber  string
	FirstName      string
	AssignmentName string
	Grade          sql.NullInt16
	Comments       sql.NullString
}

type SubmissionImport struct {
	StudentNumber string
	AssignmentId  string
	Grade         sql.NullInt16
	Comments      sql.NullString
}

type SubmissionStore struct {
	db *sql.DB
}

func (s *SubmissionStore) Create(ctx context.Context, assignmentId, studentId int, code string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db(submission.create) %w", err)
	}
	defer tx.Rollback()

	// Check if a submission already exists for this student and assignment
	var submissionId int
	err = tx.QueryRowContext(ctx,
		`SELECT id 
    FROM submission 
    WHERE student_id = $1 AND assignment_id = $2`,
		studentId, assignmentId).Scan(&submissionId)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("db(submission.create) %w", err)
	}

	if submissionId > 0 {
		// Update existing submission
		_, err = tx.ExecContext(ctx,
			`UPDATE submission
      SET code=$1, status='grading', comments='Working on it...', submitted_on = NOW()
      WHERE id = $2`, code, submissionId)
		if err != nil {
			return fmt.Errorf("db(submission.create) %w", err)
		}
	} else {
		// Insert new submission
		_, err = tx.ExecContext(ctx,
			`INSERT INTO submission (student_id, assignment_id, code, grade, comments, status)
      VALUES ($1, $2, $3, 0, 'Working on it...', 'grading')`,
			studentId,
			assignmentId,
			code)
		if err != nil {
			return fmt.Errorf("db(submission.create)  %w", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("db(submission.create)  %w", err)
	}
	return nil
}

func (s *SubmissionStore) GetNextPendingSubmission(ctx context.Context) (*Submission, error) {
	submission := &Submission{}
	err := s.db.QueryRowContext(ctx,
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
		return nil, fmt.Errorf("db(submission.GetNextPendingSubmission) %w", err)
	}

	return submission, nil
}

func (s *SubmissionStore) GetByAssignmentAndStudentIds(ctx context.Context, assignmentId, studentId int) (*Submission, error) {
	submission := &Submission{}

	err := s.db.QueryRowContext(ctx, `SELECT 
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
		return nil, fmt.Errorf("db(submission.GetByAssignmentAndStudentIds) %w", err)
	}
	return submission, nil
}

func (s *SubmissionStore) ListByAssignmentId(ctx context.Context, assignmentId int) ([]*Submission, error) {
	rows, err := s.db.QueryContext(ctx, `
  SELECT id, student_id, assignment_id, code, grade, submitted_on, comments, status, graded_on
  FROM submission
  WHERE assignment_id = $1`, assignmentId)

	if err != nil {
		return nil, fmt.Errorf("db(submission.ListByAssignmentId) %w", err)
	}
	defer rows.Close()

	submissions := []*Submission{}
	for rows.Next() {
		submission := &Submission{}
		err := rows.Scan(
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
			return nil, fmt.Errorf("db(submission.ListByAssignmentId) %w", err)
		}
		submissions = append(submissions, submission)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("db(submission.ListByAssignmentId) %w", err)
	}
	return submissions, nil
}

func (s *SubmissionStore) Update(ctx context.Context, submission *Submission) error {
	_, err := s.db.ExecContext(ctx, `UPDATE submission
  SET grade = $1, comments = $2, status = $3, graded_on = NOW() 
  WHERE id=$4`, submission.Grade, submission.Comments, submission.Status, submission.Id)

	return fmt.Errorf("db(submission.Update) %w", err)
}

func (s *SubmissionStore) UpdateAll(ctx context.Context, submissions []*SubmissionImport) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db(submission.UpdateAll) %w", err)
	}
	defer tx.Rollback()

	for _, submission := range submissions {
		_, err := tx.ExecContext(ctx, `INSERT INTO submission (student_id, assignment_id, code, grade, submitted_on, comments, status, graded_on)
		VALUES ($1, $2, 'Graded by Mr. Habib', $3, NOW(), $4, 'completed', NOW())
		ON CONFLICT (student_id, assignment_id)
		DO UPDATE SET
			code = EXCLUDED.code,
			grade = EXCLUDED.grade,
			submitted_on = EXCLUDED.submitted_on,
			comments = EXCLUDED.comments,
			status = EXCLUDED.status,
			graded_on = EXCLUDED.graded_on`,
			submission.StudentNumber,
			submission.AssignmentId,
			submission.Grade,
			submission.Comments)
		if err != nil {
			return fmt.Errorf("db(submission.UpdateAll) %w", err)
		}
	}

	return tx.Commit()
}

func (s *SubmissionStore) ListByCourseId(ctx context.Context, courseId int) ([]*SubmissionExport, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT person.id, person.first_name, assignment.name, submission.grade, submission.comments
  FROM assignment
  JOIN student_course on student_course.course_id = assignment.course_id
  JOIN person on student_course.student_id = person.id
  LEFT JOIN submission
    ON submission.student_id = person.id 
    AND submission.assignment_id = assignment.id
  WHERE assignment.course_id = $1 
  ORDER BY person.id`, courseId)

	if err != nil {
		return nil, fmt.Errorf("db(submission.ListByCourseId) %w", err)
	}
	defer rows.Close()

	submissionExport := []*SubmissionExport{}
	for rows.Next() {
		export := &SubmissionExport{}
		err := rows.Scan(&export.StudentNumber, &export.FirstName, &export.AssignmentName, &export.Grade, &export.Comments)
		if err != nil {
			return nil, err
		}
		submissionExport = append(submissionExport, export)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("db(submission.ListByCourseId) %w", err)
	}
	return submissionExport, nil
}
