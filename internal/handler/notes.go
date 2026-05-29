package handler

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/marathozin/notes-api-go/internal/middleware"
	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/service"
	"github.com/marathozin/notes-api-go/pkg/response"
)

// NoteHandler обрабатывает CRUD-запросы заметок.
type NoteHandler struct {
	notes service.NoteService
}

func NewNoteHandler(notes service.NoteService) *NoteHandler {
	return &NoteHandler{notes: notes}
}

// GetNotes возвращает список заметок текущего пользователя.
// @Summary Список заметок
// @Tags notes
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество заметок на странице" default(20)
// @Success 200 {object} object{notes=[]model.Note,pagination=model.PaginationMeta}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /notes [get]
func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	pagination, ok := parsePagination(w, r)
	if !ok {
		return
	}

	notes, total, err := h.notes.List(r.Context(), userID, pagination)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not retrieve notes")
		return
	}

	meta := model.PaginationMeta{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(pagination.Limit))),
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"notes":      notes,
		"pagination": meta,
	})
}

// GetNote возвращает заметку по ID.
// @Summary Получить заметку
// @Tags notes
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заметки"
// @Success 200 {object} object{note=model.Note}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /notes/{id} [get]
func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	userID := middleware.GetUserID(r)

	note, err := h.notes.Get(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "note not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not retrieve note")
		return
	}
	response.JSON(w, http.StatusOK, "note", note)
}

// CreateNote создает новую заметку.
// @Summary Создать заметку
// @Tags notes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body model.CreateNoteInput true "Данные заметки"
// @Success 201 {object} object{note=model.Note}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 422 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /notes [post]
func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var input model.CreateNoteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	note, err := h.notes.Create(r.Context(), userID, input)
	if err != nil {
		var validationErr service.ValidationError
		if errors.As(err, &validationErr) {
			response.Error(w, http.StatusUnprocessableEntity, validationErr.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not create note")
		return
	}
	response.JSON(w, http.StatusCreated, "note", note)
}

// UpdateNote обновляет заметку по ID.
// @Summary Обновить заметку
// @Tags notes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заметки"
// @Param input body model.UpdateNoteInput true "Новые данные заметки"
// @Success 200 {object} object{note=model.Note}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Failure 422 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /notes/{id} [put]
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
	note, err := h.notes.Update(r.Context(), id, userID, input)
	if err != nil {
		var validationErr service.ValidationError
		if errors.As(err, &validationErr) {
			response.Error(w, http.StatusUnprocessableEntity, validationErr.Error())
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "note not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not update note")
		return
	}
	response.JSON(w, http.StatusOK, "note", note)
}

// DeleteNote удаляет заметку по ID.
// @Summary Удалить заметку
// @Tags notes
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заметки"
// @Success 204
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /notes/{id} [delete]
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	userID := middleware.GetUserID(r)

	if err := h.notes.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
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

const (
	defaultNotesPage  = 1
	defaultNotesLimit = 20
	maxNotesLimit     = 100
)

// parsePagination читает page и limit из query string и валидирует их.
func parsePagination(w http.ResponseWriter, r *http.Request) (model.PaginationParams, bool) {
	params := model.PaginationParams{Page: defaultNotesPage, Limit: defaultNotesLimit}

	if raw := r.URL.Query().Get("page"); raw != "" {
		page, err := strconv.Atoi(raw)
		if err != nil || page < 1 {
			response.Error(w, http.StatusBadRequest, "page must be a positive integer")
			return model.PaginationParams{}, false
		}
		params.Page = page
	}

	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > maxNotesLimit {
			response.Error(w, http.StatusBadRequest, "limit must be an integer between 1 and 100")
			return model.PaginationParams{}, false
		}
		params.Limit = limit
	}

	return params, true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
