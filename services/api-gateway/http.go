package main

import (
	"encoding/json"
	"net/http"

	"ride-sharing/services/proto/proto"
	"ride-sharing/shared/contracts"
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

	response := contracts.APIResponse{Data: "ok"}
	writeJSON(w, http.StatusOK, response)
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
