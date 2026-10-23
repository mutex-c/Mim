package file

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mim/models"

	"gorm.io/gorm"
)

var (
	ErrFileNotFound = errors.New("file not found")
	ErrTokenExpired = errors.New("file access token expired")
	ErrInvalidToken = errors.New("invalid file access token")
)

type Config struct {
	StorageDir string
	SignSecret string
}

type Service struct {
	db  *gorm.DB
	cfg Config
}

func NewService(db *gorm.DB, cfg Config) *Service {
	_ = os.MkdirAll(cfg.StorageDir, 0755)
	return &Service{db: db, cfg: cfg}
}

// ComputeHash returns SHA-256 hex string of content
func ComputeHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// CheckFileExists checks if a file with the given SHA-256 already exists (Instant Upload / 秒传)
func (s *Service) CheckFileExists(hash string, size uint64) (*models.FileRecord, bool, error) {
	var record models.FileRecord
	err := s.db.Where("file_hash = ? AND file_size = ?", hash, size).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &record, true, nil
}

// SaveFile saves uploaded file bytes or reuses existing file record if hash matches (Instant Upload)
func (s *Service) SaveFile(uploaderID uint64, fileName, fileType string, data []byte) (*models.FileRecord, error) {
	hash := ComputeHash(data)
	size := uint64(len(data))

	// 1. Check if hash + size exists (Instant Upload / 秒传)
	if existing, found, _ := s.CheckFileExists(hash, size); found {
		return existing, nil
	}

	// 2. Physical write to storage
	ext := filepath.Ext(fileName)
	targetName := fmt.Sprintf("%s_%d%s", hash, size, ext)
	physicalPath := filepath.Join(s.cfg.StorageDir, targetName)

	if err := os.WriteFile(physicalPath, data, 0644); err != nil {
		return nil, err
	}

	record := models.FileRecord{
		FileHash:    hash,
		FileName:    fileName,
		FileSize:    size,
		FileType:    fileType,
		StoragePath: physicalPath,
		UploaderID:  uploaderID,
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

// GenerateSignToken creates a time-limited HMAC token for file access
func (s *Service) GenerateSignToken(fileID uint64, duration time.Duration) string {
	expireTime := time.Now().Add(duration).Unix()
	payload := fmt.Sprintf("%d:%d", fileID, expireTime)

	mac := hmac.New(sha256.New, []byte(s.cfg.SignSecret))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s:%s", payload, signature)
}

// VerifySignToken validates the token and expiration
func (s *Service) VerifySignToken(fileID uint64, token string) error {
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		return ErrInvalidToken
	}

	tokenFileID, _ := strconv.ParseUint(parts[0], 10, 64)
	expireTime, _ := strconv.ParseInt(parts[1], 10, 64)
	sig := parts[2]

	if tokenFileID != fileID {
		return ErrInvalidToken
	}

	if time.Now().Unix() > expireTime {
		return ErrTokenExpired
	}

	payload := fmt.Sprintf("%d:%d", fileID, expireTime)
	mac := hmac.New(sha256.New, []byte(s.cfg.SignSecret))
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return ErrInvalidToken
	}

	return nil
}
