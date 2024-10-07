package main

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/github.com/metal-oopa/distributed-ecommerce/services/user-service/userpb"
)

func TestRegisterUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &server{db: db}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("testuser", "test@example.com", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	req := &userpb.RegisterUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := s.RegisterUser(context.Background(), req)
	if err != nil {
		t.Errorf("RegisterUser failed: %v", err)
	}
	if resp.UserId != "1" {
		t.Errorf("Expected UserId '1', got '%s'", resp.UserId)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
