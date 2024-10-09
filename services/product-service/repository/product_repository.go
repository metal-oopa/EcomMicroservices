package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/metal-oopa/distributed-ecommerce/services/product-service/models"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	GetProductByID(ctx context.Context, productID int) (*models.Product, error)
	ListProducts(ctx context.Context) ([]*models.Product, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, productID int) error
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (name, description, price, quantity)
		VALUES ($1, $2, $3, $4)
		RETURNING product_id
	`

	err := r.db.QueryRowContext(ctx, query, product.Name, product.Description, product.Price, product.Quantity).Scan(&product.ProductID)
	return err
}

func (r *productRepository) GetProductByID(ctx context.Context, productID int) (*models.Product, error) {
	query := `
		SELECT product_id, name, description, price, quantity
		FROM products
		WHERE product_id = $1
	`

	product := &models.Product{}
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&product.ProductID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Quantity,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("product not found")
	}
	return product, err
}

func (r *productRepository) ListProducts(ctx context.Context) ([]*models.Product, error) {
	query := `
		SELECT product_id, name, description, price, quantity
		FROM products
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		err := rows.Scan(
			&product.ProductID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Quantity,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func (r *productRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, quantity = $4
		WHERE product_id = $5
	`

	result, err := r.db.ExecContext(ctx, query, product.Name, product.Description, product.Price, product.Quantity, product.ProductID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, productID int) error {
	query := `
		DELETE FROM products
		WHERE product_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, productID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}
