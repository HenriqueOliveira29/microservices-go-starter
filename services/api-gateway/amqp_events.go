package main

import (
	"context"
	"encoding/json"
	"log"

	"ride-sharing/shared/amqp"
	"ride-sharing/shared/contracts"
)

type RiderEventConsumer struct {
	consumer *amqp.Consumer
}

func NewRiderEventConsumer(amqpURL string) (*RiderEventConsumer, error) {
	consumer, err := amqp.NewConsumer(amqpURL, "api-gateway-rider-events", []string{contracts.TripEventDriverAssigned, contracts.TripEventNoDriversFound})
	if err != nil {
		return nil, err
	}

	return &RiderEventConsumer{consumer: consumer}, nil
}

func (c *RiderEventConsumer) Start(ctx context.Context) error {
	log.Println("Starting rider event consumer...")
	return c.consumer.Consume(ctx, c.handleEvent)
}

func (c *RiderEventConsumer) handleEvent(msg amqp.Delivery) error {
	var amqpMsg contracts.AmqpMessage
	if err := json.Unmarshal(msg.Body, &amqpMsg); err != nil {
		log.Printf("Failed to unmarshal AMQP message: %v", err)
		return err
	}

	switch msg.RoutingKey {
	case contracts.TripEventDriverAssigned:
		var event contracts.DriverAssignedEvent
		if err := json.Unmarshal(amqpMsg.Data, &event); err != nil {
			log.Printf("Failed to unmarshal driver assigned event: %v", err)
			return err
		}
		return c.forwardDriverAssignedEvent(event)
	case contracts.TripEventNoDriversFound:
		// No riders-specific route yet for this event. It can be extended
		// later if the payload contains a rider ID.
		log.Printf("Received no drivers found event, payload size=%d", len(amqpMsg.Data))
		return nil
	default:
		log.Printf("Unhandled rider event routing key: %s", msg.RoutingKey)
		return nil
	}
}

func (c *RiderEventConsumer) forwardDriverAssignedEvent(event contracts.DriverAssignedEvent) error {
	message := contracts.WSMessage{
		Type: contracts.TripEventDriverAssigned,
		Data: event,
	}

	if err := sendToRider(event.RiderID, message); err != nil {
		log.Printf("No rider websocket for driver assigned event: %v", err)
	}

	return nil
}

func (c *RiderEventConsumer) Close() error {
	return c.consumer.Close()
}
