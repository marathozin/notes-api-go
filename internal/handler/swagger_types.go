package handler

import "github.com/marathozin/notes-api-go/internal/model"

// ErrorResponse описывает ошибку API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// UserResponse обертка для пользователя.
type UserResponse struct {
	User model.User `json:"user"`
}

// NotesResponse обертка для списка заметок.
type NotesResponse struct {
	Notes []model.Note `json:"notes"`
}

// NoteResponse обертка для одной заметки.
type NoteResponse struct {
	Note model.Note `json:"note"`
}

// TokensResponse обертка для пары токенов.
type TokensResponse struct {
	Tokens model.TokenPair `json:"tokens"`
}

// RefreshInput тело запроса на обновление токена.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}
