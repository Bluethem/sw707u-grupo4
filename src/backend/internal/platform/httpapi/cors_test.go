package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithCORS_PreflightAllowedOrigin(t *testing.T) {
	h := withCORS([]string{"http://localhost:3000"}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("Allow-Origin=%q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != corsAllowMethods {
		t.Fatalf("Allow-Methods=%q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != corsAllowHeaders {
		t.Fatalf("Allow-Headers=%q", got)
	}
}

func TestWithCORS_PreflightUnknownOrigin(t *testing.T) {
	h := withCORS([]string{"http://localhost:3000"}, http.NotFoundHandler())
	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unexpected Allow-Origin for disallowed origin")
	}
}

func TestParseCORSOriginsCSV(t *testing.T) {
	got := parseCORSOriginsCSV("http://a.test, http://b.test ")
	if len(got) != 2 || got[0] != "http://a.test" || got[1] != "http://b.test" {
		t.Fatalf("parse=%v", got)
	}
}
