package main

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	"ride-sharing/services/proto/proto"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/types"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if reqBody.UserID == "" {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}

	route := buildRoute(reqBody.Pickup, reqBody.Destination)

	rideFares := []previewRideFare{
		{
			ID:                reqBody.UserID + "-sedan",
			PackageSlug:       "sedan",
			BasePrice:         500,
			TotalPriceInCents: 1500,
			ExpiresAt:         time.Now().Add(20 * time.Minute).UTC().Format(time.RFC3339),
			Route:             route,
		},
		{
			ID:                reqBody.UserID + "-suv",
			PackageSlug:       "suv",
			BasePrice:         800,
			TotalPriceInCents: 2500,
			ExpiresAt:         time.Now().Add(20 * time.Minute).UTC().Format(time.RFC3339),
			Route:             route,
		},
		{
			ID:                reqBody.UserID + "-luxury",
			PackageSlug:       "luxury",
			BasePrice:         1200,
			TotalPriceInCents: 4000,
			ExpiresAt:         time.Now().Add(20 * time.Minute).UTC().Format(time.RFC3339),
			Route:             route,
		},
	}

	response := contracts.APIResponse{Data: previewTripResponse{
		Route:     route,
		RideFares: rideFares,
	}}
	writeJSON(w, http.StatusOK, response)
}

type previewTripResponse struct {
	Route     previewRoute      `json:"route"`
	RideFares []previewRideFare `json:"rideFares"`
}

type previewRoute struct {
	Geometry []previewGeometry `json:"geometry"`
	Duration float64           `json:"duration"`
	Distance float64           `json:"distance"`
}

type previewGeometry struct {
	Coordinates []types.Coordinate `json:"coordinates"`
}

type previewRideFare struct {
	ID                string       `json:"id"`
	PackageSlug       string       `json:"packageSlug"`
	BasePrice         float64      `json:"basePrice"`
	TotalPriceInCents float64      `json:"totalPriceInCents"`
	ExpiresAt         string       `json:"expiresAt"`
	Route             previewRoute `json:"route"`
}

func buildRoute(pickup, destination types.Coordinate) previewRoute {
	distance := calculateDistanceMeters(pickup, destination)
	duration := distance / 10.0

	return previewRoute{
		Geometry: []previewGeometry{{
			Coordinates: []types.Coordinate{pickup, destination},
		}},
		Distance: distance,
		Duration: duration,
	}
}

func calculateDistanceMeters(a, b types.Coordinate) float64 {
	const earthRadiusKm = 6371.0
	lat1 := a.Latitude * math.Pi / 180
	lon1 := a.Longitude * math.Pi / 180
	lat2 := b.Latitude * math.Pi / 180
	lon2 := b.Longitude * math.Pi / 180
	deltaLat := lat2 - lat1
	deltaLon := lon2 - lon1
	aVal := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(aVal), math.Sqrt(1-aVal))
	return earthRadiusKm * c * 1000
}

func handleCreateTrip(w http.ResponseWriter, r *http.Request) {
	var reqBody createTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if reqBody.UserID == "" || reqBody.PackageSlug == "" {
		http.Error(w, "user_id and package_slug are required", http.StatusBadRequest)
		return
	}

	grpcReq := &proto.CreateTripRequest{
		UserId: reqBody.UserID,
		RideFare: &proto.RideFare{
			UserId:            reqBody.UserID,
			PackageSlug:       reqBody.PackageSlug,
			TotalPriceInCents: reqBody.TotalPriceInCents,
		},
	}

	resp, err := tripClient.CreateTrip(r.Context(), grpcReq)
	if err != nil {
		http.Error(w, "Failed to create trip: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: resp}
	writeJSON(w, http.StatusOK, response)
}
