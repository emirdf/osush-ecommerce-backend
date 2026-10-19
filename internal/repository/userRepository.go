package repository

import (
	"context"
	"errors"
	"fmt"
	"osush/internal/db"
	customError "osush/internal/errors"
	"osush/internal/log"
	"osush/internal/models"
	"time"

	"github.com/jackc/pgx/v4"
)

type UserRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewUserRepository(db *db.LoggedDB, logger log.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	r.logger.Debug("checking by id exists",
		log.Int("id", id),
	)

	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1
            	FROM users
			WHERE id = $1
        )
    `
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check id existence", log.Error(err))
		return false, errors.New("failed to check id existence")
	}

	r.logger.Debug("id existence check completed",
		log.Bool("exists", exists),
	)
	return exists, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.Users, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("retrieving user by id",
		log.Int("id", id),
	)

	query := `
		SELECT 
			id, 
			name, 
			surname, 
			phone_number, 
			avatar, 
			password, 
			created_at, 
			updated_at
		FROM users
		WHERE id = $1
	`

	var user models.Users
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.PhoneNumber,
		&user.Avatar,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Warn("user not found",
				log.Int("id", id),
			)

			return nil, customError.ErrNotFoundByID(user.ID, "user")
		}

		r.logger.Error("failed to get user by id",
			log.Error(err))

		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	r.logger.Debug("user retrieved",
		log.Int("id", user.ID),
	)
	return &user, nil
}

func (r *UserRepository) GetByName(ctx context.Context, name string) (*models.Users, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug(
		"retrieving user by name",
		log.String("name", name),
	)

	query := `
		SELECT 
			id, 
			name, 
			surname, 
			phone_number, 
			avatar, 
			password, 
			created_at, 
			updated_at
		FROM users
		WHERE name = $1
	`

	var user models.Users

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.PhoneNumber,
		&user.Avatar,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Warn("user not found",
				log.String("name", name),
			)

			return nil, customError.ErrNotFound("user")
		}

		r.logger.Error(
			"failed to get user by name",
			log.String("name", name),
			log.Error(err),
		)

		return nil, fmt.Errorf(
			"failed to get user by name: %w",
			err,
		)
	}

	r.logger.Debug(
		"user retrieved",
		log.Int("id", user.ID),
	)

	return &user, nil
}

func (r *UserRepository) Update(
	ctx context.Context,
	adminID int,
	user *models.Users,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Info(
		"updating admin profile",
		log.Int("user_id", user.ID),
	)

	query := `
		UPDATE users
		SET
			name = $1,
			surname = $2,
			phone_number = $3,
			avatar = $4,
			password = $5,
			updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
		WHERE id = $6
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Name,
		user.Surname,
		user.PhoneNumber,
		user.Avatar,
		user.Password,
		user.ID,
	)

	if err != nil {
		r.logger.Error(
			"failed to update admin profile",
			log.Int("user_id", user.ID),
			log.Error(err),
		)

		return fmt.Errorf(
			"failed to update user: %w",
			err,
		)
	}

	r.logger.Info(
		"admin profile updated successfully",
		log.Int("user_id", user.ID),
	)

	return nil
}
