package api

import (
	"encoding/json"
	"net/http"

	"backend/internal/middleware"
	"backend/internal/services"

	"github.com/google/uuid"
)

func NewFolderHandler(svc *services.FolderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.GetUser(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodPost: // Create
			var req struct {
				Name     string     `json:"name"`
				ParentID *uuid.UUID `json:"parent_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			f, err := svc.CreateFolder(user.ID, req.Name, req.ParentID)
			if err != nil {
				http.Error(w, "failed to create folder", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(f)

		case http.MethodGet: // List
			parentStr := r.URL.Query().Get("parent_id")
			var parentID *uuid.UUID
			if parentStr != "" {
				pid, err := uuid.Parse(parentStr)
				if err == nil {
					parentID = &pid
				}
			}
			folders, err := svc.ListUserFolders(user.ID, parentID)
			if err != nil {
				http.Error(w, "failed to list folders", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(folders)

		case http.MethodDelete: // Delete
			folderIDStr := r.URL.Query().Get("id")
			if folderIDStr == "" {
				http.Error(w, "missing folder id", http.StatusBadRequest)
				return
			}
			folderID, _ := uuid.Parse(folderIDStr)
			if err := svc.DeleteFolder(user.ID, folderID); err != nil {
				http.Error(w, "failed to delete folder", http.StatusForbidden)
				return
			}
			w.Write([]byte("folder deleted"))

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
