package repository

import (
	"context"
	"errors"
	"fmt"
	"osush/internal/db"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/log"
	"osush/internal/models"
	"time"

	"github.com/jackc/pgx/v4"
)

type CategoryRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewCategoryRepository(db *db.LoggedDB, logger log.Logger) *CategoryRepository {
	return &CategoryRepository{
		db:     db,
		logger: logger,
	}
}

func (r *CategoryRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	r.logger.Debug("checking by id exists",
		log.Int("id", id),
	)

	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1
            	FROM categories
			WHERE id = $1
        )
    `
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check id existence", log.Error(err))
		return false, errors.New("failed to check id existence")
	}

	r.logger.Debug("id existence check completed",
		log.Bool("exists", exists),
	)
	return exists, nil
}

func (r *CategoryRepository) ExistsByName(
	ctx context.Context,
	req request.CategoryRequest,
) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	r.logger.Debug("checking if name exists",
		log.String("name", req.Name),
	)

	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1
            	FROM categories
			WHERE name = $1
        )
    `
	err := r.db.QueryRowContext(ctx, query, req.Name).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check name existence", log.Error(err))
		return false, errors.New("failed to check name existence")
	}

	r.logger.Debug("name existence check completed",
		log.Bool("exists", exists),
	)
	return exists, nil
}

func (r *CategoryRepository) ExistsByIDAndName(
	ctx context.Context,
	name string,
	categoryID int,
) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM categories
			WHERE name = $1
			AND id != $2
		)
	`

	var exist bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		name,
		categoryID,
	).Scan(&exist)

	if err != nil {
		r.logger.Error("failed to check category name",
			log.String("name", name),
			log.Int("category_id", categoryID),
			log.Error(err),
		)

		return false, errors.New("failed to check category name")
	}

	return exist, nil
}

func (r *CategoryRepository) TotalCount(ctx context.Context) (*int, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var total int

	query := `
        	SELECT COUNT(*)
		FROM categories
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count categories",
			log.Error(err),
		)

		return nil, errors.New("failed to count categories")
	}

	r.logger.Debug("id existence check completed",
		log.Int("total", total),
	)
	return &total, nil
}

func (r *CategoryRepository) CreateCategory(
	ctx context.Context,
	category models.Category,
) (*int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("create category",
		log.String("name", category.Name),
	)

	var categoryID int

	insertCategory := `
				INSERT INTO categories (name, image_path)
				VALUES ($1, $2)
				RETURNING id
			`

	err := r.db.QueryRowContext(ctx, insertCategory, category.Name, category.ImagePath).Scan(
		&categoryID)
	if err != nil {
		r.logger.Error("failed to create category",
			log.String("name", category.Name),
			log.Error(err),
		)

		return nil, errors.New("failed create category")
	}

	r.logger.Debug("category created",
		log.String("name", category.Name),
		log.Int("category_id", categoryID),
	)

	return &categoryID, nil
}

func (r *CategoryRepository) GetCategoryByID(ctx context.Context, categoryID int) (*models.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	r.logger.Debug("fetching category by ID",
		log.Int("id", categoryID))

	var category models.Category

	query := `
		SELECT 
			id, name, image_path, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(
		&category.ID,
		&category.Name,
		&category.ImagePath,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		r.logger.Debug("category not found",
			log.Int("id", categoryID))
		return nil, customError.ErrNotFoundByID(categoryID, "category")
	}

	if err != nil {
		r.logger.Error("failed to fetch category",
			log.Int("id", categoryID),
			log.Error(err))
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	r.logger.Debug("category fetched successfully",
		log.Int("id", categoryID),
	)

	return &category, nil
}

func (r *CategoryRepository) GetCategoryList(
	ctx context.Context,
	search string,
	limit int,
	offset int,
	baseURL string,
) (*[]models.Category, *int, error) {

	r.logger.Debug("fetching all categories",
		log.String("search", search),
		log.Int("limit", limit),
		log.Int("offset", offset),
	)

	var (
		categories []models.Category
		total      int
	)

	offset = (offset - 1) * limit

	q := `
		SELECT
			id,
			name,
			image_path,
			created_at, 
			updated_at
		FROM categories
		WHERE (
			$1 = ''
			OR name ILIKE '%' || $1 || '%'
		)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, q, search, limit, offset)
	if err != nil {
		r.logger.Error("ailed to select categories",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to select categories")
	}

	defer rows.Close()

	for rows.Next() {

		var c models.Category

		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.ImagePath,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {

			r.logger.Error("failed to scan category",
				log.Error(err),
			)

			return nil, nil, errors.New("failed to scan category")
		}

		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error occurred while iterating categories",
			log.Error(err),
		)

		return nil, nil, errors.New("error occurred while iterating categories")
	}

	q = `
		SELECT COUNT(*)
		FROM categories
		WHERE (
			$1 = ''
			OR name ILIKE '%' || $1 || '%'
		)
	`

	err = r.db.QueryRowContext(ctx, q, search).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count categories",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to count categories")
	}

	r.logger.Debug("categories fetched successfully",
		log.Int("total", total),
	)

	return &categories, &total, nil
}

func (r *CategoryRepository) UpdateCategory(
	ctx context.Context,
	categoryID int,
	category models.Category,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("updating category",
		log.Int("category_id", categoryID),
		log.String("name", category.Name),
	)

	query := `
		UPDATE categories
		SET
			name = $1,
			image_path = $2,
			updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		category.Name,
		category.ImagePath,
		categoryID,
	)

	if err != nil {
		r.logger.Error("failed to update category",
			log.Int("category_id", categoryID),
			log.Error(err),
		)

		return errors.New("failed to update category")
	}

	r.logger.Info("category updated successfully",
		log.Int("category_id", categoryID),
	)

	return nil
}

func (r *CategoryRepository) DeleteCategory(
	ctx context.Context,
	categoryID int,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("deleting category",
		log.Int("category_id", categoryID),
	)

	query := `
		DELETE FROM categories
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		categoryID,
	)
	if err != nil {
		r.logger.Error("failed to delete category",
			log.Int("category_id", categoryID),
			log.Error(err),
		)

		return err
	}

	r.logger.Info("category deleted successfully",
		log.Int("category_id", categoryID),
	)

	return nil
}
