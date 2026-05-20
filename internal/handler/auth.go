package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/marathozin/notes-api-go/internal/middleware"
	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/service"
	"github.com/marathozin/notes-api-go/internal/store"
	"github.com/marathozin/notes-api-go/pkg/response"
)

// AuthHandler обрабатывает запросы авторизации.
type AuthHandler struct {
	users  store.UserStore
	tokens *service.TokenService
}

func NewAuthHandler(users store.UserStore, tokens *service.TokenService) *AuthHandler {
	return &AuthHandler{users: users, tokens: tokens}
}

// Register регистрирует пользователя.
// @Summary Регистрация пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param input body model.RegisterInput true "Данные регистрации"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input model.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Email == "" || input.Username == "" || input.Password == "" {
		response.Error(w, http.StatusUnprocessableEntity, "email, username and password are required")
		return
	}
	if len(input.Password) < 8 {
		response.Error(w, http.StatusUnprocessableEntity, "password must be at least 8 characters")
		return
	}

	user, err := h.users.Create(input)
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			response.Error(w, http.StatusConflict, "email or username already taken")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		//response.Error(w, http.StatusInternalServerError, "could not create user")
		return
	}
	response.JSON(w, http.StatusCreated, "user", user)
}

// Login аутентифицирует пользователя.
// @Summary Вход пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param input body model.LoginInput true "Данные входа"
// @Success 200 {object} TokensResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input model.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.users.GetByEmail(input.Email)
	if err != nil {
		// Одинаковый ответ для "не найден" и "неверный пароль"
		response.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.Password)); err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if !user.IsActive {
		response.Error(w, http.StatusForbidden, "account is deactivated")
		return
	}

	access, refresh, err := h.tokens.GeneratePair(user.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not generate tokens")
		return
	}
	response.JSON(w, http.StatusOK, "tokens", model.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
	})
}

// Refresh обновляет пару токенов.
// @Summary Обновление access/refresh токенов
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RefreshInput true "Refresh токен"
// @Success 200 {object} TokensResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body RefreshInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		response.Error(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	userID, err := h.tokens.ValidateRefresh(body.RefreshToken)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	access, refresh, err := h.tokens.GeneratePair(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not generate tokens")
		return
	}
	response.JSON(w, http.StatusOK, "tokens", model.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
	})
}

// Me возвращает профиль текущего пользователя.
// @Summary Профиль текущего пользователя
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	user, err := h.users.GetByID(userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}
	response.JSON(w, http.StatusOK, "user", user)
}
