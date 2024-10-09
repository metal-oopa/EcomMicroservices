package main

import (
	"log"
	"net"

	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/cartpb"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/config"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/db"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/handlers"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadConfig()

	database, err := db.Connect(cfg.DBSource)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	cartRepo := repository.NewCartRepository(database)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Port, err)
	}

	grpcServer := grpc.NewServer()

	cartpb.RegisterCartServiceServer(grpcServer, handlers.NewCartServiceServer(cartRepo, cfg.JWTSecretKey))

	reflection.Register(grpcServer)

	log.Printf("Cart Service is running on port %s", cfg.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
