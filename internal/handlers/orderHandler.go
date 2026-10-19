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

type OrderHandler struct {
	cfg          *config.Config
	orderUseCase *usecase.OrderUseCase
	logger       log.Logger
}

func NewOrderHandler(cfg *config.Config, u *usecase.OrderUseCase, logger log.Logger) *OrderHandler {
	return &OrderHandler{
		cfg:          cfg,
		orderUseCase: u,
		logger:       logger,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling CreateOrder request")

	var req request.OrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		logger.Error("failed to bind order request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})

		return
	}

	if err := req.Validate(); err != nil {

		logger.Warn("invalid order request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})

		return
	}

	logger.Debug("order request is valid")

	clientID := c.MustGet("clientId").(int)

	err := h.orderUseCase.CreateOrder(ctx, req, clientID)

	if err != nil {
		customError.HandleError(c, err)
		return
	}

	logger.Info("order created successfully",
		log.Int("client_id", clientID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "order created successfully",
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetOrder request")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Warn("invalid order id",
			log.String("id", c.Param("id")),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid order id",
		})
		return
	}

	logger.Debug("order id parsed",
		log.Int("order_id", id),
	)

	baseURL := h.cfg.MinioBaseURL()

	order, err := h.orderUseCase.GetOrderByID(
		ctx,
		id,
		baseURL,
	)
	if err != nil {
		logger.Error("failed to get order",
			log.Int("order_id", id),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("GetOrderByID request completed",
		log.Int("order_id", id),
	)

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetOrderClient(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetOrder request")

	clientID := c.MustGet("clientId").(int)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Warn("invalid order id",
			log.String("id", c.Param("id")),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid order id",
		})
		return
	}

	logger.Debug("order id parsed",
		log.Int("order_id", id),
	)

	baseURL := h.cfg.MinioBaseURL()

	order, err := h.orderUseCase.GetOrderClientByID(
		ctx,
		id,
		clientID,
		baseURL,
	)
	if err != nil {
		logger.Error("failed to get order",
			log.Int("order_id", id),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("GetOrderByID request completed",
		log.Int("order_id", id),
	)

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetOrderList(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetOrders request")

	var filter request.OrderListFilter

	if err := c.ShouldBindQuery(&filter); err != nil {
		logger.Warn("invalid order list request",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	offsetInt, limitInt, err := validation.ValidatePagination(filter.Offset, filter.Limit)
	if err != nil {
		logger.Error("failed validate pagination",
			log.String("limit", filter.Limit),
			log.String("offset", filter.Offset),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	filter.OffsetInt = *offsetInt
	filter.LimitInt = *limitInt

	err = validation.ValidateSearchQuery(filter.Search)
	if err != nil {
		logger.Error("failed validate search",
			log.String("search", filter.Search),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Debug("order list parameters",
		log.String("search", filter.Search),
		log.Int("limit", filter.LimitInt),
		log.Int("offset", filter.OffsetInt),
	)

	orders, err := h.orderUseCase.GetOrders(ctx, filter)
	if err != nil {
		logger.Error("failed to get orders",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("GetOrders request completed")

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetOrderClientList(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetOrders request")

	clientID := c.MustGet("clientId").(int)

	logger.Debug("parsed get all order parameters")

	orders, err := h.orderUseCase.GetOrdersClient(ctx, clientID)
	if err != nil {
		logger.Error("failed to get orders",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("GetOrders request completed")

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) MarkOrderAsSend(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling MarkOrderAsSend request")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Warn(
			"invalid order id",
			log.String("id", c.Param("id")),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid order id",
		})
		return
	}

	logger.Debug(
		"order id parsed",
		log.Int("order_id", id),
	)

	err = h.orderUseCase.MarkOrderAsSend(ctx, id)
	if err != nil {
		logger.Error(
			"failed to mark order as sent",
			log.Int("order_id", id),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info(
		"order marked as sent successfully",
		log.Int("order_id", id),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "order marked as sent",
	})
}
