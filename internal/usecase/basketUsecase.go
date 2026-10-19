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

type BasketUseCase struct {
	repo        *repository.BasketRepository
	productRepo *repository.ProductRepository
	clientRepo  *repository.ClientRepository
	logger      log.Logger
}

func NewBasketUseCase(
	repo *repository.BasketRepository,
	productRepo *repository.ProductRepository,
	clientRepo *repository.ClientRepository,
	logger log.Logger) *BasketUseCase {
	return &BasketUseCase{
		repo:        repo,
		productRepo: productRepo,
		clientRepo:  clientRepo,
		logger:      logger,
	}
}

func (u *BasketUseCase) AddProduct(
	ctx context.Context,
	req request.BasketRequest,
	clientID int,
) (*int, error) {

	u.logger.Info("adding product to basket",
		log.Int("client_id", clientID),
		log.Int("product_id", req.ProductID),
	)

	existClient, err := u.clientRepo.ExistsByID(ctx, clientID)
	if err != nil {
		u.logger.Warn("failed check client by id",
			log.Int("client_id", clientID),
			log.Error(err),
		)
		return nil, errors.New("failed check client")
	}

	if !existClient {
		u.logger.Warn("not found client",
			log.Int("client_id", clientID),
		)
		return nil, customError.ErrNotFound("client")
	}

	u.logger.Debug("client found",
		log.Int("client_id", clientID),
	)

	existProduct, err := u.productRepo.ExistsByID(ctx, req.ProductID)
	if err != nil {
		u.logger.Warn("failed check product by id",
			log.Int("product_id", req.ProductID),
			log.Error(err),
		)
		return nil, errors.New("failed check product")
	}

	if !existProduct {
		u.logger.Warn("not found product",
			log.Int("product_id", req.ProductID),
		)
		return nil, customError.ErrNotFoundByID(req.ProductID, "product")
	}

	u.logger.Debug("product found",
		log.Int("product_id", req.ProductID),
	)

	basket := models.Basket{
		ClientID:  clientID,
		ProductID: req.ProductID,
	}

	total, err := u.repo.AddProductInBasket(ctx, basket)
	if err != nil {
		u.logger.Error("failed to add product to basket",
			log.Int("client_id", clientID),
			log.Int("product_id", req.ProductID),
			log.Error(err),
		)

		return nil, err
	}

	u.logger.Info("product added to basket successfully",
		log.Int("client_id", clientID),
		log.Int("product_id", req.ProductID),
	)

	return total, nil
}

func (u *BasketUseCase) DeleteProduct(
	ctx context.Context,
	req request.BasketRequest,
	clientID int,
) (*int, error) {

	u.logger.Info("removing product from basket",
		log.Int("client_id", clientID),
		log.Int("product_id", req.ProductID),
	)

	existClient, err := u.clientRepo.ExistsByID(ctx, clientID)
	if err != nil {
		u.logger.Error("failed to check client by id",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, errors.New("failed to check client")
	}

	if !existClient {
		u.logger.Warn("client not found",
			log.Int("client_id", clientID),
		)

		return nil, customError.ErrNotFound("client")
	}

	u.logger.Debug("client found",
		log.Int("client_id", clientID),
	)

	existProduct, err := u.productRepo.ExistsByID(ctx, req.ProductID)
	if err != nil {
		u.logger.Error("failed to check product by id",
			log.Int("product_id", req.ProductID),
			log.Error(err),
		)

		return nil, errors.New("failed to check product")
	}

	if !existProduct {
		u.logger.Warn("product not found",
			log.Int("product_id", req.ProductID),
		)

		return nil, customError.ErrNotFoundByID(req.ProductID, "product")
	}

	u.logger.Debug("product found",
		log.Int("product_id", req.ProductID),
	)

	basket := models.Basket{
		ClientID:  clientID,
		ProductID: req.ProductID,
	}

	total, err := u.repo.DeleteProductFromBasket(ctx, basket)
	if err != nil {
		u.logger.Error("failed to remove product from basket",
			log.Int("client_id", clientID),
			log.Int("product_id", req.ProductID),
			log.Error(err),
		)

		return nil, err
	}

	u.logger.Info("product removed from basket successfully",
		log.Int("client_id", clientID),
		log.Int("product_id", req.ProductID),
	)

	return total, nil
}

func (u *BasketUseCase) CleanProduct(
	ctx context.Context,
	clientID int,
	productID int,
) error {

	u.logger.Info("deleting client product from basket",
		log.Int("client_id", clientID),
	)

	existClient, err := u.clientRepo.ExistsByID(ctx, clientID)
	if err != nil {
		u.logger.Error("failed to check client by id",
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

	u.logger.Debug("client found",
		log.Int("client_id", clientID),
	)

	existClientProduct, err := u.productRepo.ExistsByID(ctx, productID)
	if err != nil {
		u.logger.Error("failed to check product by id",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return errors.New("failed to check product")
	}

	if !existClientProduct {
		u.logger.Warn("product not found",
			log.Int("product_id", productID),
		)

		return customError.ErrNotFoundByID(productID, "product")
	}

	u.logger.Debug("product found",
		log.Int("product_id", productID),
	)

	existProductBasket, err := u.repo.ExistsByProductID(ctx, productID)
	if err != nil {
		u.logger.Error("failed to check product by id from basket",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return errors.New("failed to check product from basket")
	}

	if !existProductBasket {
		u.logger.Warn("product not found from basket",
			log.Int("product_id", productID),
		)

		return customError.ErrNotFoundByID(productID, "product from basket, product")
	}

	u.logger.Debug("product found",
		log.Int("product_id", productID),
	)

	err = u.repo.CleanBasketByID(ctx, clientID, productID)
	if err != nil {
		u.logger.Error("failed to delete client product from basket",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return err
	}

	u.logger.Info("client product from basket deleted successfully",
		log.Int("client_id", clientID),
	)

	return nil
}

func (u *BasketUseCase) GetAllBasketProducts(
	ctx context.Context,
	clientID int,
	limit int,
	offset int,
	baseURL string,
) (*response.BasketListResponse, error) {

	u.logger.Info("starting get all basket products",
		log.Int("client_id", clientID),
	)

	products, total, totalPrice, err := u.repo.GetAllBasketProducts(
		ctx,
		clientID,
		limit,
		offset,
	)
	if err != nil {
		u.logger.Error("failed to get basket products",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, err
	}

	result := response.BasketListResponse{Products: []response.ProductBasket{}}

	for _, product := range *products {

		if product.DiscountPrice != nil {
			discountPrice := *product.DiscountPrice
			discountPrice = math.Round(discountPrice)
			product.DiscountPrice = &discountPrice
		}

		p := response.ProductBasket{
			ID:            product.ID,
			Name:          product.Name,
			Price:         math.Round(product.Price),
			DiscountPrice: product.DiscountPrice,
			Image:         product.Image,
			IsPaused:      product.IsPaused,
			Total:         product.Total,
		}

		if baseURL != "" && p.Image != "" {
			cleanPath := strings.ReplaceAll(p.Image, "\\", "/")
			p.Image = fmt.Sprintf("%s/%s", baseURL, cleanPath)
		}

		result.Products = append(result.Products, p)
	}
	result.Total = *total
	result.TotalPrice = *totalPrice

	return &result, nil
}
