package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/metal-oopa/distributed-ecommerce/services/user-service/models"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/repository"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/userpb"
	"github.com/metal-oopa/distributed-ecommerce/services/user-service/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceServer struct {
	userpb.UnimplementedUserServiceServer
	repo          repository.UserRepository
	jwtSecretKey  string
	tokenDuration time.Duration
}

func NewUserServiceServer(repo repository.UserRepository, jwtSecretKey string, tokenDuration time.Duration) userpb.UserServiceServer {
	return &UserServiceServer{
		repo:          repo,
		jwtSecretKey:  jwtSecretKey,
		tokenDuration: tokenDuration,
	}
}

func (s *UserServiceServer) RegisterUser(ctx context.Context, req *userpb.RegisterUserRequest) (*userpb.RegisterUserResponse, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "all fields are required")
	}

	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "user already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &userpb.RegisterUserResponse{
		User: &userpb.User{
			UserId:   strconv.Itoa(user.UserID),
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (s *UserServiceServer) LoginUser(ctx context.Context, req *userpb.LoginUserRequest) (*userpb.LoginUserResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email and password are required")
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	token, err := utils.GenerateJWT(s.jwtSecretKey, user.UserID, s.tokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	return &userpb.LoginUserResponse{
		Token: token,
	}, nil
}

func (s *UserServiceServer) GetUserProfile(ctx context.Context, req *userpb.GetUserProfileRequest) (*userpb.GetUserProfileResponse, error) {
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	return &userpb.GetUserProfileResponse{
		User: &userpb.User{
			UserId:   strconv.Itoa(user.UserID),
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}
