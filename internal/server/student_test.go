package server

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
)


func TestHandleStudentDashboard(t *testing.T) {
	t.Run("GET /app without authentication should redirect to login", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodGet, s.URL+"/app", nil)
		if err != nil {
			t.Fatal(err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		currentPath := res.Request.URL.Path
		if currentPath != "/login" {
			t.Errorf("Expected to redirect to /login, but the path is %s", currentPath)
		}
	})

	t.Run("GET /app with authentication integration test", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodGet, s.URL+"/app", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid_session_id"})

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		// I expect to still be on /app
		currentPath := res.Request.URL.Path
		if currentPath != "/app" {
			t.Errorf("Expected to stay on /app, but was redirected to %s", currentPath)
		}
	})
}

func TestHandleStudentAssignmentGet(t *testing.T) {
	t.Run("GET /app/assignments/{assignmentId} without authentication should redirect to login", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodGet, s.URL+"/app/assignments/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", res.StatusCode)
		}

		currentPath := res.Request.URL.Path
		if currentPath != "/login" {
			t.Errorf("Expected to redirect to /login, but the path is %s", currentPath)
		}
	})


	t.Run("GET /app/assignments/{assignmentId} with authentication should stay on /app/assignments/{assignmentId}", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodGet, s.URL+"/app/assignments/1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid_session_id"})

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", res.StatusCode)
		}

		currentPath := res.Request.URL.Path
		if currentPath != "/app/assignments/1" {
			t.Errorf("Expected to stay on /app/assignments/1, but the path is %s", currentPath)
		}
	})
}

func TestHandleStudentAssignmentPost(t *testing.T) {
	t.Run("POST /app/assignments/{assignmentId} without authentication should redirect to login", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodPost, s.URL+"/app/assignments/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		currentPath := res.Request.URL.Path
		if currentPath != "/login" {
			t.Errorf("Expected to redirect to /login, but the path is %s", currentPath)
		}
	})

	t.Run("POST /app/assignments/{assignmentId} with no Content-Type Header should return 400", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodPost, s.URL+"/app/assignments/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid_session_id"})
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code 400, got %d", res.StatusCode)
		}
	})

	t.Run("POST /app/assignments/{assignmentId} with Content-Type Header but no body should return 400", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodPost, s.URL+"/app/assignments/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid_session_id"})
		req.Header.Set("Content-Type", "multipart/form-data; boundary=custom_boundary")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code 400, got %d", res.StatusCode)
		}
	})
	t.Run("POST /app/assignments/{assignmentId} with empty content should return 422", func(t *testing.T) {
		t.Parallel()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		
		// Create form file part
		part, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatal(err)
		}
		
		// Write content to the part
		part.Write([]byte("")) 
		
		// Close writer to set boundaries
		writer.Close()

		req, err := http.NewRequest(http.MethodPost, s.URL+"/app/assignments/1", body)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid_session_id"})

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		bb, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		if string(bb) != "Content of test.txt is empty\n" {
			t.Errorf("Expected error message about empty content, got %s", string(bb))
		}

		if res.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("Expected status code 422, got %d", res.StatusCode)
		}
	})

	t.Run("POST /app/assignments/{assignmentId} with valid file integration test", func(t *testing.T) {
		t.Parallel()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		
		// Create form file part
		part, err := writer.CreateFormFile("file", "solution.py")
		if err != nil {
			t.Fatal(err)
		}
		
		// Write content to the part
		part.Write([]byte("print('Hello, World!')")) 
		
		// Close writer to set boundaries
		writer.Close()

		req, err := http.NewRequest(http.MethodPost, s.URL+"/app/assignments/1", body)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid_session_id"})

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		submission, err := testStore.Submissions.GetByAssignmentAndStudentIds(1, 22222)
		if err != nil {
			t.Fatal(err)
		}

		if submission == nil {
			t.Error("Expected submission to be created, but got nil")
		}

		if submission != nil && submission.Code != "print('Hello, World!')" {
			t.Errorf("Expected submission content to be \"print('Hello, World!')\", got %s", submission.Code)
		}

		if submission != nil && submission.Status != "grading" {
			t.Errorf("Expected submission status to be \"grading\", got %s", submission.Status)
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", res.StatusCode)
		}
	})
}