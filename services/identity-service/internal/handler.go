package identity

import (
	"context"
	"errors"

	identityv1 "github.com/Guizzs26/fintrack/services/identity-service/gen/go"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	identityv1.UnimplementedIdentityServiceServer
	service *Service
}

func NewServer(s *Service) *Server {
	return &Server{service: s}
}

func (s *Server) Register(ctx context.Context, req *identityv1.RegisterRequest) (*identityv1.RegisterResponse, error) {
	if req.GetName() == "" || req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "name, email and password are rqeuired")
	}

	user, err := s.service.Register(ctx, req.GetName(), req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyInUse) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &identityv1.RegisterResponse{UserId: user.ID.String()}, nil
}

func (s *Server) Login(ctx context.Context, req *identityv1.LoginRequest) (*identityv1.LoginResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	tokenPair, err := s.service.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "failed to login user")
	}

	return &identityv1.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

func (s *Server) RefreshToken(ctx context.Context, req *identityv1.RefreshTokenRequest) (*identityv1.LoginResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	tokenPair, err := s.service.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	return &identityv1.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	userID, _ := uuid.Parse("...") // fake

	if err := s.service.Logout(ctx, userID); err != nil {
		// Logar o erro aqui
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &empty.Empty{}, nil
}
