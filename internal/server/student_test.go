package server

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
)

func TestHandleStudentDashboard(t *testing.T) {
	t.Run("GET /h without authentication should redirect to login", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodGet, s.URL+"/h", nil)
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

	t.Run("GET /h with authentication should redirect to /c/1", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodGet, s.URL+"/h", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: "amics-cookie", Value: "valid_session_id"})

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		currentPath := res.Request.URL.Path
		if currentPath != "/c/1" {
			t.Errorf("Expected to be redirected to /c/1, but was redirected to %s", currentPath)
		}
	})
}

func TestHandleStudentAssignmentGet(t *testing.T) {
	t.Run("GET /c/{courseId}/a/{assignmentId}/details without authentication should redirect to login", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodGet, s.URL+"/c/1/a/1/details", nil)
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

	t.Run("GET /c/{courseId}/a/{assignmentId}/details with authentication should stay on details page", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodGet, s.URL+"/c/1/a/1/details", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: "amics-cookie", Value: "valid_session_id"})

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", res.StatusCode)
		}

		currentPath := res.Request.URL.Path
		if currentPath != "/c/1/a/1/details" {
			t.Errorf("Expected to stay on /c/1/a/1/details, but the path is %s", currentPath)
		}
	})
}

func TestHandleStudentAssignmentPost(t *testing.T) {
	t.Run("POST /c/{courseId}/a/{assignmentId} without authentication should redirect to login", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodPost, s.URL+"/c/1/a/1", nil)
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

	t.Run("POST /c/{courseId}/a/{assignmentId} with no Content-Type header should return 400", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodPost, s.URL+"/c/1/a/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.AddCookie(&http.Cookie{Name: "amics-cookie", Value: "valid_session_id"})
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code 400, got %d", res.StatusCode)
		}
	})

	t.Run("POST /c/{courseId}/a/{assignmentId} with Content-Type header but no body should return 400", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodPost, s.URL+"/c/1/a/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.AddCookie(&http.Cookie{Name: "amics-cookie", Value: "valid_session_id"})
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
	t.Run("POST /c/{courseId}/a/{assignmentId} with empty content should return 422", func(t *testing.T) {
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

		req, err := http.NewRequest(http.MethodPost, s.URL+"/c/1/a/1", body)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.AddCookie(&http.Cookie{Name: "amics-cookie", Value: "valid_session_id"})

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

	t.Run("POST /c/{courseId}/a/{assignmentId} with valid file integration test", func(t *testing.T) {
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

		req, err := http.NewRequest(http.MethodPost, s.URL+"/c/1/a/1", body)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.AddCookie(&http.Cookie{Name: "amics-cookie", Value: "valid_session_id"})

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
