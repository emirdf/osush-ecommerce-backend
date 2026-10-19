package usecase

import (
	"context"
	"errors"
	customError "osush/internal/errors"
	"osush/internal/firebase"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/models"
	"osush/internal/repository"

	"github.com/bytedance/gopkg/util/logger"
)

type NotificationUseCase struct {
	repo            *repository.NotificationRepository
	deviceTokenRepo *repository.DeviceTokenRepository
	firebaseService *firebase.FirebaseService
	logger          log.Logger
}

func NewNotificationUseCase(
	repo *repository.NotificationRepository,
	deviceTokenRepo *repository.DeviceTokenRepository,
	firebaseService *firebase.FirebaseService,
	logger log.Logger,
) *NotificationUseCase {
	return &NotificationUseCase{
		repo:            repo,
		deviceTokenRepo: deviceTokenRepo,
		firebaseService: firebaseService,
		logger:          logger,
	}
}

func (u *NotificationUseCase) CreateNotification(
	ctx context.Context,
	req request.NotificationRequest,
) error {

	u.logger.Info("starting create notification")

	notification := models.Notification{
		TitleTm: req.Title.Tm,
		TitleRu: req.Title.Ru,
		BodyTm:  req.Title.Tm,
		BodyRu:  req.Title.Ru,
		Life:    req.Life,
	}

	err := u.repo.CreateNotification(
		ctx,
		notification,
	)
	if err != nil {
		logger.Error(
			"failed to create notification",
			log.Error(err),
		)

		return err
	}

	logger.Info("notification created")

	tokens, err := u.deviceTokenRepo.GetAllDeviceTokens(ctx)
	if err != nil {
		logger.Error("failed to get device tokens",
			log.Error(err))

		return err
	}

	logger.Info("device tokens received",
		log.Int("count", len(*tokens)),
	)

	if err := u.firebaseService.SendNotifications(
		ctx,
		*tokens,
		req.Title,
		req.Body,
	); err != nil {
		logger.Error(
			"failed to send firebase notification",
			log.Error(err),
		)
	}

	logger.Info("notification sent successfully")

	return nil
}

func (u *NotificationUseCase) GetAllNotifications(
	ctx context.Context,
	clientID int,
	deviceToken string,
) (*response.NotificationListResponse, error) {

	u.logger.Info(
		"getting all notifications",
		log.String("device_token", deviceToken),
	)

	notifications, totalNotifications, totalIsNotRead, err :=
		u.repo.GetAllNotifications(
			ctx,
			deviceToken,
		)

	if err != nil {
		u.logger.Error(
			"failed to get all notifications",
			log.Error(err),
		)

		return nil, err
	}

	notificationsResponse := make([]response.NotificationResponse, 0, len(*notifications))

	for _, notification := range *notifications {
		notificationsResponse = append(
			notificationsResponse,
			response.NotificationResponse{
				ID: notification.ID,
				Title: response.Dictionary{
					Tm: notification.TitleTm,
					Ru: notification.TitleRu,
				},
				Body: response.Dictionary{
					Tm: notification.BodyTm,
					Ru: notification.BodyRu,
				},
				IsNew:     !notification.IsRead,
				CreatedAt: notification.CreatedAt.Format("02-01-2006"),
			},
		)
	}

	result := &response.NotificationListResponse{
		TotalNotifications: *totalNotifications,
		TotalNew:           *totalIsNotRead,
		Notifications:      notificationsResponse,
	}

	u.logger.Info(
		"notifications successfully retrieved",
		log.Int("total_notifications", result.TotalNotifications),
		log.Int("total_new", result.TotalNew),
	)

	if result.TotalNew > 0 {
		err = u.repo.MarkAsRead(
			ctx,
			clientID,
			deviceToken,
		)
		if err != nil {
			u.logger.Error(
				"failed to mark notification as read",
				log.Error(err),
			)

			return nil, err
		}
	}

	return result, nil
}

func (u *NotificationUseCase) GetAllNotificationsForAdmin(
	ctx context.Context,
	limit int,
	offset int,
) (*response.NotificationListResponseForAdmin, error) {

	u.logger.Info(
		"getting all notifications for admin",
		log.Int("limit", limit),
		log.Int("offset", offset),
	)

	notifications, total, err := u.repo.GetAllNotificationsForAdmin(
		ctx,
		limit,
		offset,
	)
	if err != nil {
		u.logger.Error(
			"failed to get notifications for admin",
			log.Error(err),
		)

		return nil, err
	}

	notificationsResponse := make(
		[]response.NotificationResponseForAdmin,
		0,
		len(*notifications),
	)

	for _, notification := range *notifications {
		notificationsResponse = append(
			notificationsResponse,
			response.NotificationResponseForAdmin{
				ID: notification.ID,
				Title: response.Dictionary{
					Tm: notification.TitleTm,
					Ru: notification.TitleRu,
				},
				Body: response.Dictionary{
					Tm: notification.BodyTm,
					Ru: notification.BodyRu,
				},
				Life: notification.Life,
				CreatedAt: notification.CreatedAt.Format(
					"02-01-2006",
				),
			},
		)
	}

	return &response.NotificationListResponseForAdmin{
		Total:         *total,
		Notifications: notificationsResponse,
	}, nil
}

func (u *NotificationUseCase) DeleteNotificationByID(
	ctx context.Context,
	notificationID int,
) error {
	u.logger.Info(
		"deleting notification",
		log.Int("notification_id", notificationID),
	)

	if notificationID <= 0 {
		return errors.New("invalid notification id")
	}

	exists, err := u.repo.ExistsByID(ctx, notificationID)
	if err != nil {
		u.logger.Error("failed to check notification existence",
			log.Int("notification_id", notificationID),
			log.Error(err),
		)

		return err
	}

	if !exists {
		u.logger.Warn("notification not found",
			log.Int("notification_id", notificationID),
		)

		return customError.ErrNotFoundByID(notificationID, "notification")
	}

	err = u.repo.DeleteNotificationByID(
		ctx,
		notificationID,
	)
	if err != nil {
		u.logger.Error(
			"failed to delete notification",
			log.Error(err),
			log.Int("notification_id", notificationID),
		)

		return err
	}

	u.logger.Info(
		"notification deleted successfully",
		log.Int("notification_id", notificationID),
	)

	return nil
}

func (u *NotificationUseCase) DeleteExpiredNotifications(
	ctx context.Context,
) error {
	u.logger.Info("deleting expired notifications")

	err := u.repo.DeleteExpiredNotifications(ctx)
	if err != nil {
		u.logger.Error(
			"failed to delete expired notifications",
			log.Error(err),
		)

		return err
	}

	u.logger.Info("expired notifications deleted")

	return nil
}
