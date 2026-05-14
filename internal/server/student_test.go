package server

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/philip-h/amics/internal/storage"
)

// TestHandleStudentDashboardGet tests the student dashboard handler
func TestHandleStudentDashboardGet(t *testing.T) {
	tests := []struct {
		name              string
		studentId         int
		role              string
		assignmentCount   int
		visibleFilter     []bool
		grades            []int
		expectedStatus    int
		expectedUnitCount int
		expectError       bool
	}{
		{
			name:              "happy path: multiple assignments across units with grades",
			studentId:         1,
			role:              "student",
			assignmentCount:   3,
			visibleFilter:     []bool{true, true, true},
			grades:            []int{90, 85, 95},
			expectedStatus:    http.StatusOK,
			expectedUnitCount: 1,
			expectError:       false,
		},
		{
			name:              "no grades yet: assignments without submissions",
			studentId:         2,
			role:              "student",
			assignmentCount:   2,
			visibleFilter:     []bool{true, true},
			grades:            []int{0, 0}, // No grades
			expectedStatus:    http.StatusOK,
			expectedUnitCount: 1,
			expectError:       false,
		},
		{
			name:              "mixed visibility: visible and hidden assignments",
			studentId:         3,
			role:              "student",
			assignmentCount:   4,
			visibleFilter:     []bool{true, false, true, false},
			grades:            []int{80, 75, 85, 70},
			expectedStatus:    http.StatusOK,
			expectedUnitCount: 1,
			expectError:       false,
		},
		{
			name:              "empty dashboard: no assignments",
			studentId:         4,
			role:              "student",
			assignmentCount:   0,
			visibleFilter:     []bool{},
			grades:            []int{},
			expectedStatus:    http.StatusOK,
			expectedUnitCount: 0,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load mock templates for testing
			tmpl, err := loadMockTemplates()
			if err != nil {
				t.Fatalf("Failed to load mock templates: %v", err)
			}

			// Create mock store with predefined assignments
			mockStore := createMockStoreWithAssignments(t, tt.assignmentCount, tt.visibleFilter, tt.grades)

			// Create handler
			handler := handleStudentDashboardGet(mockStore, tmpl)

			// Create request and inject context
			req := httptest.NewRequest("GET", "/app", nil)
			req = injectContext(req, tt.studentId, tt.role)

			// Execute handler
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// Verify response
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify response body is not empty for successful requests
			if w.Code == http.StatusOK && w.Body.Len() == 0 {
				t.Errorf("expected response body, got empty")
			}
		})
	}
}

// TestHandleStudentAssignmentGet tests retrieval of a specific assignment
func TestHandleStudentAssignmentGet(t *testing.T) {
	tests := []struct {
		name               string
		assignmentId       string
		studentId          int
		role               string
		assignmentDesc     string
		submissionComments string
		mockError          error
		expectedStatus     int
		expectError        bool
	}{
		{
			name:               "happy path: assignment with markdown description and comments",
			assignmentId:       "1",
			studentId:          1,
			role:               "student",
			assignmentDesc:     "# Assignment Title\n\nThis is a test assignment",
			submissionComments: "E: Error on line 5\n✔ Good work here",
			expectedStatus:     http.StatusOK,
			expectError:        false,
		},
		{
			name:               "no submission yet: view assignment preview",
			assignmentId:       "2",
			studentId:          2,
			role:               "student",
			assignmentDesc:     "Simple assignment",
			submissionComments: "",
			expectedStatus:     http.StatusOK,
			expectError:        false,
		},
		{
			name:               "invalid assignment id: non-numeric",
			assignmentId:       "invalid",
			studentId:          3,
			role:               "student",
			assignmentDesc:     "",
			submissionComments: "",
			expectedStatus:     http.StatusBadRequest,
			expectError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := loadMockTemplates()
			if err != nil {
				t.Fatalf("Failed to load mock templates: %v", err)
			}

			logger := &Logger{
				L:      nil,
				LogLvl: nil,
			}

			mockStore := createMockStoreWithAssignmentDetails(t, tt.assignmentDesc, tt.submissionComments)

			handler := handleStudentAssignmentGet(logger, mockStore, tmpl)

			req := httptest.NewRequest("GET", fmt.Sprintf("/app/assignments/%s", tt.assignmentId), nil)
			req.SetPathValue("assignmentId", tt.assignmentId)
			req = injectContext(req, tt.studentId, tt.role)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestHandleStudentAssignmentPost tests file upload submission
func TestHandleStudentAssignmentPost(t *testing.T) {
	tests := []struct {
		name           string
		assignmentId   string
		studentId      int
		fileSize       int
		expectRedirect bool
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "happy path: valid file upload",
			assignmentId:   "1",
			studentId:      1,
			fileSize:       100,
			expectRedirect: true,
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid assignment id",
			assignmentId:   "invalid",
			studentId:      3,
			fileSize:       100,
			expectRedirect: false,
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
			logger := &Logger{
				L:      l,
				LogLvl: nil,
			}

			mockStore := createMockStoreForUpload(t)

			handler := handleStudentAssignmentPost(logger, mockStore)

			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			part, err := writer.CreateFormFile("file", "test.txt")
			if err != nil {
				t.Fatal(err)
			}

			// write tt.fileSize bytes to part
			content := bytes.Repeat([]byte("a"), tt.fileSize)
			_, err = part.Write(content)

			writer.Close()

			req := httptest.NewRequest("POST", fmt.Sprintf("/app/assignments/%s", tt.assignmentId), &buf)
			req.SetPathValue("assignmentId", tt.assignmentId)
			req = injectContext(req, tt.studentId, "student")
			// mock file content by setting a header (since we're not actually parsing multipart form in tests)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectRedirect && w.Code == http.StatusOK {
				redirect := w.Header().Get("HX-Redirect")
				if redirect == "" {
					t.Errorf("expected HX-Redirect header, got none")
				}
			}
		})
	}
}

