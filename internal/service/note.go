package service

import (
	"context"
	"errors"

	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/store"
)

type NoteService interface {
	List(ctx context.Context, userID int64, pagination model.PaginationParams) ([]*model.Note, int, error)
	Get(ctx context.Context, id, userID int64) (*model.Note, error)
	Create(ctx context.Context, userID int64, input model.CreateNoteInput) (*model.Note, error)
	Update(ctx context.Context, id, userID int64, input model.UpdateNoteInput) (*model.Note, error)
	Delete(ctx context.Context, id, userID int64) error
}

type noteService struct {
	notes store.NoteStore
}

func NewNoteService(notes store.NoteStore) NoteService {
	return &noteService{notes: notes}
}

func (s *noteService) List(ctx context.Context, userID int64, pagination model.PaginationParams) ([]*model.Note, int, error) {
	return s.notes.GetAll(ctx, userID, pagination)
}

func (s *noteService) Get(ctx context.Context, id, userID int64) (*model.Note, error) {
	note, err := s.notes.GetByID(ctx, id, userID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	return note, nil
}

func (s *noteService) Create(ctx context.Context, userID int64, input model.CreateNoteInput) (*model.Note, error) {
	if input.Title == "" || input.Content == "" {
		return nil, ValidationError{Message: "title and content are required"}
	}
	return s.notes.Create(ctx, userID, input)
}

func (s *noteService) Update(ctx context.Context, id, userID int64, input model.UpdateNoteInput) (*model.Note, error) {
	if input.Title == "" || input.Content == "" {
		return nil, ValidationError{Message: "title and content are required"}
	}
	note, err := s.notes.Update(ctx, id, userID, input)
	if err != nil {
		return nil, mapStoreError(err)
	}
	return note, nil
}

func (s *noteService) Delete(ctx context.Context, id, userID int64) error {
	if err := s.notes.Delete(ctx, id, userID); err != nil {
		return mapStoreError(err)
	}
	return nil
}

func mapStoreError(err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, store.ErrDuplicate) {
		return ErrDuplicate
	}
	return err
}
