package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/dreego-stack/dreego/internal/session"
)

func CSRF(store session.Store) func(http.Handler) http.Handler {
	return CSRFWithForbidden(store, nil)
}

func CSRFWithForbidden(store session.Store, onForbidden http.HandlerFunc) func(http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := store.Get(r, "csrf_token")
			if err != nil {
				logger.Error("csrf token read failed", "error", err)
			}
			if token == "" {
				token = generateCSRFToken()
				if err := store.Set(w, r, "csrf_token", token, nil); err != nil {
					logger.Error("csrf token write failed", "error", err)
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
			}
			secure := isSecureForCSRF(r, store)
			http.SetCookie(w, &http.Cookie{
				Name:     "csrf_token",
				Value:    token,
				Path:     csrfCookiePath(store),
				HttpOnly: false,
				Secure:   secure,
				SameSite: http.SameSiteStrictMode,
			})

			if isUnsafeMethod(r.Method) {
				clientToken := r.Header.Get("X-CSRF-Token")
				if clientToken == "" {
					if err := r.ParseForm(); err != nil {
						if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
							http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
						} else {
							http.Error(w, "invalid form body", http.StatusBadRequest)
						}
						return
					}
					clientToken = r.FormValue("csrf_token")
				}
				if subtle.ConstantTimeCompare([]byte(clientToken), []byte(token)) != 1 {
					if onForbidden != nil {
						w.WriteHeader(http.StatusForbidden)
						onForbidden(w, r)
						return
					}
					http.Error(w, "invalid csrf token", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isSecureForCSRF(r *http.Request, store session.Store) bool {
	if r.TLS != nil {
		return true
	}
	if cs, ok := store.(*session.CookieStore); ok {
		return session.IsTLS(r, cs.TrustedProxies())
	}
	return false
}

func csrfCookiePath(store session.Store) string {
	if cs, ok := store.(*session.CookieStore); ok {
		if path := cs.CookiePath(); path != "" {
			return path
		}
	}
	return "/"
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("csrf: failed to read random bytes: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		return true
	}
	return false
}
