package main

import (
	"log"
	"net"

	"ride-sharing/services/proto/proto"
	"ride-sharing/services/trip-service/internal/infrastructure/grpcserver"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"

	"google.golang.org/grpc"
)

func main() {
	inmemory := repository.NewInMemoryTripRepository()
	svc := service.NewTripService(*inmemory)

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
