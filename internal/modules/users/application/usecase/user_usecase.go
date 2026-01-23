package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/domain/repository"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/auth"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
)

type UserUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) Register(ctx context.Context, input dto.CreateUserRequest) (*dto.UserResponse, error) {
	exists, err := uc.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("erro verif. email: %w", err)
	}
	if exists != nil {
		return nil, errors.New("email já cadastrado")
	}

	user, err := entity.NewUser(input.OrganizationID, input.Name, input.Email, input.Password, input.Role)
	if err != nil {
		return nil, err
	}

	if input.StoreID != nil {
		user.StoreID = input.StoreID
	}
	if input.Phone != "" {
		user.Phone = input.Phone
	}
	if input.Timezone != "" {
		user.Timezone = input.Timezone
	}
	if input.Language != "" {
		user.Language = input.Language
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID: user.ID, OrganizationID: user.OrganizationID, StoreID: user.StoreID,
		Name: user.Name, Email: user.Email, Role: user.Role, Status: user.Status,
		Timezone: user.Timezone, Language: user.Language,
	}, nil
}

func (uc *UserUseCase) Login(ctx context.Context, input dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := uc.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("credenciais inválidas")
	}

	if !user.CheckPassword(input.Password) {
		return nil, errors.New("credenciais inválidas")
	}
	if user.Status != entity.StatusActive {
		return nil, errors.New("usuário inativo")
	}

	token, err := auth.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("erro token: %w", err)
	}

	cfg := config.Get()
	return &dto.LoginResponse{
		AccessToken: token, ExpiresIn: int(cfg.JWTExpiration.Seconds()), TokenType: "Bearer",
		User: dto.UserResponse{
			ID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role,
			AvatarURL: user.AvatarURL, Status: user.Status,
		},
	}, nil
}
