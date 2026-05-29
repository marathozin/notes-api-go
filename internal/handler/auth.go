package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/marathozin/notes-api-go/internal/middleware"
	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/service"
	"github.com/marathozin/notes-api-go/pkg/response"
)

// AuthHandler обрабатывает запросы авторизации.
type AuthHandler struct {
	auth service.AuthService
}

func NewAuthHandler(auth service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register регистрирует пользователя.
// @Summary Регистрация пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param input body model.RegisterInput true "Данные регистрации"
// @Success 201 {object} object{user=model.User}
// @Failure 400 {object} object{error=string}
// @Failure 409 {object} object{error=string}
// @Failure 422 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input model.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.auth.Register(r.Context(), input)
	if err != nil {
		var validationErr service.ValidationError
		if errors.As(err, &validationErr) {
			response.Error(w, http.StatusUnprocessableEntity, validationErr.Error())
			return
		}
		if errors.Is(err, service.ErrDuplicate) {
			response.Error(w, http.StatusConflict, "email or username already taken")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not create user")
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
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 403 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input model.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokens, err := h.auth.Login(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		if errors.Is(err, service.ErrInactiveAccount) {
			response.Error(w, http.StatusForbidden, "account is deactivated")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not generate tokens")
		return
	}
	response.JSON(w, http.StatusOK, "tokens", tokens)
}

// Refresh обновляет пару токенов.
// @Summary Обновление access/refresh токенов
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RefreshInput true "Refresh токен"
// @Success 200 {object} object{tokens=model.TokenPair}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body model.RefreshInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	tokens, err := h.auth.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		var validationErr service.ValidationError
		if errors.As(err, &validationErr) {
			response.Error(w, http.StatusBadRequest, validationErr.Error())
			return
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not generate tokens")
		return
	}
	response.JSON(w, http.StatusOK, "tokens", tokens)
}

// Me возвращает профиль текущего пользователя.
// @Summary Профиль текущего пользователя
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{user=model.User}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	user, err := h.auth.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not retrieve user")
		return
	}
	response.JSON(w, http.StatusOK, "user", user)
}
