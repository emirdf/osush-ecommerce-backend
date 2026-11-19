package usecase

import (
	"context"
	"errors"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/models"
	"osush/internal/repository"
	"time"
)

type FeedbackUseCase struct {
	repo   *repository.FeedbackRepository
	logger log.Logger
}

func NewFeedbackUseCase(
	repo *repository.FeedbackRepository,
	logger log.Logger,
) *FeedbackUseCase {
	return &FeedbackUseCase{
		repo:   repo,
		logger: logger,
	}
}

// CreateFeedback saves the feedback. clientID is -1 (or any value <= 0) for
// guests; the feedback is then saved without a client.
func (u *FeedbackUseCase) CreateFeedback(
	ctx context.Context,
	req request.FeedbackRequest,
	clientID int,
) error {

	u.logger.Info("starting create feedback",
		log.Int("client_id", clientID),
	)

	feedback := models.Feedback{
		Message: req.Message,
	}

	if clientID > 0 {
		feedback.ClientID = &clientID
	}

	if _, err := u.repo.CreateFeedback(ctx, feedback); err != nil {
		u.logger.Error("failed to create feedback",
			log.Error(err),
		)

		return err
	}

	return nil
}

func (u *FeedbackUseCase) GetFeedbackList(
	ctx context.Context,
	limit int,
	page int,
) (*response.FeedbackListResponse, error) {

	u.logger.Info("starting get feedback list",
		log.Int("limit", limit),
		log.Int("page", page),
	)

	feedbacks, total, err := u.repo.GetFeedbackList(ctx, limit, page)
	if err != nil {
		u.logger.Error("failed to get feedback list",
			log.Error(err),
		)

		return nil, err
	}

	result := &response.FeedbackListResponse{
		Total:     total,
		Feedbacks: make([]response.FeedbackResponse, 0, len(feedbacks)),
	}

	for _, feedback := range feedbacks {
		item := response.FeedbackResponse{
			ID:      feedback.ID,
			Message: feedback.Message,
			// Stored as UTC wall time; RFC3339 ("...Z") lets clients show local time.
			CreatedAt: feedback.CreatedAt.UTC().Format(time.RFC3339),
		}

		if feedback.ClientID != nil &&
			feedback.ClientUserName != nil &&
			feedback.ClientPhoneNumber != nil {
			item.Client = &response.FeedbackClientResponse{
				ID:          *feedback.ClientID,
				UserName:    *feedback.ClientUserName,
				Surname:     feedback.ClientSurname,
				PhoneNumber: *feedback.ClientPhoneNumber,
			}
		}

		result.Feedbacks = append(result.Feedbacks, item)
	}

	return result, nil
}

func (u *FeedbackUseCase) DeleteFeedbackByID(
	ctx context.Context,
	feedbackID int,
) error {

	u.logger.Info("starting delete feedback",
		log.Int("feedback_id", feedbackID),
	)

	if feedbackID <= 0 {
		return customError.ErrBadRequest("invalid feedback id")
	}

	err := u.repo.DeleteFeedbackByID(ctx, feedbackID)
	if errors.Is(err, repository.ErrFeedbackNotFound) {
		return customError.ErrNotFoundByID(feedbackID, "feedback")
	}

	if err != nil {
		u.logger.Error("failed to delete feedback",
			log.Int("feedback_id", feedbackID),
			log.Error(err),
		)

		return err
	}

	return nil
}
