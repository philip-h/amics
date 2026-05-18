package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/philip-h/amics/internal/db"
	"github.com/philip-h/amics/internal/storage"
	"github.com/philip-h/amics/templates"
)

var (
	testStore *storage.Storage
	s         *httptest.Server
)

func TestMain(m *testing.M) {
	// Setup test server and database
	tdb, err := db.NewTestDB()
	if err != nil {
		panic(err)
	}
	testStore = storage.New(tdb.Db)

	srv, err := newTestServer(testStore)
	if err != nil {
		panic(err)
	}
	s = httptest.NewServer(srv)

	m.Run()

	tdb.Drop()
	s.Close()

}

func newTestServer(testStore *storage.Storage) (http.Handler, error) {

	testLogger := &Logger{
		L:      slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})),
		LogLvl: &slog.LevelVar{},
	}

	tmpls, err := templates.LoadTemplates()

	if err != nil {
		return nil, err
	}

	testServer := NewServer(
		testLogger,
		testStore,
		tmpls,
	)

	return testServer, nil
}