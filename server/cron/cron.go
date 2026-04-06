package cron

import (
	"log"
	"opencentral/db"
	"opencentral/src"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func StartCronJobs() {
	go SyncStorageUsage()
	cleanupTicker := time.NewTicker(24 * time.Hour)
	syncTicker := time.NewTicker(1 * time.Hour)

	go func() {
		for {
			select {
			case <-cleanupTicker.C:
				RunTrashCleanup()
			case <-syncTicker.C:
				SyncStorageUsage()
			}
		}
	}()
}

func RunTrashCleanup() {
	log.Println("[Cron] Starting trash cleanup...")

	retentionStr := db.GetSetting("trash_retention_days", "0")
	days, err := strconv.Atoi(retentionStr)
	if err != nil || days <= 0 {
		log.Println("[Cron] Trash retention disabled or invalid (0 days). Skipping.")
		return
	}

	trashPath := src.TrashRoot()

	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return
	}

	now := time.Now()
	cutoff := now.AddDate(0, 0, -days)

	err = filepath.Walk(trashPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == trashPath {
			return nil
		}

		if info.ModTime().Before(cutoff) {
			err := os.RemoveAll(path)
			if err != nil {
				log.Printf("[Cron] Failed to delete expired trash file %s: %v", path, err)
			} else {
				log.Printf("[Cron] Deleted expired trash: %s", info.Name())

				db.Exec("DELETE FROM files WHERE path = ?", path)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("[Cron] Error walking trash: %v", err)
	}

	SyncStorageUsage()
}

func SyncStorageUsage() {
	log.Println("[Cron] Synchronizing storage usage...")
	var total int64

	err := filepath.Walk(src.StorageRoot(), func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})

	if err != nil {
		log.Printf("[Cron] Sync error: %v", err)
		return
	}

	_, err = db.Exec(`
			INSERT INTO system_stats (used_bytes, version) 
			VALUES (?, 1)
			ON CONFLICT(version) DO UPDATE SET used_bytes = excluded.used_bytes
	`, total)

	if err != nil {
		log.Printf("[Cron] Failed to update DB usage: %v", err)
	}
	log.Printf("[Cron] Storage sync complete: %d bytes total", total)
}
