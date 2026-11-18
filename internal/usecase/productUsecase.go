package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"mime/multipart"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/models"
	"osush/internal/repository"
	minioRepo "osush/internal/repository/minio"
	"strings"

	"github.com/bytedance/gopkg/util/logger"
)

type ProductUseCase struct {
	repo              *repository.ProductRepository
	subcategoryRepo   *repository.SubcategoryRepository
	minioRepo         *minioRepo.MinioRepository
	clientRepo        *repository.ClientRepository
	productLikeRepo   *repository.ProductLikeRepository
	orderProductsRepo *repository.OrderProductRepository
	logger            log.Logger
}

func NewProductUseCase(
	repo *repository.ProductRepository,
	subcategoryRepo *repository.SubcategoryRepository,
	minioRepo *minioRepo.MinioRepository,
	clientRepo *repository.ClientRepository,
	productLikeRepo *repository.ProductLikeRepository,
	orderProductsRepo *repository.OrderProductRepository,
	logger log.Logger) *ProductUseCase {
	return &ProductUseCase{
		repo:              repo,
		subcategoryRepo:   subcategoryRepo,
		minioRepo:         minioRepo,
		clientRepo:        clientRepo,
		productLikeRepo:   productLikeRepo,
		orderProductsRepo: orderProductsRepo,
		logger:            logger,
	}
}

func (u *ProductUseCase) CreateProduct(
	ctx context.Context,
	req request.ProductRequest,
) (*int, error) {

	u.logger.Info("starting create product",
		log.Int("subcategory_id", req.SubcategoryID),
		log.String("name", req.Name),
	)

	exist, err := u.subcategoryRepo.ExistsByID(ctx, req.SubcategoryID)
	if err != nil {
		u.logger.Error("failed to check subcategory existence",
			log.Int("subcategory_id", req.SubcategoryID),
			log.Error(err),
		)

		return nil, errors.New("failed to check subcategory existence")
	}

	u.logger.Debug("subcategory existence check completed",
		log.Int("subcategory_id", req.SubcategoryID),
		log.Bool("exists", exist),
	)

	if !exist {
		u.logger.Warn("subcategory not found",
			log.Int("subcategory_id", req.SubcategoryID),
		)

		return nil, customError.ErrNotFoundByID(req.SubcategoryID, "subcategory")
	}

	var (
		attributesTmJSON []byte
		attributesRuJSON []byte
	)

	if len(req.Attributes.Tm) != 0 {
		attributesTmJSON, err = json.Marshal(req.Attributes.Tm)
		if err != nil {
			u.logger.Error("failed to marshal attributes.tm",
				log.String("name", req.Name),
				log.Error(err),
			)

			return nil, errors.New("invalid attributes.tm JSON")
		}
	}

	if len(req.Attributes.Ru) != 0 {
		attributesRuJSON, err = json.Marshal(req.Attributes.Ru)
		if err != nil {
			u.logger.Error("failed to marshal attributes.ru",
				log.String("name", req.Name),
				log.Error(err),
			)

			return nil, errors.New("invalid attributes.ru JSON")
		}
	}

	u.logger.Debug("product attributes marshaled",
		log.String("name", req.Name),
		log.Bool("has_attributes_tm", len(attributesTmJSON) > 0),
		log.Bool("has_attributes_ru", len(attributesRuJSON) > 0),
	)

	product := models.Product{
		Name:          req.Name,
		SubcategoryID: req.SubcategoryID,
		DescriptionTm: req.Description.Tm,
		DescriptionRu: req.Description.Ru,
		Price:         req.Price,
		DiscountPrice: req.DiscountPrice,
		FeaturesTm:    string(attributesTmJSON),
		FeaturesRu:    string(attributesRuJSON),
		IsPaused:      req.IsPaused,
	}

	u.logger.Debug("creating product in repository",
		log.String("name", req.Name),
		log.Int("subcategory_id", req.SubcategoryID),
	)

	productID, err := u.repo.CreateProduct(ctx, product)
	if err != nil {
		u.logger.Error("failed to create product",
			log.Int("subcategory_id", req.SubcategoryID),
			log.String("name", req.Name),
			log.Error(err),
		)

		return nil, err
	}

	u.logger.Info("product created successfully",
		log.Int("product_id", *productID),
		log.String("name", req.Name),
	)

	return productID, nil
}

