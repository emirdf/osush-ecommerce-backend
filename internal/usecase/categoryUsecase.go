package usecase

import (
	"context"
	"errors"
	"fmt"
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

type CategoryUseCase struct {
	repo            *repository.CategoryRepository
	subcategoryRepo *repository.SubcategoryRepository
	minioRepo       *minioRepo.MinioRepository
	logger          log.Logger
}

func NewCategoryUseCase(
	repo *repository.CategoryRepository,
	subcategoryRepo *repository.SubcategoryRepository,
	minioRepo *minioRepo.MinioRepository,
	logger log.Logger) *CategoryUseCase {
	return &CategoryUseCase{
		repo:            repo,
		subcategoryRepo: subcategoryRepo,
		minioRepo:       minioRepo,
		logger:          logger,
	}
}

func (u *CategoryUseCase) CreateCategory(
	ctx context.Context,
	req request.CategoryRequest,
	image *multipart.FileHeader,
	contentType string,
	baseURL string,
) (*response.CategoryResponse, error) {
	logger.Info("starting save upload category file from minio",
		log.String("name", req.Name),
	)

	imagePath, err := GenerateFileName(image, "category")
	if err != nil {
		u.logger.Error("failed to save category image from minio",
			log.String("name", req.Name),
			log.Error(err),
		)
		return nil, errors.New("failed to save category image from minio")
	}

	file, err := image.Open()
	if err != nil {
		u.logger.Error(
			"failed to open category image",
			log.String("name", req.Name),
			log.Error(err),
		)

		return nil, errors.New("failed to open category image")
	}
	defer file.Close()

	if err := u.minioRepo.Upload(
		ctx,
		*imagePath,
		file,
		image.Size,
		contentType,
	); err != nil {

		u.logger.Error(
			"failed to upload category image to minio",
			log.String("name", req.Name),
			log.String("object_name", *imagePath),
			log.Error(err),
		)

		return nil, errors.New("failed to upload category image to minio")
	}

	logger.Debug("category image saved from minio",
		log.String("name", req.Name),
		log.String("image_path", *imagePath),
	)

	u.logger.Info("starting create category",
		log.String("name", req.Name),
	)

	exist, err := u.repo.ExistsByName(ctx, req)
	if err != nil {
		u.logger.Error("failed to check exist category by name to database",
			log.String("name", req.Name),
			log.Error(err),
		)
		return nil, errors.New("failed to check exist category by name to database")
	}

	if exist {
		u.logger.Warn("category already exists",
			log.String("name", req.Name),
		)

		return nil, customError.ErrAlreadyData("category")
	}

	u.logger.Debug("category existence check completed",
		log.String("name", req.Name),
		log.Bool("exists", exist),
	)

	category := models.Category{
		Name:      req.Name,
		ImagePath: *imagePath,
	}

	categoryID, err := u.repo.CreateCategory(ctx, category)
	if err != nil {
		u.logger.Error("failed to create category",
			log.String("name", req.Name),
			log.Error(err),
		)

		return nil, err
	}

	newCategory, err := u.GetCategoryByID(ctx, *categoryID, baseURL)
	if err != nil {
		return nil, err
	}

	return newCategory, nil
}

func (u *CategoryUseCase) GetCategoryByID(
	ctx context.Context,
	categoryID int,
	baseURL string,
) (*response.CategoryResponse, error) {

	u.logger.Info("starting get category by id",
		log.Int("id", categoryID),
	)

	category, err := u.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		u.logger.Error("failed to get category",
			log.Int("id", categoryID),
			log.Error(err),
		)

		return nil, err
	}

	if baseURL != "" && category.ImagePath != "" {
		cleanPath := strings.ReplaceAll(category.ImagePath, "\\", "/")
		category.ImagePath = fmt.Sprintf("%s/%s", baseURL, cleanPath)
	}

	resp := response.CategoryResponse{
		Id:        category.ID,
		Name:      category.Name,
		ImagePath: category.ImagePath,
	}

	return &resp, nil
}

func (u *CategoryUseCase) GetCategoryList(
	ctx context.Context,
	search string,
	limit int,
	offset int,
	baseURL string,
) (*response.CategoryListResponse, error) {

	u.logger.Info("starting get all categories",
		log.String("search", search),
		log.Int("limit", limit),
		log.Int("offset", offset),
	)

	categories, total, err := u.repo.GetCategoryList(
		ctx,
		search,
		limit,
		offset,
		baseURL,
	)
	if err != nil {
		u.logger.Error("failed to get all categories",
			log.String("search", search),
			log.Int("limit", limit),
			log.Int("offset", offset),
			log.Error(err),
		)

		return nil, err
	}

	u.logger.Info("categories fetched successfully",
		log.Int("count", len(*categories)),
	)

	resp := response.CategoryListResponse{Categories: []response.CategoryResponse{}}
	for _, category := range *categories {

		if baseURL != "" && category.ImagePath != "" {
			cleanPath := strings.ReplaceAll(category.ImagePath, "\\", "/")
			category.ImagePath = fmt.Sprintf("%s/%s", baseURL, cleanPath)
		}

		resp.Categories = append(resp.Categories,
			response.CategoryResponse{
				Id:        category.ID,
				Name:      category.Name,
				ImagePath: category.ImagePath,
			})
	}
	resp.Total = *total

	return &resp, nil
}

