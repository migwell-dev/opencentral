package src

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"opencentral/db"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ricochet2200/go-disk-usage/du"
)

type Setting struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

func GetSettings(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT key, value, updated_at FROM user_settings")
	if err != nil {
		http.Error(w, "database error", 500)
		log.Println("[OpenCentral] GetSettings error:", err)
		return
	}
	defer rows.Close()

	result := map[string]Setting{}
	for rows.Next() {
		var s Setting
		if err := rows.Scan(&s.Key, &s.Value, &s.UpdatedAt); err != nil {
			continue
		}
		result[s.Key] = s
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
	log.Println("[OpenCentral] GET /api/settings - 200 OK")
}

func UpdateSetting(w http.ResponseWriter, r *http.Request) {
	var req Setting
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", 400)
		log.Println(err)
		return
	}

	if err := validateSetting(req.Key, req.Value); err != nil {
		http.Error(w, err.Error(), 400)
		log.Printf("[OpenCentral] PATCH /api/settings - validation failed: %v", err)
		return
	}

	_, err := db.Exec(`
		INSERT INTO user_settings (key, value, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP
	`, req.Key, req.Value)
	if err != nil {
		http.Error(w, "database error", 500)
		log.Println("[OpenCentral] UpdateSetting exec error:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"key": req.Key, "value": req.Value})
	log.Printf("[OpenCentral] PATCH /api/settings - updated %s = %s", req.Key, req.Value)
}

func validateSetting(key, value string) error {
	switch key {
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil || port < 1024 || port > 65535 {
			return fmt.Errorf("port must be an integer between 1024 and 65535")
		}

	case "storage_limit_bytes":
		limit, err := strconv.ParseInt(value, 10, 64)
		if err != nil || limit < 0 {
			return fmt.Errorf("storage_limit_bytes must be a non-negative integer")
		}

		if limit > 0 {
			used := GetUsedBytes()

			if limit < used {
				return fmt.Errorf(
					"new limit (%d bytes) is lower than current usage (%d bytes). Delete files first.",
					limit, used,
				)
			}

			free, _ := diskAvailableBytes()
			log.Println("used:", used, "free:", free, "limit:", limit)
			if limit > (used + free) {
				return fmt.Errorf("limit exceeds physical disk space")
			}
		}

	case "trash_retention_days":
		days, err := strconv.Atoi(value)
		if err != nil || days < 0 {
			return fmt.Errorf("trash_retention_days must be a non-negative integer (0 = keep forever)")
		}

	default:
		return fmt.Errorf("unknown setting key: %s", key)
	}

	return nil
}

func diskAvailableBytes() (int64, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	usage := du.NewDiskUsage(cwd)
	return int64(usage.Available()), nil
}

func StorageLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limitStr := db.GetSetting("storage_limit_bytes", "0")
		limit, _ := strconv.ParseInt(limitStr, 10, 64)
		if limit <= 0 {
			next(w, r)
			return
		}

		used := GetUsedBytes()

		incoming := r.ContentLength
		if incoming > 0 && (used+incoming) > limit {
			http.Error(w, "Insufficient Storage", 507)
			return
		}

		next(w, r)
	}
}

// GetStorageUsage returns current usage as bytes and percentage.
// GET /api/storage/usage
func GetStorageUsage(w http.ResponseWriter, r *http.Request) {
	used := GetUsedBytes()

	limitStr := db.GetSetting("storage_limit_bytes", "0")
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	var percent float64
	if limit > 0 {
		percent = float64(used) / float64(limit) * 100
	}

	writeJSON(w, map[string]any{
		"usedBytes":  used,
		"limitBytes": limit,
		"percent":    percent,
	})
	log.Println("[OpenCentral] GET /api/storage/usage - 200 OK")
}

func dirSize(path string) (int64, error) {
	var total int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

func UpdateUsedBytes(delta int64) error {
	_, err := db.Exec("UPDATE system_stats SET used_bytes = used_bytes + ?", delta)
	return err
}

func GetUsedBytes() int64 {
	var used int64
	db.QueryRow("SELECT used_bytes FROM system_stats").Scan(&used)
	return used
}
