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

	t.Run("should invoke Assignment.GetByUsername", func(t *testing.T) {
		mockStore, ok := app.Store.Assignments.(*store.MockAssignmentStore)
		if !ok {
			t.Fatal("store.Assignments is not of type *MockAssignmentStore")
		}

		req, err := http.NewRequest(http.MethodGet, uri, nil)
		if err != nil {
			t.Fatal(err)
		}

		req.AddCookie(createLegitStudentCookie(t, app, "testuser"))

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if !mockStore.GetByUsernameInvoked {
			t.Errorf("expected GetByUsername() to be invoked")
		}
	})
}

func TestAssignmentDetailHandler(t *testing.T) {
	app := newTestApplication(t, Config{})
	mux := app.Mount()
	uri := "/app/assignment/1"

	t.Run("should invoke Assignment.GetById", func(t *testing.T) {
		mockStore, ok := app.Store.Assignments.(*store.MockAssignmentStore)
		if !ok {
			t.Fatal("store.Assignments is not of type *MockAssignmentStore")
		}

		req, err := http.NewRequest(http.MethodGet, uri, nil)
		if err != nil {
			t.Fatal(err)
		}

		req.AddCookie(createLegitStudentCookie(t, app, "testuser"))

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if !mockStore.GetByIdInvoked {
			t.Errorf("expected GetById() to be invoked")
		}
	})
}

func TestGETHandlersReturn200(t *testing.T) {
	app := newTestApplication(t, Config{})
	mux := app.Mount()

	tt := []struct {
		name           string
		method         string
		uri            string
		cookie 	   *http.Cookie
	}{
		{"get index returns 200", "GET", "/", nil},
		{"get login returns 200", "GET", "/login", nil},
		{"get register returns 200", "GET", "/register", nil},
		{"get dashboard returns 200 with valid cookie", "GET", "/app", createLegitStudentCookie(t, app, "test")},
		{"get assignment detail returns 200 with valid cookie", "GET", "/app/assignment/1", createLegitStudentCookie(t, app, "test")},
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
		{"get login redirects to /app with valid cookie", "GET", "/login", createLegitStudentCookie(t, app, "test"), 303, "/app"},
		{"get register redirects to /app with valid cookie", "GET", "/register", createLegitStudentCookie(t, app, "test"), 303, "/app"},
		{"get login does not redirect to /app if cookie token is invalid", "GET", "/login", createSillyCookie(t), 200, ""},
		{"get register does not redirect to /app if cookie token is invalid", "GET", "/login", createSillyCookie(t), 200, ""},
		{"get dashboard redirects to /login no cookie", "GET", "/app", nil, 303, "/login"},
		{"get dashboard redirects to /login if cookie token is invalid", "GET", "/app", createSillyCookie(t), 303, "/login"},
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

// func TestIndexGET(t *testing.T) {
// 	app := newTestApplication(t, Config{})
// 	route := "/"
// 	mux := app.Mount()

// 	t.Run("returns 200 OK", func(t *testing.T) {
// 		req, err := http.NewRequest(http.MethodGet, route, nil)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		rr := httptest.NewRecorder()
// 		mux.ServeHTTP(rr, req)

// 		checkStatusCode(t, rr.Code, http.StatusOK)
// 	})
// }

// func TestLoginGET(t *testing.T) {
// 	app := newTestApplication(t, Config{})
// 	route := "/login"

// 	t.Run("returns 200 OK", func(t *testing.T) {
// 		req, err := http.NewRequest(http.MethodGet, route, nil)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		rr := httptest.NewRecorder()
// 		app.handleLoginGet(rr, req)

// 		checkStatusCode(t, rr.Code, http.StatusOK)
// 	})

// 	t.Run("redirects to /app if user is already logged in", func(t *testing.T) {
// 		req, err := http.NewRequest(http.MethodGet, route, nil)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		// Set a dummy token cookie to simulate a logged in user
// 		req.AddCookie(&http.Cookie{Name: "token", Value: "dummy-token"})

// 		rr := httptest.NewRecorder()
// 		app.handleLoginGet(rr, req)

// 		checkStatusCode(t, rr.Code, http.StatusSeeOther)
// 		location := rr.Header().Get("Location")
// 		if location != "/app" {
// 			t.Errorf("got redirect location %s, want /app", location)
// 		}
// 	})
// }

// func TestRegisterGET(t *testing.T) {
// 	app := newTestApplication(t, Config{})
// 	route := "/register"

// 	t.Run("returns 200 OK", func(t *testing.T) {
// 		req, err := http.NewRequest(http.MethodGet, route, nil)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		rr := httptest.NewRecorder()
// 		app.handleRegisterGet(rr, req)

// 		checkStatusCode(t, rr.Code, http.StatusOK)
// 	})

// 	t.Run("redirects to /app if user is already logged in", func(t *testing.T) {
// 		req, err := http.NewRequest(http.MethodGet, route, nil)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		// Set a dummy token cookie to simulate a logged in user
// 		// testToken, err := app.Auth.CreateJwt("1","",0)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		req.AddCookie(&http.Cookie{Name: "token", Value: "dummy value"})

// 		rr := httptest.NewRecorder()
// 		app.handleRegisterGet(rr, req)

// 		checkStatusCode(t, rr.Code, http.StatusSeeOther)
// 		location := rr.Header().Get("Location")
// 		if location != "/app" {
// 			t.Errorf("got redirect location %s, want /app", location)
// 		}
// 	})
// }

