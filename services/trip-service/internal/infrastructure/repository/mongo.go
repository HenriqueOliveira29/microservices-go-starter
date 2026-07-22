package repository

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoTripRepository struct {
	collection *mongo.Collection
}

func NewMongoTripRepository(db *mongo.Database) *MongoTripRepository {
	return &MongoTripRepository{
		collection: db.Collection("trip"),
	}
}

func (r *MongoTripRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	_, err := r.collection.InsertOne(ctx, trip)

	return trip, err
}
