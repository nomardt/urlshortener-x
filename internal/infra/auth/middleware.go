package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/nomardt/urlshortener-x/internal/infra/logger"
	"go.uber.org/zap"
)

// This middleware checks if a jwt_session cookie is present and adds a new one if it's not
func WithAuth(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		authFn := func(w http.ResponseWriter, r *http.Request) {
			_, err := r.Cookie("jwt_session")
			if errors.Is(err, http.ErrNoCookie) {
				// Creating a new Cookie if not already present
				newToken, err := newJWT(secret)
				if err != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				cookie := &http.Cookie{
					Name:     "jwt_session",
					Value:    newToken,
					Path:     "/",
					Expires:  time.Now().Add(24 * time.Hour),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				}

				r.AddCookie(cookie)
				http.SetCookie(w, cookie)

				next.ServeHTTP(w, r)
				return
			} else if err != nil {
				// General error
				logger.Log.Info("Couldn't read user cookie", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// If a jwt_session cookie is already present then do nothing
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(authFn)
	}
}
