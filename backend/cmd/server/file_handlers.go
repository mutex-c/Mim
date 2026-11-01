package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mim/models"
	"mim/pkg/jwtx"
)

// File Upload Handler (SHA-256 Instant Upload / 秒传)
func (s *Server) handleFileUpload(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 32MB max in memory
	_ = r.ParseMultipartForm(32 << 20)
	fileHeader, handler, err := r.FormFile("file")
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, "Missing file parameter")
		return
	}
	defer fileHeader.Close()

	data, err := io.ReadAll(fileHeader)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to read file bytes")
		return
	}

	fileName := r.FormValue("name")
	if fileName == "" && handler != nil && handler.Filename != "" {
		fileName = handler.Filename
	}
	if fileName == "" {
		fileName = "upload_" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	record, err := s.fileService.SaveFile(userID, fileName, http.DetectContentType(data), data)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	encodedName := url.QueryEscape(record.FileName)
	signToken := s.fileService.GenerateSignToken(record.ID, 30*24*time.Hour)
	fileUrl := fmt.Sprintf("/api/file/download?file_id=%d&name=%s&sign=%s", record.ID, encodedName, url.QueryEscape(signToken))
	s.jsonResponse(w, map[string]interface{}{
		"file_id":   record.ID,
		"file_name": record.FileName,
		"file_size": record.FileSize,
		"file_hash": record.FileHash,
		"file_url":  fileUrl,
		"url":       fileUrl,
		"name":      record.FileName,
		"size":      record.FileSize,
		"mime":      record.FileType,
	})
}

// File Download Handler with authentication and anti-leech protection
func (s *Server) handleFileDownload(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.URL.Query().Get("file_id")
	fileID, _ := strconv.ParseUint(fileIDStr, 10, 64)
	if fileID == 0 {
		http.Error(w, "Invalid file_id", http.StatusBadRequest)
		return
	}

	var record models.FileRecord
	if err := s.db.First(&record, fileID).Error; err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Anti-leech & Authentication check:
	// 1. Check temporary sign token (?sign=...)
	reqSign := r.URL.Query().Get("sign")
	hasValidSign := reqSign != "" && s.fileService.VerifySignToken(record.ID, reqSign) == nil

	// 2. Or JWT authentication (Authorization: Bearer <token> or ?token=<token>)
	hasValidJWT := false
	tokenStr := ""
	if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
	} else if qToken := r.URL.Query().Get("token"); qToken != "" {
		tokenStr = qToken
	}
	if tokenStr != "" {
		if _, err := jwtx.ParseToken(s.cfg.JwtSecret, tokenStr); err == nil {
			hasValidJWT = true
		}
	}

	if !hasValidSign && !hasValidJWT {
		http.Error(w, "Unauthorized: valid download token or signed signature required", http.StatusUnauthorized)
		return
	}

	// Prevent path traversal outside storage directory
	cleanPath := filepath.Clean(record.StoragePath)
	absStorage, _ := filepath.Abs(*storageDir)
	absTarget, _ := filepath.Abs(cleanPath)
	if !strings.HasPrefix(absTarget, absStorage) {
		http.Error(w, "Forbidden: invalid storage path", http.StatusForbidden)
		return
	}

	dlName := r.URL.Query().Get("name")
	if dlName == "" {
		dlName = record.FileName
	}
	encodedName := url.PathEscape(dlName)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"; filename*=UTF-8''%s", dlName, encodedName))
	if record.FileType != "" {
		w.Header().Set("Content-Type", record.FileType)
	}
	http.ServeFile(w, r, record.StoragePath)
}
