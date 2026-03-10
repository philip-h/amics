package services

import (
	"database/sql"
	"log"
	"time"

	"github.com/philip-h/amics/internal/store"
)

type Worker struct {
	store      store.Storage
	testRunner *TestRunner
	stopChan   chan bool
}

func NewWorker(db *sql.DB) (*Worker, error) {
	store := store.New(db)
	testRunner, err := NewTestRunner()
	if err != nil {
		return nil, err
	}
	return &Worker{
		store:      store,
		testRunner: testRunner,
		stopChan:   make(chan bool),
	}, nil
}

func (w *Worker) Start() {
	log.Println("Worker started... polling db for submissions...")

	for {
		select {
		case <-w.stopChan:
			log.Println("Worker stopping...")
			return
		default:
			w.processNextSubmission()
			time.Sleep(2 * time.Second)
		}
	}
}

func (w *Worker) Stop() {
	w.stopChan <- true
}

func (w *Worker) processNextSubmission() {
	submission, err := w.store.Submissions.GetNextPendingSubmission()
	// No submission, return early
	if err != nil {
		log.Printf("We got an error processing: %v", err)
		return
	}
	if submission == nil {
		return
	}

	log.Printf("Processing submission %d for user %d and assignment %d",
		submission.Id, submission.StudentId, submission.AssignmentId)

	// Get the test code for this assignment
	assignment, err := w.store.Assignments.GetById(submission.AssignmentId)
	if err != nil {
		log.Printf("Error getting code for assignment: %v", err)
		submission.Status = sql.NullString{String: "failed", Valid: true}
		submission.Comments = sql.NullString{String: "Could not get code for the assignment", Valid: true}
		err = w.store.Submissions.Update(submission)
		if err != nil {
			log.Printf("Error updating submission: %v", err)
		}
		return
	}

	// run tests
	result, err := w.testRunner.Pytest(assignment.RequiredFilename, submission.Code, assignment.PytestCode)
	if err != nil {
		log.Printf("Error running tests: %v", err)
		submission.Status = sql.NullString{String: "failed", Valid: true}
		submission.Comments = sql.NullString{String: "Could not run pytest", Valid: true}
		err = w.store.Submissions.Update(submission)
		if err != nil {
			log.Printf("Error updating submission: %v", err)
		}
	}

	submission.Grade = result.Grade
	submission.Status = sql.NullString{String: "completed", Valid: true}
	submission.Comments = sql.NullString{String: result.Comments, Valid: true}
	err = w.store.Submissions.Update(submission)
	if err != nil {
		log.Printf("Error updating submission: %v", err)
	}
}
