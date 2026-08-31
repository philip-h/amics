package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Assignment struct {
	Id               int
	UnitName         string
	Name             string
	Description      string
	RequiredFilename string
	PytestCode       string
	Points           int
	DueDate          time.Time
	Visible          bool
	CourseId         int
}

type AssignmentSubmission struct {
	Assignment
	Submission *Submission
}

type AssignmentStore struct {
	db *sql.DB
}

func (s *AssignmentStore) Create(ctx context.Context, assignment *Assignment) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO assignment (unit_name, name, description, required_filename, pytest_code, points, due_date, visible, course_id)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		assignment.UnitName,
		assignment.Name,
		assignment.Description,
		assignment.RequiredFilename,
		assignment.PytestCode,
		assignment.Points,
		assignment.DueDate,
		assignment.Visible,
		assignment.CourseId)

	return fmt.Errorf("db(assignment.create): %w", err)
}

func (s *AssignmentStore) GetWithSubmissionByAssignmentAndStudentIds(ctx context.Context, assignmentId int, studentId int) (*AssignmentSubmission, error) {
	aws := &AssignmentSubmission{}

	// Dealing with the potential of nil values
	var (
		subId           sql.NullInt64
		subUserId       sql.NullInt64
		subAssignmentId sql.NullInt64
		subCode         sql.NullString
		subGrade        sql.NullInt64
		subSubmittedOn  sql.NullTime
		subComments     sql.NullString
		subStatus       sql.NullString
		subGradedOn     sql.NullTime
	)

	err := s.db.QueryRowContext(ctx, `SELECT 
		a.id, 
		a.unit_name, 
		a.name, 
		a.description, 
		a.required_filename, 
		a.pytest_code,
		a.points, 
		a.due_date, 
		a.visible, 
		a.course_id,
		s.id,
		s.student_id,
		s.assignment_id,
		s.code,
		s.grade,
		s.submitted_on,
		s.comments,
		s.status,
		s.graded_on
	FROM assignment a
	LEFT JOIN submission s ON s.assignment_id = a.id AND s.student_id = $1
	WHERE a.id = $2`, studentId, assignmentId).Scan(
		&aws.Assignment.Id,
		&aws.UnitName,
		&aws.Assignment.Name,
		&aws.Description,
		&aws.RequiredFilename,
		&aws.PytestCode,
		&aws.Points,
		&aws.DueDate,
		&aws.Visible,
		&aws.CourseId,
		&subId,
		&subUserId,
		&subAssignmentId,
		&subCode,
		&subGrade,
		&subSubmittedOn,
		&subComments,
		&subStatus,
		&subGradedOn)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db(assignment.getWithSubmission): %w", err)
	}

	if !subId.Valid {
		aws.Submission = nil
	} else {
		aws.Submission = &Submission{
			Id:           int(subId.Int64),
			StudentId:    int(subUserId.Int64),
			AssignmentId: int(subAssignmentId.Int64),
			Code:         subCode.String,
			Grade:        int(subGrade.Int64),
			SubmittedOn:  subSubmittedOn.Time,
			Comments:     subComments,
			Status:       subStatus.String,
			GradedOn:     subGradedOn,
		}
	}
	return aws, nil
}

