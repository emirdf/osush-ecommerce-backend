package repository

import (
	"context"
	"errors"
	"fmt"
	"osush/internal/db"
	"osush/internal/log"
	"osush/internal/models"
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/jackc/pgx/v4"
)

type NotificationRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewNotificationRepository(db *db.LoggedDB, logger log.Logger) *NotificationRepository {
	return &NotificationRepository{
		db:     db,
		logger: logger,
	}
}

func (r *NotificationRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	r.logger.Debug("checking by id exists",
		log.Int("id", id),
	)

	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1
            	FROM notifications
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

func (r *NotificationRepository) CreateNotification(
	ctx context.Context,
	req models.Notification,
) error {

	r.logger.Info("creating notification",
		log.Int("life", req.Life))

	tx, err := r.db.DB.Begin(ctx)
	if err != nil {
		r.logger.Error(
			"failed to begin notification transaction",
			log.Error(err),
		)

		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.logger.Error(
				"failed to rollback notification transaction",
				log.Error(err),
			)
		}
	}()

	titleDictionaryID, err := CreateDictionary(
		ctx,
		r.db,
		req.TitleTm,
		req.TitleRu,
	)
	if err != nil {
		r.logger.Error("failed to create notification title dictionary",
			log.Error(err),
		)

		return errors.New("failed to create notification title dictionary")
	}

	r.logger.Debug("notification title dictionary created",
		log.Int("dictionary_id", *titleDictionaryID),
	)
	req.TitleDictionaryID = *titleDictionaryID

	bodyDictionaryID, err := CreateDictionary(
		ctx,
		r.db,
		req.BodyTm,
		req.BodyRu,
	)
	if err != nil {
		r.logger.Error("failed to create notification body dictionary",
			log.Error(err),
		)

		return errors.New("failed to create notification body dictionary")
	}

	r.logger.Debug("notification body dictionary created",
		log.Int("dictionary_id", *bodyDictionaryID),
	)
	req.BodyDictionaryID = *bodyDictionaryID

	var notificationID int

	query := `
		INSERT INTO notifications (
			title_dictionary_id,
			body_dictionary_id,
			life
		)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err = tx.QueryRow(
		ctx,
		query,
		req.TitleDictionaryID,
		req.BodyDictionaryID,
		req.Life,
	).Scan(&notificationID)

	if err != nil {
		logger.Error(
			"failed to create notification",
			log.Error(err),
		)

		return fmt.Errorf("failed to create notification: %w", err)
	}

	logger.Info(
		"notification created",
		log.Int("notification_id", notificationID),
	)

	query = `
		INSERT INTO client_notifications (
			device_token_id,
			notification_id
		)
		SELECT
			id,
			$1
		FROM device_tokens
		ON CONFLICT (device_token_id, notification_id)
		DO NOTHING
	`

	_, err = tx.Exec(
		ctx,
		query,
		notificationID,
	)

	if err != nil {
		logger.Error(
			"failed to create client notifications",
			log.Int("notification_id", notificationID),
			log.Error(err),
		)

		return fmt.Errorf("failed to create client notifications: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error(
			"failed to commit notification transaction",
			log.Error(err),
		)

		return fmt.Errorf("failed to commit notification transaction: %w", err)
	}

	logger.Info(
		"notification created successfully",
		log.Int("notification_id", notificationID),
	)

	return nil
}

func (r *NotificationRepository) GetAllNotifications(
	ctx context.Context,
	deviceToken string,
) (*[]models.Notification, *int, *int, error) {

	r.logger.Info("getting all notifications")

	var (
		notifications      []models.Notification
		totalNotifications int
		totalIsNotRead     int
	)

	if deviceToken != "" {
		qTotal := `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (
				WHERE cn.is_read = FALSE
			)
		FROM client_notifications cn
		JOIN device_tokens dt
			ON dt.id = cn.device_token_id
		JOIN notifications n
			ON n.id = cn.notification_id
		WHERE dt.device_token = $1
		AND n.created_at + (n.life * INTERVAL '1 day') >= CURRENT_TIMESTAMP
	`

		err := r.db.QueryRowContext(
			ctx,
			qTotal,
			deviceToken,
		).Scan(
			&totalNotifications,
			&totalIsNotRead,
		)

		if err != nil {
			r.logger.Error(
				"failed to get notification count",
				log.Error(err),
			)

			return nil, nil, nil, errors.New(
				"failed to get notification count",
			)
		}

	} else {
		qTotal := `
		SELECT
			COUNT(*)
		FROM notifications 
		WHERE created_at + (life * INTERVAL '1 day') >= CURRENT_TIMESTAMP
	`

		err := r.db.QueryRowContext(
			ctx,
			qTotal,
		).Scan(
			&totalNotifications,
		)

		if err != nil {
			r.logger.Error(
				"failed to get notification count",
				log.Error(err),
			)

			return nil, nil, nil, errors.New(
				"failed to get notification count",
			)
		}
	}

	var (
		rows pgx.Rows
		err  error
	)

	if deviceToken != "" {
		query := `
		SELECT
			n.id,
			n.title_dictionary_id,
			n.body_dictionary_id,
			n.life,
			n.created_at,
			cn.is_read
		FROM client_notifications cn

		JOIN device_tokens dt
			ON dt.id = cn.device_token_id

		JOIN notifications n
			ON n.id = cn.notification_id

		WHERE dt.device_token = $1
		AND n.created_at + (n.life * INTERVAL '1 day') >= CURRENT_TIMESTAMP

		ORDER BY n.created_at DESC
	`

		rows, err = r.db.QueryContext(
			ctx,
			query,
			deviceToken,
		)

		if err != nil {
			r.logger.Error(
				"failed to get notifications",
				log.Error(err),
			)

			return nil, nil, nil, err
		}
	} else {
		query := `
		SELECT
			id,
			title_dictionary_id,
			body_dictionary_id,
			life,
			created_at,
			TRUE
		FROM notifications

		WHERE created_at + (life * INTERVAL '1 day') >= CURRENT_TIMESTAMP

		ORDER BY created_at DESC
	`

		rows, err = r.db.QueryContext(
			ctx,
			query,
		)

		if err != nil {
			r.logger.Error(
				"failed to get notifications",
				log.Error(err),
			)

			return nil, nil, nil, err
		}
	}

	defer rows.Close()

	for rows.Next() {

		var notification models.Notification

		err := rows.Scan(
			&notification.ID,
			&notification.TitleDictionaryID,
			&notification.BodyDictionaryID,
			&notification.Life,
			&notification.CreatedAt,
			&notification.IsRead,
		)

		if err != nil {
			r.logger.Error(
				"failed to scan notification",
				log.Error(err),
			)

			return nil, nil, nil, err
		}

		titleTm, titleRu, err := GetDictionaryByID(
			ctx,
			r.db,
			notification.TitleDictionaryID,
		)
		if err != nil {
			r.logger.Error(
				"failed to get notification title dictionary",
				log.Error(err),
			)

			return nil, nil, nil, errors.New(
				"failed to get notification title",
			)
		}

		notification.TitleTm = *titleTm
		notification.TitleRu = *titleRu

		bodyTm, bodyRu, err := GetDictionaryByID(
			ctx,
			r.db,
			notification.BodyDictionaryID,
		)
		if err != nil {
			r.logger.Error(
				"failed to get notification body dictionary",
				log.Error(err),
			)

			return nil, nil, nil, errors.New(
				"failed to get notification body",
			)
		}

		notification.BodyTm = *bodyTm
		notification.BodyRu = *bodyRu

		notifications = append(
			notifications,
			notification,
		)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(
			"failed while iterating notifications",
			log.Error(err),
		)

		return nil, nil, nil, err
	}

	r.logger.Info(
		"notifications retrieved successfully",
		log.Int("total_notifications", totalNotifications),
		log.Int("total_is_not_read", totalIsNotRead),
		log.Int("count", len(notifications)),
	)

	return &notifications, &totalNotifications, &totalIsNotRead, nil
}

func (r *NotificationRepository) MarkAsRead(
	ctx context.Context,
	clientID int,
	deviceToken string,
) error {
	r.logger.Info(
		"marking notification as read",
		log.Int("client_id", clientID),
	)

	var query string
	var args []interface{}

	if clientID == -1 && deviceToken != "" {
		query = `
			UPDATE client_notifications cn
			SET is_read = TRUE
			FROM device_tokens dt
			WHERE cn.device_token_id = dt.id
			  AND dt.device_token = $1
		`

		args = []interface{}{
			deviceToken,
		}
	} else if clientID != -1 {
		query = `
			UPDATE client_notifications cn
			SET is_read = TRUE
			FROM device_tokens dt
			WHERE cn.device_token_id = dt.id
			  AND dt.client_id = $1
		`

		args = []interface{}{
			clientID,
		}
	} else {
		r.logger.Info(
			"no client or device token provided, nothing to mark as read",
		)

		return nil
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error(
			"failed to mark notification as read",
			log.Error(err),
			log.Int("client_id", clientID),
		)

		return errors.New("failed to mark notification as read")
	}

	r.logger.Info(
		"notifications marked as read",
		log.Int("client_id", clientID),
	)

	return nil
}

func (r *NotificationRepository) GetAllNotificationsForAdmin(
	ctx context.Context,
	limit int,
	offset int,
) (*[]models.Notification, *int, error) {

	r.logger.Info(
		"getting all notifications for admin",
		log.Int("limit", limit),
		log.Int("offset", offset),
	)

	offset = (offset - 1) * limit

	var total int

	countQuery := `
		SELECT COUNT(*)
		FROM notifications n
		WHERE n.created_at + (n.life * INTERVAL '1 day') >= CURRENT_TIMESTAMP
	`

	err := r.db.QueryRowContext(
		ctx,
		countQuery,
	).Scan(&total)

	if err != nil {
		r.logger.Error(
			"failed to get notifications total",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get notifications total")
	}

	query := `
		SELECT
			n.id,
			n.title_dictionary_id,
			n.body_dictionary_id,
			n.life,
			n.created_at
		FROM notifications n
		WHERE n.created_at + (n.life * INTERVAL '1 day') >= CURRENT_TIMESTAMP
		ORDER BY n.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		limit,
		offset,
	)

	if err != nil {
		r.logger.Error(
			"failed to get notifications",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get notifications")
	}

	defer rows.Close()

	notifications := make([]models.Notification, 0)

	for rows.Next() {
		var notification models.Notification

		err := rows.Scan(
			&notification.ID,
			&notification.TitleDictionaryID,
			&notification.BodyDictionaryID,
			&notification.Life,
			&notification.CreatedAt,
		)

		if err != nil {
			r.logger.Error(
				"failed to scan notification",
				log.Error(err),
			)

			return nil, nil, errors.New("failed to scan notification")
		}

		titleTm, titleRu, err := GetDictionaryByID(
			ctx,
			r.db,
			notification.TitleDictionaryID,
		)
		if err != nil {
			r.logger.Error(
				"failed to get notification title",
				log.Error(err),
			)

			return nil, nil, err
		}

		notification.TitleTm = *titleTm
		notification.TitleRu = *titleRu

		bodyTm, bodyRu, err := GetDictionaryByID(
			ctx,
			r.db,
			notification.BodyDictionaryID,
		)
		if err != nil {
			r.logger.Error(
				"failed to get notification body",
				log.Error(err),
			)

			return nil, nil, err
		}

		notification.BodyTm = *bodyTm
		notification.BodyRu = *bodyRu

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(
			"error while iterating notifications",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get notifications")
	}

	r.logger.Info(
		"notifications successfully retrieved",
		log.Int("total", total),
	)

	return &notifications, &total, nil
}

func (r *NotificationRepository) DeleteNotificationByID(
	ctx context.Context,
	notificationID int,
) error {
	r.logger.Info(
		"deleting notification by id",
		log.Int("notification_id", notificationID),
	)

	query := `
		DELETE FROM notifications
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		notificationID,
	)
	if err != nil {
		r.logger.Error(
			"failed to delete notification",
			log.Error(err),
			log.Int("notification_id", notificationID),
		)

		return errors.New("failed to delete notification")
	}

	r.logger.Info(
		"notification successfully deleted",
		log.Int("notification_id", notificationID),
	)

	return nil
}

func (r *NotificationRepository) DeleteExpiredNotifications(
	ctx context.Context,
) error {
	r.logger.Info("deleting expired notifications from database")

	query := `
		DELETE FROM notifications
		WHERE created_at + (life * INTERVAL '1 day') < CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		r.logger.Error(
			"failed to delete expired notifications",
			log.Error(err),
		)

		return errors.New("failed to delete expired notifications")
	}

	r.logger.Info(
		"expired notifications deleted",
	)

	return nil
}
