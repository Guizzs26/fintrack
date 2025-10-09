package identity

import (
	"context"
	"errors"

	identityv1 "github.com/Guizzs26/fintrack/services/identity-service/gen/go"
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
