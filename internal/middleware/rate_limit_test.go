package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/marathozin/notes-api-go/internal/middleware"
)

// okHandler — простой хэндлер возвращающий 200.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// newRequest создаёт запрос с заданным IP через RemoteAddr.
func newRequest(ip string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = ip + ":12345"
	return r
}

// newRequestWithHeader создаёт запрос с IP в заголовке (имитация прокси).
func newRequestWithHeader(header, value string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "127.0.0.1:12345"
	r.Header.Set(header, value)
	return r
}

func do(handler http.Handler, r *http.Request) int {
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w.Code
}

// ---- IPRateLimiter.Allow ----

func TestAllow_FirstRequest(t *testing.T) {
	l := middleware.NewIPRateLimiter(rate.Every(time.Minute), 1, time.Minute)
	if !l.Allow("1.2.3.4") {
		t.Error("first request must be allowed")
	}
}

func TestAllow_ExceedsBurst(t *testing.T) {
	// burst=2: первые два запроса разрешены, третий — нет.
	l := middleware.NewIPRateLimiter(rate.Every(time.Hour), 2, time.Minute)

	if !l.Allow("1.2.3.4") {
		t.Error("request 1 must be allowed")
	}
	if !l.Allow("1.2.3.4") {
		t.Error("request 2 must be allowed")
	}
	if l.Allow("1.2.3.4") {
		t.Error("request 3 must be denied (burst exceeded)")
	}
}

func TestAllow_DifferentIPs(t *testing.T) {
	// Лимиты у разных IP независимы.
	l := middleware.NewIPRateLimiter(rate.Every(time.Hour), 1, time.Minute)

	l.Allow("1.1.1.1") // исчерпываем лимит для первого IP

	if !l.Allow("2.2.2.2") {
		t.Error("different IP must have its own independent limit")
	}
}

func TestAllow_SameIPTracked(t *testing.T) {
	l := middleware.NewIPRateLimiter(rate.Every(time.Minute), 10, time.Minute)
	l.Allow("1.2.3.4")
	l.Allow("5.6.7.8")

	if l.Len() != 2 {
		t.Errorf("expected 2 tracked IPs, got %d", l.Len())
	}
}

// ---- RateLimit middleware ----

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	l := middleware.NewIPRateLimiter(rate.Every(time.Minute), 5, time.Minute)
	handler := middleware.RateLimit(l)(okHandler)

	for i := 0; i < 5; i++ {
		code := do(handler, newRequest("1.2.3.4"))
		if code != http.StatusOK {
			t.Errorf("request %d: got %d, want 200", i+1, code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	// burst=3: четвёртый запрос должен получить 429.
	l := middleware.NewIPRateLimiter(rate.Every(time.Hour), 3, time.Minute)
	handler := middleware.RateLimit(l)(okHandler)

	for i := 0; i < 3; i++ {
		do(handler, newRequest("1.2.3.4"))
	}

	code := do(handler, newRequest("1.2.3.4"))
	if code != http.StatusTooManyRequests {
		t.Errorf("got %d, want 429", code)
	}
}

func TestRateLimit_RetryAfterHeader(t *testing.T) {
	l := middleware.NewIPRateLimiter(rate.Every(time.Hour), 1, time.Minute)
	handler := middleware.RateLimit(l)(okHandler)

	do(handler, newRequest("1.2.3.4")) // исчерпываем лимит

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest("1.2.3.4"))

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("got %d, want 429", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header must be set on 429")
	}
}

func TestRateLimit_ErrorBody(t *testing.T) {
	l := middleware.NewIPRateLimiter(rate.Every(time.Hour), 1, time.Minute)
	handler := middleware.RateLimit(l)(okHandler)

	do(handler, newRequest("1.2.3.4"))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest("1.2.3.4"))

	body := w.Body.String()
	if body == "" {
		t.Error("429 response must have a JSON body")
	}
}

func TestRateLimit_IsolatedPerIP(t *testing.T) {
	l := middleware.NewIPRateLimiter(rate.Every(time.Hour), 1, time.Minute)
	handler := middleware.RateLimit(l)(okHandler)

	// Исчерпываем лимит для первого IP.
	do(handler, newRequest("10.0.0.1"))
	if code := do(handler, newRequest("10.0.0.1")); code != http.StatusTooManyRequests {
		t.Errorf("10.0.0.1 second request: got %d, want 429", code)
	}

	// Второй IP не должен быть заблокирован.
	if code := do(handler, newRequest("10.0.0.2")); code != http.StatusOK {
		t.Errorf("10.0.0.2 first request: got %d, want 200", code)
	}
}

// ---- realIP извлечение ----

func TestRateLimit_XRealIP(t *testing.T) {
	// X-Real-IP должен использоваться вместо RemoteAddr.
	l := middleware.NewIPRateLimiter(rate.Every(time.Hour), 1, time.Minute)
	handler := middleware.RateLimit(l)(okHandler)

	// Первый запрос с X-Real-IP: 9.9.9.9 — разрешён.
	r1 := newRequestWithHeader("X-Real-IP", "9.9.9.9")
	if code := do(handler, r1); code != http.StatusOK {
		t.Errorf("first request: got %d, want 200", code)
	}

	// Второй запрос с тем же X-Real-IP — заблокирован.
	r2 := newRequestWithHeader("X-Real-IP", "9.9.9.9")
	if code := do(handler, r2); code != http.StatusTooManyRequests {
		t.Errorf("second request with same X-Real-IP: got %d, want 429", code)
	}

	// Запрос с другим X-Real-IP — разрешён (другой клиент).
	r3 := newRequestWithHeader("X-Real-IP", "8.8.8.8")
	if code := do(handler, r3); code != http.StatusOK {
		t.Errorf("request with different X-Real-IP: got %d, want 200", code)
	}
}

func TestRateLimit_XForwardedFor(t *testing.T) {
	// X-Forwarded-For: берётся первый IP в цепочке.
	l := middleware.NewIPRateLimiter(rate.Every(time.Hour), 1, time.Minute)
	handler := middleware.RateLimit(l)(okHandler)

	r1 := newRequestWithHeader("X-Forwarded-For", "5.5.5.5, 192.168.1.1, 10.0.0.1")
	do(handler, r1) // исчерпываем лимит для 5.5.5.5

	// Тот же клиент (первый IP совпадает) — заблокирован.
	r2 := newRequestWithHeader("X-Forwarded-For", "5.5.5.5, 172.16.0.1")
	if code := do(handler, r2); code != http.StatusTooManyRequests {
		t.Errorf("same client via proxy: got %d, want 429", code)
	}
}

func TestRateLimit_XRealIPTakesPrecedence(t *testing.T) {
	// X-Real-IP имеет приоритет над X-Forwarded-For.
	l := middleware.NewIPRateLimiter(rate.Every(time.Hour), 1, time.Minute)
	handler := middleware.RateLimit(l)(okHandler)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	r.Header.Set("X-Real-IP", "3.3.3.3")
	r.Header.Set("X-Forwarded-For", "4.4.4.4")

	do(handler, r) // исчерпываем лимит для 3.3.3.3

	// Второй запрос с тем же X-Real-IP — заблокирован (X-Real-IP взят первым).
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.RemoteAddr = "127.0.0.1:1234"
	r2.Header.Set("X-Real-IP", "3.3.3.3")

	if code := do(handler, r2); code != http.StatusTooManyRequests {
		t.Errorf("got %d, want 429 (X-Real-IP must take precedence)", code)
	}
}
