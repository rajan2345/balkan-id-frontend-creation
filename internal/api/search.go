package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/middleware"
	"backend/internal/services"
)

func NewSearchHandler(svc *services.SearchService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.GetUser(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		query := r.URL.Query().Get("q")
		files, err := svc.SearchFiles(user.ID, query)
		if err != nil {
			http.Error(w, "error searching", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
	}
}

func NewFilterHandler(svc *services.SearchService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.GetUser(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var mime *string
		if v := r.URL.Query().Get("mime"); v != "" {
			mime = &v
		}
		var minSize, maxSize *int64
		if v := r.URL.Query().Get("min_size"); v != "" {
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				minSize = &parsed
			}
		}

		if v := r.URL.Query().Get("max_size"); v != "" {
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				maxSize = &parsed
			}
		}

		files, err := svc.FilterFiles(user.ID, mime, minSize, maxSize)
		if err != nil {
			http.Error(w, "error filtering", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
	}
}
