package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/store"
	"github.com/marathozin/notes-api-go/pkg/response"
)

// Env хранит зависимости, доступные всем хэндлерам.
// При росте проекта сюда добавляются логгер, конфиг, сервисный слой и т.д.
type Env struct {
	Notes store.NoteStore
}

// GetNotes   GET /notes
func (e *Env) GetNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := e.Notes.GetAll()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not retrieve notes")
		return
	}
	response.JSON(w, http.StatusOK, "notes", notes)
}

// GetNote   GET /notes/{id}
func (e *Env) GetNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	note, err := e.Notes.GetByID(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "note not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not retrieve note")
		return
	}
	response.JSON(w, http.StatusOK, "note", note)
}

// CreateNote   POST /notes
func (e *Env) CreateNote(w http.ResponseWriter, r *http.Request) {
	var input model.CreateNoteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Title == "" {
		response.Error(w, http.StatusUnprocessableEntity, "title is required")
		return
	}

	note, err := e.Notes.Create(input)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not create note")
		return
	}
	response.JSON(w, http.StatusCreated, "note", note)
}

// UpdateNote   PUT /notes/{id}
func (e *Env) UpdateNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input model.UpdateNoteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Title == "" {
		response.Error(w, http.StatusUnprocessableEntity, "title is required")
		return
	}

	note, err := e.Notes.Update(id, input)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "note not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not update note")
		return
	}
	response.JSON(w, http.StatusOK, "note", note)
}

// DeleteNote   DELETE /notes/{id}
func (e *Env) DeleteNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := e.Notes.Delete(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "note not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not delete note")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
