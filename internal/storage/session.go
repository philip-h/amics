package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"
)

type Session struct {
	Id        string
	PersonId  int
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionStore struct {
	db *sql.DB
}

func generateSessionId() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *SessionStore) Create(personId int, rememberMe bool) (string, error) {
	sessionId, err := generateSessionId()
	if err != nil {
		return "", err
	}

	var expiresAt time.Time
	if rememberMe {
		// one week
		expiresAt = time.Now().Add(7 * 24 * time.Hour)
	} else {
		// 2 hours
		expiresAt = time.Now().Add(2 * time.Hour)
	}

	_, err = s.db.Exec(`
  INSERT INTO sessions (id, person_id, expires_at)
  VALUES ($1, $2, $3)`,
		sessionId,
		personId,
		expiresAt)

	if err != nil {
		return "", err
	}

	return sessionId, nil
}

func (s *SessionStore) GetById(id string) (*Session, error) {
	session := &Session{}

	err := s.db.QueryRow(`
  SELECT id, person_id, created_at, expires_at
  FROM sessions
  WHERE id = $1 AND expires_at > NOW()`, id).Scan(
		&session.Id,
		&session.PersonId,
		&session.CreatedAt,
		&session.ExpiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return session, err
}

func (s *SessionStore) Delete(id string) error {
	_, err := s.db.Exec(`
  DELETE FROM sessions
  WHERE id = $1`, id)
	return err
}
