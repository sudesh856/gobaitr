package store

import (
	"os"
	"testing"
	"time"
)

func tempDB(t *testing.T) (*Store, string, func()) {
	t.Helper()
	f, err := os.CreateTemp("", "gobaitr-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := f.Name()
	f.Close()
	
	s, err := newWithPath(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	
	return s, dbPath, func() {
		s.Close()
		os.Remove(dbPath)
	}
}

func TestSQLiteRoundtrip(t *testing.T) {
	s, _, cleanup := tempDB(t)
	defer cleanup()

	now := time.Now()
	err := s.Insert("test-id", "url", "secret333", "https://localhost/t/test-id/secret333", "note", now, nil)
	if err != nil {
		t.Fatal(err)
	}

	row, err := s.GetByID("test-id")
	if err != nil {
		t.Fatal(err)
	}

	if row["id"] != "test-id" {
		t.Errorf("expected id=test-id, got %v", row["id"])
	}
	if row["type"] != "url" {
		t.Errorf("expected type=url, got %v", row["type"])
	}
}

func TestSqliteWAL(t *testing.T) {
	s, dbPath, cleanup := tempDB(t)
	defer cleanup()
	_ = s

	s2, err := newWithPath(dbPath)
	if err != nil {
		t.Errorf("second connection failed (WAL not working?): %v", err)
	}
	if s2 != nil {
		s2.Close()
	}
}

func TestMarkTriggered(t *testing.T) {
	s, _, cleanup := tempDB(t)
	defer cleanup()

	now := time.Now()
	err := s.Insert("tok1", "url", "sec", "http://x", "", now, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = s.MarkTriggered("tok1", "1.2.3.4", "curl/1.0", "{}")
	if err != nil {
		t.Fatal(err)
	}

	row, _ := s.GetByID("tok1")
	if row["triggered"] != true {
		t.Error("expected triggered=true")
	}

	events, _ := s.GetEvents("tok1")
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestDeleteToken(t *testing.T) {
	s, _, cleanup := tempDB(t)
	defer cleanup()

	now := time.Now()
	err := s.Insert("del1", "url", "sec", "http://x", "", now, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = s.MarkTriggered("del1", "1.2.3.4", "curl", "{}")
	if err != nil {
		t.Fatal(err)
	}

	count, err := s.DeleteToken("del1")
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("expected 1 deleted event, got %d", count)
	}

	_, err = s.GetByID("del1")
	if err == nil {
		t.Error("expected error for deleted token")
	}
}