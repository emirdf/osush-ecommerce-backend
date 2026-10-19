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

type SubcategoryRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewSubcategoryRepository(db *db.LoggedDB, logger log.Logger) *SubcategoryRepository {
	return &SubcategoryRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SubcategoryRepository) ExistsByID(
	ctx context.Context,
	id int,
) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	r.logger.Debug("checking if subcategory name exists",
		log.Int("id", id),
	)

	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1
			FROM subcategories
			WHERE id = $1
		)
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)

	if err != nil {
		r.logger.Error("failed to check subcategory id existence",
			log.Error(err),
		)

		return false, errors.New("failed to check subcategory id existence")
	}

	r.logger.Debug("subcategory id existence check completed",
		log.Bool("exists", exists),
	)

	return exists, nil
}

func (r *SubcategoryRepository) ExistsByName(
	ctx context.Context,
	req request.SubcategoryRequest,
) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	r.logger.Debug("checking if subcategory name exists",
		log.Int("category_id", req.CategoryID),
		log.String("name", req.Name),
	)

	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1
			FROM subcategories
			WHERE category_id = $1 AND name = $2
		)
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		req.CategoryID,
		req.Name,
	).Scan(&exists)

	if err != nil {
		r.logger.Error("failed to check subcategory name existence",
			log.Error(err),
		)

		return false, errors.New("failed to check subcategory name existence")
	}

	r.logger.Debug("subcategory name existence check completed",
		log.Bool("exists", exists),
	)

	return exists, nil
}

func (r *SubcategoryRepository) ExistsByCategoryID(
	ctx context.Context,
	categoryID int,
) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	r.logger.Debug("checking if subcategory categoryID exists",
		log.Int("category_id", categoryID),
	)

	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1
			FROM subcategories
			WHERE category_id = $1
		)
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		categoryID,
	).Scan(&exists)

	if err != nil {
		r.logger.Error("failed to check subcategory category_id existence",
			log.Error(err),
		)

		return false, errors.New("failed to check subcategory category_id existence")
	}

	r.logger.Debug("subcategory category_id existence check completed",
		log.Bool("exists", exists),
	)

	return exists, nil
}

func (r *SubcategoryRepository) ExistsByCategoryIDAndNameNotID(
	ctx context.Context,
	categoryID int,
	name string,
	subcategoryID int,
) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM subcategories
			WHERE name = $1 AND category_id = $2
			AND id != $3
		)
	`

	var exist bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		name,
		categoryID,
		subcategoryID,
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

func (r *SubcategoryRepository) CreateSubcategory(
	ctx context.Context,
	subcategory models.Subcategory,
) (*int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("creating subcategory",
		log.Int("category_id", subcategory.CategoryID),
		log.String("name", subcategory.Name),
	)

	var subcategoryID int

	query := `
		INSERT INTO subcategories (
			category_id,
			name
		)
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		subcategory.CategoryID,
		subcategory.Name,
	).Scan(&subcategoryID)

	if err != nil {
		r.logger.Error("failed to create subcategory",
			log.Int("category_id", subcategory.CategoryID),
			log.String("name", subcategory.Name),
			log.Error(err),
		)

		return nil, errors.New("failed to create subcategory")
	}

	r.logger.Info("subcategory created",
		log.Int("subcategory_id", subcategoryID),
		log.Int("category_id", subcategory.CategoryID),
	)

	return &subcategoryID, nil
}

