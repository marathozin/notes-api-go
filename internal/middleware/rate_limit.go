package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/marathozin/notes-api-go/pkg/response"
)

// entry хранит лимитер и время последнего обращения для очистки устаревших записей.
type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter хранит лимитеры для каждого IP-адреса.
type IPRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*entry
	r       rate.Limit // токенов в секунду
	burst   int        // максимальный всплеск
	ttl     time.Duration
}

// NewIPRateLimiter создаёт хранилище лимитеров.
//
//   - r - скорость пополнения (например, rate.Every(time.Minute) = 1/60 токена в секунду)
//   - burst - максимальное число запросов за один всплеск
//   - ttl - как долго хранить лимитер неактивного IP
func NewIPRateLimiter(r rate.Limit, burst int, ttl time.Duration) *IPRateLimiter {
	l := &IPRateLimiter{
		entries: make(map[string]*entry),
		r:       r,
		burst:   burst,
		ttl:     ttl,
	}
	go l.cleanupLoop()
	return l
}

// Allow проверяет лимит для IP и возвращает true если запрос разрешён.
func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[ip]
	if !ok {
		e = &entry{limiter: rate.NewLimiter(l.r, l.burst)}
		l.entries[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter.Allow()
}

// cleanupLoop периодически удаляет устаревшие записи чтобы не росла память.
func (l *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.ttl)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		for ip, e := range l.entries {
			if time.Since(e.lastSeen) > l.ttl {
				delete(l.entries, ip)
			}
		}
		l.mu.Unlock()
	}
}

// Len возвращает текущее число отслеживаемых IP - удобно для тестов.
func (l *IPRateLimiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}

// RateLimit возвращает middleware которое ограничивает запросы через переданный лимитер.
func RateLimit(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)
			if !limiter.Allow(ip) {
				w.Header().Set("Retry-After", "60")
				response.Error(w, http.StatusTooManyRequests, "too many requests, please slow down")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// realIP извлекает реальный IP клиента с учётом прокси-заголовков.
func realIP(r *http.Request) string {
	// X-Real-IP выставляется nginx/Render.
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// X-Forwarded-For может содержать цепочку: "client, proxy1, proxy2".
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// Берём самый левый — это оригинальный клиент.
		for i := 0; i < len(ip); i++ {
			if ip[i] == ',' {
				return ip[:i]
			}
		}
		return ip
	}
	// Fallback - прямое подключение.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
