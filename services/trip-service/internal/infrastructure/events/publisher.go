package events

import (
	"context"

	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/amqp"
	"ride-sharing/shared/contracts"
)

// TripEventPublisher handles publishing trip-related events
type TripEventPublisher struct {
	publisher *amqp.Publisher
}

func NewTripEventPublisher(publisher *amqp.Publisher) *TripEventPublisher {
	return &TripEventPublisher{publisher: publisher}
}

func (p *TripEventPublisher) PublishTripCreated(ctx context.Context, trip *domain.TripModel) error {
	return p.publisher.Publish(ctx, contracts.TripEventCreated, trip)
}
