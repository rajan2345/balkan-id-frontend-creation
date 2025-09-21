package middleware

import (
	// "context"
	// "io"
	"net/http"
	"strings"

	"backend/internal/db"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Configure this secret from config.go
type AuthOptions struct {
	JWTSecret string
	DB        *gorm.DB
}

// AuthMiddleware validates Authorization: Bearer<Token>
//Expects JWT's sub claim to have user's UUID string
// On success it loads the user row from DB and attaches it to the request context using WithUser
// if token invalid -> 401 error

func AuthMiddleware(opts AuthOptions) func(next http.Handler) http.Handler {
	secret := []byte(opts.JWTSecret)
	dbConn := opts.DB

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read authorization header
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}
			tokenStr := parts[1]
			// Parse token (use standard claims)
			tok, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
				// only allow hmac signing method
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrTokenUnverifiable
				}
				return secret, nil
			})
			if err != nil || !tok.Valid {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			//load user from DB using claims.Subject userID
			var user db.User
			claims, ok := tok.Claims.(*jwt.RegisteredClaims)
			if !ok || claims.Subject == "" {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
			if err := dbConn.First(&user, "id = ?", claims.Subject).Error; err != nil {
				http.Error(w, "user not found", http.StatusUnauthorized)
				return
			}

			//attach to the context
			ctx := WithUser(r.Context(), &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
