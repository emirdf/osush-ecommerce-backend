package handlers

import (
	"net/http"
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

type SubcategoryHandler struct {
	subcategoryUseCase *usecase.SubcategoryUseCase
	logger             log.Logger
}

func NewSubcategoryHandler(u *usecase.SubcategoryUseCase, logger log.Logger) *SubcategoryHandler {
	return &SubcategoryHandler{
		subcategoryUseCase: u,
		logger:             logger,
	}
}

func (h *SubcategoryHandler) CreateSubcategory(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling CreateSubcategory request")

	var req request.SubcategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("failed data subcategory request",
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info("handling CreateSubcategory request")

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info("Subcategory request is valid")

	subcategory, err := h.subcategoryUseCase.CreateSubcategory(ctx, req)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, subcategory)
}

func (h *SubcategoryHandler) GetSubcategory(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetSubcategory by id request")

	id := c.Param("id")

	subcategoryID, err := strconv.Atoi(id)
	if err != nil {
		logger.Warn("invalid subcategory_id",
			log.String("id_str", id),
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid subcategory_id",
		})
		return
	}

	subcategory, err := h.subcategoryUseCase.GetSubcategoryByID(ctx, subcategoryID)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, subcategory)
}

func (h *SubcategoryHandler) GetSubcategoryList(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling get all subcategories request")

	search := c.Query("search")
	limit := c.Query("limit")
	offset := c.Query("offset")
	categoryID := c.Query("category_id")

	logger.Debug("parsed get all subcategories parameters",
		log.String("search", search),
		log.String("limit", limit),
		log.String("offset", offset),
		log.String("category_id", categoryID),
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

	subcategories, err := h.subcategoryUseCase.GetSubcategoryList(
		ctx,
		search,
		*limitInt,
		*offsetInt,
		categoryID,
	)
	if err != nil {
		logger.Error("failed to get all subcategories",
			log.Error(err),
		)
		customError.HandleError(c, err)
		return
	}

	logger.Info("get all subcategories request completed")

	c.JSON(http.StatusOK, subcategories)
}

func (h *SubcategoryHandler) UpdateSubcategory(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	subcategoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("invalid subcategory id",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid subcategory id",
		})
		return
	}

	logger.Info("handling UpdateSubcategory request",
		log.Int("subcategory_id", subcategoryID),
	)

	var req request.SubcategoryUpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("failed to bind subcategory request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		logger.Error("invalid subcategory request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	updatedSubcategory, err := h.subcategoryUseCase.UpdateSubcategory(
		ctx,
		subcategoryID,
		req,
	)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, updatedSubcategory)
}

func (h *SubcategoryHandler) DeleteSubcategory(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	subcategoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("invalid subcategory id",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid subcategory id",
		})
		return
	}

	logger.Info("handling DeleteSubcategory request",
		log.Int("subcategory_id", subcategoryID),
	)

	if err := h.subcategoryUseCase.DeleteSubcategory(
		ctx,
		subcategoryID,
	); err != nil {
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":"subcategory deleted successfully",
	})
}
