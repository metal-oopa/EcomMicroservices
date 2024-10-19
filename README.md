# Overview
This project is a microservices-based e-commerce platform designed to be scalable and maintainable. Each microservice is responsible for a specific domain, allowing for independent development, deployment, and scaling.

## Key Components
### Implemented Services

 - **User Service**
    - Manages user authentication and profiles.
    - Handles user registration, login, and profile management.
    
 - **Product Service**
     - Handles product listings and inventory management.
     - Supports adding, updating, and retrieving product information.

 - **Cart Service**
     - Manages user shopping carts.
     - Allows adding, updating, and removing items from the cart.
 - **Order Service**
     - Processes orders and tracks their status.
     - Manages order creation, payment status, and order history.

### Planned Services
 - **Notification Service**
     - Sends order confirmations and updates to users.
     - Supports email and SMS notifications.

## Tech Stack
 - **Programming Languages**: Go
 - **Database**: PostgreSQL
 - **API Gateway**: Envoy Proxy (Planned)
 - **Service Discovery**: Consul
 - **Inter-Service Communication**: gRPC
 - **Containerization**: Docker
 - **Orchestration**: Docker Compose (Kubernetes in future plans)
 - **Message Queue**(Planned): Kafka/RabbitMQ
 - **Monitoring and Logging**(Planned): Prometheus, Grafana

## Getting Started
### Prerequisites - Docker

### Installation

1. Clone the repository
```bash
git clone https://github.com/metal-oopa/EcomMicroservices
cd EcomMicroservices
```

2. Add necessary environment variables
3. Start Services with Docker Compose
```bash
docker-compose up --build
```
4. To test the services, you can use grpcurl or any gRPC client of your choice.
```bash
grpcurl -plaintext -d '{"username": "testuser", "email": "testuser@test.com", "password": "password"}' localhost:50051 user.UserService/RegisterUser
```

## To-Do List
    - Implement remaining services
    - Add API Gateway
    - Implement message queue
    - Add monitoring and logging
    - Implement CI/CD pipeline
    - Add Kubernetes support
    - Implement frontend