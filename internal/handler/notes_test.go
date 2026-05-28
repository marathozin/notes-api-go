package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/marathozin/notes-api-go/internal/handler"
	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/testutil"
)

const testUserID int64 = 1

func newNoteHandler() (*handler.NoteHandler, *testutil.MockNoteStore) {
	notes := testutil.NewMockNoteStore()
	return handler.NewNoteHandler(notes), notes
}

// seedNote добавляет заметку в стор и возвращает её.
func seedNote(s *testutil.MockNoteStore, userID int64, title, content string) *model.Note {
	n := &model.Note{
		ID:        1,
		Title:     title,
		Content:   content,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.Seed(n)
	return n
}

// doNote выполняет запрос к хэндлеру с userID в контексте и pathValue для {id}.
func doNote(h http.HandlerFunc, r *http.Request, userID int64) *httptest.ResponseRecorder {
	r = testutil.WithUserID(r, userID)
	return testutil.Do(h, r)
}

// setPathID добавляет {id} в PathValue запроса (Go 1.22 ServeMux).
func setPathID(r *http.Request, id int64) *http.Request {
	r.SetPathValue("id", fmt.Sprintf("%d", id))
	return r
}

// GetNotes

func TestGetNotes_Empty(t *testing.T) {
	h, _ := newNoteHandler()

	r := testutil.NewRequest(t, http.MethodGet, "/notes", nil)
	w := doNote(h.GetNotes, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusOK)
	testutil.AssertBodyContains(t, w, `"notes":[]`)
}

func TestGetNotes_ReturnsList(t *testing.T) {
	h, store := newNoteHandler()
	seedNote(store, testUserID, "First", "Body one")

	r := testutil.NewRequest(t, http.MethodGet, "/notes", nil)
	w := doNote(h.GetNotes, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusOK)
	testutil.AssertBodyContains(t, w, "First")
}

func TestGetNotes_OnlyOwnNotes(t *testing.T) {
	// Пользователь видит только свои заметки
	h, store := newNoteHandler()
	seedNote(store, testUserID, "My note", "My content")

	store.Seed(&model.Note{
		ID:      2,
		Title:   "Other note",
		Content: "Other content",
		UserID:  999, // другой пользователь
	})

	r := testutil.NewRequest(t, http.MethodGet, "/notes", nil)
	w := doNote(h.GetNotes, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusOK)
	testutil.AssertBodyContains(t, w, "My note")

	if strings.Contains(w.Body.String(), "Other note") {
		t.Errorf("response must not contain another user's note, got: %s", w.Body.String())
	}
}

func TestGetNotes_DefaultPagination(t *testing.T) {
	h, store := newNoteHandler()
	base := time.Date(2026, time.May, 28, 12, 0, 0, 0, time.UTC)
	for i := 1; i <= 25; i++ {
		store.Seed(&model.Note{
			ID:        int64(i),
			Title:     fmt.Sprintf("Note %02d", i),
			Content:   "Content",
			UserID:    testUserID,
			CreatedAt: base.Add(time.Duration(i) * time.Minute),
			UpdatedAt: base.Add(time.Duration(i) * time.Minute),
		})
	}

	r := testutil.NewRequest(t, http.MethodGet, "/notes", nil)
	w := doNote(h.GetNotes, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusOK)
	var body struct {
		Notes      []model.Note         `json:"notes"`
		Pagination model.PaginationMeta `json:"pagination"`
	}
	testutil.DecodeBody(t, w, &body)

	if len(body.Notes) != 20 {
		t.Fatalf("notes count: got %d, want 20", len(body.Notes))
	}
	if body.Notes[0].ID != 25 || body.Notes[19].ID != 6 {
		t.Fatalf("unexpected default page order: first id %d, last id %d", body.Notes[0].ID, body.Notes[19].ID)
	}
	wantMeta := model.PaginationMeta{Page: 1, Limit: 20, Total: 25, TotalPages: 2}
	if body.Pagination != wantMeta {
		t.Fatalf("pagination: got %+v, want %+v", body.Pagination, wantMeta)
	}
}

func TestGetNotes_CustomPagination(t *testing.T) {
	h, store := newNoteHandler()
	base := time.Date(2026, time.May, 28, 12, 0, 0, 0, time.UTC)
	for i := 1; i <= 5; i++ {
		store.Seed(&model.Note{
			ID:        int64(i),
			Title:     fmt.Sprintf("Note %02d", i),
			Content:   "Content",
			UserID:    testUserID,
			CreatedAt: base.Add(time.Duration(i) * time.Minute),
			UpdatedAt: base.Add(time.Duration(i) * time.Minute),
		})
	}

	r := testutil.NewRequest(t, http.MethodGet, "/notes?page=2&limit=2", nil)
	w := doNote(h.GetNotes, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusOK)
	var body struct {
		Notes      []model.Note         `json:"notes"`
		Pagination model.PaginationMeta `json:"pagination"`
	}
	testutil.DecodeBody(t, w, &body)

	if gotIDs := []int64{body.Notes[0].ID, body.Notes[1].ID}; gotIDs[0] != 3 || gotIDs[1] != 2 {
		t.Fatalf("page ids: got %v, want [3 2]", gotIDs)
	}
	wantMeta := model.PaginationMeta{Page: 2, Limit: 2, Total: 5, TotalPages: 3}
	if body.Pagination != wantMeta {
		t.Fatalf("pagination: got %+v, want %+v", body.Pagination, wantMeta)
	}
}

func TestGetNotes_PageOutOfRange(t *testing.T) {
	h, store := newNoteHandler()
	seedNote(store, testUserID, "Only note", "Body")

	r := testutil.NewRequest(t, http.MethodGet, "/notes?page=3&limit=2", nil)
	w := doNote(h.GetNotes, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusOK)
	var body struct {
		Notes      []model.Note         `json:"notes"`
		Pagination model.PaginationMeta `json:"pagination"`
	}
	testutil.DecodeBody(t, w, &body)

	if len(body.Notes) != 0 {
		t.Fatalf("notes count: got %d, want 0", len(body.Notes))
	}
	wantMeta := model.PaginationMeta{Page: 3, Limit: 2, Total: 1, TotalPages: 1}
	if body.Pagination != wantMeta {
		t.Fatalf("pagination: got %+v, want %+v", body.Pagination, wantMeta)
	}
}

func TestGetNotes_InvalidPagination(t *testing.T) {
	h, _ := newNoteHandler()

	cases := []string{
		"/notes?page=0",
		"/notes?page=abc",
		"/notes?limit=0",
		"/notes?limit=101",
		"/notes?limit=abc",
	}
	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			r := testutil.NewRequest(t, http.MethodGet, path, nil)
			w := doNote(h.GetNotes, r, testUserID)

			testutil.AssertStatus(t, w, http.StatusBadRequest)
			var body map[string]string
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("decode response body: %v", err)
			}
			if body["error"] == "" {
				t.Fatalf("expected error message, got %+v", body)
			}
		})
	}
}

