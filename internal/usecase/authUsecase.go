package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"mime/multipart"
	"osush/internal/config"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/repository"
	minioRepo "osush/internal/repository/minio"
	"strings"

	"github.com/bytedance/gopkg/util/logger"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	repo              *repository.UserRepository
	minioRepo         *minioRepo.MinioRepository
	categoryRepo      *repository.CategoryRepository
	productRepo       *repository.ProductRepository
	clientRepo        *repository.ClientRepository
	orderProductsRepo *repository.OrderProductRepository
	logger            log.Logger
}

func NewAuthUseCase(
	repo *repository.UserRepository,
	minioRepo *minioRepo.MinioRepository,
	categoryRepo *repository.CategoryRepository,
	productRepo *repository.ProductRepository,
	clientRepo *repository.ClientRepository,
	orderProductsRepo *repository.OrderProductRepository,
	logger log.Logger) *AuthUseCase {
	return &AuthUseCase{
		repo:              repo,
		minioRepo:         minioRepo,
		categoryRepo:      categoryRepo,
		productRepo:       productRepo,
		clientRepo:        clientRepo,
		orderProductsRepo: orderProductsRepo,
		logger:            logger,
	}
}

func (u *AuthUseCase) Login(ctx context.Context, cfg *config.Config, req request.LoginRequest) (*string, error) {

	u.logger.Info("starting login",
		log.String("username", req.UserName),
	)

	user, err := u.repo.GetByName(ctx, req.UserName)
	if err != nil {

		if customError.ErrNotFound("user").Is(err) {
			u.logger.Warn("authentication failed",
				log.String("user_name", req.UserName),
			)

			return nil, customError.ErrUnauthorized
		}

		u.logger.Warn(
			"failed to find user for login",
			log.String("username", req.UserName),
			log.Error(err),
		)

		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	); err != nil {
		u.logger.Warn("authentication failed",
			log.String("username", req.UserName),
		)

		return nil, customError.ErrUnauthorized
	}

	token, err := GenerateTokenPair(cfg, user.ID, "ADMIN")
	if err != nil {
		u.logger.Error(
			"failed to generate token",
			log.Int("user_id", user.ID),
			log.String("username", user.Name),
			log.Error(err),
		)

		return nil, err
	}

	u.logger.Info(
		"login successful",
		log.Int("user_id", user.ID),
		log.String("username", user.Name),
	)

	return &token, nil
}

func (u *AuthUseCase) UpdateProfile(
	ctx context.Context,
	userID int,
	req request.UpdateProfileRequest,
	avatar *multipart.FileHeader,
	baseURL string,
) (*response.UserProfileResponse, error) {

	u.logger.Info(
		"starting admin profile update",
	)

	exists, err := u.repo.ExistsByID(ctx, userID)
	if err != nil {
		u.logger.Error("failed to check user existence",
			log.Error(err),
		)

		return nil, err
	}

	if !exists {
		u.logger.Warn("user not found")

		return nil, customError.ErrNotFoundByID(userID, "user")
	}

	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Name = req.Name

	if req.Surname != "" {
		user.Surname = &req.Surname
	} else {
		user.Surname = nil
	}

	if req.PhoneNumber != "" {
		user.PhoneNumber = &req.PhoneNumber
	} else {
		user.PhoneNumber = nil
	}

	if req.NewPassword != "" {

		if req.CurrentPassword == "" {
			return nil, customError.ErrBadRequest("current password is required")
		}

		err := bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(req.CurrentPassword),
		)

		if err != nil {
			u.logger.Warn("incorrect current password")
			return nil, customError.ErrBadRequest("incorrect curent password")
		}

		hashedPassword, err := HashPassword(req.NewPassword)
		if err != nil {
			return nil, err
		}

		user.Password = string(hashedPassword)
	}

	oldAvatarPath := user.Avatar

	if avatar != nil {
		imagePath, err := GenerateFileName(avatar, "avatar")
		if err != nil {
			u.logger.Error("failed to save admin avatar from minio",
				log.String("name", req.Name),
				log.Error(err),
			)
			return nil, errors.New("failed to save admin avatar from minio")
		}

		file, err := avatar.Open()
		if err != nil {
			u.logger.Error(
				"failed to open avatar",
				log.String("name", req.Name),
				log.Error(err),
			)

			return nil, errors.New("failed to open avatar")
		}
		defer file.Close()

		contentType := avatar.Header.Get("Content-Type")

		if err := u.minioRepo.Upload(
			ctx,
			*imagePath,
			file,
			avatar.Size,
			contentType,
		); err != nil {

			u.logger.Error(
				"failed to upload avatar to minio",
				log.String("name", req.Name),
				log.Error(err),
			)

			return nil, errors.New("failed to upload avatar to minio")
		}

		logger.Debug("avatar saved from minio",
			log.String("name", req.Name),
		)

		user.Avatar = imagePath
	} else {
		user.Avatar = nil
	}

	err = u.repo.Update(ctx, userID, user)
	if err != nil {
		return nil, err
	}

	if oldAvatarPath != nil {
		if err := u.minioRepo.Delete(
			ctx,
			*oldAvatarPath,
		); err != nil {

			u.logger.Error("failed to delete old product image from minio",
				log.Error(err),
			)

			return nil, err
		}
	}

	u.logger.Info("admin profile updated successfully")

	newProfile, err := u.GetProfile(ctx, userID, baseURL)
	if err != nil {
		return nil, err
	}

	return newProfile, nil
}

