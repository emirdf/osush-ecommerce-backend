package repository

import (
	"context"
	"errors"
	"fmt"
	"osush/internal/db"
	customError "osush/internal/errors"
	"osush/internal/log"
	"osush/internal/models"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

type ClientRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewClientRepository(db *db.LoggedDB, logger log.Logger) *ClientRepository {
	return &ClientRepository{
		db:     db,
		logger: logger,
	}
}

// DeleteClient removes a client and everything that belongs only to them in
// one transaction. Orders are kept: the foreign key sets their client_id to
// NULL. It returns the client's avatar path (nil when there was none) so the
// caller can remove the file.
func (r *ClientRepository) DeleteClient(
	ctx context.Context,
	clientID int,
) (*string, error) {

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := r.db.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, errors.New("failed to begin transaction")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cleanup := []struct {
		name  string
		query string
	}{
		{"client notifications", `
			DELETE FROM client_notifications
			WHERE device_token_id IN (SELECT id FROM device_tokens WHERE client_id = $1)`},
		{"device tokens", `DELETE FROM device_tokens WHERE client_id = $1`},
		{"basket", `DELETE FROM basket WHERE client_id = $1`},
		{"product likes", `DELETE FROM product_likes WHERE client_id = $1`},
	}

	for _, step := range cleanup {
		if _, err := tx.Exec(ctx, step.query, clientID); err != nil {
			r.logger.Error("failed to delete client data",
				log.Int("client_id", clientID),
				log.String("table", step.name),
				log.Error(err),
			)

			return nil, errors.New("failed to delete client")
		}
	}

	var imagePath *string

	err = tx.QueryRow(ctx,
		`DELETE FROM clients WHERE id = $1 RETURNING image_path`,
		clientID,
	).Scan(&imagePath)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, customError.ErrNotFoundByID(clientID, "client")
	}

	if err != nil {
		r.logger.Error("failed to delete client",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, errors.New("failed to delete client")
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Error("failed to commit client deletion",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, errors.New("failed to delete client")
	}

	r.logger.Info("client deleted",
		log.Int("client_id", clientID),
	)

	return imagePath, nil
}

func (r *ClientRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	r.logger.Debug("checking by id exists",
		log.Int("id", id),
	)

	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1
            	FROM clients
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

func (r *ClientRepository) ExistsByPhoneNumber(
	ctx context.Context,
	phoneNumber string,
) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("checking client phone_number",
		log.String("phone_number", phoneNumber),
	)

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM clients
			WHERE phone_number = $1
		)
	`

	var exists bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		phoneNumber,
	).Scan(&exists)

	if err != nil {
		r.logger.Error("failed to check client phone_number",
			log.String("phone_number", phoneNumber),
			log.Error(err),
		)

		return false, errors.New("failed to check phone_number")
	}

	return exists, nil
}

func (r *ClientRepository) GetByPhoneNumber(
	ctx context.Context,
	phoneNumber string,
) (*models.Client, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting client by phone_number",
		log.String("phone_number", phoneNumber),
	)

	query := `
		SELECT
			id,
			user_name,
			surname,
			phone_number,
			address,
			image_path,
			password
		FROM clients
		WHERE phone_number = $1
	`

	var client models.Client

	err := r.db.QueryRowContext(
		ctx,
		query,
		phoneNumber,
	).Scan(
		&client.ID,
		&client.UserName,
		&client.Surname,
		&client.PhoneNumber,
		&client.Address,
		&client.ImagePath,
		&client.Password,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("client not found",
				log.String("phone_number", phoneNumber),
			)

			return nil, customError.ErrNotFound("client")
		}

		r.logger.Error("failed to get client by username",
			log.String("phone_number", phoneNumber),
			log.Error(err),
		)

		return nil, errors.New("failed to get client")
	}

	r.logger.Debug("client found",
		log.Int("client_id", client.ID),
	)

	return &client, nil
}

func (r *ClientRepository) GetByID(
	ctx context.Context,
	id int,
) (*models.Client, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting client by id",
		log.Int("client_id", id),
	)

	query := `
		SELECT
			id,
			user_name,
			surname,
			phone_number,
			address,
			image_path,
			password
		FROM clients
		WHERE id = $1
	`

	var client models.Client

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&client.ID,
		&client.UserName,
		&client.Surname,
		&client.PhoneNumber,
		&client.Address,
		&client.ImagePath,
		&client.Password,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("client not found",
				log.Int("client_id", id),
			)

			return nil, customError.ErrNotFound("client")
		}

		r.logger.Error("failed to get client by id",
			log.Int("client_id", id),
			log.Error(err),
		)

		return nil, errors.New("failed to get client")
	}

	r.logger.Debug("client found",
		log.Int("client_id", client.ID),
	)

	return &client, nil
}

func (r *ClientRepository) ExistsByPhoneNumberExceptID(
	ctx context.Context,
	phoneNumber string,
	clientID int,
) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("checking phone_number for another client",
		log.String("phone_number", phoneNumber),
		log.Int("client_id", clientID),
	)

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM clients
			WHERE phone_number = $1
			AND id <> $2
		)
	`

	var exists bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		phoneNumber,
		clientID,
	).Scan(&exists)

	if err != nil {
		r.logger.Error("failed to check phone_number",
			log.String("phone_number", phoneNumber),
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return false, errors.New("failed to check phone_number")
	}

	r.logger.Debug("phone_number check completed",
		log.Bool("exists", exists),
	)

	return exists, nil
}

func (r *ClientRepository) CreateClient(
	ctx context.Context,
	client models.Client,
) (*int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("creating client",
		log.String("user_name", client.UserName),
	)

	query := `
		INSERT INTO clients (
			user_name,
			surname,
			address,
			image_path,
			password,
			phone_number
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var clientID int

	err := r.db.QueryRowContext(
		ctx,
		query,
		client.UserName,
		client.Surname,
		client.Address,
		client.ImagePath,
		client.Password,
		client.PhoneNumber,
	).Scan(&clientID)

	if err != nil {
		r.logger.Error("failed to create client",
			log.String("user_name", client.UserName),
			log.Error(err),
		)

		return nil, errors.New("failed to create client")
	}

	r.logger.Info("client created successfully",
		log.Int("client_id", clientID),
	)

	return &clientID, nil
}

func (r *ClientRepository) UpdateProfile(
	ctx context.Context,
	client models.Client,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("updating client profile",
		log.Int("client_id", client.ID),
	)

	query := `
		UPDATE clients
		SET
			user_name = $1,
			surname = $2,
			address = $3,
			phone_number = $4,
			image_path = $5
		WHERE id = $6
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		client.UserName,
		client.Surname,
		client.Address,
		client.PhoneNumber,
		client.ImagePath,
		client.ID,
	)

	if err != nil {
		r.logger.Error("failed to update client profile",
			log.Int("client_id", client.ID),
			log.Error(err),
		)

		return errors.New("failed to update client profile")
	}

	r.logger.Info("client profile updated successfully",
		log.Int("client_id", client.ID),
	)

	return nil
}

