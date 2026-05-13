package grpcserver

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ride-sharing/services/proto/proto"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/service"
)

type TripServiceServer struct {
	proto.UnimplementedTripServiceServer
	svc *service.TripService
}

func NewTripServiceServer(svc *service.TripService) *TripServiceServer {
	return &TripServiceServer{svc: svc}
}

func (s *TripServiceServer) CreateTrip(ctx context.Context, req *proto.CreateTripRequest) (*proto.TripResponse, error) {
	if req == nil || req.RideFare == nil {
		return nil, status.Error(codes.InvalidArgument, "request and ride_fare are required")
	}

	fare := &domain.RideFareModel{
		ID:                primitive.NewObjectID(),
		UserID:            req.UserId,
		PackageSlug:       req.RideFare.PackageSlug,
		TotalPriceInCents: req.RideFare.TotalPriceInCents,
	}

	trip, err := s.svc.CreateTrip(ctx, fare)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create trip: %v", err)
	}

	return &proto.TripResponse{
		Id:     trip.ID.Hex(),
		UserId: trip.UserID,
		Status: trip.Status,
		RideFare: &proto.RideFare{
			Id:                trip.RideFare.ID.Hex(),
			UserId:            trip.RideFare.UserID,
			PackageSlug:       trip.RideFare.PackageSlug,
			TotalPriceInCents: trip.RideFare.TotalPriceInCents,
		},
	}, nil
}