func (u *ProductUseCase) AddImages(
	ctx context.Context,
	productID int,
	mainImage *multipart.FileHeader,
	images []*multipart.FileHeader,
) error {

	u.logger.Info("starting create product images",
		log.Int("product_id", productID),
		log.Int("additional_images_count", len(images)),
	)

	mainImagePath, err := GenerateFileName(mainImage, "main_img_product")
	if err != nil {
		u.logger.Error("failed to save product main image from minio",
			log.String("name", mainImage.Filename),
			log.Error(err),
		)
		return errors.New("failed to save product main image from minio")
	}

	file, err := mainImage.Open()
	if err != nil {
		u.logger.Error(
			"failed to open product main image",
			log.String("name", mainImage.Filename),
		)

		return errors.New("failed to open product main image")
	}
	defer file.Close()

	contentType := mainImage.Header.Get("Content-Type")

	if err := u.minioRepo.Upload(
		ctx,
		*mainImagePath,
		file,
		mainImage.Size,
		contentType,
	); err != nil {

		u.logger.Error(
			"failed to upload product main image to minio",
			log.String("name", mainImage.Filename),
			log.String("object_name", *mainImagePath),
			log.Error(err),
		)

		return errors.New("failed to upload product main image to minio")
	}

	logger.Debug("product main image saved from minio",
		log.String("name", mainImage.Filename),
		log.String("object_name", *mainImagePath),
	)

	var imageProducts []models.ProductImage

	imageProducts = append(imageProducts, models.ProductImage{
		ProductID: productID,
		ImagePath: *mainImagePath,
		IsMain:    true,
	})

	u.logger.Debug("main product image prepared",
		log.Int("product_id", productID),
		log.String("image_path", *mainImagePath),
	)

	for _, image := range images {

		imagePath, err := GenerateFileName(image, "product")
		if err != nil {
			u.logger.Error("failed to save product image from minio",
				log.String("name", image.Filename),
				log.Error(err),
			)
			return errors.New("failed to save product image from minio")
		}

		file, err := image.Open()
		if err != nil {
			u.logger.Error(
				"failed to open product image",
				log.String("name", image.Filename),
			)

			return errors.New("failed to open product image")
		}
		defer file.Close()

		contentType := image.Header.Get("Content-Type")

		if err := u.minioRepo.Upload(
			ctx,
			*imagePath,
			file,
			image.Size,
			contentType,
		); err != nil {

			u.logger.Error(
				"failed to upload product  image to minio",
				log.String("name", image.Filename),
				log.String("object_name", *imagePath),
				log.Error(err),
			)

			return errors.New("failed to upload product main image to minio")
		}

		imageProducts = append(imageProducts, models.ProductImage{
			ProductID: productID,
			ImagePath: *imagePath,
			IsMain:    false,
		})
	}

	u.logger.Debug("product images prepared",
		log.Int("product_id", productID),
		log.Int("total_images_count", len(imageProducts)),
		log.Int("additional_images_count", len(images)),
	)

	err = u.repo.CreateProductImages(ctx, imageProducts)
	if err != nil {
		u.logger.Error("failed to create product images",
			log.Int("product_id", productID),
			log.Int("total_images_count", len(imageProducts)),
			log.Error(err),
		)

		return err
	}

	u.logger.Info("product images created successfully",
		log.Int("product_id", productID),
		log.Int("total_images_count", len(imageProducts)),
	)

	return nil
}

