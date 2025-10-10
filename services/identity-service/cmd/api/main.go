package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	identityv1 "github.com/Guizzs26/fintrack/services/identity-service/gen/go"
	identity "github.com/Guizzs26/fintrack/services/identity-service/internal"
	"github.com/Guizzs26/fintrack/services/identity-service/internal/platform/config"
	"google.golang.org/grpc"
)

type InMemoryPublisher struct{}

// Publish simula a publicação de um evento, logando-o na saída padrão.
func (p *InMemoryPublisher) Publish(ctx context.Context, topic string, eventData []byte) error {
	slog.Info("EVENT PUBLISHED", slog.String("topic", topic), slog.String("payload", string(eventData)))
	return nil
}

func main() {
	cfg := config.Config{
		PasswordPepper: "aksdaksdasokdad",
	}

	if err := run(context.Background(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "application finished with an error: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("Starting identity-service...")

	dbClient, err := identity.NewDynamoDBClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create dynamodb client: %v", err)
	}
	publisher := &InMemoryPublisher{}

	tableName := "FintrackUsers"
	userRepo := identity.NewDynamoDBUserRepository(dbClient, tableName)
	tokenRepo := identity.NewDynamoDBTokenRepository(dbClient, tableName)

	jwtSecret := "hueheuehuhueheu"
	accessTokenTTL := time.Minute * 15
	refreshTokenTTL := time.Hour * 24 * 7
	pepper := "kkkkkkkkkkkkkkkkkkkkkkkkkkkk"

	pwdManager := identity.NewPasswordManager(pepper)
	jwtManager := identity.NewJWTManager(jwtSecret, accessTokenTTL)

	tokenService := identity.NewTokenService(tokenRepo, jwtManager, refreshTokenTTL)
	userService := identity.NewService(userRepo, tokenService, pwdManager, publisher)

	grpcHandler := identity.NewServer(userService)

	grpcServer := grpc.NewServer()
	identityv1.RegisterIdentityServiceServer(grpcServer, grpcHandler)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen on port 50051: %v", err)
	}

	go func() {
		slog.Info("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("gRPC server failed to serve", slog.String("error", err.Error()))
			cancel()
		}
	}()

	<-ctx.Done()

	slog.Info("Shutting down server gracefully...")
	grpcServer.GracefulStop()

	return nil
}
