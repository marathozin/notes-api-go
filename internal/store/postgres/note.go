package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/store"
)

// Postgres-реализация store.NoteStore.
type NoteStore struct {
	db *pgxpool.Pool
}

func NewNoteStore(db *pgxpool.Pool) *NoteStore {
	return &NoteStore{db: db}
}

const noteSelectSQL = `
	SELECT n.id, n.title, n.content, n.user_id, n.created_at, n.updated_at
	FROM notes n`

// GetAll возвращает страницу заметок пользователя и общее количество его заметок.
func (s *NoteStore) GetAll(ctx context.Context, userID int64, pagination model.PaginationParams) ([]*model.Note, int, error) {
	var total int
	if err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notes WHERE user_id = $1`,
		userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.Limit
	q := noteSelectSQL + ` WHERE n.user_id = $1 GROUP BY n.id ORDER BY n.updated_at DESC LIMIT $2 OFFSET $3`
	rows, err := s.db.Query(ctx, q, userID, pagination.Limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	notes, err := scanNotes(rows)
	if err != nil {
		return nil, 0, err
	}
	return notes, total, nil
}

// GetByID возвращает заметку, если она принадлежит пользователю.
func (s *NoteStore) GetByID(ctx context.Context, id, userID int64) (*model.Note, error) {
	q := noteSelectSQL + ` WHERE n.id = $1 AND n.user_id = $2 GROUP BY n.id`
	rows, err := s.db.Query(ctx, q, id, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes, err := scanNotes(rows)
	if err != nil {
		return nil, err
	}
	if len(notes) == 0 {
		return nil, store.ErrNotFound
	}
	return notes[0], nil
}

// Create создаёт заметку.
func (s *NoteStore) Create(ctx context.Context, userID int64, input model.CreateNoteInput) (*model.Note, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var noteID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO notes (title, content, user_id) VALUES ($1, $2, $3) RETURNING id`,
		input.Title, input.Content, userID,
	).Scan(&noteID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, noteID, userID)
}

// Update обновляет заметку.
func (s *NoteStore) Update(ctx context.Context, id, userID int64, input model.UpdateNoteInput) (*model.Note, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	ct, err := tx.Exec(ctx,
		`UPDATE notes SET title=$1, content=$2 WHERE id=$3 AND user_id=$4`,
		input.Title, input.Content, id, userID,
	)
	if err != nil {
		return nil, err
	}
	if ct.RowsAffected() == 0 {
		return nil, store.ErrNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id, userID)
}

// Delete удаляет заметку.
func (s *NoteStore) Delete(ctx context.Context, id, userID int64) error {
	ct, err := s.db.Exec(ctx,
		`DELETE FROM notes WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

// scanNotes читает строки и десериализует JSON-агрегат тегов из Postgres.
func scanNotes(rows pgx.Rows) ([]*model.Note, error) {
	var notes []*model.Note
	for rows.Next() {
		var n model.Note
		if err := rows.Scan(
			&n.ID, &n.Title, &n.Content, &n.UserID,
			&n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		notes = append(notes, &n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if notes == nil {
		notes = []*model.Note{}
	}
	return notes, nil
}
