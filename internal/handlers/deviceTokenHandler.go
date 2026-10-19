package handlers

import (
	"net/http"
	"osush/internal/config"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/middleware"
	"osush/internal/usecase"

	"github.com/gin-gonic/gin"
)

type DeviceTokenHandler struct {
	deviceTokenUseCase *usecase.DeviceTokenUseCase
	logger             log.Logger
	cfg                *config.Config
}

func NewDeviceTokenHandler(cfg *config.Config, u *usecase.DeviceTokenUseCase, logger log.Logger) *DeviceTokenHandler {
	return &DeviceTokenHandler{
		deviceTokenUseCase: u,
		logger:             logger,
		cfg:                cfg,
	}
}

func (h *DeviceTokenHandler) CreateDeviceToken(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling create device token request")

	var req request.DeviceTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(
			"failed to bind device token request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info(
		"device token request received",
		log.String("device_token", req.DeviceToken),
	)

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info("DeviceToken request is valid")

	clientID := c.MustGet("clientId").(int)

	logger.Info(
		"client id received",
		log.Int("client_id", clientID),
	)

	if err := h.deviceTokenUseCase.CreateDeviceToken(
		ctx,
		req,
		clientID,
	); err != nil {
		logger.Error(
			"failed to create or update device token",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("device token created or updated successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message": "device token created successfully",
	})
}
