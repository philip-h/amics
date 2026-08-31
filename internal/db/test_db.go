package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"golang.org/x/crypto/bcrypt"
)

// IntegrationTestDB holds test database connection and cleanup info
type IntegrationTestDB struct {
	Db     *sql.DB
	DbName string
	m      *migrate.Migrate
}

// NewTestDB connects to the test database and runs migrations
func NewTestDB() (*IntegrationTestDB, error) {
	// Use a unique database name for this test
	dbName := "amics_test"

	// Connect to test database
	testConnStr := "postgresql://postgres@127.0.0.1/" + dbName + "?sslmode=disable"
	testDB, err := sql.Open("postgres", testConnStr)
	if err != nil {
		return nil, fmt.Errorf("[NewTestDB] Cannot open db: %w", err)
	}

	m, err := migrate.New(
		"file://../db/migrations",
		testConnStr,
	)
	if err != nil {
		return nil, fmt.Errorf("[NewTestDB] Cannot create new migrations: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("[NewTestDB] Cannot run migrations: %w", err)
	}

	db := &IntegrationTestDB{
		Db:     testDB,
		DbName: dbName,
		m:      m,
	}
	if err = db.seed(); err != nil {
		return nil, fmt.Errorf("[NewTestDB] Cannot seed test data: %w", err)
	}
	return db, nil

}

func (tdb *IntegrationTestDB) seed() error {
	tx, err := tdb.Db.Begin()
	if err != nil {
		return fmt.Errorf("[seed] Cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	hashed_password, err := bcrypt.GenerateFromPassword([]byte("validpassword"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("[seed] Cannot generate hashed password: %w", err)
	}

	_, err = tx.Exec(`
INSERT INTO person (id, first_name, username, password, role) 
VALUES (11111, 'Test Teacher', 'testteacher', $1, 'teacher'),
       (22222, 'Test Student', 'teststudent', $1, 'student')`, hashed_password)
	if err != nil {
		return fmt.Errorf("[seed] Cannot insert person data: %w", err)
	}

	_, err = tx.Exec(`
INSERT INTO course (id, course_code, section, name, year, semester, join_code) 
VALUES (1, 'ICS3U', 1, 'Test Course', 2006, 1, 'JOIN')`)
	if err != nil {
		return fmt.Errorf("[seed] Cannot insert course data: %w", err)
	}

	_, err = tx.Exec(`
INSERT INTO teacher_course (teacher_id, course_id) 
VALUES (11111, 1)`)
	if err != nil {
		return fmt.Errorf("[seed] Cannot insert teacher_course data: %w", err)
	}

	_, err = tx.Exec(`
INSERT INTO student_course (student_id, course_id) 
VALUES (22222, 1)`)
	if err != nil {
		return fmt.Errorf("[seed] Cannot insert student_course data: %w", err)
	}

	_, err = tx.Exec(`
INSERT INTO assignment (unit_name, name, description, required_filename, pytest_code, points, due_date, visible, course_id) 
VALUES ('Unit 1', 'Test Assignment', 'This is a test assignment', 'solution.py', '#todo no test', 100, NOW() + INTERVAL '7 days', true, 1)`)
	if err != nil {
		return fmt.Errorf("[seed] Cannot insert assignment data: %w", err)
	}

	_, err = tx.Exec(`
INSERT INTO sessions (id, person_id, expires_at) 
VALUES ('valid_session_id', 22222, NOW() + INTERVAL '1 hour'),
       ('expired_session_id', 22222, NOW() - INTERVAL '1 hour')`)
	if err != nil {
		return fmt.Errorf("[seed] Cannot insert sessions data: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("[seed] Cannot commit transaction: %w", err)
	}
	return nil
}

// Drop drops the test database
func (tdb *IntegrationTestDB) Drop() error {
	err := tdb.m.Down()
	if err != nil {
		return fmt.Errorf("[Drop] Cannot drop test database: %w", err)
	}
	return nil
}
