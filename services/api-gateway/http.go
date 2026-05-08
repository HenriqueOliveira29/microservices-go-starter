package main

import (
	"encoding/json"
	"net/http"
	"ride-sharing/shared/contracts"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {

	var reqBoday previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBoday); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	//validation
	if reqBoday.UserID == "" {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}

	//Call trip service to get fare estimate

	//Respond with fare estimate
	response := contracts.APIResponse{Data: "ok"}
	writeJSON(w, http.StatusOK, response)
}