func (u *AuthUseCase) GetProfile(
	ctx context.Context,
	userID int,
	baseURL string,
) (*response.UserProfileResponse, error) {

	u.logger.Info("starting get user profile")

	exists, err := u.repo.ExistsByID(ctx, userID)
	if err != nil {
		u.logger.Error("failed to check user existence",
			log.Error(err),
		)

		return nil, err
	}

	if !exists {
		u.logger.Warn("user not found")

		return nil, customError.ErrNotFoundByID(userID, "user")
	}

	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		u.logger.Error("failed to get client profile",
			log.Error(err),
		)

		return nil, err
	}

	var imagePath *string

	if user.Avatar != nil {
		cleanPath := strings.ReplaceAll(*user.Avatar, "\\", "/")

		fullPath := fmt.Sprintf("%s/%s", baseURL, cleanPath)

		imagePath = &fullPath
	}

	u.logger.Info("client profile retrieved successfully")

	return &response.UserProfileResponse{
		ID:          user.ID,
		UserName:    user.Name,
		Surname:     user.Surname,
		PhoneNumber: user.PhoneNumber,
		Avatar:      imagePath,
	}, nil
}

func (u *AuthUseCase) GetAdminIndexPage(
	ctx context.Context,
	baseURL string,
) (*response.AdminIndexPageResponse, error) {

	u.logger.Info("starting get admin index page")

	var result response.AdminIndexPageResponse

	u.logger.Info("getting total categories")

	totalCategory, err := u.categoryRepo.TotalCount(ctx)
	if err != nil {
		u.logger.Error("failed to get total categories",
			log.Error(err),
		)

		return nil, err
	}

	result.TotalCategories = *totalCategory

	u.logger.Info("getting clients")

	names, totalClients, err := u.clientRepo.GetNameClients(ctx, 10, 1)
	if err != nil {
		u.logger.Error("failed to get clients",
			log.Error(err),
		)

		return nil, err
	}

	result.TotalClients = *totalClients

	if len(*names) > 0 {
		result.Clients = *names
	} else {
		result.Clients = []string{}
	}

	u.logger.Info("getting products")

	filter := request.ProductFilter{
		LimitInt:  10,
		OffsetInt: 1,
	}

	products, totalProducts, err := u.productRepo.GetProductList(ctx, filter)
	if err != nil {
		u.logger.Error("failed to get products",
			log.Error(err),
		)

		return nil, err
	}

	result.TotalProducts = *totalProducts

	for _, product := range products {

		u.logger.Info("processing product",
			log.Int("product_id", product.ID),
		)

		productImages, err := u.productRepo.GetProductImages(ctx, product.ID)
		if err != nil {
			u.logger.Error("failed to get product images",
				log.Int("product_id", product.ID),
				log.Error(err),
			)

			return nil, err
		}

		var imagePath string

		if len(productImages) > 0 {
			imagePath = productImages[0].ImagePath
		}

		if product.DiscountPrice != nil {
			discountPrice := math.Round(*product.DiscountPrice)
			product.DiscountPrice = &discountPrice
		}

		orderCount, err := u.orderProductsRepo.GetProductOrderCount(ctx, product.ID)
		if err != nil {
			u.logger.Error("failed to get product order count",
				log.Int("product_id", product.ID),
				log.Error(err),
			)

			return nil, err
		}

		productResponse := response.AdminProductsResponse{
			ID:            product.ID,
			Name:          product.Name,
			Price:         math.Round(product.Price),
			DiscountPrice: product.DiscountPrice,
			Image:         imagePath,
			OrderCount:    *orderCount,
		}

		if baseURL != "" && productResponse.Image != "" {
			cleanPath := strings.ReplaceAll(
				productResponse.Image,
				"\\",
				"/",
			)

			productResponse.Image = fmt.Sprintf(
				"%s/%s",
				strings.TrimRight(baseURL, "/"),
				strings.TrimLeft(cleanPath, "/"),
			)
		}

		result.Products = append(result.Products, productResponse)
	}

	if len(result.Products) == 0 {
		result.Products = []response.AdminProductsResponse{}
	}

	u.logger.Info("admin index page retrieved successfully",
		log.Int("total_categories", result.TotalCategories),
		log.Int("total_clients", result.TotalClients),
		log.Int("total_products", result.TotalProducts),
	)

	return &result, nil
}
