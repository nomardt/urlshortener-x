package auth

import (
	"net/http"
	"strings"
)

// This middleware returns 401 Unauthorized if no Authorization is provided
func TokenNecessary(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		tokenFn := func(w http.ResponseWriter, r *http.Request) {
			sessionCookie := r.Header.Get("Authorization")
			if sessionCookie == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if jwt_session is valid
			sessionCookie, _ = strings.CutPrefix(sessionCookie, "Bearer ")
			_, err := GetUserID(sessionCookie, secret)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
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
			sessionCookie := r.Header.Get("Authorization")
			if sessionCookie == "" {
				// Creating a new Cookie
				newToken, err := newJWT(secret)
				if err != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				newToken = "Bearer " + newToken

				r.Header.Set("Authorization", newToken)
				w.Header().Add("Authorization", newToken)

				next.ServeHTTP(w, r)
				return
			}

			// Check if jwt_session is valid
			sessionCookie, _ = strings.CutPrefix(sessionCookie, "Bearer ")
			_, err := GetUserID(sessionCookie, secret)
			if err != nil {
				// Creating a new Cookie
				newToken, err := newJWT(secret)
				if err != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				newToken = "Bearer " + newToken

				r.Header.Add("Authorization", newToken)
				w.Header().Add("Authorization", newToken)

				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(authFn)
	}
}
