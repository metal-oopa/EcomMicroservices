package handlers

import (
	"context"
	"strconv"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/cartpb"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/models"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/productpb"
	"github.com/metal-oopa/distributed-ecommerce/services/cart-service/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CartServiceServer struct {
	cartpb.UnimplementedCartServiceServer
	repo          repository.CartRepository
	jwtSecretKey  string
	productClient productpb.ProductServiceClient
	consulClient  *consulapi.Client
}

func NewCartServiceServer(repo repository.CartRepository, jwtSecretKey string, consulClient *consulapi.Client, productClient productpb.ProductServiceClient) cartpb.CartServiceServer {
	return &CartServiceServer{
		repo:          repo,
		jwtSecretKey:  jwtSecretKey,
		productClient: productClient,
		consulClient:  consulClient,
	}
}

func (s *CartServiceServer) AddItem(ctx context.Context, req *cartpb.AddItemRequest) (*cartpb.AddItemResponse, error) {
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	if req.ProductId == "" || req.Quantity <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product ID or quantity")
	}

	productID, err := strconv.Atoi(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product ID")
	}

	item := &models.CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  req.Quantity,
	}

	err = s.repo.AddItem(ctx, item)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add item to cart: %v", err)
	}

	return &cartpb.AddItemResponse{
		Message: "Item added to cart successfully",
	}, nil
}

func (s *CartServiceServer) GetCart(ctx context.Context, req *cartpb.GetCartRequest) (*cartpb.GetCartResponse, error) {
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	items, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart: %v", err)
	}

	var cartItems []*cartpb.CartItem
	for _, item := range items {
		cartItems = append(cartItems, &cartpb.CartItem{
			ProductId: strconv.Itoa(item.ProductID),
			Quantity:  item.Quantity,
		})
	}

	return &cartpb.GetCartResponse{
		Items: cartItems,
	}, nil
}

func (s *CartServiceServer) UpdateItemQuantity(ctx context.Context, req *cartpb.UpdateItemQuantityRequest) (*cartpb.UpdateItemQuantityResponse, error) {
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	productID, err := strconv.Atoi(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product ID")
	}

	if req.Quantity <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "quantity must be greater than zero")
	}

	item := &models.CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  req.Quantity,
	}

	err = s.repo.UpdateItemQuantity(ctx, item)
	if err != nil {
		if err.Error() == "item not found in cart" {
			return nil, status.Errorf(codes.NotFound, "item not found in cart")
		}
		return nil, status.Errorf(codes.Internal, "failed to update item quantity: %v", err)
	}

	return &cartpb.UpdateItemQuantityResponse{
		Message: "Item quantity updated successfully",
	}, nil
}

func (s *CartServiceServer) RemoveItem(ctx context.Context, req *cartpb.RemoveItemRequest) (*cartpb.RemoveItemResponse, error) {
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	productID, err := strconv.Atoi(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product ID")
	}

	err = s.repo.RemoveItem(ctx, userID, productID)
	if err != nil {
		if err.Error() == "item not found in cart" {
			return nil, status.Errorf(codes.NotFound, "item not found in cart")
		}
		return nil, status.Errorf(codes.Internal, "failed to remove item from cart: %v", err)
	}

	return &cartpb.RemoveItemResponse{
		Message: "Item removed from cart successfully",
	}, nil
}

func (s *CartServiceServer) ClearCart(ctx context.Context, req *cartpb.ClearCartRequest) (*cartpb.ClearCartResponse, error) {
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	err = s.repo.ClearCart(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to clear cart: %v", err)
	}

	return &cartpb.ClearCartResponse{
		Message: "Cart cleared successfully",
	}, nil
}
