package main

import (
	"log"
	"net"

	"github.com/metal-oopa/distributed-ecommerce/services/order-service/config"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/db"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/handlers"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/orderpb"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/repository"
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

	orderRepo := repository.NewOrderRepository(database)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Port, err)
	}

	grpcServer := grpc.NewServer()

	orderpb.RegisterOrderServiceServer(grpcServer, handlers.NewOrderServiceServer(orderRepo, cfg.JWTSecretKey, cfg.StripeAPIKey))

	reflection.Register(grpcServer)

	log.Printf("Order Service is running on port %s", cfg.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
