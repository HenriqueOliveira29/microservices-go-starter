package main

import "ride-sharing/shared/types"

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

type createTripRequest struct {
	UserID            string  `json:"user_id"`
	PackageSlug       string  `json:"package_slug"`
	TotalPriceInCents float64 `json:"total_price_in_cents"`
}
