package data_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zidane0000/ai-interview-platform/data"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newMockGormDB initializes a GORM DB backed by sqlmock for testing
func newMockGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		db.Close()
		t.Fatalf("failed to open gorm db with sqlmock: %v", err)
	}
	cleanup := func() { db.Close() }
	return gormDB, mock, cleanup
}

// TestInitDB_InvalidURL verifies InitDB returns an error for an invalid database URL
func TestInitDB_InvalidURL(t *testing.T) {
	_, err := data.InitDB("postgres://invalid:invalid@localhost:5432/invalid_db")
	if err == nil {
		t.Error("expected error for invalid database URL, got nil")
	}
}

// TestInitDB_ValidMock demonstrates that a valid *gorm.DB can be used (mocked)
func TestInitDB_ValidMock(t *testing.T) {
	gormDB, _, cleanup := newMockGormDB(t)
	defer cleanup()
	if gormDB == nil {
		t.Error("expected non-nil gormDB from mock")
	}
}

// TestCloseDB_WithMock ensures CloseDB works with a mock DB
func TestCloseDB_WithMock(t *testing.T) {
	gormDB, _, cleanup := newMockGormDB(t)
	defer cleanup()
	// Should not panic or error
	data.CloseDB(gormDB)
}
