package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mim/models"
)

type Config struct {
	Driver          string        // "sqlite" or "mysql"
	DSN             string        // database connection string or file path
	MaxOpenConns    int           // maximum open connections
	MaxIdleConns    int           // maximum idle connections
	ConnMaxLifetime time.Duration // connection maximum lifetime
}

// InitDB initializes database connection based on configured driver (SQLite WAL or MySQL 8)
func InitDB(cfg Config) (*gorm.DB, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	if driver == "" {
		driver = "sqlite"
	}

	var dialector gorm.Dialector

	switch driver {
	case "mysql":
		if cfg.DSN == "" {
			cfg.DSN = "root:mim_root_password_2026@tcp(127.0.0.1:3306)/mim_db?charset=utf8mb4&parseTime=True&loc=Local"
		}
		dialector = mysql.Open(cfg.DSN)

	case "sqlite":
		if cfg.DSN == "" {
			cfg.DSN = "mim.db"
		}
		// Ensure directory exists if path contains folder
		if dir := filepath.Dir(cfg.DSN); dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0755)
		}
		// Append WAL mode and concurrency flags if not already present
		dsn := cfg.DSN
		if !strings.Contains(dsn, "_journal=") {
			sep := "?"
			if strings.Contains(dsn, "?") {
				sep = "&"
			}
			dsn = fmt.Sprintf("%s%s_journal=WAL&_busy_timeout=5000&_synchronous=NORMAL&_cache_size=-64000", dsn, sep)
		}
		dialector = sqlite.Open(dsn)

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database (%s): %w", driver, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if driver == "mysql" {
		maxOpen := cfg.MaxOpenConns
		if maxOpen <= 0 {
			maxOpen = 100
		}
		maxIdle := cfg.MaxIdleConns
		if maxIdle <= 0 {
			maxIdle = 20
		}
		lifetime := cfg.ConnMaxLifetime
		if lifetime <= 0 {
			lifetime = time.Hour
		}
		sqlDB.SetMaxOpenConns(maxOpen)
		sqlDB.SetMaxIdleConns(maxIdle)
		sqlDB.SetConnMaxLifetime(lifetime)
	} else {
		// SQLite single-writer optimization for zero lock contention
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	}

	// Run auto migrations for all core schemas
	if err := models.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto migrate models: %w", err)
	}

	log.Printf("[Database] Initialized %s database successfully (DSN: %s)", strings.ToUpper(driver), maskDSN(cfg.DSN))
	return db, nil
}

func maskDSN(dsn string) string {
	if strings.Contains(dsn, ":") && strings.Contains(dsn, "@") {
		// Mask password in mysql DSN: user:pass@tcp -> user:******@tcp
		parts := strings.SplitN(dsn, "@", 2)
		userPass := parts[0]
		if idx := strings.Index(userPass, ":"); idx != -1 {
			return userPass[:idx+1] + "******@" + parts[1]
		}
	}
	return dsn
}
