package models

import (
	"time"
)

// FileRecord 文件与资源元数据表 (mim_files)
type FileRecord struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	FileHash    string    `gorm:"type:char(64);not null;uniqueIndex:uk_hash_size,priority:1;column:file_hash;comment:SHA-256 哈希值(用于秒传)" json:"file_hash"`
	FileName    string    `gorm:"type:varchar(255);not null;column:file_name;comment:原始文件名" json:"file_name"`
	FileSize    uint64    `gorm:"not null;uniqueIndex:uk_hash_size,priority:2;column:file_size;comment:文件字节大小" json:"file_size"`
	FileType    string    `gorm:"type:varchar(64);not null;column:file_type;comment:MIME类型" json:"file_type"`
	StoragePath string    `gorm:"type:varchar(255);not null;column:storage_path;comment:存储物理路径或Key" json:"storage_path"`
	UploaderID  uint64    `gorm:"not null;index:idx_uploader;column:uploader_id;comment:上传者ID" json:"uploader_id"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (FileRecord) TableName() string {
	return "mim_files"
}
