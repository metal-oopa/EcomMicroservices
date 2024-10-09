package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/userpb"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port     = ":50051"
	dbDriver = "postgres"
)

type server struct {
	userpb.UnimplementedUserServiceServer
	db *sql.DB
}

func (s *server) RegisterUser(ctx context.Context, req *userpb.RegisterUserRequest) (*userpb.RegisterUserResponse, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("all fields are required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var userID int
	err = s.db.QueryRowContext(ctx, `
        INSERT INTO users (username, email, password_hash)
        VALUES ($1, $2, $3)
        RETURNING user_id
    `, req.Username, req.Email, string(hashedPassword)).Scan(&userID)
	if err != nil {
		return nil, err
	}

	return &userpb.RegisterUserResponse{
		UserId: fmt.Sprintf("%d", userID),
	}, nil
}

func (s *server) LoginUser(ctx context.Context, req *userpb.LoginUserRequest) (*userpb.LoginUserResponse, error) {
	var (
		userID       int
		passwordHash string
	)
	err := s.db.QueryRowContext(ctx, `
        SELECT user_id, password_hash FROM users WHERE email = $1
    `, req.Email).Scan(&userID, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := createToken(req.Email)

	if err != nil {
		return &userpb.LoginUserResponse{
			Token: "tobehandled",
		}, nil
	}

	return &userpb.LoginUserResponse{
		Token: token,
	}, nil
}

func (s *server) GetUserProfile(ctx context.Context, req *userpb.GetUserProfileRequest) (*userpb.GetUserProfileResponse, error) {
	var (
		username string
		email    string
	)
	err := s.db.QueryRowContext(ctx, `
        SELECT username, email FROM users WHERE user_id = $1
    `, req.UserId).Scan(&username, &email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &userpb.GetUserProfileResponse{
		UserId:   req.UserId,
		Username: username,
		Email:    email,
	}, nil
}

func main() {
	dbSource := os.Getenv("DB_SOURCE")
	db, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}
	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, &server{db: db})

	reflection.Register(grpcServer)

	log.Printf("User Service is running on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func createToken(username string) (string, error) {
	var secretKey = []byte("secret-key")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyToken(tokenString string) error {
	var secretKey = []byte("secret-key")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}
