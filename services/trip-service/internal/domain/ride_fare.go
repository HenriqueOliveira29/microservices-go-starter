package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type RideFareModel struct {
	ID                primitive.ObjectID `json:"id"`
	UserID            string             `json:"user_id"`
	PackageSlug       string             `json:"package_slug"`
	TotalPriceInCents float64            `json:"total_price_in_cents"`
}
