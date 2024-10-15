package handlers

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/models"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/orderpb"
	productpb "github.com/metal-oopa/distributed-ecommerce/services/order-service/productpb"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/repository"
	"github.com/metal-oopa/distributed-ecommerce/services/order-service/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderServiceServer struct {
	orderpb.UnimplementedOrderServiceServer
	repo          repository.OrderRepository
	jwtSecretKey  string
	stripeAPIKey  string
	consulClient  *consulapi.Client
	productClient productpb.ProductServiceClient
}

func NewOrderServiceServer(repo repository.OrderRepository, jwtSecretKey, stripeAPIKey string, consulClient *consulapi.Client, productClient productpb.ProductServiceClient) orderpb.OrderServiceServer {
	utils.InitializeStripe(stripeAPIKey)

	return &OrderServiceServer{
		repo:          repo,
		jwtSecretKey:  jwtSecretKey,
		stripeAPIKey:  stripeAPIKey,
		consulClient:  consulClient,
		productClient: productClient,
	}
}

func (s *OrderServiceServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	if len(req.Items) == 0 || req.PaymentMethodId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "items and payment method are required")
	}

	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	// product prices from Product Service
	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range req.Items {
		productID, err := strconv.Atoi(item.ProductId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid product ID")
		}

		productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{
			ProductId: item.ProductId,
		})
		if err != nil {
			grpcErr, ok := status.FromError(err)
			if ok && grpcErr.Code() == codes.NotFound {
				return nil, status.Errorf(codes.InvalidArgument, "product not found")
			}
			return nil, status.Errorf(codes.Internal, "failed to fetch product: %v", err)
		}

		product := productResp.Product

		if product.Quantity < item.Quantity {
			return nil, status.Errorf(codes.InvalidArgument, "product %s is out of stock", product.Name)
		}

		itemTotal := product.Price * float64(item.Quantity)
		totalAmount += itemTotal

		orderItem := models.OrderItem{
			ProductID: productID,
			Quantity:  item.Quantity,
		}
		orderItems = append(orderItems, orderItem)
	}

	// Process payment with Stripe
	amountInCents := int64(totalAmount * 100) // Convert to cents
	_, err = utils.CreatePaymentIntent(amountInCents, "usd", req.PaymentMethodId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "payment failed: %v", err)
	}

	order := &models.Order{
		UserID:      userID,
		TotalAmount: totalAmount,
		Status:      "Confirmed",
		CreatedAt:   time.Now(),
		Items:       orderItems,
	}

	err = s.repo.CreateOrder(ctx, order)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	return &orderpb.CreateOrderResponse{
		Order: mapOrderToProto(order),
	}, nil
}

func (s *OrderServiceServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	orderID, err := strconv.Atoi(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order ID")
	}

	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "order not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get order: %v", err)
	}

	return &orderpb.GetOrderResponse{
		Order: mapOrderToProto(order),
	}, nil
}

func (s *OrderServiceServer) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	orders, err := s.repo.ListOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list orders: %v", err)
	}

	var orderProtos []*orderpb.Order
	for _, order := range orders {
		orderProtos = append(orderProtos, mapOrderToProto(order))
	}

	return &orderpb.ListOrdersResponse{
		Orders: orderProtos,
	}, nil
}

func mapOrderToProto(order *models.Order) *orderpb.Order {
	var items []*orderpb.OrderItem
	for _, item := range order.Items {
		items = append(items, &orderpb.OrderItem{
			ProductId: strconv.Itoa(item.ProductID),
			Quantity:  item.Quantity,
		})
	}

	return &orderpb.Order{
		OrderId:     strconv.Itoa(order.OrderID),
		UserId:      strconv.Itoa(order.UserID),
		Items:       items,
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
		CreatedAt:   order.CreatedAt.Format(time.RFC3339),
	}
}
