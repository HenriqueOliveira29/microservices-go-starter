package main

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"time"
)

func main() {
	inmemory := repository.NewInMemoryTripRepository()

	svc := service.NewTripService(*inmemory)

	// Example usage
	fare := &domain.RideFareModel{
		UserID: "42",
	}

	t, err := svc.CreateTrip(context.Background(), fare)
	if err != nil {
		log.Println(err)
	}

	log.Println(t)

	for {
		time.Sleep(time.Second)
	}
}
