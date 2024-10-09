package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/models"
)

type CartRepository interface {
	AddItem(ctx context.Context, item *models.CartItem) error
	GetCart(ctx context.Context, userID int) ([]*models.CartItem, error)
	UpdateItemQuantity(ctx context.Context, item *models.CartItem) error
	RemoveItem(ctx context.Context, userID, productID int) error
	ClearCart(ctx context.Context, userID int) error
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) AddItem(ctx context.Context, item *models.CartItem) error {
	query := `
		SELECT quantity FROM cart_items WHERE user_id = $1 AND product_id = $2
	`

	var existingQuantity int32
	err := r.db.QueryRowContext(ctx, query, item.UserID, item.ProductID).Scan(&existingQuantity)
	if err == sql.ErrNoRows {
		insertQuery := `
			INSERT INTO cart_items (user_id, product_id, quantity)
			VALUES ($1, $2, $3)
		`
		_, err := r.db.ExecContext(ctx, insertQuery, item.UserID, item.ProductID, item.Quantity)
		return err
	} else if err != nil {
		return err
	}

	updateQuery := `
		UPDATE cart_items SET quantity = $1 WHERE user_id = $2 AND product_id = $3
	`
	_, err = r.db.ExecContext(ctx, updateQuery, existingQuantity+item.Quantity, item.UserID, item.ProductID)
	return err
}

func (r *cartRepository) GetCart(ctx context.Context, userID int) ([]*models.CartItem, error) {
	query := `
		SELECT product_id, quantity FROM cart_items WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.CartItem
	for rows.Next() {
		item := &models.CartItem{UserID: userID}
		err := rows.Scan(&item.ProductID, &item.Quantity)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *cartRepository) UpdateItemQuantity(ctx context.Context, item *models.CartItem) error {
	query := `
		UPDATE cart_items SET quantity = $1 WHERE user_id = $2 AND product_id = $3
	`

	result, err := r.db.ExecContext(ctx, query, item.Quantity, item.UserID, item.ProductID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("item not found in cart")
	}

	return nil
}

func (r *cartRepository) RemoveItem(ctx context.Context, userID, productID int) error {
	query := `
		DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, userID, productID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("item not found in cart")
	}

	return nil
}

func (r *cartRepository) ClearCart(ctx context.Context, userID int) error {
	query := `
		DELETE FROM cart_items WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
