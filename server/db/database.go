package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var db *sql.DB

const DB_ROOT = "../storage/app.db"

func getDbRoot() string {
	return DB_ROOT
}

func InitDB() {
	log.Println("[OpenCentral] Initializing Database")
	var err error
	db, err = sql.Open("sqlite", DB_ROOT)
	if err != nil {
		log.Fatal("[OpenCentral] Failed to open DB:", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		log.Fatal("[OpenCentral] Failed to connect to DB:", err)
	}
	for _, pragma := range []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA foreign_keys = ON;`,
	} {
		if _, err := db.Exec(pragma); err != nil {
			log.Println("[OpenCentral] PRAGMA error:", err)
		}
	}
	log.Println("[OpenCentral] Database connected")
	createTables()
}

func InitDBAt(path string) {
	log.Println("[OpenCentral] Initializing Database")

	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatal("[OpenCentral] Failed to open DB:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("[OpenCentral] Failed to connect to DB:", err)
	}

	_, err = db.Exec(`PRAGMA journal_mode = WAL;`)
	if err != nil {
		log.Println("[OpenCentral] Failed to enable WAL:", err)
	}

	log.Println("[OpenCentral] Database connected")

	createTables()
}

func SetSetting(key, value string) {
	db.Exec("INSERT OR REPLACE INTO user_settings (key, value) VALUES (?, ?)", key, value)
}

func CloseDB() { db.Close() }

func createTables() {
	query := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		path TEXT UNIQUE,
		size INTEGER,
		is_dir BOOLEAN,
		prev_path TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER UNIQUE,
		starred BOOLEAN DEFAULT 0,
		last_opened DATETIME,
		FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS user_settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS system_stats (
		used_bytes BIGINTEGER DEFAULT 0,
		version INTEGER NOT NULL UNIQUE
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("[OpenCentral] Error while creating tables:", err)
	}

	defaults := []struct{ k, v string }{
		{"port", "8080"},
		{"storage_limit_bytes", "0"},
		{"trash_retention_days", "30"},
	}
	for _, d := range defaults {
		_, _ = db.Exec(`
			INSERT OR IGNORE INTO user_settings (key, value)
			VALUES (?, ?)
		`, d.k, d.v)
	}
}

func UpsertFile(name, path string, size int64, isDir bool) (int64, error) {
	res, err := db.Exec(`
	INSERT INTO files (name, path, size, is_dir)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET
		name=excluded.name,
		size=excluded.size,
		is_dir=excluded.is_dir
	`, name, path, size, isDir)

	if err != nil {
		log.Println("[DB] upsert error:", err)
	}

	return res.LastInsertId()
}

func GetSetting(key, fallback string) string {
	var value string
	err := QueryRow("SELECT value FROM user_settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return fallback
	}
	return value
}

// expose db api
func Exec(query string, args ...any) (sql.Result, error) {
	return db.Exec(query, args...)
}

func QueryRow(query string, args ...any) *sql.Row {
	return db.QueryRow(query, args...)
}

func Query(query string, args ...any) (*sql.Rows, error) {
	return db.Query(query, args...)
}
