package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"osush/internal/config"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/middleware"
	"osush/internal/usecase"
	"osush/internal/validation"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct {
	clientUseCase *usecase.ClientUseCase
	logger        log.Logger
	cfg           *config.Config
}

func NewClientHandler(cfg *config.Config, u *usecase.ClientUseCase, logger log.Logger) *ClientHandler {
	return &ClientHandler{
		clientUseCase: u,
		logger:        logger,
		cfg:           cfg,
	}
}

func (h *ClientHandler) Register(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling client Register request")

	var req request.RegisterRequest

	jsonData := c.PostForm("data")
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		logger.Error("failed data client request",
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info("Client request is valid")

	image, err := c.FormFile("avatar")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			logger.Error("failed to get client avatar",
				log.Error(err),
			)

			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		logger.Debug("client avatar not provided")
		image = nil
	}

	if image != nil {
		logger.Debug("client avatar received",
			log.String("user_name", req.UserName),
			log.Int64("size", image.Size),
		)

		if err := validation.ValidateImageSize(image.Size); err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		contentType := image.Header.Get("Content-Type")

		if err := validation.ValidateImageFormat(contentType); err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		logger.Info("Client avatar is valid")
	}

	baseURL := h.cfg.MinioBaseURL()

	result, err := h.clientUseCase.Register(ctx, req, image, baseURL)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	logger.Info("client Register request completed",
		log.Int("client_id", result.ID),
	)

	c.JSON(http.StatusCreated, result)
}

func (h *ClientHandler) UpdateProfile(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling client profile update request")

	clientID := c.MustGet("clientId").(int)

	var req request.UpdateRequest

	jsonData := c.PostForm("data")

	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		logger.Error("failed to parse client update request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})

		return
	}

	if err := req.Validate(); err != nil {
		logger.Warn("client update request validation failed",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})

		return
	}

	logger.Info("client update request is valid",
		log.Int("client_id", clientID),
	)

	image, err := c.FormFile("avatar")

	if err != nil {

		if !errors.Is(err, http.ErrMissingFile) {
			logger.Error("failed to get client avatar",
				log.Error(err),
			)

			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: err.Error(),
			})

			return
		}

		logger.Debug("new client avatar not provided",
			log.Int("client_id", clientID),
		)

		image = nil
	}

	if image != nil {

		logger.Debug("new client avatar received",
			log.Int("client_id", clientID),
			log.Int64("size", image.Size),
		)

		if err := validation.ValidateImageSize(image.Size); err != nil {
			logger.Warn("client avatar validation failed",
				log.Int("client_id", clientID),
				log.Error(err),
			)

			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: err.Error(),
			})

			return
		}

		contentType := image.Header.Get("Content-Type")

		if err := validation.ValidateImageFormat(contentType); err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		logger.Info("Client avatar is valid")
	}

	baseURL := h.cfg.MinioBaseURL()

	result, err := h.clientUseCase.UpdateProfile(
		ctx,
		clientID,
		req,
		image,
		baseURL,
	)

	if err != nil {
		customError.HandleError(c, err)
		return
	}

	logger.Info("client profile update request completed",
		log.Int("client_id", clientID),
	)

	c.JSON(http.StatusOK, result)
}

func (h *ClientHandler) Login(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling client Login request")

	var req request.LoginClientRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("invalid client login request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Debug("client login request received",
		log.String("phone_number", req.PhoneNumber),
	)

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

	token, err := h.clientUseCase.Login(ctx, h.cfg, req)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	logger.Info("login successful")

	c.JSON(http.StatusOK, gin.H{
		"token": *token,
	})
}

func (h *ClientHandler) Logout(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling client logout request")

	tokenString := c.GetHeader("Authorization")

	if tokenString == "" {
		logger.Warn("logout request failed: authorization header is missing")

		customError.HandleError(c, customError.ErrMissingToken)
		return
	}

	if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
		logger.Warn("logout request failed: invalid authorization header")

		customError.HandleError(c, customError.ErrBadToken)
		return
	}

	tokenString = tokenString[7:]

	logger.Debug("authorization token extracted successfully")

	if tokenString == "" {
		logger.Warn("logout request failed: empty token")

		customError.HandleError(c, customError.ErrBadToken)
		return
	}

	logger.Debug("sending logout request to usecase")

	err := h.clientUseCase.Logout(ctx, tokenString)
	if err != nil {
		logger.Error("client logout failed",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("client logout successful")

	c.JSON(http.StatusOK, gin.H{
		"message": "successful",
	})
}

// DeleteAccount deletes the signed-in client's own account.
func (h *ClientHandler) DeleteAccount(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	clientID := c.MustGet("clientId").(int)

	logger.Info("handling client account deletion request",
		log.Int("client_id", clientID),
	)

	// AuthMiddleware has already checked the "Bearer <token>" format.
	token := c.GetHeader("Authorization")[len("Bearer "):]

	if err := h.clientUseCase.DeleteAccount(ctx, clientID, token); err != nil {
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successful",
	})
}

func (h *ClientHandler) GetProfile(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling get client profile request")

	clientID := c.MustGet("clientId").(int)

	logger.Debug("client id received from context",
		log.Int("client_id", clientID),
	)

	baseURL := h.cfg.MinioBaseURL()

	result, err := h.clientUseCase.GetProfile(ctx, clientID, baseURL)

	if err != nil {
		customError.HandleError(c, err)
		return
	}

	logger.Info("get client profile request completed",
		log.Int("client_id", clientID),
	)

	c.JSON(http.StatusOK, result)
}

func (h *ClientHandler) GetClientList(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	search := c.Query("search")
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

	err = validation.ValidateSearchQuery(search)
	if err != nil {
		logger.Error("failed validate search",
			log.String("search", search),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info("client request is valid")

	result, err := h.clientUseCase.GetClientList(ctx, search, *limitInt, *offsetInt)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
