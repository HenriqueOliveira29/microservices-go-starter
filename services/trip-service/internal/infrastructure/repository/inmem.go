package repository

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
)

type InMemoryTripRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInMemoryTripRepository() *InMemoryTripRepository {
	return &InMemoryTripRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *InMemoryTripRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.trips[trip.ID.Hex()] = trip
	return trip, nil
}
