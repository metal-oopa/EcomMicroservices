package main

import (
	"log"
	"net"
	"time"

	"github.com/metal-oopa/distributed-ecommerce/services/user-service/config"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/db"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/handlers"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/repository"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/userpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to the database
	database, err := db.Connect(cfg.DBSource)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	userRepo := repository.NewUserRepository(database)

	tokenDuration, err := time.ParseDuration(cfg.TokenDuration)
	if err != nil {
		log.Fatalf("Invalid token duration: %v", err)
	}

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Port, err)
	}

	grpcServer := grpc.NewServer()

	userpb.RegisterUserServiceServer(grpcServer, handlers.NewUserServiceServer(userRepo, cfg.JWTSecretKey, tokenDuration))

	reflection.Register(grpcServer)

	log.Printf("User Service is running on port %s", cfg.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
