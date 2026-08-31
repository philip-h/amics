package server

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestHandleAuthLogin(t *testing.T) {

	statusCodeTests := []struct {
		name       string
		method     string
		formData   url.Values
		wantStatus int
	}{
		{
			name:       "GET /login returns 200",
			method:     http.MethodGet,
			formData:   url.Values{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST /login with empty body returns 422",
			method:     http.MethodPost,
			formData:   url.Values{},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "POST /login with username less than 4 characters returns 422",
			method:     http.MethodPost,
			formData:   url.Values{"username": {"abc"}, "password": {"validpassword"}},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "POST /login with password less than 10 characters returns 422",
			method:     http.MethodPost,
			formData:   url.Values{"username": {"validuser"}, "password": {"short"}},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range statusCodeTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(tt.method, s.URL+"/login", strings.NewReader(tt.formData.Encode()))
			if err != nil {
				t.Error(err)
			}

			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
			}
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, res.StatusCode)
			}
		})
	}

	t.Run("GET /login with valid session redirects to /c/1", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, s.URL+"/login", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: "amics-cookie", Value: "valid_session_id"})

		jar, _ := cookiejar.New(nil)
		client := &http.Client{
			Jar: jar,
		}
		res, err := client.Do(req)

		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", res.StatusCode)
		}

		currentPath := res.Request.URL.Path
		if currentPath != "/c/1" {
			t.Errorf("Expected to be redirected to /c/1, but was redirected to %s", currentPath)
		}
	})

	t.Run("GET /login with expired session stays on login page", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, s.URL+"/login", nil)
		if err != nil {
			t.Fatal(err)
		}
		jar, _ := cookiejar.New(nil)
		client := &http.Client{Jar: jar}

		req.AddCookie(&http.Cookie{Name: "amics-cookie", Value: "expired_session_id"})

		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		currentPath := res.Request.URL.Path
		if currentPath != "/login" {
			t.Errorf("Expected to stay on /login, but was redirected to %s", currentPath)
		}

	})

	t.Run("POST /login with valid credentials integration test ", func(t *testing.T) {
		t.Parallel()

		// valid credentials
		formData := url.Values{"username": {"teststudent"}, "password": {"validpassword"}}
		req, err := http.NewRequest(http.MethodPost, s.URL+"/login", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		jar, _ := cookiejar.New(nil)
		client := &http.Client{Jar: jar}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		cookies := jar.Cookies(res.Request.URL)
		for _, c := range jar.Cookies(res.Request.URL) {
			t.Logf("%s=%s", c.Name, c.Value)
		}

		if len(cookies) == 0 {
			t.Fatal("Expected a session cookie to be set, but got none")
		}

		if cookies[0].Name != "amics-cookie" {
			t.Errorf("Expected cookie name to be amics-cookie, got %s", cookies[0].Name)
		}

		session, err := testStore.Sessions.GetById(context.Background(), cookies[0].Value)
		if err != nil {
			t.Fatal(err)
		}
		if session == nil {
			t.Error("Expected session to be created in the database, but got nil")
		}

		if session.PersonId != 22222 {
			t.Errorf("Expected session PersonId to be 22222, got %d", session.PersonId)
		}

		currentPath := res.Request.URL.Path
		if currentPath != "/c/1" {
			t.Errorf("Expected to be redirected to /c/1, but was redirected to %s", currentPath)
		}
	})
}

type registerFormData struct {
	StudentNumber string
	FirstName     string
	Username      string
	Password      string
	JoinCode      string
}

func withShortStudentNumber() func(*registerFormData) {
	return func(data *registerFormData) {
		data.StudentNumber = "1234"
	}
}
func withLongStudentNumber() func(*registerFormData) {
	return func(data *registerFormData) {
		data.StudentNumber = "123456789"
	}
}
func withNewStudentNumber() func(*registerFormData) {
	return func(data *registerFormData) {
		data.StudentNumber = "9876543"
	}
}
func withNonNumericStudentNumber() func(*registerFormData) {
	return func(data *registerFormData) {
		data.StudentNumber = "abcdef"
	}
}
func withInvalidFirstName() func(*registerFormData) {
	return func(data *registerFormData) {
		data.FirstName = ""
	}
}
func withInvalidUsername() func(*registerFormData) {
	return func(data *registerFormData) {
		data.Username = "abc"
	}
}
func withInvalidPassword() func(*registerFormData) {
	return func(data *registerFormData) {
		data.Password = "short"
	}
}
func withInvalidJoinCode() func(*registerFormData) {
	return func(data *registerFormData) {
		data.JoinCode = "invalid"
	}
}
func validFormData(options ...func(*registerFormData)) url.Values {
	data := &registerFormData{
		StudentNumber: "1234567",
		FirstName:     "validname",
		Username:      "validusername",
		Password:      "validpassword",
		JoinCode:      "JOIN",
	}

	for _, option := range options {
		option(data)
	}

	return url.Values{
		"student-number": {data.StudentNumber},
		"first-name":     {data.FirstName},
		"username":       {data.Username},
		"password":       {data.Password},
		"join-code":      {data.JoinCode},
	}
}

