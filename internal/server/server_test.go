package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/philip-h/amics/internal/store"
)

func TestLoginHandler(t *testing.T) {
	app := newTestApplication(t, Config{})
	mux := app.Mount()
	uri := "/login"

	t.Run("should invoke Student.GetByUsername", func(t *testing.T) {
		mockStore, ok := app.Store.Students.(*store.MockStudentStore)
		if !ok {
			t.Fatal("store.Students is not of type *MockStudentStore")
		}

		req, err := http.NewRequest(http.MethodPost, uri, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Form = map[string][]string{
			"username": {"testuser"},
			"password": {"testpass"},
		}

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if !mockStore.GetByUsernameInvoked {
			t.Errorf("expected GetByUsername() to be invoked")
		}
	})
}

func TestDashboardHandler(t *testing.T) {
	app := newTestApplication(t, Config{})
	mux := app.Mount()
	uri := "/app"

	t.Run("should invoke Assignment.GetWithGradeByStudentId", func(t *testing.T) {
		mockStore, ok := app.Store.Assignments.(*store.MockAssignmentStore)
		if !ok {
			t.Fatal("store.Assignments is not of type *MockAssignmentStore")
		}

		req, err := http.NewRequest(http.MethodGet, uri, nil)
		if err != nil {
			t.Fatal(err)
		}

		req.AddCookie(createLegitStudentCookie(t, app))

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if !mockStore.GetWithGradeByStudentIdInvoked {
			t.Errorf("expected GetWithGradeByStudentId() to be invoked")
		}
	})
}

func TestAssignmentDetailHandler(t *testing.T) {
	app := newTestApplication(t, Config{})
	mux := app.Mount()
	uri := "/app/assignments/1"

	t.Run("should invoke Assignment.GetWithSubmissionByAssignmentAndStudentIds", func(t *testing.T) {
		mockStore, ok := app.Store.Assignments.(*store.MockAssignmentStore)
		if !ok {
			t.Fatal("store.Assignments is not of type *MockAssignmentStore")
		}

		req, err := http.NewRequest(http.MethodGet, uri, nil)
		if err != nil {
			t.Fatal(err)
		}

		req.AddCookie(createLegitStudentCookie(t, app))

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if !mockStore.GetWithSubmissionByAssignmentAndStudentIdsInvoked {
			t.Errorf("expected GetWithSubmissionByAssignmentAndStudentIds() to be invoked")
		}
	})
}

func TestGETHandlersReturn200(t *testing.T) {
	app := newTestApplication(t, Config{})
	mux := app.Mount()

	tt := []struct {
		name   string
		method string
		uri    string
		cookie *http.Cookie
	}{
		{"get index returns 200", "GET", "/", nil},
		{"get login returns 200", "GET", "/login", nil},
		{"get register returns 200", "GET", "/register", nil},
		{"get dashboard returns 200 with valid cookie", "GET", "/app", createLegitStudentCookie(t, app)},
		{"get assignment detail returns 200 with valid cookie", "GET", "/app/assignments/1", createLegitStudentCookie(t, app)},
		{"get teacher dashboard returns 200 with valid cookie", "GET", "/teacher", createLegitTeacherCookie(t, app)},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.uri, nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			checkStatusCode(t, rr.Code, http.StatusOK)
		})
	}
}

func TestAuthHandlers(t *testing.T) {
	app := newTestApplication(t, Config{})
	mux := app.Mount()

	tt := []struct {
		name             string
		method           string
		uri              string
		cookie           *http.Cookie
		expectedCode     int
		expectedLocation string
	}{
		{"get login redirects to /app with valid cookie", "GET", "/login", createLegitStudentCookie(t, app), 303, "/app"},
		{"get register redirects to /app with valid cookie", "GET", "/register", createLegitStudentCookie(t, app), 303, "/app"},
		{"get login does not redirect to /app if cookie token is invalid", "GET", "/login", createSillyCookie(t), 200, ""},
		{"get register does not redirect to /app if cookie token is invalid", "GET", "/login", createSillyCookie(t), 200, ""},
		{"get dashboard redirects to /login no cookie", "GET", "/app", nil, 303, "/login"},
		{"get dashboard redirects to /login if cookie token is invalid", "GET", "/app", createSillyCookie(t), 303, "/login"},
		{"get teacher dashboard returns 404 with no cookie", "GET", "/teacher", nil, 404, ""},
		{"get teacher dashboard returns 404 with invalid cookie", "GET", "/teacher", createSillyCookie(t), 404, ""},
		{"get teacher dashboard returns 404 with student cookie", "GET", "/teacher", createLegitStudentCookie(t, app), 404, ""},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.uri, nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			checkStatusCode(t, rr.Code, tc.expectedCode)
			location := rr.Header().Get("Location")
			if location != tc.expectedLocation {
				t.Errorf("got redirect location %s, want %s", location, tc.expectedLocation)
			}

		})
	}
}
