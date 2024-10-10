package handlers

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/orderpb"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/repository"
	"github.com/stripe/stripe-go/v72"
)

func TestCreateOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	stripe.Key = "sk_test_mock_key"

	// TODO: Mock Stripe payment intent creation

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO orders").
		WithArgs(1, 20.0, "Confirmed", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"order_id"}).AddRow(1))
	mock.ExpectExec("INSERT INTO order_items").
		WithArgs(1, 1, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := repository.NewOrderRepository(db)
	handler := NewOrderServiceServer(repo, "test-secret-key", "sk_test_mock_key")

	req := &orderpb.CreateOrderRequest{
		UserId:          "1",
		PaymentMethodId: "pm_mock",
		Items: []*orderpb.OrderItem{
			{ProductId: "1", Quantity: 2},
		},
	}

	resp, err := handler.CreateOrder(context.Background(), req)
	if err != nil {
		t.Errorf("CreateOrder failed: %v", err)
	}

	if resp.Order.OrderId != "1" {
		t.Errorf("Expected OrderId '1', got '%s'", resp.Order.OrderId)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
