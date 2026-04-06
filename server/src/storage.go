package src

import (
	"database/sql"
	"errors"
	"log"
	"opencentral/db"
	"os"
	"path/filepath"
	"strings"
)

var root = "../storage"

func InitStorage() {
	log.Println("[OpenCentral] Initializing Storage")
	path := StorageRoot()

	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			log.Println("[OpenCentral] Storage exists, skipping...")
		} else {
			log.Println("[OpenCentral] Exception: Storage is not a directory")
		}
	} else if errors.Is(err, os.ErrNotExist) {
		log.Println("[OpenCentral] Storage does not exist. Making...")
		err := os.MkdirAll(path+"/.trash", 0755)
		if err != nil {
			log.Println("[OpenCentral] Something went wrong while making directory:", err)
		}
	} else {
		log.Println("[OpenCentral] Something went wrong while making directory, do you have the appropriate permissions?")
	}
}

func SyncFiles() {
	root := StorageRoot()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if path == root {
			return nil
		}

		name := info.Name()

		if name == ".trash" || strings.HasSuffix(name, ".db") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		cleanPath := filepath.Clean(path)

		var fileID int64
		err = db.QueryRow(
			"SELECT id FROM files WHERE path = ?",
			cleanPath,
		).Scan(&fileID)

		if err == sql.ErrNoRows {
			fileID, err = db.UpsertFile(
				name,
				cleanPath,
				info.Size(),
				info.IsDir(),
			)
			if err != nil {
				log.Println("[OpenCentral] syncFiles - upsert error:", err)
				return nil
			}
		} else if err != nil {
			log.Println("[OpenCentral] syncFiles - query error:", err)
			return nil
		}

		_, err = db.Exec(`
			INSERT INTO metadata (file_id, starred, last_opened)
			VALUES (?, 0, NULL)
			ON CONFLICT(file_id) DO NOTHING
		`, fileID)

		if err != nil {
			log.Println("[OpenCentral] syncFiles - metadata insert error:", err)
		}

		return nil
	})

	if err != nil {
		log.Println("[OpenCentral] syncFiles - walk error:", err)
	}

	log.Println("[OpenCentral] syncFiles - done")
}

func StorageRoot() string {
	return root
}

func TrashRoot() string {
	return root + "/.trash"
}

func overrideStorageRoot(p string) { root = p }
