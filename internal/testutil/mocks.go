// Пакет testutil содержит вспомогательные структуры для тестов:
// mock-реализации стора, хелперы для HTTP-запросов и проверки ответов.
package testutil

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/store"
)

// in-memory реализация store.UserStore для тестов.
type MockUserStore struct {
	mu      sync.RWMutex
	users   map[string]*model.User // ключ: email
	counter int64
}

func NewMockUserStore() *MockUserStore {
	return &MockUserStore{users: make(map[string]*model.User)}
}

func (s *MockUserStore) Create(input model.RegisterInput) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[input.Email]; exists {
		return nil, store.ErrDuplicate
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.MinCost)
	if err != nil {
		return nil, err
	}

	s.counter++
	u := &model.User{
		ID:             s.counter,
		Email:          input.Email,
		Username:       input.Username,
		HashedPassword: string(hash),
		IsActive:       true,
		CreatedAt:      time.Now(),
	}
	s.users[u.Email] = u
	return u, nil
}

func (s *MockUserStore) GetByEmail(email string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[email]
	if !ok {
		return nil, store.ErrNotFound
	}
	return u, nil
}

func (s *MockUserStore) GetByID(id int64) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, store.ErrNotFound
}

// in-memory реализация store.NoteStore для тестов.
type MockNoteStore struct {
	mu      sync.RWMutex
	notes   map[int64]*model.Note
	counter int64
}

func NewMockNoteStore() *MockNoteStore {
	return &MockNoteStore{notes: make(map[int64]*model.Note)}
}

func (s *MockNoteStore) GetAll(userID int64, pagination model.PaginationParams) ([]*model.Note, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Note
	for _, n := range s.notes {
		if n.UserID == userID {
			result = append(result, n)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].UpdatedAt.Equal(result[j].UpdatedAt) {
			return result[i].ID > result[j].ID
		}
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})

	total := len(result)
	start := (pagination.Page - 1) * pagination.Limit
	if start >= total {
		return []*model.Note{}, total, nil
	}
	end := start + pagination.Limit
	if end > total {
		end = total
	}
	return result[start:end], total, nil
}

func (s *MockNoteStore) GetByID(id, userID int64) (*model.Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n, ok := s.notes[id]
	if !ok || n.UserID != userID {
		return nil, store.ErrNotFound
	}
	return n, nil
}

func (s *MockNoteStore) Create(userID int64, input model.CreateNoteInput) (*model.Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	now := time.Now()
	n := &model.Note{
		ID:        s.counter,
		Title:     input.Title,
		Content:   input.Content,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.notes[n.ID] = n
	return n, nil
}

func (s *MockNoteStore) Update(id, userID int64, input model.UpdateNoteInput) (*model.Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n, ok := s.notes[id]
	if !ok || n.UserID != userID {
		return nil, store.ErrNotFound
	}

	n.Title = input.Title
	n.Content = input.Content
	n.UpdatedAt = time.Now()
	return n, nil
}

func (s *MockNoteStore) Delete(id, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	n, ok := s.notes[id]
	if !ok || n.UserID != userID {
		return store.ErrNotFound
	}
	delete(s.notes, id)
	return nil
}

// Seed добавляет заметку напрямую - удобно для подготовки состояния в тестах.
func (s *MockNoteStore) Seed(note *model.Note) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.notes[note.ID] = note
	if note.ID > s.counter {
		s.counter = note.ID
	}
}

// Ошибка, которую можно вернуть из мока для симуляции сбоя БД.
var ErrStore = fmt.Errorf("store error")

// Cтор, который всегда возвращает ошибку (для тестов 500).
type MockFailUserStore struct{}

func (s *MockFailUserStore) Create(_ model.RegisterInput) (*model.User, error) {
	return nil, ErrStore
}
func (s *MockFailUserStore) GetByEmail(_ string) (*model.User, error) { return nil, ErrStore }
func (s *MockFailUserStore) GetByID(_ int64) (*model.User, error)     { return nil, ErrStore }

// Стор, который всегда возвращает ошибку (для тестов 500).
type MockFailNoteStore struct{}

func (s *MockFailNoteStore) GetAll(_ int64, _ model.PaginationParams) ([]*model.Note, int, error) {
	return nil, 0, ErrStore
}
func (s *MockFailNoteStore) GetByID(_, _ int64) (*model.Note, error) {
	return nil, ErrStore
}
func (s *MockFailNoteStore) Create(_ int64, _ model.CreateNoteInput) (*model.Note, error) {
	return nil, ErrStore
}
func (s *MockFailNoteStore) Update(_, _ int64, _ model.UpdateNoteInput) (*model.Note, error) {
	return nil, ErrStore
}
func (s *MockFailNoteStore) Delete(_, _ int64) error { return ErrStore }
