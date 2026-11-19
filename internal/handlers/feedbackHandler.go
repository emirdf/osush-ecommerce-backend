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

type FeedbackHandler struct {
	cfg             *config.Config
	feedbackUseCase *usecase.FeedbackUseCase
	logger          log.Logger
}

func NewFeedbackHandler(cfg *config.Config, u *usecase.FeedbackUseCase, logger log.Logger) *FeedbackHandler {
	return &FeedbackHandler{
		cfg:             cfg,
		feedbackUseCase: u,
		logger:          logger,
	}
}

// CreateFeedback handles POST /feedback. Works for guests and logged in
// clients (OptionalAuthMiddleware sets clientId, -1 for guests).
func (h *FeedbackHandler) CreateFeedback(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling create feedback request")

	var req request.FeedbackRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind feedback request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	if err := req.Validate(); err != nil {
		logger.Warn("feedback request validation failed",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	clientID := -1
	if value, exists := c.Get("clientId"); exists {
		if id, ok := value.(int); ok {
			clientID = id
		}
	}

	if err := h.feedbackUseCase.CreateFeedback(ctx, req, clientID); err != nil {
		logger.Error("failed to create feedback",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("feedback created successfully",
		log.Int("client_id", clientID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "feedback sent successfully",
	})
}

// GetFeedbackList handles GET /feedback/admin?limit=&offset= (offset is the page number, from 1).
func (h *FeedbackHandler) GetFeedbackList(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling get feedback list request")

	limit := c.Query("limit")
	offset := c.Query("offset")

	pageInt, limitInt, err := validation.ValidatePagination(offset, limit)
	if err != nil {
		logger.Warn("failed validate pagination",
			log.String("limit", limit),
			log.String("offset", offset),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	result, err := h.feedbackUseCase.GetFeedbackList(ctx, *limitInt, *pageInt)
	if err != nil {
		logger.Error("failed to get feedback list",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("feedback list retrieved successfully",
		log.Int("total", result.Total),
	)

	c.JSON(http.StatusOK, result)
}

// DeleteFeedback handles DELETE /feedback/admin/:id.
func (h *FeedbackHandler) DeleteFeedback(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling delete feedback request")

	id := c.Param("id")

	feedbackID, err := strconv.Atoi(id)
	if err != nil || feedbackID <= 0 {
		logger.Warn("invalid feedback id",
			log.String("id_str", id),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid feedback_id",
		})
		return
	}

	if err := h.feedbackUseCase.DeleteFeedbackByID(ctx, feedbackID); err != nil {
		logger.Error("failed to delete feedback",
			log.Int("feedback_id", feedbackID),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("feedback deleted successfully",
		log.Int("feedback_id", feedbackID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "feedback deleted successfully",
	})
}