func (r *SubcategoryRepository) GetSubcategoryByID(
	ctx context.Context,
	subcategoryID int,
) (*models.Subcategory, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("fetching subcategory by ID",
		log.Int("id", subcategoryID),
	)

	var subcategory models.Subcategory

	query := `
		SELECT
			id,
			category_id,
			name,
			created_at,
			updated_at
		FROM subcategories
		WHERE id = $1
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		subcategoryID,
	).Scan(
		&subcategory.ID,
		&subcategory.CategoryID,
		&subcategory.Name,
		&subcategory.CreatedAt,
		&subcategory.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		r.logger.Debug("subcategory not found",
			log.Int("id", subcategoryID),
		)

		return nil, customError.ErrNotFoundByID(subcategoryID, "subcategory")
	}

	if err != nil {
		r.logger.Error("failed to fetch subcategory",
			log.Int("id", subcategoryID),
			log.Error(err),
		)

		return nil, fmt.Errorf(
			"failed to fetch subcategory: %w",
			err,
		)
	}

	r.logger.Debug("subcategory fetched successfully",
		log.Int("id", subcategoryID),
	)

	return &subcategory, nil
}

func (r *SubcategoryRepository) GetSubcategoryList(
	ctx context.Context,
	search string,
	limit int,
	offset int,
	categoryID int,
) (*[]models.Subcategory, *int, error) {

	r.logger.Debug("fetching all subcategories",
		log.String("search", search),
		log.Int("limit", limit),
		log.Int("offset", offset),
	)

	var (
		subcategories  []models.Subcategory
		total          int
		categoryFilter string
	)

	if categoryID != 0 {
		categoryFilter = fmt.Sprintf("AND category_id = %d", categoryID)
	}

	offset = (offset - 1) * limit

	query := fmt.Sprintf(`
		SELECT 
			id,
			category_id,
			name, 
			created_at,
			updated_at
		FROM subcategories s
		WHERE ($1 = '' 
			OR name ILIKE '%%' || $1 || '%%')
			%s
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, categoryFilter)

	rows, err := r.db.QueryContext(ctx, query, search, limit, offset)
	if err != nil {
		r.logger.Error("failed to select subcategories",
			log.Error(err),
		)

		return nil, nil, errors.New(
			"failed to select subcategories",
		)
	}

	defer rows.Close()

	for rows.Next() {

		var subcategory models.Subcategory

		err := rows.Scan(
			&subcategory.ID,
			&subcategory.CategoryID,
			&subcategory.Name,
			&subcategory.CreatedAt,
			&subcategory.UpdatedAt,
		)

		if err != nil {
			r.logger.Error("failed to scan subcategory",
				log.Error(err),
			)

			return nil, nil, errors.New("failed to scan subcategory")
		}

		subcategories = append(subcategories, subcategory)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(
			"error occurred while iterating subcategories",
			log.Error(err),
		)

		return nil, nil, errors.New(
			"error occurred while iterating subcategories",
		)
	}

	queryCount := fmt.Sprintf(`
		SELECT count(*)
		FROM subcategories
		WHERE ($1 = '' 
			OR name ILIKE '%%' || $1 || '%%' )
			%s
	`, categoryFilter)

	err = r.db.QueryRowContext(
		ctx,
		queryCount,
		search,
	).Scan(&total)

	if err != nil {
		r.logger.Error("failed to count subcategories",
			log.Error(err),
		)

		return nil, nil, errors.New(
			"failed to count subcategories",
		)
	}

	r.logger.Debug("subcategories fetched successfully",
		log.Int("total", total),
	)

	return &subcategories, &total, nil
}

func (r *SubcategoryRepository) UpdateSubcategory(
	ctx context.Context,
	subcategoryID int,
	subcategory models.Subcategory,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("updating subcategory",
		log.Int("subcategory_id", subcategoryID),
		log.Int("category_id", subcategory.CategoryID),
		log.String("name", subcategory.Name),
	)

	query := `
		UPDATE subcategories
		SET
			name = $1,
			updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
		WHERE id = $2
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		subcategory.Name,
		subcategoryID,
	)
	if err != nil {
		r.logger.Error("failed to update subcategory",
			log.Int("subcategory_id", subcategoryID),
			log.Error(err),
		)

		return errors.New("failed to update subcategory")
	}

	r.logger.Info("subcategory updated successfully",
		log.Int("subcategory_id", subcategoryID),
	)

	return nil
}

func (r *SubcategoryRepository) DeleteSubcategory(
	ctx context.Context,
	subcategoryID int,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("deleting subcategory",
		log.Int("subcategory_id", subcategoryID),
	)

	query := `
		DELETE FROM subcategories
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		subcategoryID,
	)
	if err != nil {
		r.logger.Error("failed to delete subcategory",
			log.Int("subcategory_id", subcategoryID),
			log.Error(err),
		)

		return errors.New("failed to delete subcategory")
	}

	r.logger.Info("subcategory deleted successfully",
		log.Int("subcategory_id", subcategoryID),
	)

	return nil
}
