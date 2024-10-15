package main

import (
	"log"
	"net"
	"strconv"

	"github.com/metal-oopa/distributed-ecommerce/services/order-service/config"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/db"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/handlers"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/orderpb"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/repository"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
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

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus(cfg.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	go func() {
		log.Printf("Order Service is running on port %s", cfg.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	servicePort, err := strconv.Atoi(cfg.Port)
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	err = utils.RegisterServiceWithConsul(cfg.ServiceName, servicePort, cfg.ConsulAddress)
	if err != nil {
		log.Fatalf("Failed to register service with Consul: %v", err)
	}

	select {}
}