func (u *ProductUseCase) GetProductByID(
	ctx context.Context,
	id int,
	clientID int,
	baseURL string,
) (*response.ProductResponse, error) {

	u.logger.Info("starting get product by id",
		log.Int("product_id", id),
	)

	product, err := u.repo.GetProductByID(ctx, id)
	if err != nil {
		u.logger.Error("failed to get product",
			log.Int("product_id", id),
			log.Error(err),
		)

		return nil, err
	}

	u.logger.Debug("product retrieved",
		log.Int("product_id", product.ID),
		log.String("name", product.Name),
	)

	productImages, err := u.repo.GetProductImages(ctx, id)
	if err != nil {
		u.logger.Error("failed to get product images",
			log.Int("product_id", id),
			log.Error(err),
		)

		return nil, err
	}

	images := make([]string, 0, len(productImages))

	for _, image := range productImages {

		if baseURL != "" && image.ImagePath != "" {
			cleanPath := strings.ReplaceAll(image.ImagePath, "\\", "/")
			image.ImagePath = fmt.Sprintf("%s/%s", baseURL, cleanPath)
		}

		images = append(images, image.ImagePath)
	}

	var description *response.Dictionary
	var features *request.DictionaryMap

	if product.DescriptionTm != "" || product.DescriptionRu != "" {
		description = &response.Dictionary{
			Tm: product.DescriptionTm,
			Ru: product.DescriptionRu,
		}
	}

	if product.FeaturesTm != "" || product.FeaturesRu != "" {
		features = &request.DictionaryMap{}

		if product.FeaturesTm != "" {
			err = json.Unmarshal([]byte(product.FeaturesTm), &features.Tm)
			if err != nil {
				u.logger.Error("error unmarshal attributes tm",
					log.Error(err))
				return nil, errors.New("error unmarshal attributes tm")
			}
		}

		if product.FeaturesRu != "" {
			err = json.Unmarshal([]byte(product.FeaturesRu), &features.Ru)
			if err != nil {
				u.logger.Error("error unmarshal attributes ru",
					log.Error(err))
				return nil, errors.New("error unmarshal attributes ru")
			}
		}
	}

	if product.DiscountPrice != nil {
		discountPrice := *product.DiscountPrice
		discountPrice = math.Round(discountPrice)
		product.DiscountPrice = &discountPrice
	}

	var isLiked bool

	if clientID > 0 {
		u.logger.Info("checking product like",
			log.Int("client_id", clientID),
			log.Int("product_id", id),
		)

		isLiked, err = u.productLikeRepo.IsProductLiked(
			ctx,
			clientID,
			id,
		)

		if err != nil {
			u.logger.Error("failed to check product like",
				log.Int("client_id", clientID),
				log.Int("product_id", id),
				log.Error(err),
			)

			return nil, err
		}
	}

	result := &response.ProductResponse{
		ID:            product.ID,
		Name:          product.Name,
		IsLiked:       isLiked,
		Description:   description,
		Price:         math.Round(product.Price),
		DiscountPrice: product.DiscountPrice,
		Attributes:    features,
		IsPaused:      product.IsPaused,
		Images:        images,
		Category: &response.ProductCategory{
			ID:   product.CategoryID,
			Name: product.CategoryName,
		},
		Subcategory: &response.ProductCategory{
			ID:   product.SubcategoryID,
			Name: product.SubcategoryName,
		},
	}

	// Images are ordered cover first, the same rule the product list uses.
	if len(images) > 0 {
		result.Image = images[0]
	}

	u.logger.Info("product retrieved successfully",
		log.Int("product_id", product.ID),
		log.String("name", product.Name),
		log.Int("images_count", len(images)),
	)

	return result, nil
}

