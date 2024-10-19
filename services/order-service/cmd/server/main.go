package main

import (
	"log"
	"net"
	"strconv"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/metal-oopa/EcomMicroservices/services/order-service/auth"
	"github.com/metal-oopa/EcomMicroservices/services/order-service/config"
	"github.com/metal-oopa/EcomMicroservices/services/order-service/db"
	"github.com/metal-oopa/EcomMicroservices/services/order-service/handlers"
	"github.com/metal-oopa/EcomMicroservices/services/order-service/orderpb"
	"github.com/metal-oopa/EcomMicroservices/services/order-service/productpb"
	"github.com/metal-oopa/EcomMicroservices/services/order-service/repository"
	"github.com/metal-oopa/EcomMicroservices/services/order-service/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(auth.UnaryAuthInterceptor(cfg.JWTSecretKey)))

	// Consul client
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = cfg.ConsulAddress
	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	productServiceAddress, err := utils.GetServiceAddress(consulClient, "product-service")
	if err != nil {
		log.Fatalf("Failed to get Product Service address: %v", err)
	}

	// gRPC connection with Product Service
	productConn, err := grpc.NewClient(productServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Product Service: %v", err)
	}
	defer productConn.Close()
	productClient := productpb.NewProductServiceClient(productConn)

	orderpb.RegisterOrderServiceServer(grpcServer, handlers.NewOrderServiceServer(orderRepo, cfg.JWTSecretKey, cfg.StripeAPIKey, consulClient, productClient))

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
