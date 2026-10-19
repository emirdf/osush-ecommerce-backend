package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/models"
	"osush/internal/repository"
	"strings"
)

type OrderUseCase struct {
	repo             *repository.OrderRepository
	clientRepo       *repository.ClientRepository
	basketRepo       *repository.BasketRepository
	orderProductRepo *repository.OrderProductRepository
	productRepo      *repository.ProductRepository
	logger           log.Logger
}

func NewOrderUseCase(
	repo *repository.OrderRepository,
	clientRepo *repository.ClientRepository,
	basketRepo *repository.BasketRepository,
	orderProductRepo *repository.OrderProductRepository,
	productRepo *repository.ProductRepository,
	logger log.Logger) *OrderUseCase {
	return &OrderUseCase{
		repo:             repo,
		clientRepo:       clientRepo,
		basketRepo:       basketRepo,
		orderProductRepo: orderProductRepo,
		productRepo:      productRepo,
		logger:           logger,
	}
}

func (u *OrderUseCase) CreateOrder(
	ctx context.Context,
	req request.OrderRequest,
	clientID int,
) error {

	u.logger.Info("creating order",
		log.Int("client_id", clientID),
	)

	existClient, err := u.clientRepo.ExistsByID(ctx, clientID)
	if err != nil {
		u.logger.Error("failed to check client",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return errors.New("failed to check client")
	}

	if !existClient {
		u.logger.Warn("client not found",
			log.Int("client_id", clientID),
		)

		return customError.ErrNotFound("client")
	}

	var basket []models.Basket
	var productIDs []int

	for _, cartItem := range req.CartItems {

		basket = append(basket, models.Basket{
			ClientID:  clientID,
			ProductID: cartItem.ProductID,
			Total:     cartItem.Quantity,
		})

		productIDs = append(productIDs, cartItem.ProductID)
	}

	if len(basket) == 0 {
		return customError.ErrBadRequest("basket is empty")
	}

	err = u.basketRepo.CheckBasketProduct(ctx, clientID, basket)
	if err != nil {
		u.logger.Error("failed to check basket products",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return err
	}

	err = u.productRepo.CheckProductsNotPaused(ctx, productIDs)

	if err != nil {
		u.logger.Error("failed to check paused products",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return err
	}

	order := models.Order{
		ClientID:        clientID,
		CustomerName:    req.CustomerName,
		PhoneNumber:     req.PhoneNumber,
		City:            req.City,
		DeliveryAddress: req.DeliveryAddress,
		Notes:           req.Notes,
	}

	prices, err := u.productRepo.GetProductsPrices(ctx, productIDs)
	if err != nil {
		u.logger.Error("failed to get products prices",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return err
	}

	orderProducts := make([]models.OrderProduct, 0, len(req.CartItems))

	for _, cardItem := range req.CartItems {

		productPrice, exists := prices[cardItem.ProductID]
		if !exists {
			return customError.ErrNotFound(
				fmt.Sprintf("product with id %d", cardItem.ProductID),
			)
		}

		orderProduct := models.OrderProduct{
			ProductID:     cardItem.ProductID,
			Quantity:      cardItem.Quantity,
			Price:         productPrice.Price,
			DiscountPrice: productPrice.DiscountPrice,
		}

		orderProducts = append(orderProducts, orderProduct)
	}

	orderID, err := u.repo.CreateOrder(ctx, order, orderProducts)
	if err != nil {
		u.logger.Error("failed to create order",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return err
	}

	u.logger.Info("order and order_products created successfully",
		log.Int("client_id", clientID),
		log.Int("order_id", *orderID),
		log.Int("products_count", len(orderProducts)),
	)

	err = u.basketRepo.CleanBasket(ctx, clientID)
	if err != nil {
		u.logger.Error("failed to clean basket",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return err
	}

	u.logger.Info("basket cleaned successfully",
		log.Int("client_id", clientID),
		log.Int("order_id", *orderID),
		log.Int("products_count", len(orderProducts)),
	)

	return nil
}

func (u *OrderUseCase) GetOrderByID(
	ctx context.Context,
	orderID int,
	baseURL string,
) (*response.OrderResponse, error) {

	u.logger.Info("getting order by id",
		log.Int("order_id", orderID),
	)

	exists, err := u.repo.ExistsByID(ctx, orderID)
	if err != nil {
		u.logger.Error("failed to check order existence",
			log.Int("category_id", orderID),
			log.Error(err),
		)

		return nil, err
	}

	if !exists {
		u.logger.Warn("order not found",
			log.Int("order_id", orderID),
		)

		return nil, customError.ErrNotFoundByID(orderID, "order")
	}

	order, err := u.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		u.logger.Error("failed to get order by id",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return nil, err
	}

	result := &response.OrderResponse{
		ID:              order.ID,
		CustomerName:    order.CustomerName,
		PhoneNumber:     order.PhoneNumber,
		City:            order.City,
		DeliveryAddress: order.DeliveryAddress,
		Notes:           order.Notes,
		IsSend:          order.IsSend,
		Products:        make([]response.OrderProductResponse, 0),
	}

	orderProducts, err := u.orderProductRepo.GetOrderProductsByOrderID(
		ctx,
		orderID,
	)
	if err != nil {
		u.logger.Error("failed to get order products",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return nil, err
	}

	var totalPrice float64
	var totalProduct int

	for _, orderProduct := range *orderProducts {

		imagePath := orderProduct.ProductImagePath

		if baseURL != "" && imagePath != "" {
			cleanPath := strings.ReplaceAll(imagePath, "\\", "/")
			imagePath = fmt.Sprintf("%s/%s", baseURL, cleanPath)
		}

		price := orderProduct.Price

		if orderProduct.DiscountPrice != nil {
			price = *orderProduct.DiscountPrice
		}

		totalPrice += math.Round(price) * float64(orderProduct.Quantity)
		totalProduct += orderProduct.Quantity

		result.Products = append(
			result.Products,
			response.OrderProductResponse{
				ID:            orderProduct.ProductID,
				Name:          orderProduct.ProductName,
				Price:         orderProduct.Price,
				DiscountPrice: orderProduct.DiscountPrice,
				Image:         imagePath,
				Total:         orderProduct.Quantity,
			},
		)
	}

	result.TotalProduct = totalProduct
	result.TotalPrice = totalPrice

	u.logger.Info("order retrieved successfully",
		log.Int("order_id", orderID),
	)

	if order.IsNew {
		err := u.repo.MarkOrderAsNotNew(ctx, orderID)
		if err != nil {
			u.logger.Error("failed to mark order as not new",
				log.Int("order_id", orderID),
				log.Error(err),
			)

			return nil, err
		}
	}

	return result, nil
}

func (u *OrderUseCase) GetOrderClientByID(
	ctx context.Context,
	orderID int,
	clientID int,
	baseURL string,
) (*response.OrderResponse, error) {

	u.logger.Info("getting order by id",
		log.Int("order_id", orderID),
	)

	exists, err := u.repo.ExistsByID(ctx, orderID)
	if err != nil {
		u.logger.Error("failed to check order existence",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return nil, err
	}

	if !exists {
		u.logger.Warn("order not found",
			log.Int("order_id", orderID),
		)

		return nil, customError.ErrNotFoundByID(orderID, "order")
	}

	existsClient, err := u.repo.ExistsClientsByID(ctx, orderID, clientID)
	if err != nil {
		u.logger.Error("failed to check order existence by clientID",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return nil, err
	}

	if !existsClient {
		u.logger.Warn("order not found by client",
			log.Int("order_id", orderID),
		)

		return nil, customError.ErrNotFoundByID(orderID, "order by client, order")
	}

	order, err := u.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		u.logger.Error("failed to get order by id",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return nil, err
	}

	result := &response.OrderResponse{
		ID:              order.ID,
		CustomerName:    order.CustomerName,
		PhoneNumber:     order.PhoneNumber,
		City:            order.City,
		DeliveryAddress: order.DeliveryAddress,
		Notes:           order.Notes,
		IsSend:          order.IsSend,
		Products:        make([]response.OrderProductResponse, 0),
	}

	orderProducts, err := u.orderProductRepo.GetOrderProductsByOrderID(
		ctx,
		orderID,
	)
	if err != nil {
		u.logger.Error("failed to get order products",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return nil, err
	}

	var totalPrice float64
	var totalProduct int

	for _, orderProduct := range *orderProducts {

		imagePath := orderProduct.ProductImagePath

		if baseURL != "" && imagePath != "" {
			cleanPath := strings.ReplaceAll(imagePath, "\\", "/")
			imagePath = fmt.Sprintf("%s/%s", baseURL, cleanPath)
		}

		price := orderProduct.Price

		if orderProduct.DiscountPrice != nil {
			price = *orderProduct.DiscountPrice
		}

		totalPrice += math.Round(price) * float64(orderProduct.Quantity)
		totalProduct += orderProduct.Quantity

		result.Products = append(
			result.Products,
			response.OrderProductResponse{
				ID:            orderProduct.ProductID,
				Name:          orderProduct.ProductName,
				Price:         orderProduct.Price,
				DiscountPrice: orderProduct.DiscountPrice,
				Image:         imagePath,
				Total:         orderProduct.Quantity,
			},
		)
	}

	result.TotalProduct = totalProduct
	result.TotalPrice = totalPrice

	u.logger.Info("order retrieved successfully",
		log.Int("order_id", orderID),
	)

	if order.IsNew {
		err := u.repo.MarkOrderAsNotNew(ctx, orderID)
		if err != nil {
			u.logger.Error("failed to mark order as not new",
				log.Int("order_id", orderID),
				log.Error(err),
			)

			return nil, err
		}
	}

	return result, nil
}

func (u *OrderUseCase) GetOrders(
	ctx context.Context,
	filter request.OrderListFilter,
) (*response.OrderProductListResponse, error) {

	u.logger.Info("getting orders")

	orders, total, err := u.repo.GetOrders(ctx, filter)

	if err != nil {
		u.logger.Error("failed to get orders",
			log.Error(err),
		)

		return nil, err
	}

	result := &response.OrderProductListResponse{
		Total:  *total,
		Orders: make([]response.OrdersResponse, 0),
	}

	for _, order := range *orders {

		orderProducts, err := u.orderProductRepo.GetOrderProductsByOrderID(
			ctx,
			order.ID,
		)

		if err != nil {
			u.logger.Error("failed to get order products",
				log.Int("order_id", order.ID),
				log.Error(err),
			)

			return nil, err
		}

		var totalPrice float64
		var totalProduct int

		for _, orderProduct := range *orderProducts {
			price := orderProduct.Price

			if orderProduct.DiscountPrice != nil {
				price = *orderProduct.DiscountPrice
			}

			totalPrice += math.Round(price) * float64(orderProduct.Quantity)
			totalProduct += orderProduct.Quantity
		}

		result.Orders = append(
			result.Orders,
			response.OrdersResponse{
				ID:           order.ID,
				CustomerName: order.CustomerName,
				PhoneNumber:  order.PhoneNumber,
				City:         order.City,
				TotalProduct: totalProduct,
				TotalPrice:   totalPrice,
				IsSend:       order.IsSend,
				IsNew:        order.IsNew,
			},
		)
	}

	return result, nil
}

func (u *OrderUseCase) GetOrdersClient(
	ctx context.Context,
	clientID int,
) (*response.OrderProductClientListResponse, error) {

	u.logger.Info("getting orders")

	orders, total, err := u.repo.GetOrdersClient(ctx, clientID)

	if err != nil {
		u.logger.Error("failed to get orders",
			log.Error(err),
		)

		return nil, err
	}

	result := &response.OrderProductClientListResponse{
		Total:  *total,
		Orders: make([]response.OrdersClientResponse, 0),
	}

	for _, order := range *orders {

		orderProducts, err := u.orderProductRepo.GetOrderProductsByOrderID(
			ctx,
			order.ID,
		)

		if err != nil {
			u.logger.Error("failed to get order products",
				log.Int("order_id", order.ID),
				log.Error(err),
			)

			return nil, err
		}

		var totalPrice float64
		var totalProduct int

		for _, orderProduct := range *orderProducts {
			price := orderProduct.Price

			if orderProduct.DiscountPrice != nil {
				price = *orderProduct.DiscountPrice
			}

			totalPrice += math.Round(price) * float64(orderProduct.Quantity)
			totalProduct += orderProduct.Quantity
		}

		result.Orders = append(
			result.Orders,
			response.OrdersClientResponse{
				ID:              order.ID,
				DeliveryAddress: order.DeliveryAddress,
				TotalProduct:    totalProduct,
				TotalPrice:      totalPrice,
				IsSend:          order.IsSend,
			},
		)
	}

	return result, nil
}

func (u *OrderUseCase) MarkOrderAsSend(
	ctx context.Context,
	orderID int,
) error {

	u.logger.Info(
		"marking order as sent",
		log.Int("order_id", orderID),
	)

	exists, err := u.repo.ExistsByID(ctx, orderID)
	if err != nil {
		u.logger.Error("failed to check order existence",
			log.Int("category_id", orderID),
			log.Error(err),
		)

		return err
	}

	if !exists {
		u.logger.Warn("order not found",
			log.Int("order_id", orderID),
		)

		return customError.ErrNotFoundByID(orderID, "order")
	}

	err = u.repo.MarkOrderAsSend(ctx, orderID)
	if err != nil {
		u.logger.Error(
			"failed to mark order as sent",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return err
	}

	u.logger.Info(
		"order marked as sent successfully",
		log.Int("order_id", orderID),
	)

	return nil
}
