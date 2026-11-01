package database

import (
	"testing"
)

func TestInitDBSQLite(t *testing.T) {
	db, err := InitDB(Config{
		Driver: "sqlite",
		DSN:    ":memory:",
	})
	if err != nil {
		t.Fatalf("InitDB sqlite failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

func TestInitDBUnsupportedDriver(t *testing.T) {
	_, err := InitDB(Config{
		Driver: "oracle",
		DSN:    "oracle://localhost",
	})
	if err == nil {
		t.Fatalf("expected error for unsupported driver")
	}
}
