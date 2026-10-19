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
	"osush/internal/validation"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationUseCase *usecase.NotificationUseCase
	logger              log.Logger
	cfg                 *config.Config
}

func NewNotificationHandler(cfg *config.Config, u *usecase.NotificationUseCase, logger log.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationUseCase: u,
		logger:              logger,
		cfg:                 cfg,
	}
}

func (h *NotificationHandler) CreateNotification(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling create notification request")

	var req request.NotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		logger.Error("failed to bind notification request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})

		return
	}

	logger.Info("notification request received",
		log.Int("life", req.Life),
	)

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info("Notification request is valid")

	if err := h.notificationUseCase.CreateNotification(
		ctx,
		req,
	); err != nil {

		logger.Error(
			"failed to create notification",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info(
		"notification created successfully",
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "notification created successfully",
	})
}

func (h *NotificationHandler) GetAllNotifications(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling get all notifications request")

	clientID := c.MustGet("clientId").(int)

	deviceToken := c.Query("device_token")

	var err error

	result, err := h.notificationUseCase.GetAllNotifications(
		ctx,
		clientID,
		deviceToken,
	)
	if err != nil {
		logger.Error(
			"failed to get all notifications",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("notifications retrieved successfully")

	c.JSON(http.StatusOK, result)
}

func (h *NotificationHandler) GetAllNotificationsForAdmin(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling get all notifications for admin request")

	limit := c.Query("limit")
	offset := c.Query("offset")

	offsetInt, limitInt, err := validation.ValidatePagination(offset, limit)
	if err != nil {
		logger.Error("failed validate pagination",
			log.String("limit", limit),
			log.String("offset", offset),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	result, err := h.notificationUseCase.GetAllNotificationsForAdmin(
		ctx,
		*limitInt,
		*offsetInt,
	)
	if err != nil {
		logger.Error(
			"failed to get all notifications for admin",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info(
		"notifications successfully retrieved for admin",
	)

	c.JSON(http.StatusOK, result)
}

func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling delete notification request")

	id := c.Param("id")

	notificationID, err := strconv.Atoi(id)
	if err != nil {
		logger.Warn("invalid notification id",
			log.String("id_str", id),
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid notification_id",
		})
		return
	}

	err = h.notificationUseCase.DeleteNotificationByID(
		ctx,
		notificationID,
	)
	if err != nil {
		logger.Error(
			"failed to delete notification",
			log.Error(err),
			log.Int("notification_id", notificationID),
		)

		if err.Error() == "notification not found" {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "notification not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info(
		"notification deleted successfully",
		log.Int("notification_id", notificationID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "notification deleted successfully",
	})
}
