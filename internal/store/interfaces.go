package store

import (
	"context"
	"fmt"

	"github.com/marathozin/notes-api-go/internal/model"
)

var (
	ErrNotFound  = fmt.Errorf("record not found")
	ErrDuplicate = fmt.Errorf("record already exists")
)

// Интерфейс хранилища пользователей.
type UserStore interface {
	Create(ctx context.Context, input model.RegisterInput) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
}

// Интерфейс хранилища заметок.
type NoteStore interface {
	GetAll(ctx context.Context, userID int64, pagination model.PaginationParams) ([]*model.Note, int, error)
	GetByID(ctx context.Context, id, userID int64) (*model.Note, error)
	Create(ctx context.Context, userID int64, input model.CreateNoteInput) (*model.Note, error)
	Update(ctx context.Context, id, userID int64, input model.UpdateNoteInput) (*model.Note, error)
	Delete(ctx context.Context, id, userID int64) error
}
