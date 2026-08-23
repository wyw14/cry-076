package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestHealthAndSecurityHeaders(t *testing.T) {
	router := NewRouter(zap.NewNop(), []string{"http://localhost:5173"}, func() bool { return true }, Handlers{})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d", recorder.Code)
	}
	if recorder.Header().Get("X-Content-Type-Options") != "nosniff" || recorder.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("security headers missing: %#v", recorder.Header())
	}
	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("request id missing")
	}
}

func TestAPIRejectsMissingActorWithStableError(t *testing.T) {
	router := NewRouter(zap.NewNop(), nil, func() bool { return true }, Handlers{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", recorder.Code)
	}
	var payload struct {
		Error APIError `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error.Code != "UNAUTHENTICATED" || payload.Error.Message == "" || payload.Error.RequestID == "" {
		t.Fatalf("unexpected error: %#v", payload.Error)
	}
}
