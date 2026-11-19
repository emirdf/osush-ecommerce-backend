package repository

import (
	"context"
	"errors"
	"osush/internal/db"
	"osush/internal/log"
	"osush/internal/models"
	"time"
)

var ErrFeedbackNotFound = errors.New("feedback not found")

type FeedbackRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewFeedbackRepository(db *db.LoggedDB, logger log.Logger) *FeedbackRepository {
	return &FeedbackRepository{
		db:     db,
		logger: logger,
	}
}

// CreateFeedback saves the feedback. A client id that does not belong to an
// existing client (a guest, an admin token, a just deleted account) is stored
// as NULL instead of failing on the foreign key.
func (r *FeedbackRepository) CreateFeedback(
	ctx context.Context,
	feedback models.Feedback,
) (*int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("creating feedback")

	query := `
		INSERT INTO feedbacks (
			client_id,
			message
		)
		VALUES (
			(SELECT id FROM clients WHERE id = $1),
			$2
		)
		RETURNING id
	`

	var feedbackID int

	err := r.db.QueryRowContext(
		ctx,
		query,
		feedback.ClientID,
		feedback.Message,
	).Scan(&feedbackID)

	if err != nil {
		r.logger.Error("failed to create feedback",
			log.Error(err),
		)

		return nil, errors.New("failed to create feedback")
	}

	r.logger.Info("feedback created successfully",
		log.Int("feedback_id", feedbackID),
	)

	return &feedbackID, nil
}

// GetFeedbackList returns one page of feedbacks, newest first. page starts at 1.
func (r *FeedbackRepository) GetFeedbackList(
	ctx context.Context,
	limit int,
	page int,
) ([]models.Feedback, int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting feedbacks",
		log.Int("limit", limit),
		log.Int("page", page),
	)

	offset := (page - 1) * limit

	var total int

	if err := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM feedbacks`,
	).Scan(&total); err != nil {
		r.logger.Error("failed to count feedbacks",
			log.Error(err),
		)

		return nil, 0, errors.New("failed to get feedbacks total")
	}

	query := `
		SELECT
			f.id,
			f.client_id,
			f.message,
			f.created_at,
			c.user_name,
			c.surname,
			c.phone_number
		FROM feedbacks f
		LEFT JOIN clients c
			ON c.id = f.client_id
		ORDER BY f.created_at DESC, f.id DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("failed to get feedbacks",
			log.Error(err),
		)

		return nil, 0, errors.New("failed to get feedbacks")
	}
	defer rows.Close()

	feedbacks := make([]models.Feedback, 0)

	for rows.Next() {
		var feedback models.Feedback

		if err := rows.Scan(
			&feedback.ID,
			&feedback.ClientID,
			&feedback.Message,
			&feedback.CreatedAt,
			&feedback.ClientUserName,
			&feedback.ClientSurname,
			&feedback.ClientPhoneNumber,
		); err != nil {
			r.logger.Error("failed to scan feedback",
				log.Error(err),
			)

			return nil, 0, errors.New("failed to scan feedback")
		}

		feedbacks = append(feedbacks, feedback)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("failed while iterating feedbacks",
			log.Error(err),
		)

		return nil, 0, errors.New("failed to get feedbacks")
	}

	r.logger.Info("feedbacks retrieved successfully",
		log.Int("total", total),
		log.Int("count", len(feedbacks)),
	)

	return feedbacks, total, nil
}

// DeleteFeedbackByID returns ErrFeedbackNotFound when nothing was deleted.
func (r *FeedbackRepository) DeleteFeedbackByID(
	ctx context.Context,
	feedbackID int,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("deleting feedback",
		log.Int("feedback_id", feedbackID),
	)

	tag, err := r.db.ExecContext(
		ctx,
		`DELETE FROM feedbacks WHERE id = $1`,
		feedbackID,
	)
	if err != nil {
		r.logger.Error("failed to delete feedback",
			log.Int("feedback_id", feedbackID),
			log.Error(err),
		)

		return errors.New("failed to delete feedback")
	}

	if tag.RowsAffected() == 0 {
		return ErrFeedbackNotFound
	}

	r.logger.Info("feedback deleted successfully",
		log.Int("feedback_id", feedbackID),
	)

	return nil
}
