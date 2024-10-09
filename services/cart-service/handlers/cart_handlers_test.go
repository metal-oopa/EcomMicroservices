package handlers

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/cartpb"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/repository"
)

func TestAddItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT quantity FROM cart_items").
		WithArgs(1, 1).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec("INSERT INTO cart_items").
		WithArgs(1, 1, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := repository.NewCartRepository(db)
	handler := NewCartServiceServer(repo, "test-secret-key")

	req := &cartpb.AddItemRequest{
		UserId:    "1",
		ProductId: "1",
		Quantity:  2,
	}

	resp, err := handler.AddItem(context.Background(), req)
	if err != nil {
		t.Errorf("AddItem failed: %v", err)
	}

	if resp.Message != "Item added to cart successfully" {
		t.Errorf("Unexpected response message: %s", resp.Message)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
