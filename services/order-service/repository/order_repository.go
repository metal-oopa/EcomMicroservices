package repository

import (
	"context"
	"database/sql"

	"github.com/metal-oopa/distributed-ecommerce/services/order-service/models"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrderByID(ctx context.Context, orderID int) (*models.Order, error)
	ListOrdersByUserID(ctx context.Context, userID int) ([]*models.Order, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	orderQuery := `
		INSERT INTO orders (user_id, total_amount, status, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING order_id
	`
	err = tx.QueryRowContext(ctx, orderQuery, order.UserID, order.TotalAmount, order.Status, order.CreatedAt).Scan(&order.OrderID)
	if err != nil {
		tx.Rollback()
		return err
	}

	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity)
		VALUES ($1, $2, $3)
	`
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, itemQuery, order.OrderID, item.ProductID, item.Quantity)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *orderRepository) GetOrderByID(ctx context.Context, orderID int) (*models.Order, error) {
	orderQuery := `
		SELECT order_id, user_id, total_amount, status, created_at
		FROM orders
		WHERE order_id = $1
	`

	order := &models.Order{}
	err := r.db.QueryRowContext(ctx, orderQuery, orderID).Scan(
		&order.OrderID,
		&order.UserID,
		&order.TotalAmount,
		&order.Status,
		&order.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	itemsQuery := `
		SELECT product_id, quantity
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.OrderItem{OrderID: orderID}
		err := rows.Scan(&item.ProductID, &item.Quantity)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *orderRepository) ListOrdersByUserID(ctx context.Context, userID int) ([]*models.Order, error) {
	ordersQuery := `
		SELECT order_id, total_amount, status, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, ordersQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{UserID: userID}
		err := rows.Scan(&order.OrderID, &order.TotalAmount, &order.Status, &order.CreatedAt)
		if err != nil {
			return nil, err
		}

		itemsQuery := `
			SELECT product_id, quantity
			FROM order_items
			WHERE order_id = $1
		`

		itemRows, err := r.db.QueryContext(ctx, itemsQuery, order.OrderID)
		if err != nil {
			return nil, err
		}
		defer itemRows.Close()

		for itemRows.Next() {
			item := models.OrderItem{OrderID: order.OrderID}
			err := itemRows.Scan(&item.ProductID, &item.Quantity)
			if err != nil {
				return nil, err
			}
			order.Items = append(order.Items, item)
		}

		orders = append(orders, order)
	}

	return orders, nil
}
