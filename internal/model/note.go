package model

import "time"

// Заметка пользователя.
type Note struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Входные данные для создания заметки.
type CreateNoteInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Входные данные для обновления заметки.
type UpdateNoteInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// PaginationParams задаёт параметры постраничной выдачи.
type PaginationParams struct {
	Page  int
	Limit int
}

// PaginationMeta описывает состояние постраничной выдачи в ответе API.
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
