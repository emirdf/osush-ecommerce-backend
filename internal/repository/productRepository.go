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
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

type ProductRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewProductRepository(db *db.LoggedDB, logger log.Logger) *ProductRepository {
	return &ProductRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ProductRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	r.logger.Debug("checking by id exists",
		log.Int("id", id),
	)

	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1
            	FROM products
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

func (r *ProductRepository) ExistsBySubcategoryID(ctx context.Context, id int) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	r.logger.Debug("checking by subcategory_id exists",
		log.Int("id", id),
	)

	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1
            	FROM products
			WHERE subcategory_id = $1
        )
    `
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check subcategory_id existence", log.Error(err))
		return false, errors.New("failed to check subcategory_id existence")
	}

	r.logger.Debug("subcategory_id existence check completed",
		log.Bool("exists", exists),
	)
	return exists, nil
}

func (r *ProductRepository) CreateProduct(
	ctx context.Context,
	product models.Product,
) (*int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("creating product",
		log.String("name", product.Name),
		log.Int("subcategory_id", product.SubcategoryID),
	)

	var (
		descriptionDictionaryID *int
		featuresDictionaryID    *int
		err                     error
	)

	if product.DescriptionTm != "" || product.DescriptionRu != "" {
		r.logger.Debug("creating product description dictionary",
			log.String("product_name", product.Name),
		)

		descriptionDictionaryID, err = CreateDictionary(
			ctx,
			r.db,
			product.DescriptionTm,
			product.DescriptionRu,
		)
		if err != nil {
			r.logger.Error("failed to create product description dictionary",
				log.String("product_name", product.Name),
				log.Error(err),
			)

			return nil, errors.New("failed to create product description")
		}

		r.logger.Debug("product description dictionary created",
			log.String("product_name", product.Name),
			log.Int("dictionary_id", *descriptionDictionaryID),
		)
	}

	if product.FeaturesTm != "" || product.FeaturesRu != "" {
		r.logger.Debug("creating product features dictionary",
			log.String("product_name", product.Name),
		)

		featuresDictionaryID, err = CreateDictionary(
			ctx,
			r.db,
			product.FeaturesTm,
			product.FeaturesRu,
		)
		if err != nil {
			r.logger.Error("failed to create product features dictionary",
				log.String("product_name", product.Name),
				log.Error(err),
			)

			return nil, errors.New("failed to create product features")
		}

		r.logger.Debug("product features dictionary created",
			log.String("product_name", product.Name),
			log.Int("dictionary_id", *featuresDictionaryID),
		)
	}

	var productID int

	query := `
		INSERT INTO products (
			name,
			subcategory_id,
			description_dictionary_id,
			price,
			discount_price,
			features_dictionary_id,
			is_paused
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err = r.db.QueryRowContext(
		ctx,
		query,
		product.Name,
		product.SubcategoryID,
		descriptionDictionaryID,
		product.Price,
		product.DiscountPrice,
		featuresDictionaryID,
		product.IsPaused,
	).Scan(&productID)

	if err != nil {
		r.logger.Error("failed to create product",
			log.String("name", product.Name),
			log.Int("subcategory_id", product.SubcategoryID),
			log.Error(err),
		)

		return nil, errors.New("failed to create product")
	}

	r.logger.Info("product created successfully",
		log.Int("product_id", productID),
		log.String("name", product.Name),
	)

	return &productID, nil
}

func (r *ProductRepository) CreateProductImages(
	ctx context.Context,
	images []models.ProductImage,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(images) == 0 {
		r.logger.Warn("no product images to create")
		return nil
	}

	productID := images[0].ProductID

	r.logger.Debug("creating product images",
		log.Int("product_id", productID),
		log.Int("count", len(images)),
	)

	query := `
		INSERT INTO image_products (
			product_id,
			image_path,
			is_main
		)
		VALUES ($1, $2, $3)
	`

	for _, image := range images {

		r.logger.Debug("creating product image",
			log.Int("product_id", image.ProductID),
			log.String("image_path", image.ImagePath),
			log.Bool("is_main", image.IsMain),
		)

		_, err := r.db.ExecContext(
			ctx,
			query,
			image.ProductID,
			image.ImagePath,
			image.IsMain,
		)

		if err != nil {
			r.logger.Error("failed to create product image",
				log.Int("product_id", image.ProductID),
				log.String("image_path", image.ImagePath),
				log.Bool("is_main", image.IsMain),
				log.Error(err),
			)

			return errors.New("failed to create product image")
		}
	}

	r.logger.Info("product images created successfully",
		log.Int("product_id", productID),
		log.Int("count", len(images)),
	)

	return nil
}

