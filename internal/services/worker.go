package services

import (
	"database/sql"
	"log/slog"
	"strconv"
	"time"

	"github.com/philip-h/amics/internal/store"
)

type Worker struct {
	store      store.Storage
	logger     *slog.Logger
	testRunner *TestRunner
	stopChan   chan bool
}

func NewWorker(db *sql.DB, logger *slog.Logger) (*Worker, error) {
	store := store.New(db)
	testRunner, err := NewTestRunner()
	if err != nil {
		return nil, err
	}
	return &Worker{
		store:      store,
		logger:     logger,
		testRunner: testRunner,
		stopChan:   make(chan bool),
	}, nil
}

func (w *Worker) Start() {
	w.logger.Info("Worker started successfully")
	for {
		select {
		case <-w.stopChan:
			w.logger.Info("Worker stopping...")
			return
		default:
			w.logger.Debug("Looking for next pending submission")
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
		w.logger.Error("Could not get next pending submission", slog.String("msg", err.Error()),
			slog.Group("where",
				slog.String("function", "processNextSubmission")))
		return
	}
	if submission == nil {
		w.logger.Debug("No pending submission found")
		return
	}

	w.logger.Debug("Processing submission with id " + strconv.Itoa(submission.Id))

	// Get the test code for this assignment
	assignment, err := w.store.Assignments.GetById(submission.AssignmentId)
	if err != nil {
		w.logger.Error("Could not get the test code for the assignment", slog.String("msg", err.Error()),
			slog.Group("where",
				slog.String("function", "processNextSubmission")))
		submission.Status = "failure"
		submission.Comments = sql.NullString{String: "Could not get code for the assignment", Valid: true}
		err = w.store.Submissions.Update(submission)
		if err != nil {
			w.logger.Error("Could not update submission "+strconv.Itoa(submission.Id)+" with grader failure status", slog.String("msg", err.Error()))
		}
		return
	}

	// run tests
	result, err := w.testRunner.Pytest(assignment.RequiredFilename, submission.Code, assignment.PytestCode)
	if err != nil {
		w.logger.Error("Error running pytest", slog.String("msg", err.Error()),
			slog.Group("where",
				slog.String("function", "processNextSubmission")))
		submission.Status = "failure"
		submission.Comments = sql.NullString{String: "Could not run pytest", Valid: true}
		err = w.store.Submissions.Update(submission)
		if err != nil {
			w.logger.Error("Could not update submission "+strconv.Itoa(submission.Id)+" with grader failure status", slog.String("msg", err.Error()))
		}
	}

	submission.Grade = result.Grade
	submission.Status = "completed"
	submission.Comments = sql.NullString{String: result.Comments, Valid: true}
	err = w.store.Submissions.Update(submission)
	if err != nil {
		w.logger.Error("Could not update submission "+strconv.Itoa(submission.Id)+" with grader failure status", slog.String("msg", err.Error()))
	}
}
