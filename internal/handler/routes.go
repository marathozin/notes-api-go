package handler

import (
	"net/http"

	"github.com/marathozin/notes-api-go/internal/middleware"
)

// NewRouter собирает маршруты и оборачивает их в middleware.
// Использует стандартный ServeMux из Go 1.22 с поддержкой path parameters.
func NewRouter(env *Env) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /notes", env.GetNotes)
	mux.HandleFunc("POST /notes", env.CreateNote)
	mux.HandleFunc("GET /notes/{id}", env.GetNote)
	mux.HandleFunc("PUT /notes/{id}", env.UpdateNote)
	mux.HandleFunc("DELETE /notes/{id}", env.DeleteNote)

	// Middleware применяется снаружи внутрь: сначала RecoverPanic, потом Logging.
	return middleware.RecoverPanic(middleware.Logging(mux))
}
