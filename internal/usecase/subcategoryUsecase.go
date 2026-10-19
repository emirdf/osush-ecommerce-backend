package usecase

import (
	"context"
	"errors"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/models"
	"osush/internal/repository"
	"strconv"
)

type SubcategoryUseCase struct {
	repo         *repository.SubcategoryRepository
	categoryRepo *repository.CategoryRepository
	productRepo  *repository.ProductRepository
	logger       log.Logger
}

func NewSubcategoryUseCase(
	repo *repository.SubcategoryRepository,
	categoryRepo *repository.CategoryRepository,
	productRepo *repository.ProductRepository,
	logger log.Logger) *SubcategoryUseCase {
	return &SubcategoryUseCase{
		repo:         repo,
		categoryRepo: categoryRepo,
		productRepo:  productRepo,
		logger:       logger,
	}
}

func (u *SubcategoryUseCase) CreateSubcategory(
	ctx context.Context,
	req request.SubcategoryRequest,
) (*response.SubcategoryResponse, error) {

	u.logger.Info("starting create subcategory",
		log.Int("category_id", req.CategoryID),
		log.String("name", req.Name),
	)

	exist, err := u.repo.ExistsByName(ctx, req)
	if err != nil {
		u.logger.Error("failed to check subcategory existence",
			log.String("name", req.Name),
			log.Error(err),
		)

		return nil, errors.New("failed to check subcategory existence")
	}

	if exist {
		u.logger.Warn("subcategory already exists",
			log.String("name", req.Name),
		)

		return nil, customError.ErrAlreadyData("subcategory")
	}

	u.logger.Debug("subcategory existence check completed",
		log.String("name", req.Name),
		log.Bool("exists", exist),
	)

	existCategory, err := u.categoryRepo.ExistsByID(ctx, req.CategoryID)
	if err != nil {
		u.logger.Error("failed to check category existence",
			log.Int("category_id", req.CategoryID),
			log.Error(err),
		)

		return nil, errors.New("failed to check category existence")
	}

	if !existCategory {
		u.logger.Warn("category not found",
			log.Int("category_id", req.CategoryID),
		)

		return nil, customError.ErrNotFoundByID(req.CategoryID, "category")
	}

	subcategory := models.Subcategory{
		CategoryID: req.CategoryID,
		Name:       req.Name,
	}

	subcategoryID, err := u.repo.CreateSubcategory(ctx, subcategory)
	if err != nil {
		u.logger.Error("failed to create subcategory",
			log.Int("category_id", req.CategoryID),
			log.String("name", req.Name),
			log.Error(err),
		)

		return nil, err
	}

	newSubcategory, err := u.GetSubcategoryByID(ctx, *subcategoryID)
	if err != nil {
		return nil, err
	}

	u.logger.Info("subcategory created successfully",
		log.Int("subcategory_id", *subcategoryID),
	)

	return newSubcategory, nil
}

func (u *SubcategoryUseCase) GetSubcategoryByID(
	ctx context.Context,
	subcategoryID int,
) (*response.SubcategoryResponse, error) {

	u.logger.Info("starting get subcategory by id",
		log.Int("id", subcategoryID),
	)

	subcategory, err := u.repo.GetSubcategoryByID(
		ctx,
		subcategoryID,
	)
	if err != nil {
		u.logger.Error("failed to get subcategory",
			log.Int("id", subcategoryID),
			log.Error(err),
		)

		return nil, err
	}

	resp := response.SubcategoryResponse{
		ID:   subcategory.ID,
		Name: subcategory.Name,
	}

	return &resp, nil
}

func (u *SubcategoryUseCase) GetSubcategoryList(
	ctx context.Context,
	search string,
	limit int,
	offset int,
	categoryID string,
) (*response.SubcategoryListResponse, error) {

	u.logger.Info("starting get all subcategories",
		log.String("search", search),
		log.Int("limit", limit),
		log.Int("offset", offset),
	)
	var (
		categoryIDInt int
		err           error
	)

	if categoryID != "" {

		categoryIDInt, err = strconv.Atoi(categoryID)
		if err != nil {
			u.logger.Error("error category_id",
				log.String("category_id", categoryID),
				log.Error(err),
			)
			return nil, errors.New("error category_id")
		}

		exist, err := u.categoryRepo.ExistsByID(ctx, categoryIDInt)

		if err != nil {
			u.logger.Warn("failed check category by id",
				log.Int("category_id", categoryIDInt),
				log.Error(err),
			)
			return nil, errors.New("failed check category")
		}

		if !exist {
			u.logger.Warn("not found category",
				log.Int("category_id", categoryIDInt),
			)
			return nil, customError.ErrNotFoundByID(categoryIDInt, "category")
		}
	}

	subcategories, total, err := u.repo.GetSubcategoryList(
		ctx,
		search,
		limit,
		offset,
		categoryIDInt,
	)
	if err != nil {
		u.logger.Error("failed to get all subcategories",
			log.String("search", search),
			log.Int("limit", limit),
			log.Int("offset", offset),
			log.String("category_id", categoryID),
			log.Error(err),
		)
		return nil, err
	}

	u.logger.Info("subcategories fetched successfully",
		log.Int("count", len(*subcategories)),
		log.Int("total", *total),
	)

	resp := response.SubcategoryListResponse{
		Subcategories: []response.SubcategoryResponse{},
		Total:         *total,
	}

	for _, subcategory := range *subcategories {

		resp.Subcategories = append(
			resp.Subcategories,
			response.SubcategoryResponse{
				ID:   subcategory.ID,
				Name: subcategory.Name,
			},
		)
	}

	return &resp, nil
}