func TestHandleAuthRegister(t *testing.T) {
	statusCodeTests := []struct {
		name       string
		method     string
		formData   url.Values
		cookie     *http.Cookie
		wantStatus int
	}{
		{
			name:       "GET /register returns 200",
			method:     http.MethodGet,
			formData:   url.Values{},
			cookie:     nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST /register with empty body returns 422",
			method:     http.MethodPost,
			formData:   url.Values{},
			cookie:     nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "POST /register with studentId less than 5 characters returns 422",
			method: http.MethodPost,
			formData: validFormData(
				withShortStudentNumber(),
			),
			cookie:     nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "POST /register with studentId more than 7 characters returns 422",
			method: http.MethodPost,
			formData: validFormData(
				withLongStudentNumber(),
			),
			cookie:     nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "POST /register with studentId containing non-numeric characters returns 422",
			method: http.MethodPost,
			formData: validFormData(
				withNonNumericStudentNumber(),
			),
			cookie:     nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "POST /register with invalid first name returns 422",
			method: http.MethodPost,
			formData: validFormData(
				withInvalidFirstName(),
			),
			cookie:     nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "POST /register with username less than 4 characters returns 422",
			method: http.MethodPost,
			formData: validFormData(
				withInvalidUsername(),
			),
			cookie:     nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "POST /register with password less than 10 characters returns 422",
			method: http.MethodPost,
			formData: validFormData(
				withInvalidPassword(),
			),
			cookie:     nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "POST /register with invalid join code returns 422",
			method: http.MethodPost,
			formData: validFormData(
				withInvalidJoinCode(),
			),
			cookie:     nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range statusCodeTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var res *http.Response
			var err error

			req, err := http.NewRequest(tt.method, s.URL+"/register", strings.NewReader(tt.formData.Encode()))
			if err != nil {
				t.Error(err)
			}

			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			res, err = http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
			}
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, res.StatusCode)
			}

		})
	}

	t.Run("GET /register with valid session redirects to /c/1", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, s.URL+"/register", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: "amics-cookie", Value: "valid_session_id"})

		jar, _ := cookiejar.New(nil)
		client := &http.Client{Jar: jar}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		currentPath := res.Request.URL.Path
		if currentPath != "/c/1" {
			t.Errorf("Expected to be redirected to /c/1, but was redirected to %s", currentPath)
		}
	})

	t.Run("GET /register with expired session stays on register page", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, s.URL+"/register", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: "amics-cookie", Value: "expired_session_id"})

		jar, _ := cookiejar.New(nil)
		client := &http.Client{Jar: jar}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		currentPath := res.Request.URL.Path
		if currentPath != "/register" {
			t.Errorf("Expected to stay on /register, but was redirected to %s", currentPath)
		}
	})

	t.Run("POST /register with valid credentials integration test", func(t *testing.T) {
		t.Parallel()

		formData := validFormData(
			withNewStudentNumber(),
		)
		req, err := http.NewRequest(http.MethodPost, s.URL+"/register", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		jar, _ := cookiejar.New(nil)
		client := &http.Client{Jar: jar}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		cookies := jar.Cookies(req.URL)

		if len(cookies) == 0 {
			t.Fatalf("Expected a session cookie to be set, but got none")
		}

		if cookies[0].Name != "amics-cookie" {
			t.Errorf("Expected cookie name to be amics-cookie, got %s", cookies[0].Name)
		}

		session, err := testStore.Sessions.GetById(context.Background(), cookies[0].Value)
		if err != nil {
			t.Fatal(err)
		}
		if session == nil {
			t.Error("Expected session to be created in the database, but got nil")
		}

		if strconv.Itoa(session.PersonId) != formData.Get("student-number") {
			t.Errorf("Expected session PersonId to be %s, got %d", formData.Get("student-number"), session.PersonId)
		}

		person, err := testStore.People.GetById(context.Background(), session.PersonId)
		if err != nil {
			t.Fatal(err)
		}
		if person == nil {
			t.Error("Expected person to be created in the database, but got nil")
		}
	})
}
