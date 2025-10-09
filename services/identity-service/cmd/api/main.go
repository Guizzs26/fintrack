package main

import (
	"context"
	"log"
	"net"
	"time"

	identityv1 "github.com/Guizzs26/fintrack/services/identity-service/gen/go"
	identity "github.com/Guizzs26/fintrack/services/identity-service/internal"
	"google.golang.org/grpc"
)

type InMemoryPublisher struct{}

func (p *InMemoryPublisher) Publish(ctx context.Context, topic string, eventData []byte) error {
	log.Printf("EVENT PUBLISHED on topic %s: %s", topic, string(eventData))
	return nil
}

func main() {
	ctx := context.Background()
	dbClient, err := identity.NewDynamoDBClient(ctx)
	if err != nil {
		log.Fatalf("failed to create dynamodb client: %v", err)
	}

	jwtSecret := "as0dasoidjaodiaus0e912ijkxkkkkkkkkkk"
	accessTokenTTL := time.Minute * 15
	tokenManager := identity.NewJWTManager(jwtSecret, accessTokenTTL)
	tableName := "FintrackUsers"
	repo := identity.NewDynamoDBUserRepository(dbClient, tableName)
	publisher := &InMemoryPublisher{}
	service := identity.NewService(repo, nil, tokenManager, publisher)
	handler := identity.NewServer(service)
	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	identityv1.RegisterIdentityServiceServer(grpcServer, handler)

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
