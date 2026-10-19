package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"osush/internal/db"
	customError "osush/internal/errors"
	"osush/internal/log"
	"osush/internal/models"
	"time"

	"github.com/jackc/pgx/v4"
)

type BasketRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewBasketRepository(db *db.LoggedDB, logger log.Logger) *BasketRepository {
	return &BasketRepository{
		db:     db,
		logger: logger,
	}
}

func (r *BasketRepository) ExistsByProductID(ctx context.Context, id int) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	r.logger.Debug("checking by id exists",
		log.Int("id", id),
	)

	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1
            	FROM basket
			WHERE product_id = $1
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

func (r *BasketRepository) AddProductInBasket(
	ctx context.Context,
	basket models.Basket,
) (*int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("adding product to basket",
		log.Int("client_id", basket.ClientID),
		log.Int("product_id", basket.ProductID),
	)

	query := `
	INSERT INTO basket (
		client_id,
		product_id,
		total
	)
	VALUES ($1, $2, 1)
	ON CONFLICT (client_id, product_id)
	DO UPDATE SET
		total = basket.total + 1,
		updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
	RETURNING total
`

	var total int

	err := r.db.QueryRowContext(
		ctx,
		query,
		basket.ClientID,
		basket.ProductID,
	).Scan(&total)

	if err != nil {
		r.logger.Error("failed to add product to basket",
			log.Int("client_id", basket.ClientID),
			log.Int("product_id", basket.ProductID),
			log.Error(err),
		)

		return nil, errors.New("failed to add product to basket")
	}

	r.logger.Info("basket product count increased",
		log.Int("client_id", basket.ClientID),
		log.Int("product_id", basket.ProductID),
	)

	return &total, nil
}

func (r *BasketRepository) DeleteProductFromBasket(
	ctx context.Context,
	basket models.Basket,
) (*int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("removing product from basket",
		log.Int("client_id", basket.ClientID),
		log.Int("product_id", basket.ProductID),
	)

	query := `
		WITH current_basket AS (
			SELECT total
			FROM basket
			WHERE client_id = $1
			  AND product_id = $2
		),
		updated AS (
			UPDATE basket
			SET
				total = total - 1,
				updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
			WHERE client_id = $1
			  AND product_id = $2
			  AND total > 1
			RETURNING total
		),
		deleted AS (
			DELETE FROM basket
			WHERE client_id = $1
			  AND product_id = $2
			  AND total = 1
			RETURNING 0 AS total
		)
		SELECT total FROM updated
		UNION ALL
		SELECT total FROM deleted
	`

	var total int

	err := r.db.QueryRowContext(
		ctx,
		query,
		basket.ClientID,
		basket.ProductID,
	).Scan(&total)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("product not found in basket",
				log.Int("client_id", basket.ClientID),
				log.Int("product_id", basket.ProductID),
			)

			return nil, customError.ErrNotFound("product in basket")
		}

		r.logger.Error("failed to remove product from basket",
			log.Int("client_id", basket.ClientID),
			log.Int("product_id", basket.ProductID),
			log.Error(err),
		)

		return nil, errors.New("failed to remove product from basket")
	}

	r.logger.Info("basket product count decreased",
		log.Int("client_id", basket.ClientID),
		log.Int("product_id", basket.ProductID),
		log.Int("total", total),
	)

	return &total, nil
}

func (r *BasketRepository) CleanBasketByID(
	ctx context.Context,
	clientID int,
	productID int,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("deleting product by client id",
		log.Int("client_id", clientID),
		log.Int("product_id", productID),
	)

	query := `
		DELETE FROM basket
		WHERE client_id = $1 AND product_id = $2
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		clientID,
		productID,
	)

	if err != nil {
		r.logger.Error("failed to delete product",
			log.Int("client_id", clientID),
			log.Int("product_id", productID),
			log.Error(err),
		)

		return errors.New("failed to delete product")
	}

	r.logger.Info("product deleted successfully",
		log.Int("client_id", clientID),
		log.Int("product_id", productID),
	)

	return nil
}

func (r *BasketRepository) CleanBasket(
	ctx context.Context,
	clientID int,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("deleting basket by client id",
		log.Int("client_id", clientID),
	)

	query := `
		DELETE FROM basket
		WHERE client_id = $1 
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		clientID,
	)

	if err != nil {
		r.logger.Error("failed to delete basket",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return errors.New("failed to delete basket")
	}

	r.logger.Info("basket deleted successfully",
		log.Int("client_id", clientID),
	)

	return nil
}

