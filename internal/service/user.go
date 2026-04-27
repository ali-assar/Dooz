package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"dooz/entity"
	"dooz/internal/app/api/dto"
	"dooz/internal/pagination"
	userRepo "dooz/internal/repository/user"
	"dooz/utils/encrypt"

	"github.com/google/uuid"
)

type UserService interface {
	GetUserByID(ctx context.Context, userID string) (*dto.UserDTO, error)
	GetAllUsers(ctx context.Context, req *pagination.Request) (*pagination.Response[*dto.UserDTO], error)
	ChangePassword(ctx context.Context, userID string, newPassword string) error
	UpdateUser(ctx context.Context, userID string, req *dto.UpdateUserRequest, isAdmin bool) (*dto.UserDTO, error)
	SetOnline(ctx context.Context, userID string, online bool) error
}

type userService struct {
	userRepo userRepo.Repository
	logger   *slog.Logger
}

func NewUserService(userRepo userRepo.Repository, logger *slog.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger.With("layer", "UserService"),
	}
}

func (s *userService) GetUserByID(ctx context.Context, userID string) (*dto.UserDTO, error) {
	lg := s.logger.With("method", "GetUserByID", "userID", userID)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userRepo.ErrNotFound) {
			return nil, userRepo.ErrNotFound
		}
		lg.Error("failed to get user", "error", err)
		return nil, err
	}
	if user.DeletedAt > 0 {
		return nil, userRepo.ErrNotFound
	}
	return user.ToDTO(), nil
}

func (s *userService) GetAllUsers(ctx context.Context, req *pagination.Request) (*pagination.Response[*dto.UserDTO], error) {
	lg := s.logger.With("method", "GetAllUsers")

	queryParams, err := req.GetQueryParams()
	if err != nil {
		return nil, err
	}

	var cursorID *string
	if queryParams.HasCursor {
		cursorStr := queryParams.CursorID.String()
		cursorID = &cursorStr
	}

	users, err := s.userRepo.GetAllWithCursor(ctx, cursorID, queryParams.SortDesc, queryParams.Limit, req.Filters)
	if err != nil {
		lg.Error("failed to get users", "error", err)
		return nil, err
	}

	dtos := make([]*dto.UserDTO, len(users))
	for i, u := range users {
		dtos[i] = u.ToDTO()
	}

	return pagination.BuildResponse(dtos, req, func(u *dto.UserDTO) uuid.UUID {
		id, _ := uuid.Parse(u.ID)
		return id
	})
}

func (s *userService) ChangePassword(ctx context.Context, userID string, newPassword string) error {
	lg := s.logger.With("method", "ChangePassword", "userID", userID)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return userRepo.ErrNotFound
	}
	if user.DeletedAt > 0 {
		return userRepo.ErrNotFound
	}
	if user.PasswordHash == "" {
		return userRepo.ErrInvalidPassword
	}

	hashed := encrypt.HashSHA256(newPassword)
	if user.PasswordHash == hashed {
		return userRepo.ErrInvalidPassword
	}

	user.PasswordHash = hashed
	user.UpdatedAt = time.Now().Unix()

	if err := s.userRepo.Update(ctx, user); err != nil {
		lg.Error("failed to update password", "error", err)
		return err
	}
	return nil
}

func (s *userService) UpdateUser(ctx context.Context, userID string, req *dto.UpdateUserRequest, isAdmin bool) (*dto.UserDTO, error) {
	lg := s.logger.With("method", "UpdateUser", "userID", userID)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, userRepo.ErrNotFound
	}
	if user.DeletedAt > 0 {
		return nil, userRepo.ErrNotFound
	}

	updated := false

	if req.Fullname != "" && req.Fullname != user.Fullname {
		user.Fullname = req.Fullname
		updated = true
	}
	if req.Avatar != "" && req.Avatar != user.Avatar {
		user.Avatar = req.Avatar
		updated = true
	}

	if isAdmin {
		if req.Phone != "" && req.Phone != user.Phone {
			existing, err := s.userRepo.GetByPhone(ctx, req.Phone)
			if err == nil && existing != nil && existing.ID != userID {
				return nil, userRepo.ErrDuplicatePhone
			}
			user.Phone = req.Phone
			user.IsPhoneVerified = true
			updated = true
		}
		if req.Email != "" && req.Email != user.Email {
			existing, err := s.userRepo.GetByEmail(ctx, req.Email)
			if err == nil && existing != nil && existing.ID != userID {
				return nil, userRepo.ErrDuplicateEmail
			}
			user.Email = req.Email
			user.IsEmailVerified = true
			updated = true
		}
		if req.Role != "" {
			switch req.Role {
			case "user":
				user.Role = entity.RoleUser
			case "admin":
				user.Role = entity.RoleAdmin
			case "super_admin":
				user.Role = entity.RoleSuperAdmin
			default:
				return nil, userRepo.ErrInvalidRole
			}
			updated = true
		}
	}

	if updated {
		user.UpdatedAt = time.Now().Unix()
		if err := s.userRepo.Update(ctx, user); err != nil {
			lg.Error("failed to update user", "error", err)
			return nil, err
		}
	}

	return user.ToDTO(), nil
}

func (s *userService) SetOnline(ctx context.Context, userID string, online bool) error {
	return s.userRepo.SetOnline(ctx, userID, online)
}
