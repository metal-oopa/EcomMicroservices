package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/metal-oopa/distributed-ecommerce/services/user-service/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID int) (*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING user_id
	`

	err := r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.PasswordHash).Scan(&user.UserID)
	return err
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT user_id, username, email, password_hash
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return user, err
}

func (r *userRepository) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	query := `
		SELECT user_id, username, email
		FROM users
		WHERE user_id = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return user, err
}
