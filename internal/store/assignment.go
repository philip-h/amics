package store

import (
	"database/sql"
)

type Assignment struct {
	Id               int
	UnitName         string
	Name             string
	Description      string
	RequiredFilename string
	PytestCode       string
	Points           int
	DueDate          string
	Visible          bool
	CourseId         int
}

type AssignmentStore struct {
	db *sql.DB
}

type AssignmentSubmission struct {
	Assignment
	Submission *Submission
}

func (s *AssignmentStore) Create(assignment *Assignment) error {
	_, err := s.db.Exec(
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

	return err
}

func (s *AssignmentStore) GetWithSubmissionByAssignmentAndStudentIds(assignmentId int, studentId int) (*AssignmentSubmission, error) {
	aws := &AssignmentSubmission{}

	// Dealing with the potential of nil values
	var subId sql.NullInt64
	var subUserId sql.NullInt64
	var subAssignmentId sql.NullInt64
	var subCode sql.NullString
	var subGrade sql.NullInt64
	var subComments sql.NullString
	var subStatus sql.NullString

	err := s.db.QueryRow(`SELECT 
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
		s.comments,
    s.status
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
		&subComments,
		&subStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
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
			Comments:     subComments,
			Status:       subStatus,
		}
	}
	return aws, nil
}

type AssignmentWithGrade struct {
	Assignment
	Grade sql.NullInt64
}

func (s *AssignmentStore) GetWithGradeByStudentId(studentId int) ([]*AssignmentWithGrade, error) {
	// Get all assignments for the course the user is enrolled in, along with
	// the user's submission grade for each assignment (if it exists)
	rows, err := s.db.Query(`SELECT 
		a.id, 
		a.unit_name, 
		a.name, 
		a.description, 
		a.required_filename, 
		a.points, 
		a.due_date, 
		a.visible, 
		a.course_id ,
		s.grade
	FROM assignment a
	JOIN student ON a.course_id = student.course_id
	LEFT JOIN submission s ON s.assignment_id = a.id AND s.student_id = student.id
	WHERE student.id = $1`, studentId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := []*AssignmentWithGrade{}
	for rows.Next() {
		assignment := &AssignmentWithGrade{}
		err := rows.Scan(&assignment.Id, &assignment.UnitName, &assignment.Name, &assignment.Description, &assignment.RequiredFilename, &assignment.Points, &assignment.DueDate, &assignment.Visible, &assignment.CourseId, &assignment.Grade)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return assignments, nil
}

func (s *AssignmentStore) GetById(assignmentId int) (*Assignment, error) {
	assignment := &Assignment{}

	err := s.db.QueryRow(`SELECT 
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
		return nil, err
	}
	return assignment, nil
}

func (s *AssignmentStore) GetByCourseId(courseId int) ([]*Assignment, error) {

	rows, err := s.db.Query(`SELECT 
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
	WHERE course_id = $1`, courseId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := []*Assignment{}
	for rows.Next() {
		assignment := &Assignment{}

		err := rows.Scan(&assignment.Id, &assignment.UnitName, &assignment.Name, &assignment.Description, &assignment.RequiredFilename, &assignment.PytestCode, &assignment.Points, &assignment.DueDate, &assignment.Visible, &assignment.CourseId)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return assignments, nil
}

func (s *AssignmentStore) Update(assignment *Assignment) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE assignment 
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
		return err
	}

	return tx.Commit()
}
