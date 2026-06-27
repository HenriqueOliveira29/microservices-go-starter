package service

import (
	"context"
	"log"

	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripService struct {
	repo      repository.InMemoryTripRepository
	publisher domain.TripEventPublisher
}

func NewTripService(repo repository.InMemoryTripRepository, publisher domain.TripEventPublisher) *TripService {
	return &TripService{repo: repo, publisher: publisher}
}

func (s *TripService) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	trip := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "created",
		RideFare: fare,
	}

	createdTrip, err := s.repo.CreateTrip(ctx, trip)
	if err != nil {
		return nil, err
	}

	// Publish trip created event
	if err := s.publisher.PublishTripCreated(ctx, createdTrip); err != nil {
		// Log the error but don't fail the trip creation
		// In a production system, you might want to implement retry logic or dead letter queues
		// For now, we'll just log it
		log.Printf("Failed to publish trip created event: %v", err)
	}

	return createdTrip, nil
}
