package main

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/metal-oopa/distributed-ecommerce/services/user-service/config"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/db"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/handlers"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/repository"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/userpb"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadConfig()

	// database connection
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

	// health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus(cfg.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	go func() {
		log.Printf("User Service is running on port %s", cfg.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Register with Consul
	servicePort, err := strconv.Atoi(cfg.Port)
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	err = utils.RegisterServiceWithConsul(cfg.ServiceName, servicePort, cfg.ConsulAddress)
	if err != nil {
		log.Fatalf("Failed to register service with Consul: %v", err)
	}

	// Block main goroutine
	select {}
}
