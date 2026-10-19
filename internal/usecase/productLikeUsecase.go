package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	customError "osush/internal/errors"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/repository"
	"strings"
)

type ProductLikeUseCase struct {
	repo        *repository.ProductLikeRepository
	clientRepo  *repository.ClientRepository
	productRepo *repository.ProductRepository
	logger      log.Logger
}

func NewProductLikeUseCase(
	repo *repository.ProductLikeRepository,
	clientRepo *repository.ClientRepository,
	productRepo *repository.ProductRepository,
	logger log.Logger) *ProductLikeUseCase {
	return &ProductLikeUseCase{
		repo:        repo,
		clientRepo:  clientRepo,
		productRepo: productRepo,
		logger:      logger,
	}
}

func (u *ProductLikeUseCase) ToggleLikeProduct(
	ctx context.Context,
	clientID int,
	productID int,
) (bool, error) {

	u.logger.Info("starting toggle product like",
		log.Int("client_id", clientID),
		log.Int("product_id", productID),
	)

	existClient, err := u.clientRepo.ExistsByID(ctx, clientID)
	if err != nil {
		u.logger.Error("failed to check client by id",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return false, errors.New("failed to check client")
	}

	if !existClient {
		u.logger.Warn("client not found",
			log.Int("client_id", clientID),
		)

		return false, customError.ErrNotFound("client")
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

		return false, errors.New("failed to check product")
	}

	if !existClientProduct {
		u.logger.Warn("product not found",
			log.Int("product_id", productID),
		)

		return false, customError.ErrNotFoundByID(productID, "product")
	}

	u.logger.Debug("product found",
		log.Int("product_id", productID),
	)

	isLiked, err := u.repo.ToggleLikeProduct(
		ctx,
		clientID,
		productID,
	)

	if err != nil {
		u.logger.Error("failed to toggle product like",
			log.Int("client_id", clientID),
			log.Int("product_id", productID),
			log.Error(err),
		)

		return false, err
	}

	u.logger.Info("product like toggled successfully",
		log.Int("client_id", clientID),
		log.Int("product_id", productID),
		log.Bool("is_liked", isLiked),
	)

	return isLiked, nil
}

func (u *ProductLikeUseCase) GetLikedProducts(
	ctx context.Context,
	clientID int,
	limitInt int,
	offsetInt int,
	baseURL string,
) (*response.ProductListResponse, error) {

	u.logger.Info("starting get liked products",
		log.Int("client_id", clientID),
	)

	products, total, err := u.repo.GetLikedProducts(ctx, clientID, limitInt, offsetInt)
	if err != nil {
		u.logger.Error("failed to get liked products",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, err
	}

	result := response.ProductListResponse{Total: *total, Products: []response.Products{}}

	for _, product := range products {

		if product.DiscountPrice != nil {
			discountPrice := *product.DiscountPrice
			discountPrice = math.Round(discountPrice)
			product.DiscountPrice = &discountPrice
		}

		p := response.Products{
			ID:            product.ID,
			Name:          product.Name,
			IsLiked:       true,
			Price:         math.Round(product.Price),
			DiscountPrice: product.DiscountPrice,
			Image:         product.MainImage,
		}

		if baseURL != "" && p.Image != "" {
			cleanPath := strings.ReplaceAll(p.Image, "\\", "/")
			p.Image = fmt.Sprintf("%s/%s", baseURL, cleanPath)
		}

		result.Products = append(result.Products, p)
	}

	return &result, nil
}
