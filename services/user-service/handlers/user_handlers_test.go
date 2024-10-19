package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/metal-oopa/EcomMicroservices/services/user-service/repository"
	"github.com/metal-oopa/EcomMicroservices/services/user-service/userpb"
)

func TestRegisterUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("testuser", "test@example.com", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	repo := repository.NewUserRepository(db)
	handler := NewUserServiceServer(repo, "test-secret-key", 24*time.Hour)

	req := &userpb.RegisterUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := handler.RegisterUser(context.Background(), req)
	if err != nil {
		t.Errorf("RegisterUser failed: %v", err)
	}

	if resp.User.Username != req.Username {
		t.Errorf("Expected UserId '1', got '%s'", resp.User.Username)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
