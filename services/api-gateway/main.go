package main

import (
	"log"
	"net/http"

	"ride-sharing/services/proto/proto"
	"ride-sharing/shared/env"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	httpAddr   = env.GetString("GATEWAY_HTTP_ADDR", ":8081")
	tripAddr   = env.GetString("TRIP_SERVICE_ADDR", "trip-service:50051")
	tripClient proto.TripServiceClient
)

func main() {
	log.Println("Starting API Gateway")

	conn, err := grpc.Dial(tripAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to trip service: %v", err)
	}
	defer conn.Close()

	tripClient = proto.NewTripServiceClient(conn)

	mux := http.NewServeMux()
	mux.HandleFunc("/trip/preview", handleTripPreview)
	mux.HandleFunc("/trip/create", handleCreateTrip)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	log.Printf("API Gateway is running on %s\n", httpAddr)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