func TestGetNotes_StoreError(t *testing.T) {
	h := handler.NewNoteHandler(&testutil.MockFailNoteStore{})

	r := testutil.NewRequest(t, http.MethodGet, "/notes", nil)
	w := doNote(h.GetNotes, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusInternalServerError)
}

// GetNote

func TestGetNote_Success(t *testing.T) {
	h, store := newNoteHandler()
	note := seedNote(store, testUserID, "My note", "My content")

	r := testutil.NewRequest(t, http.MethodGet, "/notes/1", nil)
	r = setPathID(r, note.ID)
	w := doNote(h.GetNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusOK)
	testutil.AssertBodyContains(t, w, "My note")
}

func TestGetNote_NotFound(t *testing.T) {
	h, _ := newNoteHandler()

	r := testutil.NewRequest(t, http.MethodGet, "/notes/99", nil)
	r = setPathID(r, 99)
	w := doNote(h.GetNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusNotFound)
}

func TestGetNote_AnotherUsersNote(t *testing.T) {
	// Нельзя получить чужую заметку.
	h, store := newNoteHandler()
	note := seedNote(store, 999, "Secret", "Hidden")

	r := testutil.NewRequest(t, http.MethodGet, "/notes/1", nil)
	r = setPathID(r, note.ID)
	w := doNote(h.GetNote, r, testUserID) // testUserID != 999

	testutil.AssertStatus(t, w, http.StatusNotFound)
}

func TestGetNote_InvalidID(t *testing.T) {
	h, _ := newNoteHandler()

	r := testutil.NewRequest(t, http.MethodGet, "/notes/abc", nil)
	r.SetPathValue("id", "abc")
	w := doNote(h.GetNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusBadRequest)
}

func TestGetNote_StoreError(t *testing.T) {
	h := handler.NewNoteHandler(&testutil.MockFailNoteStore{})

	r := testutil.NewRequest(t, http.MethodGet, "/notes/1", nil)
	r = setPathID(r, 1)
	w := doNote(h.GetNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusInternalServerError)
}

// CreateNote

func TestCreateNote_Success(t *testing.T) {
	h, _ := newNoteHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/notes", model.CreateNoteInput{
		Title:   "New note",
		Content: "Some content",
	})
	w := doNote(h.CreateNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusCreated)
	testutil.AssertBodyContains(t, w, "New note")
	testutil.AssertBodyContains(t, w, "Some content")
}

func TestCreateNote_MissingTitle(t *testing.T) {
	h, _ := newNoteHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/notes", model.CreateNoteInput{
		Content: "Some content",
	})
	w := doNote(h.CreateNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusUnprocessableEntity)
}

