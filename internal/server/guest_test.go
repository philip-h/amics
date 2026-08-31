package server

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHandleGuestGet(t *testing.T) {
	t.Run("GET / returns 200", func(t *testing.T) {
		t.Parallel()

		res, err := http.Get(s.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}
	})

	t.Run("GET / displays guest home", func(t *testing.T) {
		t.Parallel()

		res, err := http.Get(s.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		body, _ := io.ReadAll(res.Body)

		if !strings.Contains(strings.ToLower(string(body)), "welcome to amics") {
			t.Error("Expected body to contain 'welcome to amics'")
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", res.StatusCode)
		}
	})

}
