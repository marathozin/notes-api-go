package response

import (
	"encoding/json"
	"net/http"
)

// envelope оборачивает данные в именованный ключ - удобно для расширения ответа.
type envelope map[string]any

// JSON пишет успешный JSON-ответ с заданным HTTP-статусом.
func JSON(w http.ResponseWriter, status int, key string, data any) {
	writeJSON(w, status, envelope{key: data})
}

// Error пишет JSON-ответ с ошибкой.
func Error(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, envelope{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