func (s *AssignmentStore) GetAllWithSubmissionByCourseAndStudentIds(ctx context.Context, courseId, studentId int) ([]*AssignmentSubmission, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT
		a.id,
		a.course_id,
		a.unit_name,
		a.name,
		a.description,
		a.required_filename,
		a.points,
		a.due_date,
		a.visible,
		s.id,
		s.code, 		     
		s.grade,          
		s.submitted_on,
		s.comments,		 
		s.status,		 
		s.graded_on      
	FROM assignment a
	LEFT JOIN submission s
		ON s.assignment_id = a.id
		AND s.student_id = $1
	WHERE a.course_id = $2
	ORDER BY a.unit_name, a.due_date;`, studentId, courseId)
	if err != nil {
		return nil, fmt.Errorf("db(assignment.getAllWithSubmission): %w", err)
	}
	defer rows.Close()

	assignments := []*AssignmentSubmission{}
	for rows.Next() {
		var a Assignment

		// sql.Null* for every submission column — they're NULL when not submitted
		var (
			submissionID sql.NullInt64
			code         sql.NullString
			grade        sql.NullInt64
			submittedOn  sql.NullTime
			comments     sql.NullString
			status       sql.NullString
			gradedOn     sql.NullTime
		)

		err := rows.Scan(
			&a.Id,
			&a.CourseId,
			&a.UnitName,
			&a.Name,
			&a.Description,
			&a.RequiredFilename,
			&a.Points,
			&a.DueDate,
			&a.Visible,
			&submissionID,
			&code,
			&grade,
			&submittedOn,
			&comments,
			&status,
			&gradedOn,
		)
		if err != nil {
			return nil, fmt.Errorf("db(assignment.getAllWithSubmission): %w", err)
		}

		aws := AssignmentSubmission{Assignment: a}

		// only populate the submission if one exists
		if submissionID.Valid {
			aws.Submission = &Submission{
				Id:           int(submissionID.Int64),
				StudentId:    studentId,
				AssignmentId: a.Id,
				Code:         code.String,
				Grade:        int(grade.Int64),
				SubmittedOn:  submittedOn.Time,
				Comments:     comments,
				Status:       status.String,
				GradedOn:     gradedOn,
			}
		}

		assignments = append(assignments, &aws)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db(assignment.getAllWithSubmission): %w", err)
	}

	return assignments, nil
}

type AssignmentWithGrade struct {
	Assignment
	Grade sql.NullInt64
}

func (s *AssignmentStore) GetWithGrade(ctx context.Context, studentId, courseId int) ([]*AssignmentWithGrade, error) {
	// Get all assignments for the given course, along with
	// the user's submission grade for each assignment (if it exists)
	rows, err := s.db.QueryContext(ctx, `SELECT 
		a.id, 
		a.unit_name, 
		a.name, 
		a.description, 
		a.required_filename, 
		a.points, 
		a.due_date, 
		a.visible, 
		a.course_id,
		s.grade
	FROM assignment a
	JOIN student_course sc ON sc.course_id = a.course_id
	JOIN person p ON sc.student_id = p.id
	LEFT JOIN submission s ON s.assignment_id = a.id AND s.student_id = p.id
	WHERE p.id = $1 AND a.course_id = $2
  ORDER BY a.due_date DESC, a.unit_name DESC`, studentId, courseId)

	if err != nil {
		return nil, fmt.Errorf("db(assignment.getWithGrade): %w", err)
	}
	defer rows.Close()

	assignments := []*AssignmentWithGrade{}
	for rows.Next() {
		assignment := &AssignmentWithGrade{}
		err := rows.Scan(&assignment.Id, &assignment.UnitName, &assignment.Name, &assignment.Description, &assignment.RequiredFilename, &assignment.Points, &assignment.DueDate, &assignment.Visible, &assignment.CourseId, &assignment.Grade)
		if err != nil {
			return nil, fmt.Errorf("db(getWithGrade): %w", err)
		}
		assignments = append(assignments, assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("db(assignment.getWithGrade): %w", err)
	}
	return assignments, nil
}

func (s *AssignmentStore) GetById(ctx context.Context, assignmentId int) (*Assignment, error) {
	assignment := &Assignment{}

	err := s.db.QueryRowContext(ctx, `SELECT 
		id, 
		unit_name, 
		name, 
		description, 
		required_filename, 
    pytest_code,
		points, 
		due_date, 
		visible, 
		course_id
	FROM assignment
	WHERE id = $1`, assignmentId).Scan(
		&assignment.Id,
		&assignment.UnitName,
		&assignment.Name,
		&assignment.Description,
		&assignment.RequiredFilename,
		&assignment.PytestCode,
		&assignment.Points,
		&assignment.DueDate,
		&assignment.Visible,
		&assignment.CourseId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db(assignment.getById): %w", err)
	}
	return assignment, nil
}

func (s *AssignmentStore) GetByCourseId(ctx context.Context, courseId int) ([]*Assignment, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT 
		id, 
		unit_name, 
		name, 
		description, 
		required_filename, 
		pytest_code,
		points, 
		due_date, 
		visible, 
		course_id
	FROM assignment
	WHERE course_id = $1
  ORDER BY due_date`, courseId)

	if err != nil {
		return nil, fmt.Errorf("db(assignment.getByCourseId): %w", err)
	}
	defer rows.Close()

	assignments := []*Assignment{}
	for rows.Next() {
		assignment := &Assignment{}

		err := rows.Scan(&assignment.Id, &assignment.UnitName, &assignment.Name, &assignment.Description, &assignment.RequiredFilename, &assignment.PytestCode, &assignment.Points, &assignment.DueDate, &assignment.Visible, &assignment.CourseId)
		if err != nil {
			return nil, fmt.Errorf("db(assignment.getByCourseId): %w", err)
		}
		assignments = append(assignments, assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("db(assignment.getByCourseId): %w", err)
	}
	return assignments, nil
}

func (s *AssignmentStore) Update(ctx context.Context, assignment *Assignment) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db(assignment.update): %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE assignment 
  SET unit_name=$1, name=$2, description=$3, required_filename=$4, points=$5, pytest_code=$6, due_date=$7, visible=$8, course_id=$9
  WHERE id=$10`,
		assignment.UnitName,
		assignment.Name,
		assignment.Description,
		assignment.RequiredFilename,
		assignment.Points,
		assignment.PytestCode,
		assignment.DueDate,
		assignment.Visible,
		assignment.CourseId,
		assignment.Id)

	if err != nil {
		return fmt.Errorf("db(assignment.update): %w", err)
	}

	return tx.Commit()
}

func (s *AssignmentStore) GetUnitNamesByCourseId(ctx context.Context, courseId int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT unit_name FROM assignment WHERE course_id = $1`, courseId)
	if err != nil {
		return nil, fmt.Errorf("db(assignment.getUnitNames): %w", err)
	}
	defer rows.Close()

	unitNames := []string{}
	for rows.Next() {
		var unitName string
		err := rows.Scan(&unitName)
		if err != nil {
			return nil, fmt.Errorf("db(assignment.getUnitNames): %w", err)
		}
		unitNames = append(unitNames, unitName)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("db(assignment.getUnitNames): %w", err)
	}
	return unitNames, nil
}
