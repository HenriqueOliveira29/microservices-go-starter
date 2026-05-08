package service

import (
	"context"

	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripService struct {
	repo repository.InMemoryTripRepository
}

func NewTripService(repo repository.InMemoryTripRepository) *TripService {
	return &TripService{repo: repo}
}

func (s *TripService) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	trip := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "created",
		RideFare: fare,
	}
	return s.repo.CreateTrip(ctx, trip)
}
