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
	Points           int
	DueDate          string
	Visible          bool
	PytestFileId     int
	CourseId         int
}

type Submission struct {
	Id           int
	StudentId    int
	AssignmentId int
	FileId       int
	Grade        int
	Comments     string
}

type PyFile struct {
	Id      int
	Name    string
	Content string
}

type AssignmentStore struct {
	db *sql.DB
}

type AssignmentSubmission struct {
	Assignment
	Submission *Submission
	PyFile     *PyFile
}

func (s *AssignmentStore) GetById(id int, username string) (*AssignmentSubmission, error) {
	assignment := &AssignmentSubmission{}

	// Dealing with the potential of nil values
	var subId sql.NullInt64
	var subUserId sql.NullInt64
	var subAssignmentId sql.NullInt64
	var subFileId sql.NullInt64
	var subGrade sql.NullInt64
	var subComments sql.NullString
	var pyFileId sql.NullInt64
	var pyFileName sql.NullString
	var pyFileContent sql.NullString

	err := s.db.QueryRow(`SELECT 
		a.id, 
		a.unit_name, 
		a.name, 
		a.description, 
		a.required_filename, 
		a.points, 
		a.due_date, 
		a.visible, 
		a.pytest_file_id, 
		a.course_id,
		s.id,
		s.student_id,
		s.assignment_id,
		s.file_id,
		s.grade,
		s.comments,
		f.id,
		f.name,
		f.content
	FROM assignment a
	JOIN student ON a.course_id = student.course_id
	LEFT JOIN submission s ON s.assignment_id = a.id
	LEFT JOIN file f ON s.file_id = f.id
	WHERE a.id = $1 AND student.username = $2`, id, username).Scan(
		&assignment.Assignment.Id,
		&assignment.UnitName,
		&assignment.Assignment.Name,
		&assignment.Description,
		&assignment.RequiredFilename,
		&assignment.Points,
		&assignment.DueDate,
		&assignment.Visible,
		&assignment.PytestFileId,
		&assignment.CourseId,
		&subId,
		&subUserId,
		&subAssignmentId,
		&subFileId,
		&subGrade,
		&subComments,
		&pyFileId,
		&pyFileName,
		&pyFileContent)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if !subId.Valid {
		assignment.Submission = nil
		assignment.PyFile = nil
	} else {
		assignment.Submission = &Submission{
			Id:           int(subId.Int64),
			StudentId:    int(subUserId.Int64),
			AssignmentId: int(subAssignmentId.Int64),
			FileId:       int(subFileId.Int64),
			Grade:        int(subGrade.Int64),
			Comments:     subComments.String,
		}
		assignment.PyFile = &PyFile{
			Id:      int(pyFileId.Int64),
			Name:    pyFileName.String,
			Content: pyFileContent.String,
		}
	}
	return assignment, nil
}

type AssignmentWithGrade struct {
	Assignment
	Grade sql.NullInt64 `json:"grade" db:"grade"`
}

func (s *AssignmentStore) GetByUsername(username string) ([]*AssignmentWithGrade, error) {
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
		a.pytest_file_id, 
		a.course_id ,
		s.grade
	FROM assignment a
	JOIN student ON a.course_id = student.course_id
	LEFT JOIN submission s ON s.assignment_id = a.id AND s.student_id = student.id
	WHERE student.username = $1`, username)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := []*AssignmentWithGrade{}
	for rows.Next() {
		assignment := &AssignmentWithGrade{}
		err := rows.Scan(&assignment.Id, &assignment.UnitName, &assignment.Name, &assignment.Description, &assignment.RequiredFilename, &assignment.Points, &assignment.DueDate, &assignment.Visible, &assignment.PytestFileId, &assignment.CourseId, &assignment.Grade)
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

func (s *AssignmentStore) Submit(assignmentId int, username string, pyFile *PyFile) error {
	// For now, just replace any existing submission with the new one... no grading
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var studentId int
	err = tx.QueryRow("SELECT id FROM student WHERE username = $1", username).Scan(&studentId)
	if err != nil {
		return err
	}

	row, err := tx.Exec("INSERT INTO file (name, content) VALUES ($1, $2)", pyFile.Name, pyFile.Content)
	if err != nil {
		return err
	}

	id, err := row.LastInsertId()
	fileId := int(id)

	// Check if a submission already exists for this student and assignment
	var existingSubmissionId int
	err = tx.QueryRow("SELECT id FROM submission WHERE student_id = $1 AND assignment_id = $2", studentId, assignmentId).Scan(&existingSubmissionId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if existingSubmissionId > 0 {
		// Update existing submission
		_, err = tx.Exec("UPDATE submission SET file_id = $1, grade = NULL, comments = NULL WHERE id = $2", fileId, existingSubmissionId)
		if err != nil {
			return err
		}
	} else {
		// Insert new submission
		_, err = tx.Exec("INSERT INTO submission (student_id, assignment_id, file_id) VALUES ($1, $2, $3)", studentId, assignmentId, fileId)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}
