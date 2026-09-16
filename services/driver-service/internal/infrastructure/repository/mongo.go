package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ride-sharing/services/driver-service/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoDriverRepository struct {
	collection *mongo.Collection
}

func NewMongoDriverRepository(db *mongo.Database) *MongoDriverRepository {
	return &MongoDriverRepository{
		collection: db.Collection("driver"),
	}
}

func (r *MongoTripRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	_, err := r.collection.InsertOne(ctx, trip)

	return trip, err
}

func (r *MongoDriverRepository) CreateDriver(ctx context.Context, driver *domain.Driver) (*domain.Driver, error) {

	if driver.ID.IsZero() {
		driver.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, driver)

	return &driver, err
}

func (r *MongoDriverRepository) GetDriver(ctx context.Context, id string) (*domain.Driver, error) {

	filter := bson.M{"Id": id}

	var driver domain.Driver

	err := r.collection.FindOne(ctx, filter).Decode(&driver)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("driver with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to fetch driver: %w", err)
	}

	return &driver, nil
}

func (r *MongoDriverRepository) UpdateDriver(ctx context.Context, driver *domain.Driver) error {

	filter := bson.M{"Id", driver.ID}

	result, err := r.collection.ReplaceOne(ctx, filter, driver)

	if err != nil {
		return fmt.Errorf("failed to replace driver: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Error("driver with id %s not found", driver.ID)
	}

	return nil
}

func (r *MongoDriverRepository) GetAvailableDrivers(ctx context.Context, location domain.Coordinate, radiusKm float64) ([]*domain.Driver, error) {

	radiusMeters := radiusKm * 1000

	filter := bson.M{
		"status": "available",
		"location": bson.MP{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{loc.Longitude, loc.Latitude},
				},
				"$maxDistance": radiusMeters,
			},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)

	if err != nil {
		return nil, fmt.Errorf("failed to query nearby drivers: &w", err)
	}

	defer cursor.Close(ctx)

	var availableDrivers []*domain.Driver

	if err := cursor.All(ctx, &availableDrivers); err != nil {
		return nil, fmt.Errorf("failed to decode drivers %w", err)
	}

	return availableDrivers, nil
}

func (r *MongoDriverRepository) UpdateDriverStatus(ctx context.Context, driverID string, status string) error {

	filter := bson.M{"_id", driverID}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"UpdatedAt": time.Now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return fmt.Errorf("failed to update the driver: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("driver with id %s not found", driverID)
	}

	return nil
}

type InMemoryTripAssignmentRepository struct {
	mu          sync.RWMutex
	assignments map[string]*domain.TripAssignment
}

func NewInMemoryTripAssignmentRepository() *InMemoryTripAssignmentRepository {
	return &InMemoryTripAssignmentRepository{
		assignments: make(map[string]*domain.TripAssignment),
	}
}

func (r *InMemoryTripAssignmentRepository) CreateAssignment(ctx context.Context, assignment *domain.TripAssignment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.assignments[assignment.TripID] = assignment
	return nil
}

func (r *InMemoryTripAssignmentRepository) GetAssignment(ctx context.Context, tripID string) (*domain.TripAssignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	assignment, exists := r.assignments[tripID]
	if !exists {
		return nil, fmt.Errorf("assignment not found")
	}

	return assignment, nil
}

func (r *InMemoryTripAssignmentRepository) UpdateAssignmentStatus(ctx context.Context, tripID string, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	assignment, exists := r.assignments[tripID]
	if !exists {
		return fmt.Errorf("assignment not found")
	}

	assignment.Status = status
	return nil
}
