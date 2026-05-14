package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ride-sharing/services/driver-service/internal/domain"
	"ride-sharing/services/driver-service/internal/infrastructure/events"
	"ride-sharing/services/driver-service/internal/infrastructure/repository"
	"ride-sharing/services/driver-service/internal/service"
	"ride-sharing/shared/amqp"
	"ride-sharing/shared/env"
)

func main() {
	log.Println("Starting Driver Service")

	// Get configuration
	amqpURL := env.GetString("AMQP_URL", "amqp://guest:guest@localhost:5672/")

	// Initialize repositories
	driverRepo := repository.NewInMemoryDriverRepository()
	assignmentRepo := repository.NewInMemoryTripAssignmentRepository()

	// Initialize AMQP publisher
	publisher, err := amqp.NewPublisher(amqpURL)
	if err != nil {
		log.Fatalf("Failed to create AMQP publisher: %v", err)
	}
	defer publisher.Close()

	// Initialize event publisher
	eventPublisher := service.NewAMQPDriverEventPublisher(publisher)

	// Initialize service
	svc := service.NewDriverService(driverRepo, assignmentRepo, eventPublisher)

	// Initialize event consumer
	consumer, err := events.NewTripEventConsumer(amqpURL, svc)
	if err != nil {
		log.Fatalf("Failed to create trip event consumer: %v", err)
	}
	defer consumer.Close()

	// Seed some sample drivers
	if err := seedDrivers(context.Background(), svc); err != nil {
		log.Printf("Failed to seed drivers: %v", err)
	}

	// Start event consumer in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Event consumer stopped: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Driver service is running. Press Ctrl+C to stop.")

	<-sigChan
	log.Println("Shutting down driver service...")
	cancel()
}

// seedDrivers creates some sample drivers for testing
func seedDrivers(ctx context.Context, svc *service.DriverService) error {
	drivers := []*domain.Driver{
		{
			UserID: "driver-1",
			Name:   "John Doe",
			Email:  "john@example.com",
			Phone:  "+1234567890",
			Location: domain.Coordinate{
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
			Vehicle: domain.Vehicle{
				Make:         "Toyota",
				Model:        "Camry",
				Year:         2020,
				LicensePlate: "ABC123",
				Color:        "Black",
			},
		},
		{
			UserID: "driver-2",
			Name:   "Jane Smith",
			Email:  "jane@example.com",
			Phone:  "+1234567891",
			Location: domain.Coordinate{
				Latitude:  40.7589,
				Longitude: -73.9851,
			},
			Vehicle: domain.Vehicle{
				Make:         "Honda",
				Model:        "Civic",
				Year:         2019,
				LicensePlate: "XYZ789",
				Color:        "White",
			},
		},
	}

	for _, driver := range drivers {
		if _, err := svc.RegisterDriver(ctx, driver); err != nil {
			return err
		}
		log.Printf("Seeded driver: %s (%s)", driver.Name, driver.UserID)
	}

	return nil
}
