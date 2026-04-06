package src

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"opencentral/db"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type RenameRequest struct {
	OldPath string `json:"oldPath"`
	NewName string `json:"newName"`
}

type MoveRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type MkdirRequest struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

func ListFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	fullPath, err := resolvePath(path)
	if err != nil {
		http.Error(w, "invalid path", 400)
		log.Println("[OpenCentral] route /api/files - 400 err:", err)
		return
	}

	files, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println("[OpenCentral] route /api/files - 400 err:", err)
		return
	}

	var result []map[string]any

	for _, f := range files {
		if !strings.Contains(f.Name(), ".db") && f.Name() != ".trash" {
			info, _ := f.Info()
			full := filepath.Join(fullPath, f.Name())

			var starred bool
			var lastOpened sql.NullString

			_ = db.QueryRow(
				`SELECT COALESCE(m.starred, 0), COALESCE(m.last_opened, '')
				 FROM files f
				 LEFT JOIN metadata m ON m.file_id = f.id
				 WHERE f.path = ?`,
				full,
			).Scan(&starred, &lastOpened)

			result = append(result, map[string]any{
				"name":    f.Name(),
				"isDir":   f.IsDir(),
				"size":    info.Size(),
				"starred": starred,
				"last_opened": func() any {
					if !lastOpened.Valid || lastOpened.String == "" {
						return nil
					}
					t, err := time.Parse("2006-01-02 15:04:05", lastOpened.String)
					if err != nil {
						return nil
					}
					return t.Format(time.RFC3339)
				}(),
			})
		}
	}

	if result == nil {
		result = []map[string]any{}
	}

	writeJSON(w, result)
	log.Println("[OpenCentral] route /api/files - 200 OK")
}

func ListTrashFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	fullPath, err := resolveTrashPath(path)
	if err != nil {
		http.Error(w, "invalid path", 400)
		log.Println("[OpenCentral] route /api/trash - 400 err:", err)
		return
	}

	files, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println("[OpenCentral] route /api/trash - 500 err:", err)
		return
	}

	var result []map[string]any

	for _, f := range files {
		info, _ := f.Info()
		full := filepath.Join(fullPath, f.Name())

		var starred bool
		var lastOpened sql.NullString

		_ = db.QueryRow(
			`SELECT starred, last_opened 
			 FROM metadata m 
			 JOIN files f ON m.file_id = f.id 
			 WHERE f.path = ?`,
			full,
		).Scan(&starred, &lastOpened)

		result = append(result, map[string]any{
			"name":    f.Name(),
			"isDir":   f.IsDir(),
			"size":    info.Size(),
			"starred": starred,
			"last_opened": func() any {
				if !lastOpened.Valid || lastOpened.String == "" {
					return nil
				}
				t, err := time.Parse("2006-01-02 15:04:05", lastOpened.String)
				if err != nil {
					return nil
				}
				return t.Format(time.RFC3339)
			}(),
		})
	}

	if result == nil {
		result = []map[string]any{}
	}

	writeJSON(w, result)
	log.Println("[OpenCentral] route /api/trash - 200 OK")
}

func UploadFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	fullPath, err := resolvePath(path)
	if err != nil {
		http.Error(w, "invalid path", 400)
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	limitStr := db.GetSetting("storage_limit_bytes", "0")
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if part.FileName() == "" {
			continue
		}

		used := GetUsedBytes()
		var partReader io.Reader = part
		if limit > 0 {
			remaining := limit - used
			if remaining <= 0 {
				http.Error(w, "storage limit reached", 507)
				return
			}
			partReader = io.LimitReader(part, remaining)
		}

		dstPath := filepath.Join(fullPath, filepath.Base(part.FileName()))
		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		written, copyErr := io.Copy(dst, partReader)
		dst.Close()

		if copyErr != nil {
			os.Remove(dstPath)
			http.Error(w, "upload failed", 500)
			return
		}

		UpdateUsedBytes(written)

		fileID, err := db.UpsertFile(part.FileName(), dstPath, written, false)
		if err != nil {
			log.Printf("[OpenCentral] DB Upsert error: %v", err)
			continue
		}

		_, err = db.Exec(`
            INSERT INTO metadata (file_id, last_opened) 
            VALUES (?, CURRENT_TIMESTAMP)
            ON CONFLICT(file_id) DO UPDATE SET last_opened = CURRENT_TIMESTAMP
        `, fileID)

		if err != nil {
			log.Printf("[OpenCentral] Metadata init error: %v", err)
		}
	}

	w.Write([]byte("uploaded"))
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	fullPath, err := resolvePath(path)
	if err != nil {
		http.Error(w, "invalid path", 400)
		return
	}

	sizeToDelete, err := dirSize(fullPath)
	if err != nil {
		http.Error(w, "file not found", 404)
		return
	}

	err = os.RemoveAll(fullPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if err := UpdateUsedBytes(-sizeToDelete); err != nil {
		log.Printf("[OpenCentral] Error updating stats after delete: %v", err)
	}

	_, err = db.Exec("DELETE FROM files WHERE path = ?", fullPath)

	w.Write([]byte("deleted"))
}

func RenameFile(w http.ResponseWriter, r *http.Request) {
	var req RenameRequest
	json.NewDecoder(r.Body).Decode(&req)

	oldFull, err := resolvePath(req.OldPath)
	if err != nil {
		http.Error(w, "invalid path", 400)
		log.Println("[OpenCentral] route /api/rename - 400 err:", err)
		return
	}

	newFull := filepath.Join(filepath.Dir(oldFull), req.NewName)

	err = os.Rename(oldFull, newFull)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println("[OpenCentral] route /api/rename - 500 err:", err)
		return
	}

	_, err = db.Exec(
		"UPDATE files SET name = ?, path = ? WHERE path = ?",
		req.NewName,
		newFull,
		oldFull,
	)

	w.Write([]byte("renamed"))
	log.Println("[OpenCentral] route /api/rename - 200 OK")
}

func MoveFile(w http.ResponseWriter, r *http.Request) {
	var req MoveRequest
	json.NewDecoder(r.Body).Decode(&req)

	srcPath, err := resolvePath(req.Source)
	if err != nil {
		http.Error(w, "invalid source", 400)
		log.Println("[OpenCentral] route /api/move - 400 err:", err)
		return
	}

	dstPath, err := resolvePath(req.Destination)
	if err != nil {
		http.Error(w, "invalid destination", 400)
		log.Println("[OpenCentral] route /api/move - 400 err:", err)
		return
	}

	err = os.Rename(srcPath, dstPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println("[OpenCentral] route /api/move - 500 err:", err)
		return
	}
	_, err = db.Exec(
		"UPDATE files SET path = ? WHERE path = ?",
		dstPath,
		srcPath,
	)

	w.Write([]byte("moved"))
	log.Println("[OpenCentral] route /api/move - 200 OK")
}

func CreateDirectory(w http.ResponseWriter, r *http.Request) {
	var req MkdirRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid body", 400)
		log.Println("[OpenCentral] route /api/mkdir - 400 err:", err)
		return
	}

	if strings.Contains(req.Name, "/") || strings.Contains(req.Name, "\\") {
		http.Error(w, "directory name required", 400)
		log.Println("[OpenCentral] route /api/mkdir - 400 err:", err)
		return
	}

	parentPath, err := resolvePath(req.Path)
	if err != nil {
		http.Error(w, "invalid path", 400)
		log.Println("[OpenCentral] route /api/mkdir - 400 err:", err)
		return
	}

	newDir := filepath.Join(parentPath, req.Name)

	cleanDir := filepath.Clean(newDir)

	if cleanDir != newDir {
		http.Error(w, "invalid directory name", 400)
		log.Println("[OpenCentral] route /api/mkdir - 400 err:", err)
		return
	}

	err = os.Mkdir(cleanDir, 0755)
	if err != nil {
		if os.IsExist(err) {
			http.Error(w, "directory already exists", 400)
			log.Println("[OpenCentral] route /api/mkdir - 400 err:", err)
			return
		}
		http.Error(w, err.Error(), 500)
		log.Println("[OpenCentral] route /api/mkdir - 500 err:", err)
		return
	}

	db.UpsertFile(
		req.Name,
		cleanDir,
		0,
		true,
	)

	w.Write([]byte("directory created"))
	log.Println("[OpenCentral] route /api/mkdir - 200 OK")
}

func ServeFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	path := r.URL.Query().Get("path")
	fullPath, err := resolvePath(path)
	if err != nil {
		http.Error(w, "invalid path", 400)
		log.Println("[OpenCentral] route /api/file - 400:", err)
		return
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "file not found", 404)
		log.Println("[OpenCentral] route /api/file - 404:", err)
		return
	}
	if info.IsDir() {
		http.Error(w, "cannot serve directory", 400)
		return
	}

	if r.Header.Get("Range") == "" {
		log.Println("[OpenCentral] ServeFile - looking up path:", fullPath)

		var fileID int
		err := db.QueryRow("SELECT id FROM files WHERE path = ?", fullPath).Scan(&fileID)
		if err != nil {
			log.Println("[OpenCentral] ServeFile - file not found in DB:", err, "path:", fullPath)
		} else {
			log.Println("[OpenCentral] ServeFile - found fileID:", fileID)
			_, err = db.Exec(`
                INSERT INTO metadata (file_id, last_opened)
                VALUES (?, CURRENT_TIMESTAMP)
                ON CONFLICT(file_id) DO UPDATE SET
                last_opened = CURRENT_TIMESTAMP
            `, fileID)
			if err != nil {
				log.Println("[OpenCentral] ServeFile - metadata update error:", err)
			} else {
				log.Println("[OpenCentral] ServeFile - last_opened updated for fileID:", fileID)
			}
		}
	}

	http.ServeFile(w, r, fullPath)
	log.Println("[OpenCentral] route /api/file - 200 OK:", fullPath)
}

func StarFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	fullPath, err := resolvePath(path)
	if err != nil {
		http.Error(w, "invalid path", 400)
		log.Println("[OpenCentral] route /api/files/star - 400:", err)
		return
	}
	_, err = os.Stat(fullPath)
	if err != nil {
		http.Error(w, "file not found", 404)
		log.Println("[OpenCentral] route /api/files/star - 404:", err)
		return
	}

	var fileID int
	err = db.QueryRow("SELECT id FROM files WHERE path = ?", fullPath).Scan(&fileID)
	if err != nil {
		http.Error(w, "database error", 500)
		log.Println("[OpenCentral] route /api/files/star - file not in db:", err)
		return
	}

	var starred bool
	err = db.QueryRow("SELECT starred FROM metadata WHERE file_id = ?", fileID).Scan(&starred)
	if err != nil {
		http.Error(w, "database error", 500)
		log.Println("[OpenCentral] route /api/files/star - metadata not found:", err)
		return
	}

	newStarred := !starred
	_, err = db.Exec(`
		UPDATE metadata SET starred = ? WHERE file_id = ?
	`, newStarred, fileID)
	if err != nil {
		http.Error(w, "database error", 500)
		log.Println("[OpenCentral] route /api/files/star 500:", err)
		return
	}

	if newStarred {
		w.Write([]byte("starred"))
	} else {
		w.Write([]byte("unstarred"))
	}
	log.Println("[OpenCentral] route /api/files/star 200: starred=", newStarred)
}

