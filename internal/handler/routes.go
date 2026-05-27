package handler

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"github.com/marathozin/notes-api-go/internal/middleware"
	"github.com/marathozin/notes-api-go/internal/service"

	_ "github.com/marathozin/notes-api-go/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter собирает маршруты и оборачивает их в middleware.
// Использует стандартный ServeMux из Go 1.22 с поддержкой path parameters.
func NewRouter(
	auth *AuthHandler,
	notes *NoteHandler,
	ts service.TokenService,
) http.Handler {
	// Глобальный лимитер: 100 запросов в минуту, всплеск до 20.
	globalLimiter := middleware.NewIPRateLimiter(
		rate.Every(time.Minute/100),
		20,
		5*time.Minute,
	)

	// Строгий лимитер для auth-эндпоинтов: 10 запросов в минуту, всплеск до 5.
	authLimiter := middleware.NewIPRateLimiter(
		rate.Every(time.Minute/10),
		5,
		10*time.Minute,
	)

	globalRL := middleware.RateLimit(globalLimiter)
	authRL := middleware.RateLimit(authLimiter)
	authMW := middleware.Auth(ts)

	mux := http.NewServeMux()

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// public
	mux.Handle("POST /auth/register", authRL(http.HandlerFunc(auth.Register)))
	mux.Handle("POST /auth/login", authRL(http.HandlerFunc(auth.Login)))
	mux.HandleFunc("POST /auth/refresh", auth.Refresh)

	// protected
	protected := http.NewServeMux()
	protected.HandleFunc("GET /auth/me", auth.Me)

	protected.HandleFunc("GET /notes", notes.GetNotes)
	protected.HandleFunc("POST /notes", notes.CreateNote)
	protected.HandleFunc("GET /notes/{id}", notes.GetNote)
	protected.HandleFunc("PUT /notes/{id}", notes.UpdateNote)
	protected.HandleFunc("DELETE /notes/{id}", notes.DeleteNote)

	// auth middleware
	mux.Handle("/auth/me", authMW(protected))
	mux.Handle("/notes", authMW(protected))
	mux.Handle("/notes/", authMW(protected))

	// Глобальные middleware: CORS -> RecoverPanic -> Logging -> глобальный rate limit.
	return middleware.CORS(middleware.RecoverPanic(middleware.Logging(globalRL(mux))))
}
