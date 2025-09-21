package api

import (
	"encoding/json"
	"net/http"

	"backend/internal/middleware"
	"backend/internal/services"

	"github.com/google/uuid"
)

type CreateShareRequest struct {
	FileID     string  `json:"file_id"`
	IsPublic   bool    `json:"is_public"`
	SharedWith *string `json:"shared_with,omitempty"`
}

// post/shares
func NewShareHandler(svc *services.ShareService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.GetUser(r.Context())

		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}

		var req CreateShareRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		fileID, err := uuid.Parse(req.FileID)
		if err != nil {
			http.Error(w, "invalid file ID", http.StatusBadRequest)
			return
		}

		share, err := svc.CreateShare(user.ID, fileID, req.IsPublic, req.SharedWith)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		w.Header().Set("Content-type", "application/json")
		json.NewEncoder(w).Encode(share)
	}
}

// Get shares/{id}
func NewGetShareHandler(svc *services.ShareService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shareIDstr := r.URL.Query().Get("id")
		if shareIDstr == "" {
			http.Error(w, "missing share id", http.StatusBadRequest)
			return
		}
		shareID, err := uuid.Parse(shareIDstr)
		if err != nil {
			http.Error(w, "invalid shrer id ", http.StatusBadRequest)
			return
		}
		share, err := svc.GetShare(shareID)
		if err != nil {
			http.Error(w, "share not found", http.StatusNotFound)
			return
		}

		// Increment download counter for files
		if share.IsPublic {
			_ = svc.IncrementDownloads(share.ID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(share)
	}
}
