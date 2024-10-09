package handlers

import (
	"context"
	"strconv"

	"github.com/metal-oopa/distributed-ecommerce/services/product-service/models"
	"github.com/metal-oopa/distributed-ecommerce/services/product-service/productpb"
	"github.com/metal-oopa/distributed-ecommerce/services/product-service/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductServiceServer struct {
	productpb.UnimplementedProductServiceServer
	repo repository.ProductRepository
}

func NewProductServiceServer(repo repository.ProductRepository) productpb.ProductServiceServer {
	return &ProductServiceServer{repo: repo}
}

func (s *ProductServiceServer) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	if req.Name == "" || req.Price <= 0 || req.Quantity < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input")
	}

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Quantity:    req.Quantity,
	}

	err := s.repo.CreateProduct(ctx, product)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	return &productpb.CreateProductResponse{
		Product: mapProductToProto(product),
	}, nil
}

func (s *ProductServiceServer) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	productID, err := strconv.Atoi(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product ID")
	}

	product, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found")
	}

	return &productpb.GetProductResponse{
		Product: mapProductToProto(product),
	}, nil
}

func (s *ProductServiceServer) ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
	products, err := s.repo.ListProducts(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	var productProtos []*productpb.Product
	for _, product := range products {
		productProtos = append(productProtos, mapProductToProto(product))
	}

	return &productpb.ListProductsResponse{
		Products: productProtos,
	}, nil
}

func (s *ProductServiceServer) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.UpdateProductResponse, error) {
	productID, err := strconv.Atoi(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product ID")
	}

	product := &models.Product{
		ProductID:   productID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Quantity:    req.Quantity,
	}

	err = s.repo.UpdateProduct(ctx, product)
	if err != nil {
		if err.Error() == "product not found" {
			return nil, status.Errorf(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	return &productpb.UpdateProductResponse{
		Product: mapProductToProto(product),
	}, nil
}

func (s *ProductServiceServer) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.DeleteProductResponse, error) {
	productID, err := strconv.Atoi(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product ID")
	}

	err = s.repo.DeleteProduct(ctx, productID)
	if err != nil {
		if err.Error() == "product not found" {
			return nil, status.Errorf(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}

	return &productpb.DeleteProductResponse{
		Message: "Product deleted successfully",
	}, nil
}

func mapProductToProto(product *models.Product) *productpb.Product {
	return &productpb.Product{
		ProductId:   strconv.Itoa(product.ProductID),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Quantity:    product.Quantity,
	}
}
