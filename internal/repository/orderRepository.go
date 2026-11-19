package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"osush/internal/db"
	"osush/internal/handlers/request"
	"osush/internal/log"
	"osush/internal/models"
	"time"

	"github.com/jackc/pgx/v4"

	customError "osush/internal/errors"
)

type OrderRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewOrderRepository(db *db.LoggedDB, logger log.Logger) *OrderRepository {
	return &OrderRepository{
		db:     db,
		logger: logger,
	}
}

func (r *OrderRepository) ExistsByID(ctx context.Context, id int) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	r.logger.Debug("checking by id exists",
		log.Int("id", id),
	)

	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1
            	FROM orders
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

func (r *OrderRepository) ExistsClientsByID(ctx context.Context, id int, clientID int) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	r.logger.Debug("checking by id exists",
		log.Int("id", id),
	)

	var exists bool
	query := `
        SELECT EXISTS(
            SELECT 1
            	FROM orders
			WHERE id = $1 AND client_id = $2
        )
    `
	err := r.db.QueryRowContext(ctx, query, id, clientID).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check id existence", log.Error(err))
		return false, errors.New("failed to check id existence")
	}

	r.logger.Debug("id existence check completed",
		log.Bool("exists", exists),
	)
	return exists, nil
}

func (r *OrderRepository) CreateOrder(
	ctx context.Context,
	order models.Order,
	orderProducts []models.OrderProduct,
) (*int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.db.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, errors.New("failed to begin transaction")
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO orders (
			client_id,
			customer_name,
			phone_number,
			city,
			delivery_address,
			notes
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var orderID int

	err = tx.QueryRow(
		ctx,
		query,
		order.ClientID,
		order.CustomerName,
		order.PhoneNumber,
		order.City,
		order.DeliveryAddress,
		order.Notes,
	).Scan(&orderID)

	if err != nil {
		r.logger.Error("failed to create order",
			log.Int("client_id", order.ClientID),
			log.Error(err),
		)

		return nil, errors.New("failed to create order")
	}

	err = r.CreateOrderProduct(
		ctx,
		tx,
		orderID,
		orderProducts,
	)

	if err != nil {
		r.logger.Error("failed to create order products",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.logger.Error("failed to commit order transaction",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return nil, errors.New("failed to create order")
	}

	return &orderID, nil
}

func (r *OrderRepository) CreateOrderProduct(
	ctx context.Context,
	tx pgx.Tx,
	orderID int,
	orderProducts []models.OrderProduct,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO order_products (
			order_id,
			product_id,
			quantity,
			price,
			discount_price
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	for _, orderProduct := range orderProducts {

		_, err := tx.Exec(
			ctx,
			query,
			orderID,
			orderProduct.ProductID,
			orderProduct.Quantity,
			orderProduct.Price,
			orderProduct.DiscountPrice,
		)

		if err != nil {
			r.logger.Error("failed to create order product",
				log.Int("order_id", orderID),
				log.Int("product_id", orderProduct.ProductID),
				log.Error(err),
			)

			return errors.New("failed to create order product")
		}
	}

	return nil
}

func (r *OrderRepository) GetOrderByID(
	ctx context.Context,
	orderID int,
) (*models.Order, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting order by id",
		log.Int("order_id", orderID),
	)

	var order models.Order

	query := `
		SELECT
			id,
			COALESCE(client_id, 0),
			customer_name,
			phone_number,
			city,
			delivery_address,
			notes,
			is_new,
			is_send,
			created_at,
			updated_at
		FROM orders
		WHERE id = $1
	`

	err := r.db.DB.QueryRow(ctx, query, orderID).Scan(
		&order.ID,
		&order.ClientID,
		&order.CustomerName,
		&order.PhoneNumber,
		&order.City,
		&order.DeliveryAddress,
		&order.Notes,
		&order.IsNew,
		&order.IsSend,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		r.logger.Debug("order not found",
			log.Int("order_id", orderID),
		)

		return nil, customError.ErrNotFoundByID(orderID, "order")
	}

	if err != nil {
		r.logger.Error("failed to get order",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return nil, errors.New("failed to get order")
	}

	r.logger.Debug("order retrieved successfully",
		log.Int("order_id", orderID),
	)

	return &order, nil
}

func (r *OrderRepository) GetClientTotalOrderAndSpend(
	ctx context.Context,
	clientID int,
) (*int, *float64, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var orderCount int

	rows, err := r.db.QueryContext(
		ctx,
		`
			SELECT
				op.quantity,
				op.price,
				op.discount_price
			FROM orders o
			JOIN order_products op ON op.order_id = o.id
			WHERE o.client_id = $1
		`,
		clientID,
	)

	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var totalSpent float64

	for rows.Next() {
		var (
			quantity      int
			price         float64
			discountPrice *float64
		)

		if err := rows.Scan(
			&quantity,
			&price,
			&discountPrice,
		); err != nil {
			return nil, nil, err
		}

		productPrice := price

		if discountPrice != nil {
			productPrice = *discountPrice
		}

		totalSpent += float64(quantity) * math.Round(productPrice)
		orderCount++
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return &orderCount, &totalSpent, nil
}

func (r *OrderRepository) MarkOrderAsNotNew(
	ctx context.Context,
	orderID int,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("marking order as not new",
		log.Int("order_id", orderID),
	)

	query := `
		UPDATE orders
		SET
			is_new = FALSE,
			updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
		WHERE id = $1
	`

	_, err := r.db.DB.Exec(ctx, query, orderID)
	if err != nil {
		r.logger.Error("failed to mark order as not new",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return errors.New("failed to mark order as not new")
	}

	r.logger.Debug("order marked as not new",
		log.Int("order_id", orderID),
	)

	return nil
}

func (r *OrderRepository) GetOrders(
	ctx context.Context,
	filter request.OrderListFilter,
) (*[]models.Order, *int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting orders",
		log.String("search", filter.Search),
		log.Int("limit", filter.LimitInt),
		log.Int("offset", filter.OffsetInt),
	)

	var args []interface{}
	where := ""
	argIndex := 1

	if filter.Search != "" {
		where += fmt.Sprintf(
			` WHERE LOWER(COALESCE(customer_name, '')) LIKE LOWER($%d) `,
			argIndex,
		)

		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if filter.IsSend != nil {
		if where == "" {
			where = " WHERE "
		} else {
			where += " AND "
		}

		where += fmt.Sprintf(
			`is_send = $%d`,
			argIndex,
		)

		args = append(args, *filter.IsSend)
		argIndex++
	}

	countQuery := `
		SELECT COUNT(*)
		FROM orders
	` + where

	var total int

	err := r.db.DB.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total)

	if err != nil {
		r.logger.Error("failed to count orders",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to count orders")
	}

	filter.OffsetInt = (filter.OffsetInt - 1) * filter.LimitInt

	query := `
		SELECT
			o.id,
			o.customer_name,
			o.phone_number,
			o.city,
			o.is_send,
			o.is_new
		FROM orders o
	` + where + `
		ORDER BY o.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + `
		OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, filter.LimitInt, filter.OffsetInt)

	rows, err := r.db.DB.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to get orders",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get orders")
	}
	defer rows.Close()

	orders := make([]models.Order, 0)

	for rows.Next() {

		var order models.Order

		err := rows.Scan(
			&order.ID,
			&order.CustomerName,
			&order.PhoneNumber,
			&order.City,
			&order.IsSend,
			&order.IsNew,
		)

		if err != nil {
			r.logger.Error("failed to scan order",
				log.Error(err),
			)

			return nil, nil, errors.New("failed to scan order")
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("failed to iterate orders",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to iterate orders")
	}

	return &orders, &total, nil
}

func (r *OrderRepository) GetOrdersClient(
	ctx context.Context,
	clientID int,
) (*[]models.Order, *int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting orders")

	countQuery := `
		SELECT COUNT(*)
		FROM orders
		WHERE client_id = $1
	`

	var total int

	err := r.db.DB.QueryRow(
		ctx,
		countQuery,
		clientID,
	).Scan(&total)

	if err != nil {
		r.logger.Error("failed to count orders",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to count orders")
	}

	query := `
		SELECT
			id,
			delivery_address,
			is_send
		FROM orders
		WHERE client_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.DB.Query(ctx, query, clientID)
	if err != nil {
		r.logger.Error("failed to get orders",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get orders")
	}
	defer rows.Close()

	orders := make([]models.Order, 0)

	for rows.Next() {

		var order models.Order

		err := rows.Scan(
			&order.ID,
			&order.DeliveryAddress,
			&order.IsSend,
		)

		if err != nil {
			r.logger.Error("failed to scan order",
				log.Error(err),
			)

			return nil, nil, errors.New("failed to scan order")
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("failed to iterate orders",
			log.Error(err),
		)

		return nil, nil, errors.New("failed to iterate orders")
	}

	return &orders, &total, nil
}

func (r *OrderProductRepository) GetOrderTotals(
	ctx context.Context,
	orderID int,
) (*int, *float64, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting order totals",
		log.Int("order_id", orderID),
	)

	query := `
		SELECT
			COALESCE(SUM(quantity), 0),
			COALESCE(
				SUM(
					COALESCE(discount_price, price) * quantity
				),
				0
			)
		FROM order_products
		WHERE order_id = $1
	`

	var totalProduct int
	var totalPrice float64

	err := r.db.DB.QueryRow(
		ctx,
		query,
		orderID,
	).Scan(
		&totalProduct,
		&totalPrice,
	)

	if err != nil {
		r.logger.Error("failed to get order totals",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return nil, nil, errors.New("failed to get order totals")
	}

	r.logger.Debug("order totals retrieved",
		log.Int("order_id", orderID),
		log.Int("total_product", totalProduct),
		log.Float64("total_price", totalPrice),
	)

	return &totalProduct, &totalPrice, nil
}

// MarkOrderAsSend sets is_send = TRUE and returns the order's client id.
// changed is false when the order had already been marked as sent, so the
// caller does not notify the client twice if the admin presses again.
func (r *OrderRepository) MarkOrderAsSend(
	ctx context.Context,
	orderID int,
) (clientID int, changed bool, err error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug(
		"marking order as sent",
		log.Int("order_id", orderID),
	)

	// The CTE reads the old is_send value before the update is applied.
	query := `
		WITH previous AS (
			SELECT id, is_send FROM orders WHERE id = $1
		)
		UPDATE orders o
		SET
			is_send = TRUE,
			updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
		FROM previous p
		WHERE o.id = p.id
		RETURNING o.client_id, NOT p.is_send
	`

	err = r.db.DB.QueryRow(ctx, query, orderID).Scan(&clientID, &changed)
	if err != nil {
		r.logger.Error(
			"failed to mark order as sent",
			log.Int("order_id", orderID),
			log.Error(err),
		)

		return 0, false, errors.New("failed to mark order as sent")
	}

	r.logger.Debug(
		"order marked as sent successfully",
		log.Int("order_id", orderID),
		log.Bool("changed", changed),
	)

	return clientID, changed, nil
}