// TestHandleStudentAssignmentPoll tests submission status polling
func TestHandleStudentAssignmentPoll(t *testing.T) {
	tests := []struct {
		name             string
		assignmentId     string
		studentId        int
		submissionStatus string
		comments         string
		expectedStatus   int
		expectError      bool
	}{
		{
			name:             "happy path: poll pending submission",
			assignmentId:     "1",
			studentId:        1,
			submissionStatus: "pending",
			comments:         "",
			expectedStatus:   http.StatusOK,
			expectError:      false,
		},
		{
			name:             "poll graded submission with feedback",
			assignmentId:     "2",
			studentId:        2,
			submissionStatus: "graded",
			comments:         "E: Syntax error\n✔ Logic looks good",
			expectedStatus:   http.StatusOK,
			expectError:      false,
		},
		{
			name:             "no submission yet",
			assignmentId:     "3",
			studentId:        3,
			submissionStatus: "",
			comments:         "",
			expectedStatus:   http.StatusOK,
			expectError:      false,
		},
		{
			name:             "invalid assignment id",
			assignmentId:     "invalid",
			studentId:        4,
			submissionStatus: "",
			comments:         "",
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := loadMockTemplates()
			if err != nil {
				t.Fatalf("Failed to load mock templates: %v", err)
			}

			mockStore := createMockStoreForPolling(t, tt.submissionStatus, tt.comments)

			handler := handleStudentAssignmentPoll(mockStore, tmpl)

			req := httptest.NewRequest("GET", fmt.Sprintf("/app/assignments/%s/poll", tt.assignmentId), nil)
			req.SetPathValue("assignmentId", tt.assignmentId)
			req = injectContext(req, tt.studentId, "student")

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// ============================================================================
// Mock Helpers
// ============================================================================

// loadMockTemplates loads templates for testing
func loadMockTemplates() (*template.Template, error) {
	tmpl := template.New("base")
	tmpl, err := tmpl.Parse(`
		{{define "base"}}
		<html>
			<body>
				{{template "app" .}}
			</body>
		</html>
		{{end}}
		{{define "app"}}
			<div>Dashboard: {{.StudentAverage}}</div>
			{{range $unit, $assignments := .Assignments}}
				<div>{{$unit}}: {{len $assignments}} assignments</div>
			{{end}}
		{{end}}
		{{define "assignment"}}
			<div>{{.Assignment.Name}}</div>
			<div>{{.Description}}</div>
			{{if .Submission}}<div>{{.Comments}}</div>{{end}}
		{{end}}
		{{define "submission_overview"}}
			<div>Submission Status</div>
			{{if .Submission}}<div>{{.Comments}}</div>{{end}}
		{{end}}
	`)
	return tmpl, err
}

// createMockStoreWithAssignments creates a mock storage with test assignments
func createMockStoreWithAssignments(t *testing.T, count int, visibleFilter []bool, grades []int) *storage.Storage {
	store := storage.NewMockStore()

	// Create mock assignments
	assignments := make([]*storage.AssignmentWithGrade, count)
	for i := 0; i < count; i++ {
		visible := true
		if i < len(visibleFilter) {
			visible = visibleFilter[i]
		}

		grade := sql.NullInt64{}
		if i < len(grades) && grades[i] > 0 {
			grade = sql.NullInt64{Int64: int64(grades[i]), Valid: true}
		}

		assignments[i] = &storage.AssignmentWithGrade{
			Assignment: storage.Assignment{
				Id:       i + 1,
				UnitName: "Unit 1",
				Name:     fmt.Sprintf("Assignment %d", i+1),
				Points:   100,
				Visible:  visible,
			},
			Grade: grade,
		}
	}

	// Patch the mock to return assignments
	mockAssignmentStore := store.Assignments.(*storage.MockAssignmentStore)
	mockAssignmentStore.GetWithGradeResult = assignments

	return &store
}

// createMockStoreWithAssignmentDetails creates a mock storage with assignment and submission details
func createMockStoreWithAssignmentDetails(t *testing.T, desc, comments string) *storage.Storage {
	store := storage.NewMockStore()

	submission := (*storage.Submission)(nil)
	if comments != "" {
		submission = &storage.Submission{
			Code:     "test code",
			Comments: sql.NullString{String: comments, Valid: true},
			Status:   "graded",
			Grade:    85,
		}
	}

	aws := &storage.AssignmentSubmission{
		Assignment: storage.Assignment{
			Id:          1,
			Name:        "Test Assignment",
			Description: desc,
			Points:      100,
		},
		Submission: submission,
	}

	mockAssignmentStore := store.Assignments.(*storage.MockAssignmentStore)
	mockAssignmentStore.GetWithSubmissionResult = aws

	return &store
}

// createMockStoreForUpload creates a mock storage for upload testing
func createMockStoreForUpload(t *testing.T) *storage.Storage {
	store := storage.NewMockStore()
	return &store
}

// createMockStoreForPolling creates a mock storage for polling tests
func createMockStoreForPolling(t *testing.T, status, comments string) *storage.Storage {
	store := storage.NewMockStore()

	submission := (*storage.Submission)(nil)
	if status != "" {
		submission = &storage.Submission{
			Code:     "test code",
			Status:   status,
			Comments: sql.NullString{String: comments, Valid: comments != ""},
		}
	}

	aws := &storage.AssignmentSubmission{
		Assignment: storage.Assignment{
			Id:   1,
			Name: "Test Assignment",
		},
		Submission: submission,
	}

	mockAssignmentStore := store.Assignments.(*storage.MockAssignmentStore)
	mockAssignmentStore.GetWithSubmissionResult = aws

	return &store
}

// ============================================================================
// Integration Tests
// ============================================================================

// TestStudentDashboardIntegration tests the full student dashboard flow
// This starts the entire application and makes real HTTP requests
func TestStudentDashboardIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	testDB := setupIntegrationTestDB(t)
	defer testDB.cleanup()

	// Seed test data
	seedIntegrationTestData(t, testDB.db)

	// Start the full application
	app := startIntegrationTestApp(t, testDB.db)
	defer app.shutdown()

	// Create a student session (simulate login)
	sessionID := createTestSession(t, testDB.db, 1) // student ID 1

	// Make HTTP request to student dashboard
	req, err := http.NewRequest("GET", app.url+"/app", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add session cookie
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: sessionID,
		Path:  "/app",
	})

	// Execute request
	resp, err := app.client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Verify response contains expected content
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "Student Dashboard") {
		t.Errorf("Expected response to contain 'Student Dashboard', got: %s", bodyStr)
	}

	// Verify we can access student-specific data
	if !strings.Contains(bodyStr, "Test Course") {
		t.Errorf("Expected response to contain course data, got: %s", bodyStr)
	}
}

