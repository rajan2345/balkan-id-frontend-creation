package api

import (
	"encoding/json"
	"net/http"

	"backend/internal/middleware"
	"backend/internal/services"
)

func NewAdminHandler(svc *services.AdminService) http.Handler {
	mux := http.NewServeMux()

	// GET /admin/users → list users
	mux.Handle("/users", middleware.AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		users, err := svc.ListUsers()
		if err != nil {
			http.Error(w, "failed to fetch users", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(users)
	})))

	// DELETE /admin/users?id=UUID → delete user
	mux.Handle("/users/delete", middleware.AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("id")
		if userID == "" {
			http.Error(w, "missing user id", http.StatusBadRequest)
			return
		}
		if err := svc.DeleteUser(userID); err != nil {
			http.Error(w, "failed to delete user", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("user deleted"))
	})))

	// GET /admin/stats → system storage stats
	mux.Handle("/stats", middleware.AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats, err := svc.GetSystemStats()
		if err != nil {
			http.Error(w, "failed to fetch stats", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(stats)
	})))

	return mux
}
