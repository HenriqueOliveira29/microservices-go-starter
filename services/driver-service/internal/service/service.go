package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"ride-sharing/services/driver-service/internal/domain"
	"ride-sharing/shared/amqp"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/types"
)

type DriverService struct {
	driverRepo     domain.DriverRepository
	assignmentRepo domain.TripAssignmentRepository
	publisher      domain.DriverEventPublisher
}

func NewDriverService(
	driverRepo domain.DriverRepository,
	assignmentRepo domain.TripAssignmentRepository,
	publisher domain.DriverEventPublisher,
) *DriverService {
	return &DriverService{
		driverRepo:     driverRepo,
		assignmentRepo: assignmentRepo,
		publisher:      publisher,
	}
}

func (s *DriverService) RegisterDriver(ctx context.Context, driver *domain.Driver) (*domain.Driver, error) {
	// Set default values
	now := time.Now()
	driver.Status = "available"
	driver.CreatedAt = now
	driver.UpdatedAt = now
	driver.Rating = 5.0 // Default rating for new drivers

	return s.driverRepo.CreateDriver(ctx, driver)
}

var ErrNoAvailableDrivers = errors.New("no available drivers found")

func (s *DriverService) AssignDriverToTrip(ctx context.Context, tripID string, riderID string) (*domain.TripAssignment, error) {
	// For now, we'll assign to a hardcoded location (could be extracted from trip data)
	// In a real implementation, you'd get the trip location from the trip service
	tripLocation := domain.Coordinate{Latitude: 40.7128, Longitude: -74.0060} // NYC coordinates

	// Find available drivers within 10km radius
	availableDrivers, err := s.driverRepo.GetAvailableDrivers(ctx, tripLocation, 10.0)
	if err != nil {
		return nil, fmt.Errorf("failed to find available drivers: %w", err)
	}

	if len(availableDrivers) == 0 {
		// Publish event that no drivers were found
		if err := s.publisher.PublishNoDriversFound(ctx, tripID); err != nil {
			log.Printf("Failed to publish no drivers found event: %v", err)
		}
		return nil, ErrNoAvailableDrivers
	}

	// For simplicity, assign to the first available driver
	// In a real implementation, you might use more sophisticated logic
	selectedDriver := availableDrivers[0]

	// Create assignment
	assignment := &domain.TripAssignment{
		TripID:     tripID,
		DriverID:   selectedDriver.ID.Hex(),
		Status:     "assigned",
		AssignedAt: time.Now(),
	}

	// Save assignment
	if err := s.assignmentRepo.CreateAssignment(ctx, assignment); err != nil {
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	// Update driver status to busy
	if err := s.driverRepo.UpdateDriverStatus(ctx, selectedDriver.ID.Hex(), "busy"); err != nil {
		log.Printf("Failed to update driver status: %v", err)
		// Don't fail the assignment for this
	}

	// Publish assignment event
	assignedEvent := &contracts.DriverAssignedEvent{
		TripID:  tripID,
		RiderID: riderID,
		Status:  "assigned",
		Driver: contracts.DriverDetails{
			ID:     selectedDriver.ID.Hex(),
			UserID: selectedDriver.UserID,
			Name:   selectedDriver.Name,
			Email:  selectedDriver.Email,
			Phone:  selectedDriver.Phone,
			Location: types.Coordinate{
				Latitude:  selectedDriver.Location.Latitude,
				Longitude: selectedDriver.Location.Longitude,
			},
			Vehicle: contracts.DriverVehicle{
				Make:         selectedDriver.Vehicle.Make,
				Model:        selectedDriver.Vehicle.Model,
				Year:         selectedDriver.Vehicle.Year,
				LicensePlate: selectedDriver.Vehicle.LicensePlate,
				Color:        selectedDriver.Vehicle.Color,
			},
			Rating: selectedDriver.Rating,
		},
	}

	if err := s.publisher.PublishTripAssigned(ctx, assignedEvent); err != nil {
		log.Printf("Failed to publish trip assigned event: %v", err)
		// Don't fail the assignment for this
	}

	return assignment, nil
}

func (s *DriverService) UpdateDriverLocation(ctx context.Context, driverID string, location domain.Coordinate) error {
	// Get current driver
	driver, err := s.driverRepo.GetDriver(ctx, driverID)
	if err != nil {
		return fmt.Errorf("failed to get driver: %w", err)
	}

	// Update location and timestamp
	driver.Location = location
	driver.UpdatedAt = time.Now()

	return s.driverRepo.UpdateDriver(ctx, driver)
}

func (s *DriverService) GetDriver(ctx context.Context, driverID string) (*domain.Driver, error) {
	return s.driverRepo.GetDriver(ctx, driverID)
}

// calculateDistance calculates the distance between two coordinates using Haversine formula
func calculateDistance(coord1, coord2 domain.Coordinate) float64 {
	const earthRadiusKm = 6371.0

	lat1Rad := coord1.Latitude * math.Pi / 180
	lon1Rad := coord1.Longitude * math.Pi / 180
	lat2Rad := coord2.Latitude * math.Pi / 180
	lon2Rad := coord2.Longitude * math.Pi / 180

	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// DriverEventPublisher implementation
type AMQPDriverEventPublisher struct {
	publisher *amqp.Publisher
}

func NewAMQPDriverEventPublisher(publisher *amqp.Publisher) *AMQPDriverEventPublisher {
	return &AMQPDriverEventPublisher{publisher: publisher}
}

func (p *AMQPDriverEventPublisher) PublishTripAssigned(ctx context.Context, assignment *contracts.DriverAssignedEvent) error {
	return p.publisher.Publish(ctx, contracts.TripEventDriverAssigned, assignment)
}

func (p *AMQPDriverEventPublisher) PublishNoDriversFound(ctx context.Context, tripID string) error {
	data := map[string]string{"trip_id": tripID}
	return p.publisher.Publish(ctx, contracts.TripEventNoDriversFound, data)
}
