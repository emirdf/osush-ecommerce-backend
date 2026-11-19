package repository

import (
	"context"
	"fmt"
	"osush/internal/db"
	"osush/internal/log"
	"osush/internal/models"

	"github.com/bytedance/gopkg/util/logger"
)

type DeviceTokenRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewDeviceTokenRepository(db *db.LoggedDB, logger log.Logger) *DeviceTokenRepository {
	return &DeviceTokenRepository{
		db:     db,
		logger: logger,
	}
}

func (r *DeviceTokenRepository) CreateOrUpdateDeviceToken(
	ctx context.Context,
	deviceToken models.DeviceToken,
) error {

	query := `
		INSERT INTO device_tokens (
			client_id,
			device_token
		)
		VALUES ($1, $2)
		ON CONFLICT (device_token)
		DO UPDATE SET
			client_id = EXCLUDED.client_id,
			updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
	`

	_, err := r.db.ExecContext(ctx, query, deviceToken.ClientID, deviceToken.DeviceToken)
	if err != nil {
		return fmt.Errorf("failed to create or update device token: %w", err)
	}

	return nil
}

func (r *DeviceTokenRepository) GetAllDeviceTokens(
	ctx context.Context,
) (*[]string, error) {

	r.logger.Info("getting all device tokens")

	query := `
		SELECT device_token
		FROM device_tokens
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("failed to get device tokens",
			log.Error(err),
		)

		return nil, fmt.Errorf("failed to get device tokens: %w", err)
	}
	defer rows.Close()

	var deviceTokens []string

	for rows.Next() {
		var deviceToken string

		if err := rows.Scan(&deviceToken); err != nil {
			logger.Error("failed to scan device token",
				log.Error(err),
			)

			return nil, fmt.Errorf("failed to scan device token: %w",
				err,
			)
		}

		deviceTokens = append(deviceTokens, deviceToken)
	}

	if err := rows.Err(); err != nil {
		logger.Error(
			"failed while iterating device tokens",
			log.Error(err),
		)

		return nil, fmt.Errorf(
			"failed while iterating device tokens: %w",
			err,
		)
	}

	logger.Info(
		"device tokens retrieved successfully",
		log.Int("count", len(deviceTokens)),
	)

	return &deviceTokens, nil
}

// GetDeviceTokensByClientID returns every FCM token linked to the client
// (one client may be logged in on several phones).
func (r *DeviceTokenRepository) GetDeviceTokensByClientID(
	ctx context.Context,
	clientID int,
) ([]string, error) {
	query := `
		SELECT device_token
		FROM device_tokens
		WHERE client_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client device tokens: %w", err)
	}
	defer rows.Close()

	var deviceTokens []string
	for rows.Next() {
		var deviceToken string
		if err := rows.Scan(&deviceToken); err != nil {
			return nil, fmt.Errorf("failed to scan device token: %w", err)
		}
		deviceTokens = append(deviceTokens, deviceToken)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating device tokens: %w", err)
	}

	return deviceTokens, nil
}

// DeleteDeviceTokens removes tokens that FCM reported as no longer valid
// (app uninstalled, token rotated), so they are not used again.
func (r *DeviceTokenRepository) DeleteDeviceTokens(
	ctx context.Context,
	deviceTokens []string,
) error {
	if len(deviceTokens) == 0 {
		return nil
	}

	query := `
		DELETE FROM device_tokens
		WHERE device_token = ANY($1)
	`

	if _, err := r.db.ExecContext(ctx, query, deviceTokens); err != nil {
		return fmt.Errorf("failed to delete device tokens: %w", err)
	}

	return nil
}
