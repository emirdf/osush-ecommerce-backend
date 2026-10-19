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

type BasketHandler struct {
	cfg           *config.Config
	basketUseCase *usecase.BasketUseCase
	logger        log.Logger
}

func NewBasketHandler(cfg *config.Config, u *usecase.BasketUseCase, logger log.Logger) *BasketHandler {
	return &BasketHandler{
		cfg:           cfg,
		basketUseCase: u,
		logger:        logger,
	}
}

func (h *BasketHandler) AddProductFromBasket(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling AddProductToBasket request")

	var req request.BasketRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("failed data basket request",
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

	logger.Info("Subcategory request is valid")

	clientID := c.MustGet("clientId").(int)

	total, err := h.basketUseCase.AddProduct(ctx, req, clientID)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	logger.Info("product added to basket",
		log.Int("client_id", clientID),
		log.Int("product_id", req.ProductID),
		log.Int("total", *total),
	)

	c.JSON(http.StatusOK, gin.H{
		"total": *total,
	})
}

func (h *BasketHandler) DeleteProductFromBasket(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling DeleteProductFromBasket request")

	var req request.BasketRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("failed data basket request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		logger.Error("invalid basket request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Debug("basket request is valid",
		log.Int("product_id", req.ProductID),
	)

	clientID := c.MustGet("clientId").(int)

	total, err := h.basketUseCase.DeleteProduct(ctx, req, clientID)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	logger.Info("product removed from basket",
		log.Int("client_id", clientID),
		log.Int("product_id", req.ProductID),
		log.Int("total", *total),
	)

	c.JSON(http.StatusOK, gin.H{
		"total": *total,
	})
}

func (h *BasketHandler) CleanProductFromBasket(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling DeleteBasket request")

	clientID := c.MustGet("clientId").(int)

	id := c.Param("id")
	productID, err := strconv.Atoi(id)
	if err != nil {
		logger.Warn("invalid product id",
			log.String("id_str", id),
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid product id",
		})
		return
	}

	err = h.basketUseCase.CleanProduct(ctx, clientID, productID)
	if err != nil {
		customError.HandleError(c, err)
		return
	}

	logger.Info("basket deleted successfully",
		log.Int("client_id", clientID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "basket deleted successfully",
	})
}

func (h *BasketHandler) GetAllBasketProducts(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetAllBasketProducts request")

	clientID := c.MustGet("clientId").(int)

	limit := c.Query("limit")
	offset := c.Query("offset")

	logger.Debug("parsed get all basket products parameters",
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

	baseURL := h.cfg.MinioBaseURL()

	result, err := h.basketUseCase.GetAllBasketProducts(
		ctx,
		clientID,
		*limitInt,
		*offsetInt,
		baseURL,
	)
	if err != nil {
		logger.Error("failed to get basket products",
			log.Int("client_id", clientID),
			log.Error(err),
		)
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
