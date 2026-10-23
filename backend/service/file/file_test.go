package file_test

import (
	"os"
	"testing"
	"time"

	"mim/models"
	"mim/service/file"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFileTest(t *testing.T) (*file.Service, func()) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "mim_file_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cfg := file.Config{
		StorageDir: tempDir,
		SignSecret: "file_sign_secret_key_2026",
	}

	svc := file.NewService(db, cfg)
	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	return svc, cleanup
}

func TestFileUploadAndDeduplication(t *testing.T) {
	svc, cleanup := setupFileTest(t)
	defer cleanup()

	data := []byte("Mim Instant Messaging File Content Demo 2026")

	// 1. First upload
	rec1, err := svc.SaveFile(101, "demo.txt", "text/plain", data)
	if err != nil {
		t.Fatalf("first upload failed: %v", err)
	}
	if rec1.FileHash == "" || rec1.FileSize != uint64(len(data)) {
		t.Errorf("unexpected record: %+v", rec1)
	}

	// 2. Second upload (Different user, same file content -> Instant Upload / 秒传)
	rec2, err := svc.SaveFile(102, "demo_copy.txt", "text/plain", data)
	if err != nil {
		t.Fatalf("second upload failed: %v", err)
	}

	if rec2.FileHash != rec1.FileHash {
		t.Errorf("expected same hash %s, got %s", rec1.FileHash, rec2.FileHash)
	}
	if rec2.StoragePath != rec1.StoragePath {
		t.Errorf("expected same storage path for deduplicated file")
	}

	// 3. Test Signed Token Generation & Verification
	token := svc.GenerateSignToken(rec1.ID, 5*time.Second)
	if err := svc.VerifySignToken(rec1.ID, token); err != nil {
		t.Fatalf("token verification failed: %v", err)
	}

	// Wrong file ID should fail
	if err := svc.VerifySignToken(rec1.ID+99, token); err == nil {
		t.Errorf("expected failure for wrong file ID")
	}
}
