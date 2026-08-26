package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestUnauthorizedResponseKeepsIncomingRequestID(t *testing.T) {
	router := NewRouter(zap.NewNop(), nil, func() bool { return true }, Handlers{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
	request.Header.Set("X-Request-ID", "support-case-076")
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
	if recorder.Header().Get("X-Request-ID") != "support-case-076" {
		t.Fatalf("response header request id=%q", recorder.Header().Get("X-Request-ID"))
	}
	if payload.Error.RequestID != "support-case-076" {
		t.Fatalf("error body request id=%q, want incoming id", payload.Error.RequestID)
	}
}