func TrashFile(w http.ResponseWriter, r *http.Request) {
	var req MoveRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	srcPath, err := resolvePath(req.Source)
	if err != nil {
		http.Error(w, "invalid source", 400)
		log.Println("[OpenCentral] route /api/trash - 400 err:", err)
		return
	}

	if _, err := os.Stat(srcPath); err != nil {
		http.Error(w, "file not found", 404)
		return
	}

	trashDir, err := resolveTrashPath("/")
	if err != nil {
		http.Error(w, "invalid trash directory", 500)
		log.Println("[OpenCentral] resolve trash error:", err)
		return
	}

	rel, err := filepath.Rel(trashDir, srcPath)
	if err == nil && !strings.HasPrefix(rel, "..") {
		http.Error(w, "file already in trash", 400)
		return
	}

	fileName := filepath.Base(srcPath)
	dstPath := filepath.Join(trashDir, fileName)

	if _, err := os.Stat(dstPath); err == nil {
		ext := filepath.Ext(fileName)
		name := strings.TrimSuffix(fileName, ext)
		dstPath = filepath.Join(
			trashDir,
			fmt.Sprintf("%s_%d%s", name, time.Now().Unix(), ext),
		)
	}

	err = os.Rename(srcPath, dstPath)
	if err != nil {
		input, copyErr := os.ReadFile(srcPath)
		if copyErr != nil {
			http.Error(w, "failed to move file", 500)
			log.Println("[OpenCentral] read fallback error:", copyErr)
			return
		}

		copyErr = os.WriteFile(dstPath, input, 0644)
		if copyErr != nil {
			http.Error(w, "failed to move file", 500)
			log.Println("[OpenCentral] write fallback error:", copyErr)
			return
		}

		_ = os.Remove(srcPath)
	}

	_, err = db.Exec(
		"UPDATE files SET path = ?, prev_path = ? WHERE path = ?",
		dstPath,
		srcPath,
		srcPath,
	)
	if err != nil {
		log.Println("[DB] update error:", err)
	}

	w.Write([]byte("moved to trash"))
	log.Println("[OpenCentral] route /api/trash - moved:", dstPath)
}

func RestoreFile(w http.ResponseWriter, r *http.Request) {
	var req MoveRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	srcPath, err := resolveTrashPath(req.Source)
	if err != nil {
		http.Error(w, "invalid source", 400)
		return
	}

	if _, err := os.Stat(srcPath); err != nil {
		http.Error(w, "file not found", 404)
		return
	}

	var prevPath string
	err = db.QueryRow(
		"SELECT prev_path FROM files WHERE path = ?",
		srcPath,
	).Scan(&prevPath)

	if err != nil || prevPath == "" {
		http.Error(w, "original path not found", 400)
		return
	}

	dstPath := filepath.Clean(prevPath)
	if !strings.HasPrefix(dstPath, filepath.Clean(StorageRoot())) {
		http.Error(w, "invalid destination", 400)
		return
	}

	err = os.MkdirAll(filepath.Dir(dstPath), os.ModePerm)
	if err != nil {
		http.Error(w, "failed to create destination directory", 500)
		return
	}

	if _, err := os.Stat(dstPath); err == nil {
		ext := filepath.Ext(dstPath)
		name := strings.TrimSuffix(filepath.Base(dstPath), ext)
		dstPath = filepath.Join(
			filepath.Dir(dstPath),
			fmt.Sprintf("%s_restored_%d%s", name, time.Now().Unix(), ext),
		)
	}

	err = os.Rename(srcPath, dstPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = db.Exec(
		"UPDATE files SET path = ?, prev_path = NULL WHERE path = ?",
		dstPath,
		srcPath,
	)
	if err != nil {
		log.Println("[DB] update error:", err)
	}

	w.Write([]byte("restored"))
	log.Println("[OpenCentral] restored:", dstPath)
}

func BatchDownload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Paths []string `json:"paths"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Paths) == 0 {
		http.Error(w, "invalid request", 400)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="download.zip"`)

	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, p := range req.Paths {
		fullPath, err := resolvePath(p)
		if err != nil {
			continue
		}
		if err := addToZip(zw, fullPath, filepath.Base(fullPath)); err != nil {
			log.Println("[OpenCentral] BatchDownload zip error:", err)
		}
	}

	log.Println("[OpenCentral] POST /api/download/batch - 200 OK")
}

func addToZip(zw *zip.Writer, path, name string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return filepath.Walk(path, func(filePath string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(filepath.Dir(path), filePath)
			fw, err := zw.Create(rel)
			if err != nil {
				return err
			}
			f, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(fw, f)
			return err
		})
	}

	fw, err := zw.Create(name)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(fw, f)
	return err
}