func (r *ProductRepository) GetProductByID(
	ctx context.Context,
	id int,
) (*models.Product, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting product by id",
		log.Int("product_id", id),
	)

	query := `
		SELECT
			p.id,
			p.name,
			p.subcategory_id,
			s.name,
			c.id,
			c.name,
			p.description_dictionary_id,
			p.price,
			p.discount_price,
			p.features_dictionary_id,
			p.is_paused
		FROM products p
		JOIN subcategories s ON s.id = p.subcategory_id
		JOIN categories c ON c.id = s.category_id
		WHERE p.id = $1
	`

	var product models.Product

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&product.ID,
		&product.Name,
		&product.SubcategoryID,
		&product.SubcategoryName,
		&product.CategoryID,
		&product.CategoryName,
		&product.DescriptionDictionaryID,
		&product.Price,
		&product.DiscountPrice,
		&product.FeaturesDictionaryID,
		&product.IsPaused,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("product not found",
				log.Int("product_id", id),
			)

			return nil, customError.ErrNotFoundByID(id, "product")
		}

		r.logger.Error("failed to get product by id",
			log.Int("product_id", id),
			log.Error(err),
		)

		return nil, errors.New("failed to get product")
	}

	if product.DescriptionDictionaryID != nil {
		descriptionTm, descriptionRu, err := GetDictionaryByID(
			ctx,
			r.db,
			*product.DescriptionDictionaryID,
		)
		if err != nil {
			r.logger.Error("failed to get product description dictionary",
				log.String("product_name", product.Name),
				log.Error(err),
			)

			return nil, errors.New("failed to get product description")
		}
		product.DescriptionTm = *descriptionTm
		product.DescriptionRu = *descriptionRu
	}

	if product.FeaturesDictionaryID != nil {
		featuresTm, featuresRu, err := GetDictionaryByID(
			ctx,
			r.db,
			*product.FeaturesDictionaryID,
		)
		if err != nil {
			r.logger.Error("failed to get product features dictionary",
				log.String("product_name", product.Name),
				log.Error(err),
			)

			return nil, errors.New("failed to get product features")
		}
		product.FeaturesTm = *featuresTm
		product.FeaturesRu = *featuresRu
	}

	r.logger.Debug("product retrieved from database",
		log.Int("product_id", product.ID),
		log.String("name", product.Name),
	)

	return &product, nil
}