func (u *SubcategoryUseCase) UpdateSubcategory(
	ctx context.Context,
	subcategoryID int,
	req request.SubcategoryUpdateRequest,
) (*response.SubcategoryResponse, error) {

	u.logger.Info("starting update subcategory",
		log.Int("subcategory_id", subcategoryID),
		log.String("name", req.Name),
	)

	exists, err := u.repo.ExistsByID(ctx, subcategoryID)
	if err != nil {
		u.logger.Error("failed to check subcategory existence",
			log.Int("subcategory_id", subcategoryID),
			log.Error(err),
		)

		return nil, errors.New("failed to check subcategory existence")
	}

	if !exists {
		u.logger.Warn("subcategory not found",
			log.Int("subcategory_id", subcategoryID),
		)

		return nil, customError.ErrNotFoundByID(
			subcategoryID,
			"subcategory",
		)
	}

	subcategory, err := u.repo.GetSubcategoryByID(ctx, subcategoryID)
	if err != nil {
		u.logger.Error("failed to get subcategory",
			log.Int("id", subcategoryID),
			log.Error(err),
		)

		return nil, err
	}

	exist, err := u.repo.ExistsByCategoryIDAndNameNotID(
		ctx,
		subcategory.CategoryID,
		req.Name,
		subcategoryID,
	)
	if err != nil {
		u.logger.Error("failed to check subcategory name",
			log.Int("subcategory_id", subcategoryID),
			log.String("name", req.Name),
			log.Error(err),
		)

		return nil, errors.New("failed to check subcategory name")
	}

	if exist {
		u.logger.Warn("subcategory name already exists",
			log.Int("subcategory_id", subcategoryID),
			log.String("name", req.Name),
		)

		return nil, customError.ErrAlreadyData("subcategory")
	}

	newSubcategory := models.Subcategory{
		Name: req.Name,
	}

	if err := u.repo.UpdateSubcategory(
		ctx,
		subcategoryID,
		newSubcategory,
	); err != nil {
		u.logger.Error("failed to update subcategory",
			log.Int("subcategory_id", subcategoryID),
			log.Error(err),
		)

		return nil, err
	}

	updatedSubcategory, err := u.GetSubcategoryByID(ctx, subcategoryID)
	if err != nil {
		return nil, err
	}

	u.logger.Info("subcategory updated successfully",
		log.Int("subcategory_id", subcategoryID),
	)

	return updatedSubcategory, nil
}

func (u *SubcategoryUseCase) DeleteSubcategory(
	ctx context.Context,
	subcategoryID int,
) error {

	u.logger.Info("starting delete subcategory",
		log.Int("subcategory_id", subcategoryID),
	)

	exists, err := u.repo.ExistsByID(ctx, subcategoryID)
	if err != nil {
		u.logger.Error("failed to check subcategory existence",
			log.Int("subcategory_id", subcategoryID),
			log.Error(err),
		)

		return err
	}

	if !exists {
		u.logger.Warn("subcategory not found",
			log.Int("subcategory_id", subcategoryID),
		)

		return customError.ErrNotFoundByID(
			subcategoryID,
			"subcategory",
		)
	}

	existsProduct, err := u.productRepo.ExistsBySubcategoryID(ctx, subcategoryID)
	if err != nil {
		u.logger.Error("failed to check product existence by subcategory_id",
			log.Int("subcategory_id", subcategoryID),
			log.Error(err),
		)

		return err
	}

	if existsProduct {
		u.logger.Warn("cannot delete subcategory because it has product",
			log.Int("subcategory_id", subcategoryID),
		)

		return customError.ErrBadRequest(
			"cannot delete subcategory because it has product",
		)
	}

	if err := u.repo.DeleteSubcategory(ctx, subcategoryID); err != nil {
		u.logger.Error("failed to delete subcategory",
			log.Int("subcategory_id", subcategoryID),
			log.Error(err),
		)

		return err
	}

	u.logger.Info("subcategory deleted successfully",
		log.Int("subcategory_id", subcategoryID),
	)

	return nil
}
