package main

import (
	"log"
	"net"
	"strconv"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/auth"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/cartpb"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/config"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/db"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/handlers"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/productpb"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/repository"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/utils"
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

	cartRepo := repository.NewCartRepository(database)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Port, err)
	}

	// Initialize Consul client
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = cfg.ConsulAddress
	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(auth.UnaryAuthInterceptor(cfg.JWTSecretKey)))

	productServiceAddress, err := utils.GetServiceAddress(consulClient, "product-service")
	if err != nil {
		log.Fatalf("Failed to get Product Service address: %v", err)
	}

	productConn, err := grpc.NewClient(productServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Product Service: %v", err)
	}

	defer productConn.Close()
	productClient := productpb.NewProductServiceClient(productConn)

	cartpb.RegisterCartServiceServer(grpcServer, handlers.NewCartServiceServer(cartRepo, cfg.JWTSecretKey, consulClient, productClient))

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus(cfg.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	go func() {
		log.Printf("Cart Service is running on port %s", cfg.Port)
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
