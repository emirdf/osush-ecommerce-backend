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
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryUseCase *usecase.CategoryUseCase
	cfg             *config.Config
	logger          log.Logger
}

func NewCategoryHandler(
	cfg *config.Config,
	u *usecase.CategoryUseCase,
	logger log.Logger) *CategoryHandler {
	return &CategoryHandler{
		cfg:             cfg,
		categoryUseCase: u,
		logger:          logger,
	}
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling CreateCategory request")

	var req request.CategoryRequest

	jsonData := c.PostForm("data")
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		logger.Error("failed data category request",
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info("handling CreateCategory request")

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info("Category request is valid")

	image, err := c.FormFile("image")
	if err != nil {
		logger.Error("failed image category request",
			log.Error(err),
		)
		if errors.Is(err, http.ErrMissingFile) {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: "image is required",
			})
			return
		}

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Debug("category image received",
		log.String("name", req.Name),
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

	logger.Info("Category image is valid")

	baseURL := h.cfg.MinioBaseURL()

	category, err := h.categoryUseCase.CreateCategory(ctx, req, image, contentType, baseURL)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *CategoryHandler) GetCategory(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetCategory by id request")

	id := c.Param("id")

	categoryID, err := strconv.Atoi(id)
	if err != nil {
		logger.Warn("invalid category id",
			log.String("id_str", id),
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid category_id",
		})
		return
	}

	baseURL := h.cfg.MinioBaseURL()

	category, err := h.categoryUseCase.GetCategoryByID(ctx, categoryID, baseURL)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) GetCategoryList(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling get all categories request")

	search := c.Query("search")
	limit := c.Query("limit")
	offset := c.Query("offset")

	logger.Debug("parsed get all categories parameters",
		log.String("search", search),
		log.String("limit", limit),
		log.String("offset", offset),
	)

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
	baseURL := h.cfg.MinioBaseURL()

	categories, err := h.categoryUseCase.GetCategoryList(
		ctx,
		search,
		*limitInt,
		*offsetInt,
		baseURL,
	)
	if err != nil {
		logger.Error("failed to get all categories",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("get all categories request completed")

	c.JSON(http.StatusOK, categories)
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling UpdateCategory request")

	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("failed to parse category id",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid category id",
		})
		return
	}

	var req request.CategoryRequest

	jsonData := c.PostForm("data")
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		logger.Error("failed to parse category request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		logger.Error("category request validation failed",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var image *multipart.FileHeader
	contentType := ""

	file, err := c.FormFile("image")
	if err == nil {
		image = file
		contentType = image.Header.Get("Content-Type")

		logger.Debug("category image received",
			log.String("name", req.Name),
			log.Int64("size", image.Size),
		)

		if err := validation.ValidateImageSize(image.Size); err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		if err := validation.ValidateImageFormat(contentType); err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
	}

	baseURL := h.cfg.MinioBaseURL()

	category, err := h.categoryUseCase.UpdateCategory(
		ctx,
		categoryID,
		req,
		image,
		contentType,
		baseURL,
	)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling DeleteCategory by id request")

	id := c.Param("id")

	categoryID, err := strconv.Atoi(id)
	if err != nil {
		logger.Warn("invalid category id",
			log.String("id_str", id),
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid category_id",
		})
		return
	}

	err = h.categoryUseCase.DeleteCategory(ctx, categoryID)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "category deleted successfully",
	})
}