func (r *ProductRepository) GetProductImages(
	ctx context.Context,
	productID int,
) ([]models.ProductImage, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting product images",
		log.Int("product_id", productID),
	)

	query := `
		SELECT
			id,
			product_id,
			image_path,
			is_main
		FROM image_products
		WHERE product_id = $1
		ORDER BY is_main DESC, id ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		productID,
	)
	if err != nil {
		r.logger.Error("failed to get product images",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return nil, errors.New("failed to get product images")
	}
	defer rows.Close()

	var images []models.ProductImage

	for rows.Next() {
		var image models.ProductImage

		if err := rows.Scan(
			&image.ID,
			&image.ProductID,
			&image.ImagePath,
			&image.IsMain,
		); err != nil {
			r.logger.Error("failed to scan product image",
				log.Int("product_id", productID),
				log.Error(err),
			)

			return nil, errors.New("failed to scan product image")
		}

		images = append(images, image)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("failed while iterating product images",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return nil, errors.New("failed to get product images")
	}

	r.logger.Debug("product images retrieved",
		log.Int("product_id", productID),
		log.Int("count", len(images)),
	)

	return images, nil
}

func (r *ProductRepository) DeleteProductImages(
	ctx context.Context,
	productID int,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		DELETE FROM image_products
		WHERE product_id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		productID,
	)

	if err != nil {
		r.logger.Error("failed to delete product images",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return errors.New("failed to delete product images")
	}

	return nil
}

// DeleteProductImagesByIDs deletes the given image rows of one product.
func (r *ProductRepository) DeleteProductImagesByIDs(
	ctx context.Context,
	productID int,
	imageIDs []int,
) error {

	if len(imageIDs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ids := make([]int64, 0, len(imageIDs))
	for _, id := range imageIDs {
		ids = append(ids, int64(id))
	}

	query := `
		DELETE FROM image_products
		WHERE product_id = $1
			AND id = ANY($2::bigint[])
	`

	if _, err := r.db.ExecContext(ctx, query, productID, ids); err != nil {
		r.logger.Error("failed to delete product images by ids",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return errors.New("failed to delete product images")
	}

	return nil
}

func (r *ProductRepository) GetProductList(
	ctx context.Context,
	filter request.ProductFilter,
) ([]models.Product, *int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting products",
		log.String("search", filter.Search),
		log.String("sort_by_value", filter.SortByValue),
		log.Int("subcategory_id", filter.SubcategoryID),
		log.Int("category_id", filter.CategoryID),
		log.Int("limit", filter.LimitInt),
		log.Int("offset", filter.OffsetInt),
	)

	query := `
		SELECT
			p.id,
			p.name,
			p.subcategory_id,
			p.description_dictionary_id,
			p.price,
			p.discount_price,
			p.features_dictionary_id,
			p.is_paused,
			p.created_at
		FROM products p
		JOIN subcategories s ON s.id = p.subcategory_id
		JOIN categories c ON c.id = s.category_id
	`

	var (
		conditions []string
		args       []interface{}
		argIndex   = 1
	)

	if filter.Search != "" {
		conditions = append(
			conditions,
			fmt.Sprintf("LOWER(p.name) LIKE LOWER($%d)", argIndex),
		)

		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if filter.SubcategoryID != 0 {
		conditions = append(
			conditions,
			fmt.Sprintf("p.subcategory_id = $%d", argIndex),
		)

		args = append(args, filter.SubcategoryID)
		argIndex++
	}

	if filter.CategoryID != 0 {
		conditions = append(
			conditions,
			fmt.Sprintf("s.category_id = $%d", argIndex),
		)

		args = append(args, filter.CategoryID)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	queryTotal := `
		SELECT COUNT(*)
		FROM products p
		JOIN subcategories s ON s.id = p.subcategory_id
		JOIN categories c ON c.id = s.category_id
	`

	if len(conditions) > 0 {
		queryTotal += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int

	err := r.db.QueryRowContext(
		ctx,
		queryTotal,
		args...,
	).Scan(&total)

	if err != nil {
		r.logger.Error("failed to get products total",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get products total")
	}

	r.logger.Debug("products total retrieved",
		log.Int("total", total),
	)

	if filter.SortByValue == "ASC" || filter.SortByValue == "DESC" {
		query += " ORDER BY p.price " + filter.SortByValue
	} else {
		query += " ORDER BY p.created_at DESC"
	}

	filter.OffsetInt = (filter.OffsetInt - 1) * filter.LimitInt

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)

	args = append(args, filter.LimitInt)
	args = append(args, filter.OffsetInt)

	r.logger.Debug("executing get products query")

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to get products",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get products")
	}
	defer rows.Close()

	var products []models.Product

	for rows.Next() {

		var product models.Product

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.SubcategoryID,
			&product.DescriptionDictionaryID,
			&product.Price,
			&product.DiscountPrice,
			&product.FeaturesDictionaryID,
			&product.IsPaused,
			&product.CreatedAt,
		)

		if err != nil {
			r.logger.Error("failed to scan product",
				log.Error(err),
			)

			return nil, nil, errors.New("failed to scan product")
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("failed while iterating products",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get products")
	}

	r.logger.Info("products retrieved successfully",
		log.Int("count", len(products)),
		log.Int("total", total),
	)

	return products, &total, nil
}

func (r *ProductRepository) GetIndexProducts(
	ctx context.Context,
) (*[]models.IndexProduct, *[]models.IndexProduct, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting index products")

	randomQuery := `
		SELECT
			p.id,
			p.name,
			p.price,
			p.discount_price,
			COALESCE(img.image_path, ''),
			p.is_paused
		FROM products p
		LEFT JOIN image_products img
			ON img.product_id = p.id
			AND img.is_main = true
		ORDER BY RANDOM()
		LIMIT 5
	`

	rows, err := r.db.QueryContext(ctx, randomQuery)
	if err != nil {
		r.logger.Error("failed to get random products",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get random products")
	}

	var products []models.IndexProduct

	for rows.Next() {

		var product models.IndexProduct

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.DiscountPrice,
			&product.MainImage,
			&product.IsPaused,
		)

		if err != nil {
			rows.Close()

			r.logger.Error("failed to scan random product",
				log.Error(err),
			)

			return nil, nil, errors.New("failed to scan random product")
		}

		products = append(products, product)
	}

	rows.Close()

	if err := rows.Err(); err != nil {
		r.logger.Error("failed while iterating random products",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get random products")
	}

	newProductsQuery := `
		SELECT
			p.id,
			p.name,
			p.price,
			p.discount_price,
			COALESCE(img.image_path, ''),
			p.is_paused
		FROM products p
		LEFT JOIN image_products img
			ON img.product_id = p.id
			AND img.is_main = true
		ORDER BY p.created_at DESC
		LIMIT 5
	`

	rows, err = r.db.QueryContext(ctx, newProductsQuery)
	if err != nil {
		r.logger.Error("failed to get new products",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get new products")
	}

	var newProducts []models.IndexProduct

	for rows.Next() {

		var product models.IndexProduct

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.DiscountPrice,
			&product.MainImage,
			&product.IsPaused,
		)

		if err != nil {
			rows.Close()

			r.logger.Error("failed to scan new product",
				log.Error(err),
			)

			return nil, nil, errors.New("failed to scan new product")
		}

		newProducts = append(newProducts, product)
	}

	rows.Close()

	if err := rows.Err(); err != nil {
		r.logger.Error("failed while iterating new products",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get new products")
	}

	r.logger.Info("index products retrieved successfully",
		log.Int("products_count", len(products)),
		log.Int("new_products_count", len(newProducts)),
	)

	return &products, &newProducts, nil
}

func (r *ProductRepository) UpdateProduct(
	ctx context.Context,
	productID int,
	product models.Product,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("updating product",
		log.Int("product_id", productID),
		log.String("name", product.Name),
	)

	var oldDescriptionDictionaryID *int
	var oldFeaturesDictionaryID *int

	selectQuery := `
		SELECT
			description_dictionary_id,
			features_dictionary_id
		FROM products
		WHERE id = $1
	`

	err := r.db.QueryRowContext(
		ctx,
		selectQuery,
		productID,
	).Scan(
		&oldDescriptionDictionaryID,
		&oldFeaturesDictionaryID,
	)

	if err != nil {
		r.logger.Error("failed to get old product dictionary ids",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return errors.New("failed to get old product dictionary ids")
	}

	if product.DescriptionTm != "" || product.DescriptionRu != "" {

		if oldDescriptionDictionaryID != nil {

			err := UpdateDictionary(
				ctx,
				r.db,
				*oldDescriptionDictionaryID,
				product.DescriptionTm,
				product.DescriptionRu,
			)
			if err != nil {
				r.logger.Error("failed to update product description dictionary",
					log.Int("product_id", productID),
					log.Error(err),
				)

				return errors.New("failed to update product description")
			}

			product.DescriptionDictionaryID = oldDescriptionDictionaryID

		} else {

			descriptionDictionaryID, err := CreateDictionary(
				ctx,
				r.db,
				product.DescriptionTm,
				product.DescriptionRu,
			)
			if err != nil {
				r.logger.Error("failed to create product description dictionary",
					log.Int("product_id", productID),
					log.Error(err),
				)

				return errors.New("failed to create product description")
			}

			product.DescriptionDictionaryID = descriptionDictionaryID
		}

	} else {
		product.DescriptionDictionaryID = nil
	}

	if product.FeaturesTm != "" || product.FeaturesRu != "" {

		if oldFeaturesDictionaryID != nil {

			err := UpdateDictionary(
				ctx,
				r.db,
				*oldFeaturesDictionaryID,
				product.FeaturesTm,
				product.FeaturesRu,
			)
			if err != nil {
				r.logger.Error("failed to update product features dictionary",
					log.Int("product_id", productID),
					log.Error(err),
				)

				return errors.New("failed to update product features")
			}

			product.FeaturesDictionaryID = oldFeaturesDictionaryID

		} else {

			featuresDictionaryID, err := CreateDictionary(
				ctx,
				r.db,
				product.FeaturesTm,
				product.FeaturesRu,
			)
			if err != nil {
				r.logger.Error("failed to create product features dictionary",
					log.Int("product_id", productID),
					log.Error(err),
				)

				return errors.New("failed to create product features")
			}

			product.FeaturesDictionaryID = featuresDictionaryID
		}

	} else {
		product.FeaturesDictionaryID = nil
	}

	query := `
		UPDATE products
		SET
			name = $1,
			description_dictionary_id = $2,
			price = $3,
			discount_price = $4,
			features_dictionary_id = $5,
			is_paused = $6,
			subcategory_id = COALESCE(NULLIF($8::bigint, 0), subcategory_id),
			updated_at = NOW()
		WHERE id = $7
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		product.Name,
		product.DescriptionDictionaryID,
		product.Price,
		product.DiscountPrice,
		product.FeaturesDictionaryID,
		product.IsPaused,
		productID,
		product.SubcategoryID,
	)

	if err != nil {
		r.logger.Error("failed to update product",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return errors.New("failed to update product")
	}

	if (product.DescriptionTm == "" && product.DescriptionRu == "") &&
		oldDescriptionDictionaryID != nil {

		err := DeleteDictionary(
			ctx,
			r.db,
			*oldDescriptionDictionaryID,
		)
		if err != nil {
			r.logger.Error("failed to delete old product description dictionary",
				log.Int("product_id", productID),
				log.Int("dictionary_id", *oldDescriptionDictionaryID),
				log.Error(err),
			)

			return errors.New("failed to delete product description")
		}
	}

	if (product.FeaturesTm == "" && product.FeaturesRu == "") &&
		oldFeaturesDictionaryID != nil {

		err := DeleteDictionary(
			ctx,
			r.db,
			*oldFeaturesDictionaryID,
		)
		if err != nil {
			r.logger.Error("failed to delete old product features dictionary",
				log.Int("product_id", productID),
				log.Int("dictionary_id", *oldFeaturesDictionaryID),
				log.Error(err),
			)

			return errors.New("failed to delete product features")
		}
	}

	r.logger.Info("product updated successfully",
		log.Int("product_id", productID),
	)

	return nil
}

