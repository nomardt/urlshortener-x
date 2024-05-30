package auth

import (
	"errors"
	"net/http"
	"time"
)

// This middleware returns 401 Unauthorized if no jwt_session is provided
func TokenNecessary(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		tokenFn := func(w http.ResponseWriter, r *http.Request) {
			sessionCookie, err := r.Cookie("jwt_session")
			if errors.Is(err, http.ErrNoCookie) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if jwt_session is valid
			_, err = GetUserID(sessionCookie.Value, secret)
			if err != nil {
				http.Error(w, "Unathorized", http.StatusUnauthorized)
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(tokenFn)
	}
}

// This middleware checks if a jwt_session cookie is present and adds a new one if it's not
func WithAuth(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		authFn := func(w http.ResponseWriter, r *http.Request) {
			// Check if jwt_session exists
			sessionCookie, err := r.Cookie("jwt_session")
			if err != nil {
				// Creating a new Cookie
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
			}

			// Check if jwt_session is valid
			_, err = GetUserID(sessionCookie.Value, secret)
			if err != nil {
				// Creating a new Cookie
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
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(authFn)
	}
}
