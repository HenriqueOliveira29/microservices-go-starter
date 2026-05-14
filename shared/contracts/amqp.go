package contracts

import "ride-sharing/shared/types"

// AmqpMessage is the message structure for AMQP.
type AmqpMessage struct {
	OwnerID string `json:"ownerId"`
	Data    []byte `json:"data"`
}

type DriverVehicle struct {
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	LicensePlate string `json:"license_plate"`
	Color        string `json:"color"`
}

type DriverDetails struct {
	ID       string           `json:"id"`
	UserID   string           `json:"user_id"`
	Name     string           `json:"name"`
	Email    string           `json:"email"`
	Phone    string           `json:"phone"`
	Location types.Coordinate `json:"location"`
	Vehicle  DriverVehicle    `json:"vehicle"`
	Rating   float64          `json:"rating"`
}

type DriverAssignedEvent struct {
	TripID  string        `json:"trip_id"`
	RiderID string        `json:"rider_id"`
	Driver  DriverDetails `json:"driver"`
	Status  string        `json:"status"`
}

// Routing keys - using consistent event/command patterns
const (
	// Trip events (trip.event.*)
	TripEventCreated             = "trip.event.created"
	TripEventDriverAssigned      = "trip.event.driver_assigned"
	TripEventNoDriversFound      = "trip.event.no_drivers_found"
	TripEventDriverNotInterested = "trip.event.driver_not_interested"

	// Driver commands (driver.cmd.*)
	DriverCmdTripRequest = "driver.cmd.trip_request"
	DriverCmdTripAccept  = "driver.cmd.trip_accept"
	DriverCmdTripDecline = "driver.cmd.trip_decline"
	DriverCmdLocation    = "driver.cmd.location"
	DriverCmdRegister    = "driver.cmd.register"

	// Payment events (payment.event.*)
	PaymentEventSessionCreated = "payment.event.session_created"
	PaymentEventSuccess        = "payment.event.success"
	PaymentEventFailed         = "payment.event.failed"
	PaymentEventCancelled      = "payment.event.cancelled"

	// Payment commands (payment.cmd.*)
	PaymentCmdCreateSession = "payment.cmd.create_session"
)