func (u *ProductUseCase) GetProductList(
	ctx context.Context,
	filter request.ProductFilter,
	clientID int,
	baseURL string,
) (*response.ProductListResponse, error) {

	u.logger.Info("starting get all products")

	products, total, err := u.repo.GetProductList(ctx, filter)
	if err != nil {
		u.logger.Error("failed to get products",
			log.Error(err),
		)

		return nil, err
	}

	u.logger.Debug("products retrieved from repository",
		log.Int("count", len(products)),
	)

	result := response.ProductListResponse{
		Products: []response.Products{},
		Total:    *total,
	}

	for _, product := range products {

		productImages, err := u.repo.GetProductImages(ctx, product.ID)
		if err != nil {
			u.logger.Error("failed to get product images",
				log.Int("product_id", product.ID),
				log.Error(err),
			)

			return nil, err
		}

		if product.DiscountPrice != nil {
			discountPrice := *product.DiscountPrice
			discountPrice = math.Round(discountPrice)
			product.DiscountPrice = &discountPrice
		}

		var isLiked bool

		if clientID > 0 {
			u.logger.Info("checking product like",
				log.Int("client_id", clientID),
				log.Int("product_id", product.ID),
			)

			isLiked, err = u.productLikeRepo.IsProductLiked(
				ctx,
				clientID,
				product.ID,
			)

			if err != nil {
				u.logger.Error("failed to check product like",
					log.Int("client_id", clientID),
					log.Int("product_id", product.ID),
					log.Error(err),
				)

				return nil, err
			}
		}

		orderCount, err := u.orderProductsRepo.GetProductOrderCount(ctx, product.ID)

		if err != nil {
			u.logger.Error("failed to get order count",
				log.Int("client_id", clientID),
				log.Int("product_id", product.ID),
				log.Error(err),
			)

			return nil, err
		}
		createdAt := product.CreatedAt.Format("02-01-2006")

		var imagePath string
		if len(productImages) > 0 {
			imagePath = productImages[0].ImagePath
		}

		product := response.Products{
			ID:            product.ID,
			Name:          product.Name,
			IsLiked:       isLiked,
			Price:         math.Round(product.Price),
			DiscountPrice: product.DiscountPrice,
			Image:         imagePath,
			OrderCount:    *orderCount,
			IsPaused:      product.IsPaused,
			CreatedAt:     createdAt,
		}

		if baseURL != "" && product.Image != "" {
			cleanPath := strings.ReplaceAll(product.Image, "\\", "/")
			product.Image = fmt.Sprintf("%s/%s", baseURL, cleanPath)
		}

		result.Products = append(result.Products, product)
	}

	u.logger.Info("products retrieved successfully",
		log.Int("count", len(result.Products)),
	)

	return &result, nil
}

