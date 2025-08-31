package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/toluhikay/fx-exchange/internal/dtos"
	customError "github.com/toluhikay/fx-exchange/internal/errors"
	"github.com/toluhikay/fx-exchange/internal/models"
	"github.com/toluhikay/fx-exchange/internal/repository"
)

type UserServiceImpl struct {
	userRepo repository.UserDbRepo
}

func NewUserService(ur repository.UserDbRepo) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: ur,
	}
}

func (us *UserServiceImpl) CreateUser(ctx context.Context, req dtos.RegisterUser) (*models.User, error) {
	if req.Email == "" {
		return &models.User{}, errors.New("email required")
	}

	if len(req.Password) < 8 {
		return &models.User{}, customError.ErrPasswordTooShort
	}
	if len(req.Password) > 32 {
		return &models.User{}, customError.ErrPasswordTooLong
	}

	if req.Password != req.ConfirmPassword {
		return &models.User{}, customError.ErrPasswordMismatch

	}

	uId, err := uuid.NewV7()
	if err != nil {
		fmt.Println("err generating uuid")
		return &models.User{}, customError.ErrInternalServer
	}

	user := models.User{
		ID:    uId,
		Name:  req.Name,
		Email: req.Email,
	}
	hashedPassword, err := user.HashPassword(req.Password)
	if err != nil {
		fmt.Println("error hashing user password")
		return &models.User{}, customError.ErrInternalServer
	}

	user.Password = hashedPassword

	newUser, err := us.userRepo.CreateUser(ctx, user)

	if err != nil {
		fmt.Println(err, "error at create new user")
		if strings.Contains(err.Error(), "duplicate key value") {
			fmt.Println("error duplicate")
			return nil, customError.ErrDuplicateEmail
		}
		return nil, customError.ErrInternalServer
	}

	return newUser, nil
}

func (us *UserServiceImpl) GetUserById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if id.String() == "" {
		return &models.User{}, customError.ErrInvalidPayload
	}

	user, err := us.userRepo.GetUserById(ctx, id)
	if err != nil {
		fmt.Println(err, "here 1")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customError.ErrInvalidCredentials
		}
		return nil, customError.ErrInternalServer
	}

	return user, nil
}

func (us *UserServiceImpl) GetUserByEmail(ctx context.Context, req dtos.UserLogin) (*models.User, error) {
	if req.Email == "" {
		return &models.User{}, customError.ErrInvalidPayload
	}
	fmt.Println(req.Email)
	user, err := us.userRepo.GetUserByEmail(ctx, req.Email)

	if err != nil {
		fmt.Println(err, "here 1")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customError.ErrInvalidCredentials
		}
		return nil, customError.ErrInternalServer
	}

	return user, nil
}

func (us *UserServiceImpl) UserLogin(ctx context.Context, email, password string) (*models.User, error) {
	if email == "" {
		return nil, customError.ErrInvalidPayload
	}
	if password == "" {
		return nil, customError.ErrInvalidPayload
	}

	user, err := us.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		fmt.Println(err, "here 1")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customError.ErrInvalidCredentials
		}
		return nil, customError.ErrInternalServer
	}

	ok, err := user.CompareHashedPassword(password)
	if !ok {
		return nil, customError.ErrInvalidCredentials
	}
	if err != nil {
		return nil, customError.ErrInvalidCredentials
	}

	return user, nil

}
