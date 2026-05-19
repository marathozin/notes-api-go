package model

import "time"

// Note - основная сущность приложения.
type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateNoteInput - входные данные для создания заметки.
type CreateNoteInput struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// UpdateNoteInput - входные данные для обновления заметки.
type UpdateNoteInput struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}