func (r *ProductRepository) CheckProductsNotPaused(
	ctx context.Context,
	productIDs []int,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT id
		FROM products
		WHERE id = ANY($1)
		  AND is_paused = TRUE
	`

	rows, err := r.db.QueryContext(ctx, query, productIDs)
	if err != nil {
		r.logger.Error("failed to check paused products",
			log.Error(err),
		)

		return errors.New("failed to check products")
	}
	defer rows.Close()

	var productIsPaused string

	for rows.Next() {
		var productID int

		if err := rows.Scan(&productID); err != nil {
			r.logger.Error("failed to scan paused product",
				log.Error(err),
			)

			return errors.New("failed to check products")
		}

		r.logger.Warn("product is paused",
			log.Int("product_id", productID),
		)

		productIsPaused = productIsPaused + " " + strconv.Itoa(productID)
	}

	if err := rows.Err(); err != nil {
		return errors.New("failed to check products")
	}

	if productIsPaused != "" {
		return customError.ErrBadRequest("product is paused, product_id =" + productIsPaused)
	}

	return nil
}

func (r *ProductRepository) GetProductsPrices(
	ctx context.Context,
	productIDs []int,
) (map[int]models.Product, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT
			id,
			price,
			discount_price
		FROM products
		WHERE id = ANY($1)
	`

	rows, err := r.db.DB.Query(ctx, query, productIDs)
	if err != nil {
		r.logger.Error("failed to get products prices",
			log.Error(err),
		)

		return nil, errors.New("failed to get products prices")
	}
	defer rows.Close()

	prices := make(map[int]models.Product, len(productIDs))

	for rows.Next() {
		var productID int
		var price float64
		var discountPrice *float64

		err := rows.Scan(
			&productID,
			&price,
			&discountPrice,
		)
		if err != nil {
			r.logger.Error("failed to scan product price",
				log.Error(err),
			)

			return nil, errors.New("failed to get products prices")
		}

		prices[productID] = models.Product{
			Price:         price,
			DiscountPrice: discountPrice,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("failed to get products prices")
	}

	return prices, nil
}

func (r *ProductRepository) DeleteProduct(
	ctx context.Context,
	productID int,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("deleting product",
		log.Int("product_id", productID),
	)

	query := `
		DELETE FROM products
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		productID,
	)
	if err != nil {
		r.logger.Error("failed to delete product",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return errors.New("failed to delete product")
	}

	r.logger.Info("product deleted successfully",
		log.Int("product_id", productID),
	)

	return nil
}
