package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/marathozin/notes-api-go/internal/model"
	"github.com/marathozin/notes-api-go/internal/store"
)

var (
	ErrDuplicate          = errors.New("resource already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveAccount    = errors.New("inactive account")
	ErrNotFound           = errors.New("resource not found")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type AuthService interface {
	Register(ctx context.Context, input model.RegisterInput) (*model.User, error)
	Login(ctx context.Context, input model.LoginInput) (model.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (model.TokenPair, error)
	Me(ctx context.Context, userID int64) (*model.User, error)
}

type authService struct {
	users  store.UserStore
	tokens TokenService
}

func NewAuthService(users store.UserStore, tokens TokenService) AuthService {
	return &authService{users: users, tokens: tokens}
}

func (s *authService) Register(ctx context.Context, input model.RegisterInput) (*model.User, error) {
	if input.Email == "" || input.Username == "" || input.Password == "" {
		return nil, ValidationError{Message: "email, username and password are required"}
	}
	if len(input.Password) < 8 {
		return nil, ValidationError{Message: "password must be at least 8 characters"}
	}

	user, err := s.users.Create(ctx, input)
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(ctx context.Context, input model.LoginInput) (model.TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return model.TokenPair{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.Password)); err != nil {
		return model.TokenPair{}, ErrInvalidCredentials
	}
	if !user.IsActive {
		return model.TokenPair{}, ErrInactiveAccount
	}

	return s.generatePair(user.ID)
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (model.TokenPair, error) {
	if refreshToken == "" {
		return model.TokenPair{}, ValidationError{Message: "refresh_token is required"}
	}

	userID, err := s.tokens.ValidateRefresh(refreshToken)
	if err != nil {
		return model.TokenPair{}, ErrInvalidCredentials
	}
	return s.generatePair(userID)
}

func (s *authService) Me(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *authService) generatePair(userID int64) (model.TokenPair, error) {
	access, refresh, err := s.tokens.GeneratePair(userID)
	if err != nil {
		return model.TokenPair{}, err
	}
	return model.TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}
