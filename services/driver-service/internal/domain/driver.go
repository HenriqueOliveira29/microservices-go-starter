package domain

import (
	"context"
	"time"

	"ride-sharing/shared/contracts"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Driver represents a driver in the system
type Driver struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    string             `json:"user_id" bson:"user_id"`
	Name      string             `json:"name" bson:"name"`
	Email     string             `json:"email" bson:"email"`
	Phone     string             `json:"phone" bson:"phone"`
	Status    string             `json:"status" bson:"status"` // available, busy, offline
	Location  GeoJSONPoint       `json:"location" bson:"location"`
	Vehicle   Vehicle            `json:"vehicle" bson:"vehicle"`
	Rating    float64            `json:"rating" bson:"rating"`
	TripCount int                `json:"trip_count" bson:"trip_count"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// Vehicle represents a driver's vehicle
type Vehicle struct {
	Make         string `json:"make" bson:"make"`
	Model        string `json:"model" bson:"model"`
	Year         int    `json:"year" bson:"year"`
	LicensePlate string `json:"license_plate" bson:"license_plate"`
	Color        string `json:"color" bson:"color"`
}

type GeoJSONPoint struct {
	Type string `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

// Coordinate represents a geographic coordinate
type Coordinate struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

// TripAssignment represents the assignment of a driver to a trip
type TripAssignment struct {
	TripID     string    `json:"trip_id"`
	DriverID   string    `json:"driver_id"`
	Status     string    `json:"status"` // assigned, accepted, declined, completed
	AssignedAt time.Time `json:"assigned_at"`
}

// DriverRepository defines the interface for driver data operations
type DriverRepository interface {
	CreateDriver(ctx context.Context, driver *Driver) (*Driver, error)
	GetDriver(ctx context.Context, id string) (*Driver, error)
	UpdateDriver(ctx context.Context, driver *Driver) error
	GetAvailableDrivers(ctx context.Context, location Coordinate, radiusKm float64) ([]*Driver, error)
	UpdateDriverStatus(ctx context.Context, driverID string, status string) error
}

// TripAssignmentRepository defines the interface for trip assignment data operations
type TripAssignmentRepository interface {
	CreateAssignment(ctx context.Context, assignment *TripAssignment) error
	GetAssignment(ctx context.Context, tripID string) (*TripAssignment, error)
	UpdateAssignmentStatus(ctx context.Context, tripID string, status string) error
}

// DriverService defines the business logic interface for driver operations
type DriverService interface {
	RegisterDriver(ctx context.Context, driver *Driver) (*Driver, error)
	AssignDriverToTrip(ctx context.Context, tripID string, riderID string) (*TripAssignment, error)
	UpdateDriverLocation(ctx context.Context, driverID string, location Coordinate) error
	GetDriver(ctx context.Context, driverID string) (*Driver, error)
}

// DriverEventPublisher defines the interface for publishing driver-related events
type DriverEventPublisher interface {
	PublishTripAssigned(ctx context.Context, assignment *contracts.DriverAssignedEvent) error
	PublishNoDriversFound(ctx context.Context, tripID string) error
}