func (r *ClientRepository) Logout(
	ctx context.Context,
	token string,
) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	r.logger.Debug("creating blacklist record")

	query := `
		INSERT INTO blacklist (
			token
		)
		VALUES ($1)
	`

	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		r.logger.Error("failed to create blacklist record",
			log.Error(err),
		)

		return errors.New("failed to create blacklist")
	}

	r.logger.Info("blacklist record created successfully")

	return nil
}

func (r *ClientRepository) CheckInBlacklist(
	ctx context.Context,
	token string,
) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	r.logger.Debug("checking whether token exists in blacklist")

	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1
			FROM blacklist
			WHERE token = $1
		)
	`

	err := r.db.QueryRowContext(ctx, query, token).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check token in blacklist",
			log.Error(err),
		)

		return false, errors.New("failed to check token existence")
	}

	if exists {
		r.logger.Warn("token is already blacklisted")
	} else {
		r.logger.Debug("token is not blacklisted")
	}

	r.logger.Debug("blacklist check completed",
		log.Bool("exists", exists),
	)

	return exists, nil
}

func (r *ClientRepository) GetClientList(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) (*[]models.Client, *int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug(
		"getting clients",
		log.String("search", search),
		log.Int("limit", limit),
		log.Int("offset", offset),
	)

	var (
		conditions []string
		args       []interface{}
		argIndex   = 1
	)

	if search != "" {
		conditions = append(
			conditions,
			fmt.Sprintf(`
				(
					LOWER(c.user_name) LIKE LOWER($%d)
					OR LOWER(c.surname) LIKE LOWER($%d)
					OR c.phone_number LIKE $%d
					OR LOWER(CONCAT(c.surname, ' ', c.user_name)) LIKE LOWER($%d)
				)
			`,
				argIndex,
				argIndex,
				argIndex,
				argIndex,
			),
		)

		args = append(args, "%"+search+"%")
		argIndex++
	}

	queryTotal := `
		SELECT COUNT(*)
		FROM clients c
	`

	if len(conditions) > 0 {
		queryTotal += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int

	if err := r.db.QueryRowContext(
		ctx,
		queryTotal,
		args...,
	).Scan(&total); err != nil {

		r.logger.Error(
			"failed to get clients total",
			log.Error(err),
		)

		return nil, nil, err
	}

	query := `
		SELECT
			c.id,
			c.user_name,
			c.surname,
			c.phone_number,
			c.address,
			c.image_path,
			c.created_at,
			c.updated_at
		FROM clients c
	`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += fmt.Sprintf(
		" ORDER BY c.created_at DESC LIMIT $%d OFFSET $%d",
		argIndex,
		argIndex+1,
	)

	offset = (offset - 1) * limit

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error(
			"failed to get clients",
			log.Error(err),
		)

		return nil, nil, err
	}
	defer rows.Close()

	clients := make([]models.Client, 0)

	for rows.Next() {
		var client models.Client

		if err := rows.Scan(
			&client.ID,
			&client.UserName,
			&client.Surname,
			&client.PhoneNumber,
			&client.Address,
			&client.ImagePath,
			&client.CreatedAt,
			&client.UpdatedAt,
		); err != nil {

			r.logger.Error(
				"failed to scan client",
				log.Error(err),
			)

			return nil, nil, err
		}

		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(
			"failed while iterating clients",
			log.Error(err),
		)

		return nil, nil, err
	}

	return &clients, &total, nil
}

func (r *ClientRepository) GetNameClients(
	ctx context.Context,
	limit int,
	offset int,
) (*[]string, *int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug(
		"getting clients",
		log.Int("limit", limit),
		log.Int("offset", offset),
	)

	queryTotal := `
		SELECT COUNT(*)
		FROM clients
	`

	var total int

	if err := r.db.QueryRowContext(ctx, queryTotal).Scan(&total); err != nil {

		r.logger.Error(
			"failed to get clients total",
			log.Error(err),
		)

		return nil, nil, err
	}

	query := `
		SELECT
			user_name,
			surname
		FROM clients
		LIMIT $1 OFFSET $2
	`

	offset = (offset - 1) * limit

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error(
			"failed to get clients",
			log.Error(err),
		)

		return nil, nil, err
	}
	defer rows.Close()

	var names []string

	for rows.Next() {
		var userName string
		var surname *string

		if err := rows.Scan(
			&userName,
			&surname,
		); err != nil {

			r.logger.Error(
				"failed to scan client",
				log.Error(err),
			)

			return nil, nil, err
		}

		if surname != nil {
			userName = fmt.Sprintf("%s %s", *surname, userName)
		}

		names = append(names, userName)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(
			"failed while iterating clients",
			log.Error(err),
		)

		return nil, nil, err
	}

	return &names, &total, nil
}
