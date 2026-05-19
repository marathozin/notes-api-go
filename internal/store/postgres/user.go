package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/store"
)

// Postgres-реализация store.UserStore.
type UserStore struct {
	db *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(input model.RegisterInput) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	const q = `
		INSERT INTO users (email, username, hashed_password)
		VALUES ($1, $2, $3)
		RETURNING id, email, username, hashed_password, is_active, created_at`

	var u model.User
	err = s.db.QueryRow(context.Background(), q,
		input.Email, input.Username, string(hash),
	).Scan(&u.ID, &u.Email, &u.Username, &u.HashedPassword, &u.IsActive, &u.CreatedAt)
	if err != nil {
		if isDuplicateError(err) {
			return nil, store.ErrDuplicate
		}
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) GetByEmail(email string) (*model.User, error) {
	const q = `
		SELECT id, email, username, hashed_password, is_active, created_at
		FROM users WHERE email = $1`

	var u model.User
	err := s.db.QueryRow(context.Background(), q, email).
		Scan(&u.ID, &u.Email, &u.Username, &u.HashedPassword, &u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) GetByID(id int64) (*model.User, error) {
	const q = `
		SELECT id, email, username, hashed_password, is_active, created_at
		FROM users WHERE id = $1`

	var u model.User
	err := s.db.QueryRow(context.Background(), q, id).
		Scan(&u.ID, &u.Email, &u.Username, &u.HashedPassword, &u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// isDuplicateError проверяет, является ли ошибка нарушением уникального ключа (код 23505).
func isDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
