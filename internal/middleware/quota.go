package middleware

import (
	"fmt"
	"net/http"

	"gorm.io/gorm"

	"backend/internal/db"
	// "context"
)

// Quota Middleware :
// - read user-Id header
// -loads user record from db and puts it into request context
// -do a pre-check using content-length (-1 for chunked/multipart transfer)
// -finally for quota , quota check + update is performed automatically in the db
func QuotaMiddleware(dbConn *gorm.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If auth middleware has already set the user, use it.
			if user := GetUser(r.Context()); user != nil {
				// pre-check content-length
				if r.ContentLength > 0 {
					if user.UsedStorage+r.ContentLength > user.Quota {
						msg := fmt.Sprintf("quota exceeded. Used: %d + request %d > Quota %d", user.UsedStorage, r.ContentLength, user.Quota)
						http.Error(w, msg, http.StatusForbidden)
						return
					}
				}
				next.ServeHTTP(w, r)
				return
			}

			// fallback: read header X-user-Id and load user (compatible with older clients)
			userID := r.Header.Get("X-user-Id")
			if userID == "" {
				http.Error(w, "Missing user", http.StatusUnauthorized)
				return
			}
			var user db.User
			if err := dbConn.First(&user, "id = ?", userID).Error; err != nil {
				http.Error(w, "user not found", http.StatusUnauthorized)
				return
			}

			// precheck user content length
			if r.ContentLength > 0 {
				if user.UsedStorage+r.ContentLength > user.Quota {
					msg := fmt.Sprintf("quota exceeded. Used: %d + request %d > Quota %d", user.UsedStorage, r.ContentLength, user.Quota)
					http.Error(w, msg, http.StatusForbidden)
					return
				}
			}

			//attach user to context for handlers or services to use
			ctx := WithUser(r.Context(), &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