func (u *ProductUseCase) GetIndexProducts(
	ctx context.Context,
	baseURL string,
	clientID int,
) (*response.IndexProductResponse, error) {

	u.logger.Info("starting get index products")

	products, newProducts, err := u.repo.GetIndexProducts(ctx)
	if err != nil {
		u.logger.Error("failed to get index products",
			log.Error(err),
		)

		return nil, err
	}

	result := &response.IndexProductResponse{
		Products:    make([]response.ProductIndex, 0, len(*products)),
		NewProducts: make([]response.ProductIndex, 0, len(*newProducts)),
	}

	for _, product := range *products {

		if product.DiscountPrice != nil {
			discountPrice := *product.DiscountPrice
			discountPrice = math.Round(discountPrice)
			product.DiscountPrice = &discountPrice
		}

		var isLiked bool

		if clientID > 0 {
			u.logger.Info("checking product like",
				log.Int("client_id", clientID),
				log.Int("product_id", product.ID),
			)

			isLiked, err = u.productLikeRepo.IsProductLiked(
				ctx,
				clientID,
				product.ID,
			)

			if err != nil {
				u.logger.Error("failed to check product like",
					log.Int("client_id", clientID),
					log.Int("product_id", product.ID),
					log.Error(err),
				)

				return nil, err
			}
		}

		p := response.ProductIndex{
			ID:            product.ID,
			Name:          product.Name,
			IsLiked:       isLiked,
			Price:         math.Round(product.Price),
			DiscountPrice: product.DiscountPrice,
			Image:         product.MainImage,
			IsPaused:      product.IsPaused,
		}

		if baseURL != "" && p.Image != "" {
			cleanPath := strings.ReplaceAll(p.Image, "\\", "/")
			p.Image = fmt.Sprintf("%s/%s", baseURL, cleanPath)
		}

		result.Products = append(result.Products, p)
	}

	for _, product := range *newProducts {

		if product.DiscountPrice != nil {
			discountPrice := *product.DiscountPrice
			discountPrice = math.Round(discountPrice)
			product.DiscountPrice = &discountPrice
		}

		var isLiked bool

		if clientID > 0 {
			u.logger.Info("checking product like",
				log.Int("client_id", clientID),
				log.Int("product_id", product.ID),
			)

			isLiked, err = u.productLikeRepo.IsProductLiked(
				ctx,
				clientID,
				product.ID,
			)

			if err != nil {
				u.logger.Error("failed to check product like",
					log.Int("client_id", clientID),
					log.Int("product_id", product.ID),
					log.Error(err),
				)

				return nil, err
			}
		}

		p := response.ProductIndex{
			ID:            product.ID,
			Name:          product.Name,
			IsLiked:       isLiked,
			Price:         math.Round(product.Price),
			DiscountPrice: product.DiscountPrice,
			Image:         product.MainImage,
			IsPaused:      product.IsPaused,
		}

		if baseURL != "" && p.Image != "" {
			cleanPath := strings.ReplaceAll(p.Image, "\\", "/")
			p.Image = fmt.Sprintf("%s/%s", baseURL, cleanPath)
		}

		result.NewProducts = append(result.NewProducts, p)
	}

	u.logger.Info("index products retrieved successfully",
		log.Int("products_count", len(result.Products)),
		log.Int("new_products_count", len(result.NewProducts)),
	)

	return result, nil
}

