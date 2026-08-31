package services

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"
	"time"

	"github.com/philip-h/amics/internal/server"
	"github.com/philip-h/amics/internal/storage"
)

type Worker struct {
	store      *storage.Storage
	logger     *slog.Logger
	testRunner *TestRunner
	done       chan struct{}
}

func NewWorker(db *sql.DB, logger *server.Logger) (*Worker, error) {
	store := storage.New(db)
	testRunner, err := NewTestRunner()
	if err != nil {
		return nil, err
	}
	return &Worker{
		store:      store,
		logger:     logger.L,
		testRunner: testRunner,
		done:       make(chan struct{}),
	}, nil
}

// Start runs the worker loop until ctx is cancelled. Call Wait afterward to
// block until the current iteration has finished and the loop has exited.
func (w *Worker) Start(ctx context.Context) {
	defer close(w.done)
	w.logger.Info("Worker started successfully")
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Worker stopping...")
			return
		default:
			w.logger.Debug("Looking for next pending submission")
			w.processNextSubmission(ctx)

			select {
			case <-ctx.Done():
				w.logger.Info("Worker stopping...")
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// Wait blocks until Start has returned after ctx was cancelled.
func (w *Worker) Wait() {
	<-w.done
}

func (w *Worker) processNextSubmission(ctx context.Context) {
	submission, err := w.store.Submissions.GetNextPendingSubmission(ctx)
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
	assignment, err := w.store.Assignments.GetById(ctx, submission.AssignmentId)
	if err != nil {
		w.logger.Error("Could not get the test code for the assignment", slog.String("msg", err.Error()),
			slog.Group("where",
				slog.String("function", "processNextSubmission")))
		submission.Status = "failure"
		submission.Comments = sql.NullString{String: "Could not get code for the assignment", Valid: true}
		err = w.store.Submissions.Update(ctx, submission)
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
		err = w.store.Submissions.Update(ctx, submission)
		if err != nil {
			w.logger.Error("Could not update submission "+strconv.Itoa(submission.Id)+" with grader failure status", slog.String("msg", err.Error()))
		}
	}

	submission.Grade = result.Grade
	submission.Status = "completed"
	submission.Comments = sql.NullString{String: result.Comments, Valid: true}
	err = w.store.Submissions.Update(ctx, submission)
	if err != nil {
		w.logger.Error("Could not update submission "+strconv.Itoa(submission.Id)+" with grader failure status", slog.String("msg", err.Error()))
	}
}
