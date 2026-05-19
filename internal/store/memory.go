package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/marathozin/notes-api-go/internal/model"
)

// ErrNotFound возвращается, когда заметка не найдена.
var ErrNotFound = fmt.Errorf("note not found")

// NoteStore - интерфейс хранилища заметок.
// Позволяет в будущем заменить in-memory реализацию на БД без изменения хэндлеров.
type NoteStore interface {
	GetAll() ([]*model.Note, error)
	GetByID(id string) (*model.Note, error)
	Create(input model.CreateNoteInput) (*model.Note, error)
	Update(id string, input model.UpdateNoteInput) (*model.Note, error)
	Delete(id string) error
}

// InMemoryStore - потокобезопасное in-memory хранилище.
type InMemoryStore struct {
	mu      sync.RWMutex
	notes   map[string]*model.Note
	counter int
}

// NewInMemoryStore создаёт новое пустое хранилище.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		notes: make(map[string]*model.Note),
	}
}

func (s *InMemoryStore) GetAll() ([]*model.Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*model.Note, 0, len(s.notes))
	for _, n := range s.notes {
		list = append(list, n)
	}
	return list, nil
}

func (s *InMemoryStore) GetByID(id string) (*model.Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n, ok := s.notes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return n, nil
}

func (s *InMemoryStore) Create(input model.CreateNoteInput) (*model.Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	now := time.Now()
	n := &model.Note{
		ID:        fmt.Sprintf("%d", s.counter),
		Title:     input.Title,
		Body:      input.Body,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.notes[n.ID] = n
	return n, nil
}

func (s *InMemoryStore) Update(id string, input model.UpdateNoteInput) (*model.Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n, ok := s.notes[id]
	if !ok {
		return nil, ErrNotFound
	}

	n.Title = input.Title
	n.Body = input.Body
	n.UpdatedAt = time.Now()
	return n, nil
}

func (s *InMemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.notes[id]; !ok {
		return ErrNotFound
	}
	delete(s.notes, id)
	return nil
}
