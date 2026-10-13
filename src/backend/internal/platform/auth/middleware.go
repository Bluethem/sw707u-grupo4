package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var ErrSessionNotFound = errors.New("sesion no encontrada")

type SessionLookup func(context.Context, string) (User, error)

func Authenticate(lookup SessionLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "no autorizado")
				return
			}
			user, err := lookup(r.Context(), token)
			if errors.Is(err, ErrSessionNotFound) {
				writeAuthError(w, http.StatusUnauthorized, "no autorizado")
				return
			}
			if err != nil {
				writeAuthError(w, http.StatusServiceUnavailable, "servicio no disponible")
				return
			}
			user.Token = token
			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "no autorizado")
				return
			}
			if _, ok := allowed[user.Rol]; !ok {
				writeAuthError(w, http.StatusForbidden, "prohibido")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(header string) (string, bool) {
	scheme, token, found := strings.Cut(strings.TrimSpace(header), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", false
	}
	token = strings.TrimSpace(token)
	if strings.ContainsAny(token, " \t\r\n") {
		return "", false
	}
	return token, true
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}