func (u *ProductUseCase) UpdateProduct(
	ctx context.Context,
	productID int,
	req request.UpdateProductRequest,
	mainImage *multipart.FileHeader,
	additionalImages []*multipart.FileHeader,
) error {

	u.logger.Info("starting update product",
		log.Int("product_id", productID),
	)

	exists, err := u.repo.ExistsByID(ctx, productID)
	if err != nil {
		u.logger.Error("failed to check product existence",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return err
	}

	if !exists {
		u.logger.Warn("product not found",
			log.Int("product_id", productID),
		)

		return customError.ErrNotFoundByID(productID, "product")
	}

	if req.SubcategoryID > 0 {
		exists, err := u.subcategoryRepo.ExistsByID(ctx, req.SubcategoryID)
		if err != nil {
			u.logger.Error("failed to check subcategory existence",
				log.Int("subcategory_id", req.SubcategoryID),
				log.Error(err),
			)

			return errors.New("failed to check subcategory existence")
		}

		if !exists {
			u.logger.Warn("subcategory not found",
				log.Int("subcategory_id", req.SubcategoryID),
			)

			return customError.ErrNotFoundByID(req.SubcategoryID, "subcategory")
		}
	}

	// The request carries the product's complete new set of images. The old
	// ones are removed only after the new ones are stored, so a failed upload
	// never leaves the product without images.
	oldImages, err := u.repo.GetProductImages(ctx, productID)
	if err != nil {
		u.logger.Error("failed to get old product images",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return err
	}

	var (
		attributesTmJSON []byte
		attributesRuJSON []byte
	)

	if len(req.Attributes.Tm) != 0 {
		attributesTmJSON, err = json.Marshal(req.Attributes.Tm)
		if err != nil {
			u.logger.Error("failed to marshal attributes.tm",
				log.String("name", req.Name),
				log.Error(err),
			)

			return errors.New("invalid attributes.tm JSON")
		}
	}

	if len(req.Attributes.Ru) != 0 {
		attributesRuJSON, err = json.Marshal(req.Attributes.Ru)
		if err != nil {
			u.logger.Error("failed to marshal attributes.ru",
				log.String("name", req.Name),
				log.Error(err),
			)

			return errors.New("invalid attributes.ru JSON")
		}
	}

	u.logger.Debug("product attributes marshaled",
		log.String("name", req.Name),
		log.Bool("has_attributes_tm", len(attributesTmJSON) > 0),
		log.Bool("has_attributes_ru", len(attributesRuJSON) > 0),
	)

	product := models.Product{
		Name:          req.Name,
		SubcategoryID: req.SubcategoryID,
		DescriptionTm: req.Description.Tm,
		DescriptionRu: req.Description.Ru,
		Price:         req.Price,
		DiscountPrice: req.DiscountPrice,
		FeaturesTm:    string(attributesTmJSON),
		FeaturesRu:    string(attributesRuJSON),
		IsPaused:      req.IsPaused,
	}

	if err := u.repo.UpdateProduct(ctx, productID, product); err != nil {
		u.logger.Error("failed to update product",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return err
	}

	if err := u.AddImages(
		ctx,
		productID,
		mainImage,
		additionalImages,
	); err != nil {

		u.logger.Error("failed to add new product images",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return err
	}

	u.removeOldImages(ctx, productID, oldImages)

	u.logger.Info("product updated successfully",
		log.Int("product_id", productID),
	)

	return nil
}

// removeOldImages deletes images that were replaced by an update. The new
// images are already stored, so a failure here only leaves an orphaned file
// and is logged instead of failing the request.
func (u *ProductUseCase) removeOldImages(
	ctx context.Context,
	productID int,
	oldImages []models.ProductImage,
) {
	if len(oldImages) == 0 {
		return
	}

	ids := make([]int, 0, len(oldImages))
	for _, image := range oldImages {
		ids = append(ids, image.ID)
	}

	if err := u.repo.DeleteProductImagesByIDs(ctx, productID, ids); err != nil {
		u.logger.Error("failed to delete replaced product images",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return
	}

	for _, image := range oldImages {
		if err := u.minioRepo.Delete(ctx, image.ImagePath); err != nil {
			u.logger.Warn("failed to delete replaced product image from minio",
				log.Int("product_id", productID),
				log.String("image_path", image.ImagePath),
				log.Error(err),
			)
		}
	}
}

func (u *ProductUseCase) DeleteProduct(
	ctx context.Context,
	productID int,
) error {

	u.logger.Info("starting delete product",
		log.Int("product_id", productID),
	)

	exists, err := u.repo.ExistsByID(ctx, productID)
	if err != nil {
		u.logger.Error("failed to check product existence",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return err
	}

	if !exists {
		u.logger.Warn("product not found",
			log.Int("product_id", productID),
		)

		return customError.ErrNotFoundByID(
			productID,
			"product",
		)
	}

	images, err := u.repo.GetProductImages(
		ctx,
		productID,
	)
	if err != nil {
		u.logger.Error("failed to get product images",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return err
	}

	if err := u.repo.DeleteProduct(
		ctx,
		productID,
	); err != nil {

		u.logger.Error("failed to delete product",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return err
	}

	u.logger.Info("product deleted successfully",
		log.Int("product_id", productID),
	)

	for _, image := range images {

		if image.ImagePath == "" {
			continue
		}

		if err := u.minioRepo.Delete(
			ctx,
			image.ImagePath,
		); err != nil {

			u.logger.Error("failed to delete product image from minio",
				log.Int("product_id", productID),
				log.String("image_path", image.ImagePath),
				log.Error(err),
			)

			return err
		}
	}
	u.logger.Info("product images deleted successfully from minio",
		log.Int("product_id", productID),
	)

	return nil
}
