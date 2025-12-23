package repository

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"provider_management/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AcceptedServiceRepo struct {
	col               *mongo.Collection
	serviceRequestCol *mongo.Collection
}

func NewAcceptedServiceRepo(db *mongo.Database) *AcceptedServiceRepo {
	return &AcceptedServiceRepo{
		col:               db.Collection("acceptedservices"),
		serviceRequestCol: db.Collection("servicerequests"),
	}
}

func (r *AcceptedServiceRepo) FindByID(ctx context.Context, id string) (*domain.AcceptedService, error) {
	var res domain.AcceptedService

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *AcceptedServiceRepo) CountByUserID(ctx context.Context, userID string, filter bson.M) (int64, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	filter["user"] = userObjID
	return r.col.CountDocuments(ctx, filter)
}

func (r *AcceptedServiceRepo) SumExpensesByUserID(ctx context.Context, userID string) (int64, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"user":          userObjID,
			"status":        "completed",
			"paymentStatus": "paid",
		}},
		{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$finalPrice"},
		}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Total int64 `bson:"total"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.Total, nil
	}

	return 0, nil
}

func (r *AcceptedServiceRepo) CountCompletedInDateRange(ctx context.Context, userID primitive.ObjectID, start, end time.Time) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{
		"user":          userID,
		"status":        "completed",
		"paymentStatus": "paid",
		"createdAt": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
}

func (r *AcceptedServiceRepo) Aggregate(ctx context.Context, pipeline []bson.M) (*mongo.Cursor, error) {
	return r.col.Aggregate(ctx, pipeline)
}

func (r *AcceptedServiceRepo) CountByProviderID(ctx context.Context, providerID string, filter bson.M) (int64, error) {
	if objID, err := primitive.ObjectIDFromHex(providerID); err == nil {
		filter["provider"] = objID
	} else {
		filter["provider"] = providerID
	}

	return r.col.CountDocuments(ctx, filter)
}

func (r *AcceptedServiceRepo) FindByProviderID(ctx context.Context, providerID string, limit int64) ([]domain.Service, error) {
	var providerObjectID interface{}
	if objID, err := primitive.ObjectIDFromHex(providerID); err == nil {
		providerObjectID = objID
	} else {
		providerObjectID = providerID
	}

	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "provider", Value: providerObjectID},
			}},
		},
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "servicerequests"},
				{Key: "localField", Value: "serviceRequest"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "serviceRequestDetails"},
			}},
		},
		bson.D{
			{Key: "$sort", Value: bson.D{
				{Key: "createdAt", Value: -1},
			}},
		},
		bson.D{
			{Key: "$limit", Value: limit},
		},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate services: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		domain.AcceptedService `bson:",inline"`
		ServiceRequestDetails  []domain.ServiceRequest `bson:"serviceRequestDetails"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode services: %w", err)
	}

	var services []domain.Service
	for _, result := range results {
		service := domain.Service{
			ID:               result.ID.Hex(),
			ServiceRequestID: result.ServiceRequestID.Hex(),
			Status:           result.Status,
			ServiceType:      result.ServiceType,
		}

		if !result.CreatedAt.IsZero() {
			service.Date = result.CreatedAt.Format("2006-01-02 15:04")
		}

		service.Amount = result.FinalPrice
		if service.Amount == 0 {
			service.Amount = result.BasePrice
		}

		if len(result.ServiceRequestDetails) > 0 {
			sr := result.ServiceRequestDetails[0]
			service.VehicleNumber = sr.VehicleNumber

			if len(sr.Problems) > 0 {
				service.Problem = strings.Join(sr.Problems, ", ")
			} else if sr.Description != "" {
				service.Problem = sr.Description
			} else {
				service.Problem = "N/A"
			}

			if service.Problem == "" || service.Problem == "N/A" {
				service.Problem = sr.ServiceType
			}
		} else {
			service.VehicleNumber = "N/A"
			service.Problem = "N/A"
		}

		services = append(services, service)
	}

	return services, nil
}

