package handler_test

import (
	"net/http"
	"testing"

	"github.com/marathozin/notes-api-go/internal/handler"
	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/testutil"
)

func newAuthHandler() (*handler.AuthHandler, *testutil.MockUserStore) {
	users := testutil.NewMockUserStore()
	return handler.NewAuthHandler(users, &testutil.MockTokenService{}), users
}

// Register
func TestRegister_Success(t *testing.T) {
	h, _ := newAuthHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/auth/register", model.RegisterInput{
		Email:    "user@example.com",
		Username: "user",
		Password: "secret123",
	})
	w := testutil.Do(h.Register, r)

	testutil.AssertStatus(t, w, http.StatusCreated)
	testutil.AssertBodyContains(t, w, `"email":"user@example.com"`)
	testutil.AssertBodyContains(t, w, `"username":"user"`)
}

func TestRegister_MissingFields(t *testing.T) {
	h, _ := newAuthHandler()

	cases := []model.RegisterInput{
		{Email: "", Username: "u", Password: "secret123"},
		{Email: "u@example.com", Username: "", Password: "secret123"},
		{Email: "u@example.com", Username: "u", Password: ""},
	}

	for _, input := range cases {
		r := testutil.NewRequest(t, http.MethodPost, "/auth/register", input)
		w := testutil.Do(h.Register, r)
		testutil.AssertStatus(t, w, http.StatusUnprocessableEntity)
	}
}

func TestRegister_PasswordTooShort(t *testing.T) {
	h, _ := newAuthHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/auth/register", model.RegisterInput{
		Email:    "user@example.com",
		Username: "user",
		Password: "short",
	})
	w := testutil.Do(h.Register, r)

	testutil.AssertStatus(t, w, http.StatusUnprocessableEntity)
	testutil.AssertBodyContains(t, w, "8 characters")
}

func TestRegister_DuplicateEmail(t *testing.T) {
	h, _ := newAuthHandler()

	input := model.RegisterInput{
		Email:    "dup@example.com",
		Username: "user1",
		Password: "secret123",
	}
	testutil.Do(h.Register, testutil.NewRequest(t, http.MethodPost, "/auth/register", input))

	// Второй запрос с тем же email.
	input.Username = "user2"
	w := testutil.Do(h.Register, testutil.NewRequest(t, http.MethodPost, "/auth/register", input))

	testutil.AssertStatus(t, w, http.StatusConflict)
}

func TestRegister_InvalidJSON(t *testing.T) {
	h, _ := newAuthHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/auth/register", nil)
	r.Body = http.NoBody
	w := testutil.Do(h.Register, r)

	testutil.AssertStatus(t, w, http.StatusBadRequest)
}

func TestRegister_StoreError(t *testing.T) {
	h := handler.NewAuthHandler(&testutil.MockFailUserStore{}, &testutil.MockTokenService{})

	r := testutil.NewRequest(t, http.MethodPost, "/auth/register", model.RegisterInput{
		Email:    "user@example.com",
		Username: "user",
		Password: "secret123",
	})
	w := testutil.Do(h.Register, r)

	testutil.AssertStatus(t, w, http.StatusInternalServerError)
}

// Login

func TestLogin_Success(t *testing.T) {
	h, users := newAuthHandler()

	// Сначала создание пользователя.
	_, err := users.Create(model.RegisterInput{
		Email:    "login@example.com",
		Username: "loginuser",
		Password: "secret123",
	})
	if err != nil {
		t.Fatal(err)
	}

	r := testutil.NewRequest(t, http.MethodPost, "/auth/login", model.LoginInput{
		Email:    "login@example.com",
		Password: "secret123",
	})
	w := testutil.Do(h.Login, r)

	testutil.AssertStatus(t, w, http.StatusOK)
	testutil.AssertBodyContains(t, w, `"access_token":"access-token"`)
	testutil.AssertBodyContains(t, w, `"refresh_token":"refresh-token"`)
}

func TestLogin_WrongPassword(t *testing.T) {
	h, users := newAuthHandler()

	_, err := users.Create(model.RegisterInput{
		Email:    "user@example.com",
		Username: "user",
		Password: "secret123",
	})
	if err != nil {
		t.Fatal(err)
	}

	r := testutil.NewRequest(t, http.MethodPost, "/auth/login", model.LoginInput{
		Email:    "user@example.com",
		Password: "wrongpassword",
	})
	w := testutil.Do(h.Login, r)

	testutil.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestLogin_UserNotFound(t *testing.T) {
	h, _ := newAuthHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/auth/login", model.LoginInput{
		Email:    "nobody@example.com",
		Password: "secret123",
	})
	w := testutil.Do(h.Login, r)

	testutil.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestLogin_InvalidJSON(t *testing.T) {
	h, _ := newAuthHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/auth/login", nil)
	r.Body = http.NoBody
	w := testutil.Do(h.Login, r)

	testutil.AssertStatus(t, w, http.StatusBadRequest)
}

// Refresh

func TestRefresh_Success(t *testing.T) {
	h, _ := newAuthHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": "refresh-token",
	})
	w := testutil.Do(h.Refresh, r)

	testutil.AssertStatus(t, w, http.StatusOK)
	testutil.AssertBodyContains(t, w, "access_token")
}

func TestRefresh_InvalidToken(t *testing.T) {
	h, _ := newAuthHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": "not.a.valid.token",
	})
	w := testutil.Do(h.Refresh, r)

	testutil.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestRefresh_AccessTokenRejected(t *testing.T) {
	// Нельзя обновить сессию через access-токен - только через refresh.
	h, _ := newAuthHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": "access-token", // access вместо refresh
	})
	w := testutil.Do(h.Refresh, r)

	testutil.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestRefresh_MissingToken(t *testing.T) {
	h, _ := newAuthHandler()

	r := testutil.NewRequest(t, http.MethodPost, "/auth/refresh", map[string]string{})
	w := testutil.Do(h.Refresh, r)

	testutil.AssertStatus(t, w, http.StatusBadRequest)
}

// Me

func TestMe_Success(t *testing.T) {
	h, users := newAuthHandler()

	user, err := users.Create(model.RegisterInput{
		Email:    "me@example.com",
		Username: "meuser",
		Password: "secret123",
	})
	if err != nil {
		t.Fatal(err)
	}

	r := testutil.NewRequest(t, http.MethodGet, "/auth/me", nil)
	r = testutil.WithUserID(r, user.ID)
	w := testutil.Do(h.Me, r)

	testutil.AssertStatus(t, w, http.StatusOK)
	testutil.AssertBodyContains(t, w, `"email":"me@example.com"`)
}

func TestMe_UserNotFound(t *testing.T) {
	h, _ := newAuthHandler()

	r := testutil.NewRequest(t, http.MethodGet, "/auth/me", nil)
	r = testutil.WithUserID(r, 9999)
	w := testutil.Do(h.Me, r)

	testutil.AssertStatus(t, w, http.StatusNotFound)
}
