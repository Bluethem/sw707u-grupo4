package httpapi

import (
	"net/http"
	"strings"
)

const (
	corsAllowMethods = "GET, POST, PATCH, OPTIONS"
	corsAllowHeaders = "Authorization, Content-Type, Idempotency-Key"
)

func defaultCORSOrigins() []string {
	return []string{"http://localhost:3000"}
}

func withCORS(origins []string, next http.Handler) http.Handler {
	if len(origins) == 0 {
		origins = defaultCORSOrigins()
	}
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allowed[o] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)
				w.Header().Set("Access-Control-Allow-Headers", corsAllowHeaders)
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func parseCORSOriginsCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
