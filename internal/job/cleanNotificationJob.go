package job

import (
	"context"

	"osush/internal/log"
	"osush/internal/usecase"
)

type CleanNotificationJob struct {
	notificationUseCase *usecase.NotificationUseCase
	logger              log.Logger
}

func NewCleanNotificationJob(
	notificationUseCase *usecase.NotificationUseCase,
	logger log.Logger,
) *CleanNotificationJob {
	return &CleanNotificationJob{
		notificationUseCase: notificationUseCase,
		logger:              logger,
	}
}

func (j *CleanNotificationJob) Run() {
	j.logger.Info("starting clean notification job")

	err := j.notificationUseCase.DeleteExpiredNotifications(
		context.Background(),
	)
	if err != nil {
		j.logger.Error(
			"clean notification job failed",
			log.Error(err),
		)

		return
	}

	j.logger.Info("clean notification job completed")
}