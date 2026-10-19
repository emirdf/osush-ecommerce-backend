package repository

import (
	"context"
	"errors"
	"osush/internal/db"
	"osush/internal/log"
	"osush/internal/models"
	"time"
)

type ProductLikeRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewProductLikeRepository(db *db.LoggedDB, logger log.Logger) *ProductLikeRepository {
	return &ProductLikeRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ProductLikeRepository) ToggleLikeProduct(
	ctx context.Context,
	clientID int,
	productID int,
) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exists bool

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM product_likes
			WHERE client_id = $1
			  AND product_id = $2
		)
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		clientID,
		productID,
	).Scan(&exists)

	if err != nil {
		return false, errors.New("failed to check product like")
	}

	if exists {
		query = `
			DELETE FROM product_likes
			WHERE client_id = $1
			  AND product_id = $2
		`

		_, err = r.db.ExecContext(
			ctx,
			query,
			clientID,
			productID,
		)

		if err != nil {
			return false, errors.New("failed to remove product like")
		}

		return false, nil
	}

	query = `
		INSERT INTO product_likes (client_id, product_id)
		VALUES ($1, $2)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		clientID,
		productID,
	)

	if err != nil {
		return false, errors.New("failed to add product like")
	}

	return true, nil
}

func (r *ProductLikeRepository) IsProductLiked(
	ctx context.Context,
	clientID int,
	productID int,
) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM product_likes
			WHERE client_id = $1
			  AND product_id = $2
		)
	`

	var isLiked bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		clientID,
		productID,
	).Scan(&isLiked)

	if err != nil {
		r.logger.Error("failed to check product like",
			log.Int("client_id", clientID),
			log.Int("product_id", productID),
			log.Error(err),
		)

		return false, errors.New("failed to check product like")
	}

	return isLiked, nil
}

func (r *ProductLikeRepository) GetLikedProducts(
	ctx context.Context,
	clientID int,
	limit int,
	offset int,
) ([]models.IndexProduct, *int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	offset = (offset - 1) * limit

	query := `
		SELECT
			p.id,
			p.name,
			p.price,
			p.discount_price,
			img.image_path,
			COUNT(*) OVER() AS total
		FROM product_likes pl
		JOIN products p ON p.id = pl.product_id
		LEFT JOIN image_products img
			ON img.product_id = p.id
			AND img.is_main = true
		WHERE pl.client_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		clientID,
		limit,
		offset,
	)
	if err != nil {
		r.logger.Error("failed to get liked products",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get liked products")
	}
	defer rows.Close()

	products := make([]models.IndexProduct, 0)
	var total int

	for rows.Next() {
		var product models.IndexProduct

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.DiscountPrice,
			&product.MainImage,
			&total,
		)

		if err != nil {
			r.logger.Error("failed to scan liked product",
				log.Int("client_id", clientID),
				log.Error(err),
			)

			return nil, nil, errors.New("failed to scan liked product")
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("failed to read liked products",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, nil, errors.New("failed to read liked products")
	}

	return products, &total, nil
}