func (u *CategoryUseCase) UpdateCategory(
	ctx context.Context,
	categoryID int,
	req request.CategoryRequest,
	image *multipart.FileHeader,
	contentType string,
	baseURL string,
) (*response.CategoryResponse, error) {

	u.logger.Info("starting update category",
		log.Int("category_id", categoryID),
		log.String("name", req.Name),
	)

	exists, err := u.repo.ExistsByID(ctx, categoryID)
	if err != nil {
		u.logger.Error("failed to check category existence",
			log.Int("category_id", categoryID),
			log.Error(err),
		)

		return nil, err
	}

	if !exists {
		u.logger.Warn("category not found",
			log.Int("category_id", categoryID),
		)

		return nil, customError.ErrNotFoundByID(categoryID, "category")
	}

	oldCategory, err := u.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		u.logger.Error("failed to get category before update",
			log.Int("category_id", categoryID),
			log.Error(err),
		)

		return nil, err
	}

	exist, err := u.repo.ExistsByIDAndName(
		ctx,
		req.Name,
		categoryID,
	)
	if err != nil {
		u.logger.Error("failed to check category name",
			log.Int("category_id", categoryID),
			log.String("name", req.Name),
			log.Error(err),
		)

		return nil, errors.New("failed to check category name")
	}

	if exist {
		u.logger.Warn("category name already exists",
			log.Int("category_id", categoryID),
			log.String("name", req.Name),
		)

		return nil, customError.ErrAlreadyData("category")
	}

	newImagePath := oldCategory.ImagePath

	if image != nil {

		u.logger.Debug("starting upload new category image",
			log.Int("category_id", categoryID),
			log.String("name", req.Name),
		)

		imagePath, err := GenerateFileName(image, "category")
		if err != nil {
			u.logger.Error("failed to generate category image name",
				log.Int("category_id", categoryID),
				log.Error(err),
			)

			return nil, errors.New("failed to generate category image name")
		}

		file, err := image.Open()
		if err != nil {
			u.logger.Error("failed to open category image",
				log.Int("category_id", categoryID),
				log.Error(err),
			)

			return nil, errors.New("failed to open category image")
		}
		defer file.Close()

		if err := u.minioRepo.Upload(
			ctx,
			*imagePath,
			file,
			image.Size,
			contentType,
		); err != nil {

			u.logger.Error("failed to upload new category image",
				log.Int("category_id", categoryID),
				log.String("object_name", *imagePath),
				log.Error(err),
			)

			return nil, errors.New("failed to upload new category image")
		}

		newImagePath = *imagePath

		u.logger.Debug("new category image uploaded",
			log.Int("category_id", categoryID),
			log.String("image_path", newImagePath),
		)
	}

	category := models.Category{
		Name:      req.Name,
		ImagePath: newImagePath,
	}

	if err := u.repo.UpdateCategory(ctx, categoryID, category); err != nil {
		u.logger.Error("failed to update category",
			log.Int("category_id", categoryID),
			log.Error(err),
		)

		if image != nil {
			if deleteErr := u.minioRepo.Delete(ctx, newImagePath); deleteErr != nil {
				u.logger.Error("failed to delete uploaded category image after update failure",
					log.Int("category_id", categoryID),
					log.String("object_name", newImagePath),
					log.Error(deleteErr),
				)
			}
		}

		return nil, err
	}

	if image != nil && oldCategory.ImagePath != "" {
		if err := u.minioRepo.Delete(ctx, oldCategory.ImagePath); err != nil {
			u.logger.Error("failed to delete old category image",
				log.Int("category_id", categoryID),
				log.String("object_name", oldCategory.ImagePath),
				log.Error(err),
			)
		}
	}

	newCategory, err := u.GetCategoryByID(
		ctx,
		categoryID,
		baseURL,
	)
	if err != nil {
		return nil, err
	}

	u.logger.Info("category updated successfully",
		log.Int("category_id", categoryID),
	)

	return newCategory, nil
}

func (u *CategoryUseCase) DeleteCategory(
	ctx context.Context,
	categoryID int,
) error {

	u.logger.Info("starting delete category",
		log.Int("category_id", categoryID),
	)

	exists, err := u.repo.ExistsByID(ctx, categoryID)
	if err != nil {
		u.logger.Error("failed to check category existence",
			log.Int("category_id", categoryID),
			log.Error(err),
		)

		return err
	}

	if !exists {
		u.logger.Warn("category not found",
			log.Int("category_id", categoryID),
		)

		return customError.ErrNotFoundByID(categoryID, "category")
	}

	existsSubcategory, err := u.subcategoryRepo.ExistsByCategoryID(ctx, categoryID)
	if err != nil {
		u.logger.Error("failed to check subcategory existence by category_id",
			log.Int("category_id", categoryID),
			log.Error(err),
		)

		return err
	}

	if existsSubcategory {
		u.logger.Warn("cannot delete category because it has subcategories",
			log.Int("category_id", categoryID),
		)

		return customError.ErrBadRequest(
			"cannot delete category because it has subcategories",
		)
	}

	category, err := u.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		u.logger.Error("failed to get category",
			log.Int("id", categoryID),
			log.Error(err),
		)

		return err
	}

	err = u.repo.DeleteCategory(ctx, categoryID)
	if err != nil {
		u.logger.Error("failed to delete category",
			log.Int("category_id", categoryID),
			log.Error(err),
		)

		return err
	}

	if err := u.minioRepo.Delete(ctx, category.ImagePath); err != nil {
		u.logger.Error("failed to delete old category image",
			log.Int("category_id", categoryID),
			log.String("object_name", category.ImagePath),
			log.Error(err),
		)
	}

	u.logger.Info("category deleted successfully",
		log.Int("category_id", categoryID),
	)

	return nil
}
