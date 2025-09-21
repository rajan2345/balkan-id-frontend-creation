package middleware

import (
	"net/http"
)

// Admin only ensures the user is an admin
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if user == nil || !user.IsAdmin {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
