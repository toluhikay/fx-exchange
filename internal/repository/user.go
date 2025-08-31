package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/toluhikay/fx-exchange/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u models.User) (*models.User, error)
}

type UserDbRepo struct {
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) *UserDbRepo {
	return &UserDbRepo{
		DB: db,
	}
}

func (m *UserDbRepo) CreateUser(ctx context.Context, u models.User) (*models.User, error) {
	stmt := `INSERT INTO users (name, email, password)
			VALUES ($1, $2, $3) RETURNING id, name, email`

	var newUser models.User

	err := m.DB.QueryRowContext(ctx, stmt,
		u.Name,
		u.Email,
		u.Password,
	).Scan(
		&newUser.ID,
		&newUser.Name,
		&newUser.Email,
	)

	if err != nil {
		return nil, err
	}

	return &newUser, nil

}

func (m *UserDbRepo) GetUserById(ctx context.Context, u uuid.UUID) (*models.User, error) {

	query := `
				SELECT * from users
				WHERE id = $1
				FOR UPDATE
	`
	var user models.User
	if err := m.DB.QueryRowContext(ctx, query, u).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *UserDbRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {

	query := `
				SELECT id, name, email, password,  created_at, updated_at from users
				WHERE email = $1
				FOR UPDATE
	`
	var user models.User
	if err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		fmt.Println(err, email)
		return nil, err
	}

	return &user, nil

}
