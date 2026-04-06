package src

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
)

func resolvePath(userPath string) (string, error) {
	if userPath == "" {
		userPath = "/"
	}

	clean := filepath.Clean("/" + userPath)
	full := filepath.Join(StorageRoot(), clean)

	if !strings.HasPrefix(full, filepath.Clean(StorageRoot())) {
		return "", http.ErrNotSupported
	}

	return full, nil
}

func resolveTrashPath(userPath string) (string, error) {
	trashBase := filepath.Join(StorageRoot(), ".trash")

	if userPath == "" {
		userPath = "/"
	}

	clean := filepath.Clean("/" + userPath)

	full := filepath.Join(trashBase, clean)

	if !strings.HasPrefix(full, filepath.Clean(trashBase)) {
		return "", http.ErrNotSupported
	}

	return full, nil
}

func writeJSON(w http.ResponseWriter, v any) error {
    w.Header().Set("Content-Type", "application/json")
    return json.NewEncoder(w).Encode(v)
}