// integrationTestApp holds the running application for integration tests
type integrationTestApp struct {
	testServer *httptest.Server
	client     *http.Client
	url        string
}

// startIntegrationTestApp starts the full application for integration testing
func startIntegrationTestApp(t *testing.T, db *sql.DB) *integrationTestApp {
	// Create a test server on a random port
	mux := http.NewServeMux()

	// Setup application components (similar to main.run())
	logger := &Logger{
		L:      slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		LogLvl: nil,
	}

	store := storage.New(db)

	// Load templates (simplified for testing)
	templates := make(map[string]*template.Template)
	mockTmpl, err := loadMockTemplates()
	if err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}
	templates["app"] = mockTmpl
	templates["assignment"] = mockTmpl
	templates["submission_overview"] = mockTmpl

	// Setup routes
	addRoutes(mux, logger, store, templates)

	// Use httptest.Server for integration testing
	testServer := httptest.NewServer(mux)

	return &integrationTestApp{
		testServer: testServer,
		client:     &http.Client{Timeout: 10 * time.Second},
		url:        testServer.URL,
	}
}

// shutdown stops the integration test application
func (app *integrationTestApp) shutdown() {
	if app.testServer != nil {
		app.testServer.Close()
	}
}

// setupIntegrationTestDB creates a test database for integration tests
func setupIntegrationTestDB(t *testing.T) *integrationTestDB {
	// Use a unique database name for this test
	dbName := fmt.Sprintf("amics_test_%d", time.Now().UnixNano())

	// Connect to postgres to create test database
	adminConnStr := "postgresql://postgres@127.0.0.1/?sslmode=disable"
	adminDB, err := sql.Open("postgres", adminConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to admin database: %v", err)
	}
	defer adminDB.Close()

	// Create test database
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Connect to test database
	testConnStr := fmt.Sprintf("postgresql://postgres@127.0.0.1/%s?sslmode=disable", dbName)
	testDB, err := sql.Open("postgres", testConnStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations/schema setup
	err = runTestMigrations(t, testDB)
	if err != nil {
		testDB.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return &integrationTestDB{
		db:     testDB,
		dbName: dbName,
	}
}

// integrationTestDB holds test database connection and cleanup info
type integrationTestDB struct {
	db     *sql.DB
	dbName string
}

// cleanup drops the test database
func (tdb *integrationTestDB) cleanup() {
	if tdb.db != nil {
		tdb.db.Close()

		// Drop the test database
		adminConnStr := "postgresql://postgres@127.0.0.1/?sslmode=disable"
		adminDB, err := sql.Open("postgres", adminConnStr)
		if err == nil {
			defer adminDB.Close()
			// adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", tdb.dbName))
		}
	}
}

// runTestMigrations sets up the database schema for testing
func runTestMigrations(t *testing.T, db *sql.DB) error {
	// Run your migration SQL here
	// This is simplified - in reality you'd run your actual migration files
	migrationSQL := `
		CREATE TABLE person (
			id SERIAL PRIMARY KEY,
			first_name TEXT,
			username TEXT UNIQUE,
			password TEXT,
			role TEXT
		);

		CREATE TABLE course (
			id SERIAL PRIMARY KEY,
			year INTEGER,
			semester INTEGER,
			name TEXT,
			join_code TEXT UNIQUE
		);

		CREATE TABLE assignment (
			id SERIAL PRIMARY KEY,
			unit_name TEXT,
			name TEXT,
			description TEXT,
			points INTEGER,
			visible BOOLEAN DEFAULT true,
			course_id INTEGER REFERENCES course(id)
		);

		CREATE TABLE submission (
			id SERIAL PRIMARY KEY,
			student_id INTEGER REFERENCES person(id),
			assignment_id INTEGER REFERENCES assignment(id),
			code TEXT,
			status TEXT DEFAULT 'pending',
			comments TEXT,
			grade INTEGER
		);

		CREATE TABLE sessions (
			id VARCHAR(64) PRIMARY KEY,
			person_id INTEGER NOT NULL REFERENCES person(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL
		);
	`

	_, err := db.Exec(migrationSQL)
	return err
}

// seedIntegrationTestData populates the test database with sample data
func seedIntegrationTestData(t *testing.T, db *sql.DB) {
	// Insert test course
	_, err := db.Exec(`
		INSERT INTO course (id, year, semester, name, join_code) 
		VALUES (1, 2024, 1, 'Test Course', 'TEST123')
	`)
	if err != nil {
		t.Fatalf("Failed to seed course: %v", err)
	}

	// Insert test student
	_, err = db.Exec(`
		INSERT INTO person (id, first_name, username, password, role) 
		VALUES (1, 'Test Student', 'student', 'password', 'student')
	`)
	if err != nil {
		t.Fatalf("Failed to seed student: %v", err)
	}

	// Insert test assignment
	_, err = db.Exec(`
		INSERT INTO assignment (id, unit_name, name, description, points, visible, course_id) 
		VALUES (1, 'Unit 1', 'Test Assignment', 'Test description', 100, true, 1)
	`)
	if err != nil {
		t.Fatalf("Failed to seed assignment: %v", err)
	}
}

// createTestSession creates a session for the given student ID
func createTestSession(t *testing.T, db *sql.DB, studentID int) string {
	sessionID := fmt.Sprintf("test_session_%d", studentID)

	// Insert session (assuming you have a sessions table)
	_, err := db.Exec(`
		INSERT INTO sessions (id, person_id, expires_at) 
		VALUES ($1, $2, NOW() + INTERVAL '24 hours')
	`, sessionID, studentID)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return sessionID
}
