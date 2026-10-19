package handlers

import (
	"encoding/json"
	"net/http"
	"osush/internal/config"
	"osush/internal/errors"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/middleware"
	"osush/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
	logger      log.Logger
	cfg         *config.Config
}

func NewAuthHandler(cfg *config.Config, u *usecase.AuthUseCase, logger log.Logger) *AuthHandler {
	return &AuthHandler{
		authUseCase: u,
		logger:      logger,
		cfg:         cfg,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling login request")

	var req request.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn(
			"failed to parse login request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		logger.Warn(
			"login request validation failed",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Debug("login request validated")

	token, err := h.authUseCase.Login(ctx, h.cfg, req)
	if err != nil {
		logger.Error("login failed",
			log.Error(err),
		)
		errors.HandleError(c, err)
		return
	}

	logger.Info("login successful")

	c.JSON(http.StatusOK, gin.H{
		"token": *token,
	})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling update admin profile request")

	userID := c.MustGet("adminId").(int)

	data := c.PostForm("data")

	var req request.UpdateProfileRequest

	if err := json.Unmarshal([]byte(data), &req); err != nil {
		logger.Warn(
			"failed to parse update profile request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		logger.Error("user request validation failed",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	avatar, err := c.FormFile("avatar")
	if err != nil {
		avatar = nil
	}

	baseURL := h.cfg.MinioBaseURL()

	userProfile, err := h.authUseCase.UpdateProfile(ctx, userID, req, avatar, baseURL)

	if err != nil {
		logger.Error(
			"failed to update admin profile",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info(
		"admin profile updated successfully",
	)

	c.JSON(http.StatusOK, userProfile)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling get client profile request")

	userID := c.MustGet("adminId").(int)

	logger.Debug("client id received from context",
		log.Int("client_id", userID),
	)

	baseURL := h.cfg.MinioBaseURL()

	result, err := h.authUseCase.GetProfile(ctx, userID, baseURL)

	if err != nil {
		customError.HandleError(c, err)
		return
	}

	logger.Info("get client profile request completed")

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) GetAdminIndexPage(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling get admin index page request")

	baseURL := h.cfg.MinioBaseURL()

	result, err := h.authUseCase.GetAdminIndexPage(ctx, baseURL)
	if err != nil {
		logger.Error("failed to get admin index page",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("get admin index page request completed")

	c.JSON(http.StatusOK, result)
}
