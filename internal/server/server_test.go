package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/philip-h/amics/internal/db"
	"github.com/philip-h/amics/internal/storage"
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

	testServer := NewServer(
		testLogger,
		testStore,
	)

	return testServer, nil
}
