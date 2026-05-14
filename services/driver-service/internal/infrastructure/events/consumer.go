package events

import (
	"context"
	"encoding/json"
	"log"

	"ride-sharing/services/driver-service/internal/service"
	"ride-sharing/shared/amqp"
	"ride-sharing/shared/contracts"
)

// TripCreatedEvent represents the data from a trip creation event
type TripCreatedEvent struct {
	TripID   string `json:"id"`
	UserID   string `json:"user_id"`
	Status   string `json:"status"`
	RideFare struct {
		ID                string  `json:"id"`
		UserID            string  `json:"user_id"`
		PackageSlug       string  `json:"package_slug"`
		TotalPriceInCents float64 `json:"total_price_in_cents"`
	} `json:"ride_fare"`
}

// TripEventConsumer handles consuming trip-related events
type TripEventConsumer struct {
	consumer *amqp.Consumer
	service  *service.DriverService
}

// NewTripEventConsumer creates a new trip event consumer
func NewTripEventConsumer(amqpURL string, service *service.DriverService) (*TripEventConsumer, error) {
	consumer, err := amqp.NewConsumer(amqpURL, "driver-service-trip-events", []string{contracts.TripEventCreated})
	if err != nil {
		return nil, err
	}

	return &TripEventConsumer{
		consumer: consumer,
		service:  service,
	}, nil
}

// Start starts consuming trip events
func (c *TripEventConsumer) Start(ctx context.Context) error {
	log.Println("Starting trip event consumer...")

	return c.consumer.Consume(ctx, c.handleTripEvent)
}

// handleTripEvent processes incoming trip events
func (c *TripEventConsumer) handleTripEvent(msg amqp.Delivery) error {
	var amqpMsg contracts.AmqpMessage
	if err := json.Unmarshal(msg.Body, &amqpMsg); err != nil {
		log.Printf("Failed to unmarshal AMQP message: %v", err)
		log.Printf("Raw message body: %s", string(msg.Body))
		return err
	}

	var tripEvent TripCreatedEvent
	if err := json.Unmarshal(amqpMsg.Data, &tripEvent); err != nil {
		log.Printf("Failed to unmarshal trip event: %v", err)
		return err
	}

	log.Printf("Received trip created event for trip: %s", tripEvent.TripID)

	// Assign a driver to the trip
	assignment, err := c.service.AssignDriverToTrip(context.Background(), tripEvent.TripID, tripEvent.UserID)
	if err != nil {
		log.Printf("Failed to assign driver to trip %s: %v", tripEvent.TripID, err)
		return err
	}

	log.Printf("Successfully assigned driver %s to trip %s", assignment.DriverID, assignment.TripID)
	return nil
}

// Close closes the consumer
func (c *TripEventConsumer) Close() error {
	return c.consumer.Close()
}