func (r *AcceptedServiceRepo) GetServiceStats(ctx context.Context, providerID string) (total, completed int64, err error) {
	var providerObjectID interface{}
	if objID, err := primitive.ObjectIDFromHex(providerID); err == nil {
		providerObjectID = objID
	} else {
		providerObjectID = providerID
	}

	total, err = r.col.CountDocuments(ctx, bson.M{"provider": providerObjectID})
	if err != nil {
		return 0, 0, err
	}

	completed, err = r.col.CountDocuments(ctx, bson.M{
		"provider": providerObjectID,
		"status":   "completed",
	})

	return total, completed, err
}

func (r *AcceptedServiceRepo) FindWithServiceRequest(ctx context.Context, id string) (*domain.AcceptedService, *domain.ServiceRequest, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil, err
	}

	var acceptedService domain.AcceptedService
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&acceptedService)
	if err != nil {
		return nil, nil, err
	}

	var serviceRequest domain.ServiceRequest
	srObjID, err := primitive.ObjectIDFromHex(acceptedService.ServiceRequestID.Hex())
	if err != nil {
		return &acceptedService, nil, nil
	}

	err = r.serviceRequestCol.FindOne(ctx, bson.M{"_id": srObjID}).Decode(&serviceRequest)
	if err != nil {
		return &acceptedService, nil, nil
	}

	return &acceptedService, &serviceRequest, nil
}

func (r *AcceptedServiceRepo) FindCompletedPaidBetween(
	ctx context.Context,
	from, to time.Time,
) ([]domain.AcceptedService, error) {

	filter := bson.M{
		"status":        "completed",
		"paymentStatus": "paid",
		"completedAt": bson.M{
			"$gte": from,
			"$lt":  to,
		},
	}

	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var services []domain.AcceptedService
	log.Println("servicesss",services)
	if err := cursor.All(ctx, &services); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *AcceptedServiceRepo) MarkAsSettled(
	ctx context.Context,
	serviceIDs []primitive.ObjectID,
	settlementID primitive.ObjectID,
) error {

	now := time.Now()

	filter := bson.M{
		"_id": bson.M{"$in": serviceIDs},
		"$or": []bson.M{
			{"isSettled": false},
			{"isSettled": bson.M{"$exists": false}},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"isSettled":    true,
			"settlementId": settlementID,
			"settledAt":    now,
			"updatedAt":    now,
		},
	}

	res, err := r.col.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return fmt.Errorf("no unsettled services found to settle")
	}

	if res.MatchedCount != int64(len(serviceIDs)) {
		return fmt.Errorf("some services are already settled or not found")
	}

	return nil
}

func (r *AcceptedServiceRepo) FindCompletedPaidByProvider(ctx context.Context, providerID primitive.ObjectID) ([]*domain.AcceptedService, error) {
	filter := bson.M{
		"provider":      providerID,
		"status":        "completed",
		"paymentStatus": "paid",
	}
	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var services []*domain.AcceptedService
	if err := cursor.All(ctx, &services); err != nil {
		return nil, err
	}
	return services, nil
}

func (r *AcceptedServiceRepo) FindUnsettledByProvider(ctx context.Context, providerID primitive.ObjectID) ([]*domain.AcceptedService, error) {
	filter := bson.M{
		"provider":      providerID,
		"status":        "completed",
		"paymentStatus": "paid",
		"isSettled":     false,
	}
	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var services []*domain.AcceptedService
	if err := cursor.All(ctx, &services); err != nil {
		return nil, err
	}
	return services, nil
}

func (r *AcceptedServiceRepo) CountUnsettledByIDs(
	ctx context.Context,
	serviceIDs []primitive.ObjectID,
) (int64, error) {
	log.Println("service idsss", serviceIDs)
	filter := bson.M{
		"_id": bson.M{"$in": serviceIDs},
		"$or": []bson.M{
			{"isSettled": false},
			{"isSettled": bson.M{"$exists": false}},
		},
	}

	return r.col.CountDocuments(ctx, filter)
}

func (r *AcceptedServiceRepo) CountSettledByIDs(
	ctx context.Context,
	serviceIDs []primitive.ObjectID,
) (int64, error) {
	filter := bson.M{
		"_id":       bson.M{"$in": serviceIDs},
		"isSettled": true,
	}

	return r.col.CountDocuments(ctx, filter)
}

