package handler

import (
	"net/http"

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
	mux := http.NewServeMux()

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// public
	mux.HandleFunc("POST /auth/register", auth.Register)
	mux.HandleFunc("POST /auth/login", auth.Login)
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
	authMW := middleware.Auth(ts)
	mux.Handle("/auth/me", authMW(protected))
	mux.Handle("/notes", authMW(protected))

	// CORS -> RecoverPanic -> Logging
	return middleware.CORS(middleware.RecoverPanic(middleware.Logging(mux)))
}
