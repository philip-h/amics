package store

// import "database/sql"

// type Submission struct {
// 	Id           int
// 	StudentId    int
// 	AssignmentId int
// 	FileId       int
// 	Grade        int
// 	Comments     string
// }

// type PyFile struct {
// 	Id      int
// 	Name    string
// 	Content string
// }

// type SubmissionStore struct {
// 	db *sql.DB
// }

// // id INTEGER PRIMARY KEY AUTOINCREMENT,
// // student_id INTEGER NOT NULL,
// // assignment_id INTEGER NOT NULL,
// // file_id INTEGER NOT NULL,
// // grade INTEGER,
// // comments TEXT,

// func (s *SubmissionStore) GetByAssignmentStudentIds(assignmentId int, studentId int) (*Submission, error) {
// 	submission := &Submission{}

// 	err := s.db.QueryRow(`
//     SELECT id, student_id, assignment_id, file_id, grade, comments
//     FROM submission
//     WHERE assignment_id = $1
//     AND student_id = $2`, assignmentId, studentId).Scan(
// 		&submission.Id,
// 		&submission.StudentId,
// 		&submission.AssignmentId,
// 		&submission.FileId,
// 		&submission.Grade,
// 		&submission.Comments)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return submission, nil

// }
// func (s *SubmissionStore) SubmitByAssignmentStudentIds(assignmentId, studentId int, pyFile *PyFile) error {
// 	// For now, just replace any existing submission with the new one... no grading
// 	tx, err := s.db.Begin()
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()

// 	row, err := tx.Exec("INSERT INTO file (name, content) VALUES ($1, $2)", pyFile.Name, pyFile.Content)
// 	if err != nil {
// 		return err
// 	}

// 	id, err := row.LastInsertId()
// 	fileId := int(id)

// 	// Check if a submission already exists for this student and assignment
// 	var existingSubmissionId int
// 	err = tx.QueryRow("SELECT id FROM submission WHERE student_id = $1 AND assignment_id = $2", studentId, assignmentId).Scan(&existingSubmissionId)
// 	if err != nil && err != sql.ErrNoRows {
// 		return err
// 	}

// 	if existingSubmissionId > 0 {
// 		// Update existing submission
// 		_, err = tx.Exec("UPDATE submission SET file_id = $1, grade = NULL, comments = NULL WHERE id = $2", fileId, existingSubmissionId)
// 		if err != nil {
// 			return err
// 		}
// 	} else {
// 		// Insert new submission
// 		_, err = tx.Exec("INSERT INTO submission (student_id, assignment_id, file_id) VALUES ($1, $2, $3)", studentId, assignmentId, fileId)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	err = tx.Commit()
// 	return err
// }
