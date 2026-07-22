package main

import (
	"context"
	"log"
	"net"

	"ride-sharing/services/proto/proto"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	"ride-sharing/services/trip-service/internal/infrastructure/grpcserver"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/amqp"
	"ride-sharing/shared/env"

	"google.golang.org/grpc"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	ctx := context.Background()

	//In memory repository
	// repo := repository.NewInMemoryTripRepository()

	//connect to mongo
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(env.GetString("MONGO_URI", "mongodb://admin:SuperSecurePassword123!@mongodb-service:27017/?authSource=admin")))

	if err != nil {
		log.Fatal(err)
	}

	db := client.Database("rides")

	repo := repository.NewMongoTripRepository(db)

	// Initialize AMQP publisher for events
	amqpURL := env.GetString("AMQP_URL", "amqp://guest:guest@localhost:5672/")
	publisher, err := amqp.NewPublisher(amqpURL)
	if err != nil {
		log.Fatalf("Failed to create AMQP publisher: %v", err)
	}
	defer publisher.Close()

	// Initialize event publisher
	eventPublisher := events.NewTripEventPublisher(publisher)

	svc := service.NewTripService(*repo, eventPublisher)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterTripServiceServer(grpcServer, grpcserver.NewTripServiceServer(svc))

	log.Println("Trip service gRPC server listening on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("grpc serve: %v", err)
	}
}
