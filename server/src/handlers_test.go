// src/handlers_test.go
package src

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"opencentral/db"
)

// ---------------------------------------------------------------------------
// Harness
// ---------------------------------------------------------------------------

func setupTest(t *testing.T) func() {
	t.Helper()

	// Temp storage directory
	tmpStorage, err := os.MkdirTemp("", "oc-storage-*")
	if err != nil {
		t.Fatalf("create temp storage: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpStorage, ".trash"), 0755); err != nil {
		t.Fatalf("create temp trash dir: %v", err)
	}
	overrideStorageRoot(tmpStorage) // swaps the package-level StorageRoot var

	// Temp SQLite DB
	tmpDB, err := os.CreateTemp("", "oc-test-*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	tmpDB.Close()
	db.InitDBAt(tmpDB.Name()) // same as InitDB but accepts a path

	return func() {
		os.RemoveAll(tmpStorage)
		os.Remove(tmpDB.Name())
		db.CloseDB()
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeRequest(method, target string, body io.Reader) *http.Request {
	r := httptest.NewRequest(method, target, body)
	return r
}

func do(handler http.HandlerFunc, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	handler(w, r)
	return w
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

func multipartUpload(t *testing.T, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("multipart create: %v", err)
	}
	fw.Write([]byte(content))
	mw.Close()
	return &buf, mw.FormDataContentType()
}

// ---------------------------------------------------------------------------
// ListFiles
// ---------------------------------------------------------------------------

func TestListFiles_EmptyDir(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	r := makeRequest("GET", "/api/files?path=/", nil)
	w := do(ListFiles, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d items", len(result))
	}
}

func TestListFiles_WithFiles(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	writeFile(t, filepath.Join(StorageRoot(), "hello.txt"), "hello")

	r := makeRequest("GET", "/api/files?path=/", nil)
	w := do(ListFiles, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result))
	}
	if result[0]["name"] != "hello.txt" {
		t.Errorf("unexpected name: %v", result[0]["name"])
	}
}

func TestListFiles_InvalidPath(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	r := makeRequest("GET", "/api/files?path=/nonexistent-dir", nil)
	w := do(ListFiles, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for missing dir, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// UploadFile
// ---------------------------------------------------------------------------

func TestUploadFile_OK(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	buf, ct := multipartUpload(t, "test.txt", "hello world")
	r := makeRequest("POST", "/api/upload?path=/", buf)
	r.Header.Set("Content-Type", ct)

	w := do(UploadFile, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if _, err := os.Stat(filepath.Join(StorageRoot(), "test.txt")); err != nil {
		t.Errorf("uploaded file not found on disk: %v", err)
	}
}

func TestUploadFile_StorageLimitExceeded(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	// Set limit to 1 byte
	db.SetSetting("storage_limit_bytes", "1")

	buf, ct := multipartUpload(t, "big.txt", "this is definitely more than 1 byte")
	r := makeRequest("POST", "/api/upload?path=/", buf)
	r.Header.Set("Content-Type", ct)

	w := do(UploadFile, r)

	if w.Code != http.StatusInsufficientStorage {
		t.Errorf("expected 507, got %d", w.Code)
	}

	// Partial file must not remain on disk
	if _, err := os.Stat(filepath.Join(StorageRoot(), "big.txt")); !os.IsNotExist(err) {
		t.Errorf("partial file should have been removed")
	}
}

func TestUploadFile_InvalidPath(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	r := makeRequest("POST", "/api/upload?path=/nonexistent/subdir", nil)
	buf, ct := multipartUpload(t, "x.txt", "x") // rebuild buf inside
	r = makeRequest("POST", "/api/upload?path=/nonexistent/subdir", buf)
	r.Header.Set("Content-Type", ct)
	w := do(UploadFile, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for missing subdir, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// DeleteFile
// ---------------------------------------------------------------------------

func TestDeleteFile_OK(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	p := filepath.Join(StorageRoot(), "todelete.txt")
	writeFile(t, p, "bye")

	r := makeRequest("DELETE", "/api/delete?path=/todelete.txt", nil)
	w := do(DeleteFile, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("file should have been deleted")
	}
}

func TestDeleteFile_NotFound(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	r := makeRequest("DELETE", "/api/delete?path=/ghost.txt", nil)
	w := do(DeleteFile, r)

	// os.RemoveAll on a non-existent path returns nil, so this is still 200.
	// If you want 404 behaviour, add an os.Stat check before RemoveAll.
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// RenameFile
// ---------------------------------------------------------------------------

func TestRenameFile_OK(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	p := filepath.Join(StorageRoot(), "old.txt")
	writeFile(t, p, "content")

	body, _ := json.Marshal(map[string]string{
		"oldPath": "/old.txt",
		"newName": "new.txt",
	})
	r := makeRequest("POST", "/api/rename", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := do(RenameFile, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(StorageRoot(), "new.txt")); err != nil {
		t.Errorf("renamed file not found: %v", err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("old file should be gone")
	}
}

func TestRenameFile_InvalidPath(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	body, _ := json.Marshal(map[string]string{
		"oldPath": "/doesnotexist.txt",
		"newName": "new.txt",
	})
	r := makeRequest("POST", "/api/rename", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := do(RenameFile, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// MoveFile
// ---------------------------------------------------------------------------

func TestMoveFile_OK(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	src := filepath.Join(StorageRoot(), "move-me.txt")
	dst := filepath.Join(StorageRoot(), "subdir")
	os.Mkdir(dst, 0755)
	writeFile(t, src, "moving")

	body, _ := json.Marshal(map[string]string{
		"source":      "/move-me.txt",
		"destination": "/subdir/move-me.txt",
	})
	r := makeRequest("POST", "/api/move", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := do(MoveFile, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dst, "move-me.txt")); err != nil {
		t.Errorf("file not at destination: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CreateDirectory
// ---------------------------------------------------------------------------

func TestCreateDirectory_OK(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	body, _ := json.Marshal(map[string]string{
		"path": "/",
		"name": "newdir",
	})
	r := makeRequest("POST", "/api/mkdir", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := do(CreateDirectory, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	info, err := os.Stat(filepath.Join(StorageRoot(), "newdir"))
	if err != nil || !info.IsDir() {
		t.Errorf("directory not created")
	}
}

func TestCreateDirectory_AlreadyExists(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	os.Mkdir(filepath.Join(StorageRoot(), "dupe"), 0755)

	body, _ := json.Marshal(map[string]string{"path": "/", "name": "dupe"})
	r := makeRequest("POST", "/api/mkdir", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := do(CreateDirectory, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateDirectory_SlashInName(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	body, _ := json.Marshal(map[string]string{"path": "/", "name": "a/b"})
	r := makeRequest("POST", "/api/mkdir", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := do(CreateDirectory, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// ServeFile
// ---------------------------------------------------------------------------

func TestServeFile_OK(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	writeFile(t, filepath.Join(StorageRoot(), "serve-me.txt"), "served!")

	r := makeRequest("GET", "/api/file?path=/serve-me.txt", nil)
	w := do(ServeFile, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "served!" {
		t.Errorf("unexpected body: %q", w.Body.String())
	}
}

func TestServeFile_Directory(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	os.Mkdir(filepath.Join(StorageRoot(), "adir"), 0755)
	r := makeRequest("GET", "/api/file?path=/adir", nil)
	w := do(ServeFile, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for directory, got %d", w.Code)
	}
}

func TestServeFile_NotFound(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	r := makeRequest("GET", "/api/file?path=/nope.txt", nil)
	w := do(ServeFile, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// StarFiles
// ---------------------------------------------------------------------------

func TestStarFiles_Toggle(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	p := filepath.Join(StorageRoot(), "star-me.txt")
	writeFile(t, p, "⭐")
	db.UpsertFile("star-me.txt", p, 2, false)

	r := makeRequest("POST", "/api/files/star?path=/star-me.txt", nil)
	w := do(StarFiles, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "starred" {
		t.Errorf("expected 'starred', got %q", w.Body.String())
	}

	// Toggle back
	r2 := makeRequest("POST", "/api/files/star?path=/star-me.txt", nil)
	w2 := do(StarFiles, r2)
	if w2.Body.String() != "unstarred" {
		t.Errorf("expected 'unstarred', got %q", w2.Body.String())
	}
}

// ---------------------------------------------------------------------------
// TrashFile + RestoreFile
// ---------------------------------------------------------------------------

func TestTrashAndRestore(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	p := filepath.Join(StorageRoot(), "trash-me.txt")
	writeFile(t, p, "going to trash")
	db.UpsertFile("trash-me.txt", p, 14, false)

	// Trash it
	body, _ := json.Marshal(map[string]string{"source": "/trash-me.txt"})
	r := makeRequest("POST", "/api/trash", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := do(TrashFile, r)

	if w.Code != http.StatusOK {
		t.Fatalf("trash: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("original file should be gone after trash")
	}

	// Restore it
	body2, _ := json.Marshal(map[string]string{"source": "/trash-me.txt"})
	r2 := makeRequest("POST", "/api/restore", bytes.NewReader(body2))
	r2.Header.Set("Content-Type", "application/json")
	w2 := do(RestoreFile, r2)

	if w2.Code != http.StatusOK {
		t.Fatalf("restore: expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
	if _, err := os.Stat(p); err != nil {
		t.Errorf("restored file should be back at original path: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

func TestGetSettings(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	r := makeRequest("GET", "/api/settings", nil)
	w := do(GetSettings, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]Setting
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := result["port"]; !ok {
		t.Errorf("expected default 'port' key in settings")
	}
}

func TestUpdateSetting_Port(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	body, _ := json.Marshal(map[string]string{"key": "port", "value": "9090"})
	r := makeRequest("PATCH", "/api/settings/update", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := do(UpdateSetting, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if got := db.GetSetting("port", ""); got != "9090" {
		t.Errorf("expected port 9090, got %q", got)
	}
}

func TestUpdateSetting_InvalidPort(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	for _, bad := range []string{"80", "99999", "abc", "-1"} {
		body, _ := json.Marshal(map[string]string{"key": "port", "value": bad})
		r := makeRequest("PATCH", "/api/settings/update", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := do(UpdateSetting, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("port %q: expected 400, got %d", bad, w.Code)
		}
	}
}

func TestUpdateSetting_StorageLimitExceedsDisk(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	// 999 petabytes — should always exceed actual disk
	body, _ := json.Marshal(map[string]string{
		"key":   "storage_limit_bytes",
		"value": "999000000000000000",
	})
	r := makeRequest("PATCH", "/api/settings/update", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := do(UpdateSetting, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateSetting_UnknownKey(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	body, _ := json.Marshal(map[string]string{"key": "banana", "value": "yes"})
	r := makeRequest("PATCH", "/api/settings/update", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := do(UpdateSetting, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GetStorageUsage
// ---------------------------------------------------------------------------

func TestGetStorageUsage_NoLimit(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	writeFile(t, filepath.Join(StorageRoot(), "a.txt"), "12345")

	r := makeRequest("GET", "/api/storage/usage", nil)
	w := do(GetStorageUsage, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)

	if result["usedBytes"].(float64) < 5 {
		t.Errorf("expected usedBytes >= 5, got %v", result["usedBytes"])
	}
	if result["limitBytes"].(float64) != 0 {
		t.Errorf("expected limitBytes 0 when no limit set")
	}
}

func TestGetStorageUsage_WithLimit(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	db.SetSetting("storage_limit_bytes", "1000")
	writeFile(t, filepath.Join(StorageRoot(), "b.txt"), "hello")

	r := makeRequest("GET", "/api/storage/usage", nil)
	w := do(GetStorageUsage, r)

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)

	pct := result["percent"].(float64)
	if pct <= 0 || pct > 100 {
		t.Errorf("expected percent in (0, 100], got %v", pct)
	}
}
