package repository

import (
	"context"
	"errors"
	"fmt"
	"osush/internal/db"
	"osush/internal/log"
	"osush/internal/models"
	"time"
)

type OrderProductRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewOrderProductRepository(db *db.LoggedDB, logger log.Logger) *OrderProductRepository {
	return &OrderProductRepository{
		db:     db,
		logger: logger,
	}
}

func (r *OrderProductRepository) GetOrderProductsByOrderID(
	ctx context.Context,
	orderID int,
) (*[]models.OrderProduct, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT
			op.id,
			op.order_id,
			op.product_id,
			op.quantity,
			op.price,
			op.discount_price,
			p.name,
			img.image_path
		FROM order_products op
		JOIN products p ON p.id = op.product_id
		LEFT JOIN image_products img
			ON img.product_id = p.id
			AND img.is_main = true
		WHERE op.order_id = $1
		ORDER BY op.id
	`

	rows, err := r.db.DB.Query(ctx, query, orderID)
	if err != nil {
		return nil, errors.New("failed to get order products")
	}
	defer rows.Close()

	orderProducts := make([]models.OrderProduct, 0)

	for rows.Next() {
		var orderProduct models.OrderProduct

		err := rows.Scan(
			&orderProduct.ID,
			&orderProduct.OrderID,
			&orderProduct.ProductID,
			&orderProduct.Quantity,
			&orderProduct.Price,
			&orderProduct.DiscountPrice,
			&orderProduct.ProductName,
			&orderProduct.ProductImagePath,
		)
		if err != nil {
			return nil, errors.New("failed to scan order product")
		}

		orderProducts = append(orderProducts, orderProduct)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("failed to iterate order products")
	}

	return &orderProducts, nil
}

func (r *OrderProductRepository) GetProductOrderCount(ctx context.Context, productID int) (*int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
        SELECT COUNT(*)
        FROM order_products
        WHERE product_id = $1
    `

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		productID,
	).Scan(&count)

	if err != nil {
		r.logger.Error(
			"failed to get product order count",
			log.Int("product_id", productID),
			log.Error(err),
		)

		return nil, fmt.Errorf(
			"failed to get product order count: %w",
			err,
		)
	}

	return &count, nil
}
