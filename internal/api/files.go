package api

import (
	"encoding/json"
	"net/http"

	// "balkanid-capstone/backend/internal/db"
	"backend/internal/services"

	"github.com/google/uuid"
)

// List user files
// GET `/files`
func ListUserFiles(w http.ResponseWriter, r *http.Request) {
	userIDstr := r.Header.Get("X-user-Id")
	if userIDstr == "" {
		http.Error(w, "Missing user ", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(userIDstr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	files, err := services.ListUserFiles(userID)
	if err != nil {
		http.Error(w, "Error fetching files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// DELETE /files/:id -> delete a file reference
func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	userIDstr := r.Header.Get("X-user-Id")
	if userIDstr == "" {
		http.Error(w, "Missing user: ", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(userIDstr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// using query params for now later will be using mux/path params
	fileIDstr := r.URL.Query().Get("id")
	if fileIDstr == "" {
		http.Error(w, "Missing file ID", http.StatusBadRequest)
		return
	}
	fileID, err := uuid.Parse(fileIDstr)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}
	err = services.DeleteUserFile(userID, fileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File deleted successufully"))
}
