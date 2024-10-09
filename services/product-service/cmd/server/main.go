package main

import (
	"log"
	"net"

	"github.com/metal-oopa/distributed-ecommerce/services/product-service/config"
	"github.com/metal-oopa/distributed-ecommerce/services/product-service/db"
	"github.com/metal-oopa/distributed-ecommerce/services/product-service/handlers"
	"github.com/metal-oopa/distributed-ecommerce/services/product-service/productpb"
	"github.com/metal-oopa/distributed-ecommerce/services/product-service/repository"
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

	productRepo := repository.NewProductRepository(database)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Port, err)
	}

	grpcServer := grpc.NewServer()

	productpb.RegisterProductServiceServer(grpcServer, handlers.NewProductServiceServer(productRepo))

	reflection.Register(grpcServer)

	log.Printf("Product Service is running on port %s", cfg.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
