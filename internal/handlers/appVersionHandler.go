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
	"strings"

	"github.com/gin-gonic/gin"
)

type AppVersionHandler struct {
	cfg               *config.Config
	appVersionUseCase *usecase.AppVersionUseCase
	logger            log.Logger
}

func NewAppVersionHandler(cfg *config.Config, u *usecase.AppVersionUseCase, logger log.Logger) *AppVersionHandler {
	return &AppVersionHandler{cfg: cfg, appVersionUseCase: u, logger: logger}
}

// CheckAppVersion handles GET /app-version?platform=android|ios&version=1.0.0+4.
// Public: the mobile app asks before login.
func (h *AppVersionHandler) CheckAppVersion(c *gin.Context) {
	logger := middleware.GetLogger(c)

	platform := strings.ToLower(strings.TrimSpace(c.Query("platform")))
	current := strings.TrimSpace(c.Query("version"))

	result, err := h.appVersionUseCase.Check(c.Request.Context(), platform, current)
	if err != nil {
		logger.Warn("failed to check app version",
			log.String("platform", platform),
			log.String("version", current),
			log.Error(err),
		)
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAppVersions handles GET /app-version/admin: both platforms for the panel.
func (h *AppVersionHandler) GetAppVersions(c *gin.Context) {
	logger := middleware.GetLogger(c)

	result, err := h.appVersionUseCase.GetAll(c.Request.Context())
	if err != nil {
		logger.Error("failed to get app versions", log.Error(err))
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateAppVersion handles PUT /app-version/admin.
//   - raise only latest_version to tell older apps a new release exists;
//   - raise min_version too to force every older app to update.
func (h *AppVersionHandler) UpdateAppVersion(c *gin.Context) {
	logger := middleware.GetLogger(c)

	var req request.AppVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind app version request", log.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := req.Validate(); err != nil {
		logger.Warn("app version request validation failed", log.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.appVersionUseCase.Update(c.Request.Context(), req)
	if err != nil {
		logger.Error("failed to update app version", log.Error(err))
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
