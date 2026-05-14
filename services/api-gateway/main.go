package main

import (
	"context"
	"log"
	"net/http"

	"ride-sharing/services/proto/proto"
	"ride-sharing/shared/env"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	httpAddr      = env.GetString("GATEWAY_HTTP_ADDR", ":8081")
	tripAddr      = env.GetString("TRIP_SERVICE_ADDR", "trip-service:50051")
	amqpURL       = env.GetString("AMQP_URL", "amqp://guest:guest@rabbitmq:5672/")
	allowedOrigin = env.GetString("ALLOWED_ORIGIN", "http://localhost:3000")
	tripClient    proto.TripServiceClient
)

func main() {
	log.Println("Starting API Gateway")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	riderConsumer, err := NewRiderEventConsumer(amqpURL)
	if err != nil {
		log.Fatalf("failed to create rider event consumer: %v", err)
	}
	defer riderConsumer.Close()

	conn, err := grpc.Dial(tripAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to trip service: %v", err)
	}
	defer conn.Close()

	tripClient = proto.NewTripServiceClient(conn)

	go func() {
		if err := riderConsumer.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Rider event consumer stopped: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/trip/preview", handleTripPreview)
	mux.HandleFunc("/trip/create", handleCreateTrip)
	mux.HandleFunc("/ws/riders", handleRiderWs)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: withCORS(mux),
	}

	log.Printf("API Gateway is running on %s\n", httpAddr)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func withCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