func (r *BasketRepository) CheckBasketProduct(
	ctx context.Context,
	clientID int,
	req []models.Basket,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("checking basket products",
		log.Int("client_id", clientID),
	)

	query := `
		SELECT
			b.product_id,
			b.total
		FROM basket b
		WHERE b.client_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		r.logger.Error("failed to check basket products",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return errors.New("failed to check basket products")
	}
	defer rows.Close()

	basket := make(map[int]int)

	for rows.Next() {
		var (
			productID int
			total     int
		)

		if err := rows.Scan(
			&productID,
			&total,
		); err != nil {
			r.logger.Error("failed to scan basket product",
				log.Int("client_id", clientID),
				log.Error(err),
			)

			return errors.New("failed to check basket products")
		}

		basket[productID] = total
	}

	if err := rows.Err(); err != nil {
		return errors.New("failed to check basket products")
	}

	for _, product := range req {

		basketTotal, exists := basket[product.ProductID]

		if !exists {
			r.logger.Warn("product not found in basket",
				log.Int("client_id", clientID),
				log.Int("product_id", product.ProductID),
			)

			return customError.ErrNotFound(
				fmt.Sprintf(
					"product not found in basket, product_id = %d",
					product.ProductID,
				),
			)
		}

		if basketTotal != product.Total {
			r.logger.Warn("unequal product quantity in basket",
				log.Int("client_id", clientID),
				log.Int("product_id", product.ProductID),
				log.Int("requested_quantity", product.Total),
				log.Int("basket_quantity", basketTotal),
			)

			return customError.ErrBadRequest(
				fmt.Sprintf(
					"unequal product quantity in basket, product_id = %d",
					product.ProductID,
				),
			)
		}
	}

	if len(basket) != len(req) {
		return customError.ErrBadRequest("basket products do not match request")
	}

	return nil
}

func (r *BasketRepository) GetAllBasketProducts(
	ctx context.Context,
	clientID int,
	limit int,
	offset int,
) (*[]models.BasketProduct, *int, *float64, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	offset = (offset - 1) * limit

	query := `
		SELECT
			p.id,
			p.name,
			p.price,
			p.discount_price,
			COALESCE(img.image_path, '') AS image,
			p.is_paused,
			b.total
		FROM basket b
		JOIN products p
			ON p.id = b.product_id
		LEFT JOIN image_products img
			ON img.product_id = p.id
			AND img.is_main = true
		WHERE b.client_id = $1
		ORDER BY b.created_at DESC
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
		r.logger.Error("failed to get basket products",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, nil, nil, errors.New("failed to get basket products")
	}
	defer rows.Close()

	products := make([]models.BasketProduct, 0)

	for rows.Next() {

		var product models.BasketProduct

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.DiscountPrice,
			&product.Image,
			&product.IsPaused,
			&product.Total,
		)

		if err != nil {
			r.logger.Error("failed to scan basket product",
				log.Int("client_id", clientID),
				log.Error(err),
			)

			return nil, nil, nil, errors.New("failed to scan basket product")
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("failed to read basket products",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, nil, nil, errors.New("failed to read basket products")
	}

	query = `
		SELECT
			p.price,
			p.discount_price,
			b.total
		FROM basket b
		JOIN products p
			ON p.id = b.product_id
		WHERE b.client_id = $1
	`

	rows, err = r.db.QueryContext(
		ctx,
		query,
		clientID,
	)
	if err != nil {
		r.logger.Error("failed to get basket prices",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, nil, nil, errors.New("failed to get basket prices")
	}
	defer rows.Close()

	var totalPrice float64
	var total int

	for rows.Next() {
		var price float64
		var discountPrice *float64
		var count int

		err := rows.Scan(
			&price,
			&discountPrice,
			&count,
		)
		if err != nil {
			r.logger.Error("failed to scan basket price",
				log.Int("client_id", clientID),
				log.Error(err),
			)

			return nil, nil, nil, errors.New("failed to scan basket price")
		}

		total++

		finalPrice := price

		if discountPrice != nil {
			finalPrice = *discountPrice
		}

		totalPrice += math.Round(finalPrice) * float64(count)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, nil, errors.New("failed to iterate basket prices")
	}

	return &products, &total, &totalPrice, nil
}