func TestCreateNote_MissingContent(t *testing.T) {
	h, _ := newNoteHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/notes", model.CreateNoteInput{
		Title: "A title",
	})
	w := doNote(h.CreateNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusUnprocessableEntity)
}

func TestCreateNote_InvalidJSON(t *testing.T) {
	h, _ := newNoteHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/notes", nil)
	r.Body = http.NoBody
	w := doNote(h.CreateNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusBadRequest)
}

func TestCreateNote_StoreError(t *testing.T) {
	h := handler.NewNoteHandler(&testutil.MockFailNoteStore{})

	r := testutil.NewRequest(t, http.MethodPost, "/notes", model.CreateNoteInput{
		Title:   "Title",
		Content: "Content",
	})
	w := doNote(h.CreateNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusInternalServerError)
}

// UpdateNote

func TestUpdateNote_Success(t *testing.T) {
	h, store := newNoteHandler()
	note := seedNote(store, testUserID, "Old title", "Old content")

	r := testutil.NewRequest(t, http.MethodPut, "/notes/1", model.UpdateNoteInput{
		Title:   "New title",
		Content: "New content",
	})
	r = setPathID(r, note.ID)
	w := doNote(h.UpdateNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusOK)
	testutil.AssertBodyContains(t, w, "New title")
	testutil.AssertBodyContains(t, w, "New content")
}

func TestUpdateNote_NotFound(t *testing.T) {
	h, _ := newNoteHandler()

	r := testutil.NewRequest(t, http.MethodPut, "/notes/99", model.UpdateNoteInput{
		Title:   "Title",
		Content: "Content",
	})
	r = setPathID(r, 99)
	w := doNote(h.UpdateNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusNotFound)
}

func TestUpdateNote_AnotherUsersNote(t *testing.T) {
	h, store := newNoteHandler()
	note := seedNote(store, 999, "Secret", "Private")

	r := testutil.NewRequest(t, http.MethodPut, "/notes/1", model.UpdateNoteInput{
		Title:   "Hacked",
		Content: "Hacked",
	})
	r = setPathID(r, note.ID)
	w := doNote(h.UpdateNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusNotFound)
}

func TestUpdateNote_MissingFields(t *testing.T) {
	h, store := newNoteHandler()
	note := seedNote(store, testUserID, "Title", "Content")

	cases := []model.UpdateNoteInput{
		{Content: "Content"}, // нет title
		{Title: "Title"},     // нет content
	}

	for _, input := range cases {
		r := testutil.NewRequest(t, http.MethodPut, "/notes/1", input)
		r = setPathID(r, note.ID)
		w := doNote(h.UpdateNote, r, testUserID)
		testutil.AssertStatus(t, w, http.StatusUnprocessableEntity)
	}
}

func TestUpdateNote_InvalidID(t *testing.T) {
	h, _ := newNoteHandler()

	r := testutil.NewRequest(t, http.MethodPut, "/notes/abc", model.UpdateNoteInput{
		Title:   "Title",
		Content: "Content",
	})
	r.SetPathValue("id", "abc")
	w := doNote(h.UpdateNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusBadRequest)
}

// DeleteNote

func TestDeleteNote_Success(t *testing.T) {
	h, store := newNoteHandler()
	note := seedNote(store, testUserID, "To delete", "Gone soon")

	r := testutil.NewRequest(t, http.MethodDelete, "/notes/1", nil)
	r = setPathID(r, note.ID)
	w := doNote(h.DeleteNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusNoContent)
}

func TestDeleteNote_NotFound(t *testing.T) {
	h, _ := newNoteHandler()

	r := testutil.NewRequest(t, http.MethodDelete, "/notes/99", nil)
	r = setPathID(r, 99)
	w := doNote(h.DeleteNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusNotFound)
}

func TestDeleteNote_AnotherUsersNote(t *testing.T) {
	h, store := newNoteHandler()
	note := seedNote(store, 999, "Secret", "Private")

	r := testutil.NewRequest(t, http.MethodDelete, "/notes/1", nil)
	r = setPathID(r, note.ID)
	w := doNote(h.DeleteNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusNotFound)
}

func TestDeleteNote_InvalidID(t *testing.T) {
	h, _ := newNoteHandler()

	r := testutil.NewRequest(t, http.MethodDelete, "/notes/xyz", nil)
	r.SetPathValue("id", "xyz")
	w := doNote(h.DeleteNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusBadRequest)
}

func TestDeleteNote_StoreError(t *testing.T) {
	h := handler.NewNoteHandler(&testutil.MockFailNoteStore{})

	r := testutil.NewRequest(t, http.MethodDelete, "/notes/1", nil)
	r = setPathID(r, 1)
	w := doNote(h.DeleteNote, r, testUserID)

	testutil.AssertStatus(t, w, http.StatusInternalServerError)
}
