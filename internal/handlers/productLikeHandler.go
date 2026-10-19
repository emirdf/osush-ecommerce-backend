package handlers

import (
	"net/http"
	"osush/internal/config"
	customError "osush/internal/errors"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/middleware"
	"osush/internal/usecase"
	"osush/internal/validation"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductLikeHandler struct {
	cfg                *config.Config
	productLikeUseCase *usecase.ProductLikeUseCase
	logger             log.Logger
}

func NewProductLikeHandler(
	cfg *config.Config,
	u *usecase.ProductLikeUseCase,
	logger log.Logger,
) *ProductLikeHandler {
	return &ProductLikeHandler{
		cfg:                cfg,
		productLikeUseCase: u,
		logger:             logger,
	}
}

func (h *ProductLikeHandler) ToggleLikeProduct(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling ToggleLikeProduct request")

	clientID := c.MustGet("clientId").(int)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Warn("invalid product id",
			log.String("id", c.Param("id")),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid product id",
		})
		return
	}

	logger.Debug("toggling product like",
		log.Int("client_id", clientID),
		log.Int("product_id", id),
	)

	isLiked, err := h.productLikeUseCase.ToggleLikeProduct(
		ctx,
		clientID,
		id,
	)

	if err != nil {
		logger.Error("failed to toggle product like",
			log.Int("client_id", clientID),
			log.Int("product_id", id),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_liked": isLiked,
	})
}

func (h *ProductLikeHandler) GetLikedProducts(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetLikedProducts request")

	limit := c.Query("limit")
	offset := c.Query("offset")

	logger.Debug("parsed get all product is liked parameters",
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

	clientID := c.MustGet("clientId").(int)

	baseURL := h.cfg.MinioBaseURL()

	result, err := h.productLikeUseCase.GetLikedProducts(
		ctx,
		clientID,
		*limitInt,
		*offsetInt,
		baseURL,
	)

	if err != nil {
		logger.Error("failed to get liked products",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("liked products retrieved successfully",
		log.Int("client_id", clientID),
		log.Int("products_total", result.Total),
	)

	c.JSON(http.StatusOK, result)
}
