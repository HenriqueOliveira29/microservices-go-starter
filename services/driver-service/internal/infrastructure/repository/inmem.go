package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ride-sharing/services/driver-service/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InMemoryDriverRepository struct {
	mu      sync.RWMutex
	drivers map[string]*domain.Driver
}

func NewInMemoryDriverRepository() *InMemoryDriverRepository {
	return &InMemoryDriverRepository{
		drivers: make(map[string]*domain.Driver),
	}
}

func (r *InMemoryDriverRepository) CreateDriver(ctx context.Context, driver *domain.Driver) (*domain.Driver, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if driver.ID.IsZero() {
		driver.ID = primitive.NewObjectID()
	}

	r.drivers[driver.ID.Hex()] = driver
	return driver, nil
}

func (r *InMemoryDriverRepository) GetDriver(ctx context.Context, id string) (*domain.Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	driver, exists := r.drivers[id]
	if !exists {
		return nil, fmt.Errorf("driver not found")
	}

	return driver, nil
}

func (r *InMemoryDriverRepository) UpdateDriver(ctx context.Context, driver *domain.Driver) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.drivers[driver.ID.Hex()] = driver
	return nil
}

func (r *InMemoryDriverRepository) GetAvailableDrivers(ctx context.Context, location domain.Coordinate, radiusKm float64) ([]*domain.Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var availableDrivers []*domain.Driver
	for _, driver := range r.drivers {
		if driver.Status == "available" {
			// For simplicity, include all available drivers
			// In a real implementation, you'd calculate distance and filter by radius
			availableDrivers = append(availableDrivers, driver)
		}
	}

	return availableDrivers, nil
}

func (r *InMemoryDriverRepository) UpdateDriverStatus(ctx context.Context, driverID string, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	driver, exists := r.drivers[driverID]
	if !exists {
		return fmt.Errorf("driver not found")
	}

	driver.Status = status
	driver.UpdatedAt = time.Now()
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
