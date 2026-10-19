package handlers

import (
	"encoding/json"
	"errors"
	"mime/multipart"
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

type BannerHandler struct {
	cfg           *config.Config
	bannerUseCase *usecase.BannerUseCase
	logger        log.Logger
}

func NewBannerHandler(cfg *config.Config, u *usecase.BannerUseCase, logger log.Logger) *BannerHandler {
	return &BannerHandler{
		cfg:           cfg,
		bannerUseCase: u,
		logger:        logger,
	}
}

func (h *BannerHandler) CreateBanner(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	var (
		additionalImages []*multipart.FileHeader
		req              request.BannerRequest
	)

	logger.Info("handling CreateBanner request")

	jsonData := c.PostForm("data")
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		logger.Error("failed to parse banner request data",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		logger.Warn("banner request validation failed",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info("banner request is valid")

	mainImage, err := c.FormFile("mainImage")
	if err != nil {
		logger.Error("failed to get banner main image",
			log.Error(err),
		)

		if errors.Is(err, http.ErrMissingFile) {
			logger.Warn("banner main image is missing")

			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: "mainImage is required",
			})
			return
		}

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Debug("banner main image received",
		log.String("file_name", mainImage.Filename),
		log.Int64("size", mainImage.Size),
	)

	if err := validation.ValidateImageSize(mainImage.Size); err != nil {
		logger.Warn("banner main image size validation failed",
			log.Int64("size", mainImage.Size),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	contentType := mainImage.Header.Get("Content-Type")

	if err := validation.ValidateImageFormat(contentType); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		logger.Warn("failed to get multipart form for additional images",
			log.Error(err),
		)
	} else {
		additionalFiles := form.File["additionImages"]

		logger.Debug("additional banner images received",
			log.Int("count", len(additionalFiles)),
		)

		for _, image := range additionalFiles {
			logger.Debug("validating banner additional image",
				log.String("file_name", image.Filename),
				log.Int64("size", image.Size),
			)

			if err := validation.ValidateImageSize(image.Size); err != nil {
				logger.Warn("banner additional image size validation failed",
					log.String("file_name", image.Filename),
					log.Int64("size", image.Size),
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

			additionalImages = append(additionalImages, image)
		}
	}

	bannerID, err := h.bannerUseCase.CreateBanner(ctx, req)
	if err != nil {
		logger.Error("failed to create banner",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("banner created successfully",
		log.Int("banner_id", *bannerID),
	)

	logger.Info("starting save banner images",
		log.Int("banner_id", *bannerID),
		log.String("file_name", mainImage.Filename),
	)

	err = h.bannerUseCase.AddImages(
		ctx,
		*bannerID,
		mainImage,
		additionalImages,
	)

	if err != nil {
		logger.Error("failed to save banner images",
			log.Int("banner_id", *bannerID),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("banner images saved successfully",
		log.Int("banner_id", *bannerID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "banner created successfully",
	})
}

func (h *BannerHandler) GetBannerList(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetBannerList request")

	baseURL := h.cfg.MinioBaseURL()

	banners, err := h.bannerUseCase.GetBannerList(ctx, baseURL)
	if err != nil {
		logger.Error("failed to get banners",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("banners retrieved successfully",
		log.Int("total", banners.Total),
	)

	c.JSON(http.StatusOK, banners)
}
