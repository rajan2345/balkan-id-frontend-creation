package api

import (
	"encoding/json"
	"net/http"

	"backend/internal/middleware"
	"backend/internal/services"
)

func NewStatsHandler(svc *services.StatsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.GetUser(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		stats, err := svc.GetUserStats(user.ID)
		if err != nil {
			http.Error(w, "error getting stats", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}
