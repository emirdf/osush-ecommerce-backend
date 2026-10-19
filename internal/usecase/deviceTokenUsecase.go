package usecase

import (
	"context"
	"osush/internal/handlers/request"
	"osush/internal/log"
	"osush/internal/models"
	"osush/internal/repository"

	"github.com/bytedance/gopkg/util/logger"
)

type DeviceTokenUseCase struct {
	repo   *repository.DeviceTokenRepository
	logger log.Logger
}

func NewDeviceTokenUseCase(
	repo *repository.DeviceTokenRepository,
	logger log.Logger,
) *DeviceTokenUseCase {
	return &DeviceTokenUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (u *DeviceTokenUseCase) CreateDeviceToken(
	ctx context.Context,
	req request.DeviceTokenRequest,
	clientID int,
) error {

	u.logger.Info(
		"creating or updating device token",
		log.String("device_token", req.DeviceToken),
		log.Int("client_id", clientID),
	)

	var clientIDPtr *int

	if clientID == -1 {
		u.logger.Info(
			"anonymous client detected, saving client_id as NULL",
		)

		clientIDPtr = nil
	} else {
		u.logger.Info(
			"authenticated client detected",
			log.Int("client_id", clientID),
		)

		clientIDPtr = &clientID
	}

	deviceToken := models.DeviceToken{
		ClientID:    clientIDPtr,
		DeviceToken: req.DeviceToken,
	}

	if err := u.repo.CreateOrUpdateDeviceToken(
		ctx,
		deviceToken,
	); err != nil {
		logger.Error(
			"failed to create or update device token in repository",
			log.Error(err),
		)

		return err
	}

	logger.Info("device token created or updated successfully")

	return nil
}
