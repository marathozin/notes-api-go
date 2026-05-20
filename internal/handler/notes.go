package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/marathozin/notes-api-go/internal/middleware"
	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/store"
	"github.com/marathozin/notes-api-go/pkg/response"
)

// NoteHandler обрабатывает CRUD-запросы заметок.
type NoteHandler struct {
	notes store.NoteStore
}

func NewNoteHandler(notes store.NoteStore) *NoteHandler {
	return &NoteHandler{notes: notes}
}

// GetNotes   GET /notes
func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	notes, err := h.notes.GetAll(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not retrieve notes")
		return
	}
	response.JSON(w, http.StatusOK, "notes", notes)
}

// GetNote   GET /notes/{id}
func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	userID := middleware.GetUserID(r)

	note, err := h.notes.GetByID(id, userID)
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
func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var input model.CreateNoteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Title == "" || input.Content == "" {
		response.Error(w, http.StatusUnprocessableEntity, "title and content are required")
		return
	}

	note, err := h.notes.Create(userID, input)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		//response.Error(w, http.StatusInternalServerError, "could not create note")
		return
	}
	response.JSON(w, http.StatusCreated, "note", note)
}

// UpdateNote   PUT /notes/{id}
func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	userID := middleware.GetUserID(r)

	var input model.UpdateNoteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Title == "" || input.Content == "" {
		response.Error(w, http.StatusUnprocessableEntity, "title and content are required")
		return
	}

	note, err := h.notes.Update(id, userID, input)
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
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	userID := middleware.GetUserID(r)

	if err := h.notes.Delete(id, userID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "note not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not delete note")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseID читает {id} из пути и валидирует как int64.
func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "id must be an integer")
		return 0, false
	}
	return id, true
}
