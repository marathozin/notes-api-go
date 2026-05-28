package store

import (
	"fmt"

	"github.com/marathozin/notes-api-go/internal/model"
)

var (
	ErrNotFound  = fmt.Errorf("record not found")
	ErrDuplicate = fmt.Errorf("record already exists")
)

// Интерфейс хранилища пользователей.
type UserStore interface {
	Create(input model.RegisterInput) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
}

// Интерфейс хранилища заметок.
type NoteStore interface {
	GetAll(userID int64, pagination model.PaginationParams) ([]*model.Note, int, error)
	GetByID(id, userID int64) (*model.Note, error)
	Create(userID int64, input model.CreateNoteInput) (*model.Note, error)
	Update(id, userID int64, input model.UpdateNoteInput) (*model.Note, error)
	Delete(id, userID int64) error
}
