package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/marathozin/notes-api-go/internal/middleware"
	"github.com/marathozin/notes-api-go/internal/service"
)

// Токен-сервис с коротким TTL для тестов.
func TokenSvc() *service.TokenService {
	return service.NewTokenService("test-secret", time.Minute, time.Hour)
}

// NewRequest создаёт *http.Request с JSON-телом и опциональными заголовками.
func NewRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	r := httptest.NewRequest(method, path, &buf)
	r.Header.Set("Content-Type", "application/json")
	return r
}

// WithUserID кладёт userID в контекст запроса - имитирует прохождение Auth middleware.
func WithUserID(r *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	return r.WithContext(ctx)
}

// WithBearer добавляет Authorization: Bearer заголовок к запросу.
func WithBearer(r *http.Request, token string) *http.Request {
	r.Header.Set("Authorization", "Bearer "+token)
	return r
}

// Do выполняет хэндлер и возвращает записанный ответ.
func Do(handler http.HandlerFunc, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	handler(w, r)
	return w
}

// DecodeBody десериализует JSON-тело ответа в переданную структуру.
func DecodeBody(t *testing.T, w *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(dst); err != nil {
		t.Fatalf("decode response body: %v\nbody: %s", err, w.Body.String())
	}
}

// AssertStatus проверяет HTTP-статус ответа.
func AssertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status: got %d, want %d\nbody: %s", w.Code, want, w.Body.String())
	}
}

// AssertBodyContains проверяет, что тело ответа содержит подстроку.
func AssertBodyContains(t *testing.T, w *httptest.ResponseRecorder, substr string) {
	t.Helper()
	if body := w.Body.String(); !bytes.Contains([]byte(body), []byte(substr)) {
		t.Errorf("body %q does not contain %q", body, substr)
	}
}
