package handlers

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/metal-oopa/EcomMicroservices/services/product-service/productpb"
	"github.com/metal-oopa/EcomMicroservices/services/product-service/repository"
)

func TestCreateProduct(t *testing.T) {
	// Mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Prepare mock expectations
	mock.ExpectQuery("INSERT INTO products").
		WithArgs("Test Product", "Description", 99.99, 10).
		WillReturnRows(sqlmock.NewRows([]string{"product_id"}).AddRow(1))

	repo := repository.NewProductRepository(db)
	handler := NewProductServiceServer(repo)

	req := &productpb.CreateProductRequest{
		Name:        "Test Product",
		Description: "Description",
		Price:       99.99,
		Quantity:    10,
	}

	resp, err := handler.CreateProduct(context.Background(), req)
	if err != nil {
		t.Errorf("CreateProduct failed: %v", err)
	}

	if resp.Product.ProductId != "1" {
		t.Errorf("Expected ProductId '1', got '%s'", resp.Product.ProductId)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
