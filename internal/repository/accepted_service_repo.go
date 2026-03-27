package repository

import (
	"context"
	"fmt"
	"log"
	"math"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (r *AcceptedServiceRepo) FindByObjectIDs(ctx context.Context, id primitive.ObjectID) (*domain.AcceptedService, error) {
	var res domain.AcceptedService

	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&res)
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
			ServiceRequestID: result.ServiceNumber,
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

func (r *AcceptedServiceRepo) FindCompletedPaidBetween(
	ctx context.Context,
	from, to time.Time,
) ([]domain.AcceptedService, error) {

	filter := bson.M{
		"status":        "completed",
		"paymentStatus": "paid",
		"timestamps.completedAt": bson.M{
			"$gte": from,
			"$lte": to,
		},
		"$or": []bson.M{
			{"payoutCreated": bson.M{"$exists": false}},
			{"payoutCreated": false},
		},
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"complaintUser": bson.M{"$exists": false}},
					{"complaintUser": ""},
				},
			},
			{
				"$or": []bson.M{
					{"complaintProvider": bson.M{"$exists": false}},
					{"complaintProvider": ""},
				},
			},
		},
	}

	log.Println("Fetching services between:", from, "and", to)
	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var services []domain.AcceptedService
	if err := cursor.All(ctx, &services); err != nil {
		return nil, err
	}

	log.Println("Found services count:", len(services))
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
			{
				"$or": []bson.M{
					{"settlementStatus": bson.M{"$exists": false}},
					{"settlementStatus": ""},
				},
			},
			{
				"settlementStatus": bson.M{
					"$in": []string{
						string(domain.SettleStatusPending),
						string(domain.SettleStatusSettled),
					},
				},
				"hasComplaintAdjustment": true,
			},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"settlementStatus": domain.SettleStatusPending,
			"settlementId":     settlementID,
			"updatedAt":        now,
		},
		"$unset": bson.M{
			"hasComplaintAdjustment": "",
			"pendingDeductionAmount": "",
		},
	}

	res, err := r.col.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return fmt.Errorf("no eligible services found for settlement")
	}

	if res.MatchedCount != int64(len(serviceIDs)) {
		return fmt.Errorf("some services are not eligible for settlement")
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
		"$or": []bson.M{
			{"settlementStatus": bson.M{"$exists": false}},
			{"settlementStatus": ""},
		},
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

func (r *AcceptedServiceRepo) CountUnsettledByIDs(ctx context.Context, serviceIDs []primitive.ObjectID) (int64, error) {

	filter := bson.M{
		"_id": bson.M{"$in": serviceIDs},
		"$or": []bson.M{
			{
				"$or": []bson.M{
					{"settlementStatus": bson.M{"$exists": false}},
					{"settlementStatus": ""},
				},
			},

			{
				"$and": []bson.M{
					{"settlementStatus": string(domain.SettleStatusSettled)},
					{"payoutStatus": domain.PayoutStatusComplaintAfterSettlement},
					{
						"$or": []bson.M{
							{"isSettledAfterComplaint": false},
							{"isSettledAfterComplaint": bson.M{"$exists": false}},
						},
					},
				},
			},
		},
	}

	return r.col.CountDocuments(ctx, filter)
}

func (r *AcceptedServiceRepo) CountSettledByIDs(ctx context.Context, serviceIDs []primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"_id": bson.M{"$in": serviceIDs},
		"settlementStatus": bson.M{
			"$in": []string{
				string(domain.SettleStatusPending),
				string(domain.SettleStatusSettled),
			},
		},
	}

	return r.col.CountDocuments(ctx, filter)
}

func (r *AcceptedServiceRepo) MarkAsSettledAfterComplaint(ctx context.Context, serviceID primitive.ObjectID, settledAt *time.Time) error {
	filter := bson.M{"_id": serviceID}
	update := bson.M{
		"$set": bson.M{
			"settlementStatus":        domain.SettleStatusPending,
			"isSettledAfterComplaint": true,
			"settledAfterComplaintAt": settledAt,
			"updatedAt":               time.Now(),
		},
	}

	_, err := r.col.UpdateOne(ctx, filter, update)
	return err
}

func (r *AcceptedServiceRepo) MarkPayoutCreated(ctx context.Context, serviceIDs []primitive.ObjectID) error {
	filter := bson.M{
		"_id": bson.M{"$in": serviceIDs},
	}

	update := bson.M{
		"$set": bson.M{
			"payoutCreated":   true,
			"payoutCreatedAt": time.Now(),
			"payoutStatus":    domain.PayoutStatusRegular,
		},
	}

	_, err := r.col.UpdateMany(ctx, filter, update)
	return err
}

func (r *AcceptedServiceRepo) FindByInternalID(ctx context.Context, internalID int64) (*domain.AcceptedService, error) {

	var svc domain.AcceptedService
	err := r.col.FindOne(ctx, bson.M{
		"internalId": internalID,
	}).Decode(&svc)

	if err != nil {
		return nil, err
	}

	return &svc, nil
}

func (r *AcceptedServiceRepo) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]domain.AcceptedService, error) {
	var services []domain.AcceptedService

	filter := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &services); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *AcceptedServiceRepo) GetBookingStats(ctx context.Context) (dto.BookingsStats, error) {
	pipeline := []bson.M{
		{
			"$facet": bson.M{
				"total": []bson.M{
					{"$count": "count"},
				},
				"completed": []bson.M{
					{"$match": bson.M{"status": "completed"}},
					{"$count": "count"},
				},
				"ongoing": []bson.M{
					{"$match": bson.M{"status": bson.M{"$in": []string{"confirmed", "started", "reached_location", "otp_verified", "in_progress"}}}},
					{"$count": "count"},
				},
			},
		},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return dto.BookingsStats{}, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		Total     []struct{ Count int64 } `bson:"total"`
		Completed []struct{ Count int64 } `bson:"completed"`
		Ongoing   []struct{ Count int64 } `bson:"ongoing"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return dto.BookingsStats{}, err
	}

	stats := dto.BookingsStats{}
	if len(result) > 0 {
		if len(result[0].Total) > 0 {
			stats.Total = result[0].Total[0].Count
		}
		if len(result[0].Completed) > 0 {
			stats.Completed = result[0].Completed[0].Count
		}
		if len(result[0].Ongoing) > 0 {
			stats.Ongoing = result[0].Ongoing[0].Count
		}
	}

	return stats, nil
}

func (r *AcceptedServiceRepo) GetTopServices(ctx context.Context) ([]dto.TopService, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"status":      "completed",
				"serviceType": bson.M{"$ne": ""},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$serviceType",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
		{
			"$limit": 5,
		},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ServiceType string `bson:"_id"`
		Count       int64  `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	var totalCount int64
	for _, r := range results {
		totalCount += r.Count
	}

	services := make([]dto.TopService, 0, len(results))
	for _, r := range results {
		percentage := 0.0
		if totalCount > 0 {
			percentage = math.Round((float64(r.Count) / float64(totalCount)) * 100)
		}

		services = append(services, dto.TopService{
			Name:       r.ServiceType,
			Count:      r.Count,
			Percentage: percentage,
		})
	}

	return services, nil
}

func (r *AcceptedServiceRepo) UpdateComplaintFlags(ctx context.Context, serviceID string, flags map[string]any) error {
	objID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return fmt.Errorf("invalid service ID: %w", err)
	}

	update := bson.M{"$set": flags}

	_, err = r.col.UpdateByID(ctx, objID, update)
	if err != nil {
		return fmt.Errorf("failed to update accepted service: %w", err)
	}

	return nil
}

func (r *AcceptedServiceRepo) UpdatePayoutStatus(ctx context.Context, serviceID string, fields map[string]any) error {
	objID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return err
	}

	update := bson.M{"$set": fields}
	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *AcceptedServiceRepo) UpdatePayoutCancellation(ctx context.Context, serviceID string, cancelled bool) error {
	objID, _ := primitive.ObjectIDFromHex(serviceID)
	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{
			"isPayoutCancelled": cancelled,
			"PayoutCancelledAt": time.Now(),
		}},
	)
	return err
}

func (r *AcceptedServiceRepo) FindAll(ctx context.Context, filter bson.M) ([]domain.AcceptedService, error) {
	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var services []domain.AcceptedService
	if err := cursor.All(ctx, &services); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *AcceptedServiceRepo) MarkServicesAsSettled(ctx context.Context, serviceIDs []primitive.ObjectID, settlementID primitive.ObjectID, settledAt *time.Time,
) error {
	filter := bson.M{"_id": bson.M{"$in": serviceIDs}}
	update := bson.M{
		"$set": bson.M{
			"settlementStatus": domain.SettleStatusSettled,
			"settlementId":     settlementID,
			"settledAt":        settledAt,
			"updatedAt":        time.Now(),
		},
	}

	_, err := r.col.UpdateMany(ctx, filter, update)
	return err
}

func (r *AcceptedServiceRepo) FindByServiceID(
	ctx context.Context,
	serviceID string,
) (*domain.AcceptedService, error) {

	serviceObjID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid serviceID: %w", err)
	}

	var res domain.AcceptedService

	err = r.col.FindOne(ctx, bson.M{
		"_id": serviceObjID,
	}).Decode(&res)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *AcceptedServiceRepo) FindByServiceRequestID(
	ctx context.Context,
	serviceRequestID string,
) (*domain.AcceptedService, error) {

	objID, err := primitive.ObjectIDFromHex(serviceRequestID)
	if err != nil {
		return nil, fmt.Errorf("invalid serviceRequestID: %w", err)
	}

	var res domain.AcceptedService
	err = r.col.FindOne(ctx, bson.M{
		"_id": objID,
	}).Decode(&res)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *AcceptedServiceRepo) PromoCodeUsage(
	ctx context.Context,
	page, limit int,
) ([]domain.AcceptedService, int64, error) {

	filter := bson.M{
		"totalDiscount": bson.M{"$gt": 0},
	}

	skip := (page - 1) * limit

	cur, err := r.col.Find(ctx, filter, options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"createdAt": -1}),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var services []domain.AcceptedService
	if err := cur.All(ctx, &services); err != nil {
		return nil, 0, err
	}

	total, _ := r.col.CountDocuments(ctx, filter)

	return services, total, nil
}

func (r *AcceptedServiceRepo) GetUsageList(
    ctx context.Context,
    promoID, discountID, userID string,
    page, limit int,
) ([]domain.AcceptedService, int64, error) {
    filter := bson.M{
        "appliedPromo": bson.M{"$exists": true, "$ne": nil},
    }

    if promoID != "" {
        filter["appliedPromo.promoId"] = promoID
    }

    if userID != "" {
        userObjID, err := primitive.ObjectIDFromHex(userID)
        if err != nil {
            return nil, 0, fmt.Errorf("invalid user id")
        }
        filter["user"] = userObjID
    }

    if discountID != "" {
        filter["appliedDiscount.discountId"] = discountID
    }

    total, err := r.col.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, err
    }

    skip := int64((page - 1) * limit)
    opts := options.Find().
        SetSkip(skip).
        SetLimit(int64(limit)).
        SetSort(bson.D{{Key: "createdAt", Value: -1}})

    cursor, err := r.col.Find(ctx, filter, opts)
    if err != nil {
        return nil, 0, err
    }
    defer cursor.Close(ctx)

    var results []domain.AcceptedService
    if err := cursor.All(ctx, &results); err != nil {
        return nil, 0, err
    }

    return results, total, nil
}

func (r *AcceptedServiceRepo) GetPromoUserCount(ctx context.Context, promoID string) (int64, error) {
    pipeline := mongo.Pipeline{
        {{Key: "$match", Value: bson.M{
            "appliedPromo.promoId": promoID,
        }}},
        {{Key: "$group", Value: bson.M{
            "_id": "$user",
        }}},
        {{Key: "$count", Value: "uniqueUsers"}},
    }

    cursor, err := r.col.Aggregate(ctx, pipeline)
    if err != nil {
        return 0, err
    }
    defer cursor.Close(ctx)

    var result []struct {
        UniqueUsers int64 `bson:"uniqueUsers"`
    }
    if err := cursor.All(ctx, &result); err != nil {
        return 0, err
    }
    if len(result) == 0 {
        return 0, nil
    }
    return result[0].UniqueUsers, nil
}

func (r *AcceptedServiceRepo) GetServicesByPromoAndUser(
	ctx context.Context,
	promoID, userID, createdAt string,
	skip, limit int64,
) ([]domain.AcceptedService, int64, error) {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"appliedPromo.promoId": promoID,
		"user":                 userOID,
	}
	
	if createdAt != "" {
		t, err := time.Parse("2006-01-02", createdAt)
		if err == nil {
	
			startOfDay := t
			endOfDay := t.Add(24 * time.Hour)
	
			filter["createdAt"] = bson.M{
				"$gte": primitive.NewDateTimeFromTime(startOfDay),
				"$lt":  primitive.NewDateTimeFromTime(endOfDay),
			}
		}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var services []domain.AcceptedService
	if err := cursor.All(ctx, &services); err != nil {
		return nil, 0, err
	}
	return services, total, nil
}


func (r *AcceptedServiceRepo) GetPromoUsersAggregated(
	ctx context.Context,
	promoID, createdAt string,
	skip, limit int64,
) ([]dto.PromoUserUsageRow, int64, error) {
	matchStage := bson.M{"appliedPromo.promoId": promoID}

	if createdAt != "" {
		t, err := time.Parse("2006-01-02", createdAt)
		if err == nil {
			startOfDay := t
			endOfDay := t.Add(24 * time.Hour)
	
			matchStage["createdAt"] = bson.M{
				"$gte": primitive.NewDateTimeFromTime(startOfDay),
				"$lt":  primitive.NewDateTimeFromTime(endOfDay),
			}
		}
	}
	
	countPipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$group", Value: bson.M{"_id": "$user"}}},
		{{Key: "$count", Value: "count"}},
	}
	total, err := r.runCountPipeline(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$group", Value: bson.M{
			"_id":           "$user",
			"usage_count":   bson.M{"$sum": 1},
			"total_discount": bson.M{"$sum": "$appliedPromo.discountAmt"},
			"last_used_at":  bson.M{"$max": "$createdAt"},
		}}},
		{{Key: "$sort", Value: bson.M{"usage_count": -1}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var rows []dto.PromoUserUsageRow
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}


func (r *AcceptedServiceRepo) runCountPipeline(ctx context.Context, pipeline mongo.Pipeline) (int64, error) {
	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)
	var results []struct {
		Count int64 `bson:"count"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}
	return results[0].Count, nil
}

func (r *AcceptedServiceRepo) GetDiscountUsersAggregated(
	ctx context.Context,
	discountID, createdAt string,
	skip, limit int64,
) ([]dto.DiscountUserUsageRow, int64, error) {
	matchStage := bson.M{"appliedDiscount.discountId": discountID}

	if createdAt != "" {
		t, err := time.Parse("2006-01-02", createdAt)
		if err == nil {
			startOfDay := t
			endOfDay := t.Add(24 * time.Hour)
	
			matchStage["createdAt"] = bson.M{
				"$gte": primitive.NewDateTimeFromTime(startOfDay),
				"$lt":  primitive.NewDateTimeFromTime(endOfDay),
			}
		}
	}
	
	countPipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$group", Value: bson.M{"_id": "$user"}}},
		{{Key: "$count", Value: "count"}},
	}
	total, err := r.runCountPipeline(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$group", Value: bson.M{
			"_id":            "$user",
			"usage_count":    bson.M{"$sum": 1},
			"total_discount": bson.M{"$sum": "$appliedDiscount.discountAmt"},
			"last_used_at":   bson.M{"$max": "$createdAt"},
		}}},
		{{Key: "$sort", Value: bson.M{"usage_count": -1}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var rows []dto.DiscountUserUsageRow
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *AcceptedServiceRepo) GetServicesByDiscountAndUser(
	ctx context.Context,
	discountID, userID, createdAt string,
	skip, limit int64,
) ([]domain.AcceptedService, int64, error) {
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}
   log.Println("cdkjbhbshjbhjdsbcdcscdsxs",createdAt)
	filter := bson.M{
		"appliedDiscount.discountId": discountID,
		"user":                       userOID,
	}
    log.Println("createdAt", createdAt)
	if createdAt != "" {
		t, err := time.Parse("2006-01-02", createdAt)
		if err == nil {
	
			startOfDay := t
			endOfDay := t.Add(24 * time.Hour)
	
			filter["createdAt"] = bson.M{
				"$gte": primitive.NewDateTimeFromTime(startOfDay),
				"$lt":  primitive.NewDateTimeFromTime(endOfDay),
			}
		}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var services []domain.AcceptedService
	if err := cursor.All(ctx, &services); err != nil {
		return nil, 0, err
	}
	return services, total, nil
}

func (r *AcceptedServiceRepo) GetDiscountUserCount(ctx context.Context, discountID string) (int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"appliedDiscount.discountId": discountID}}},
		{{Key: "$group", Value: bson.M{"_id": "$user"}}},
		{{Key: "$count", Value: "count"}},
	}
	return r.runCountPipeline(ctx, pipeline)
}

func (r *AcceptedServiceRepo) GetDiscountUsageStats(
	ctx context.Context,
	discountIDs []string,
) (map[string]struct {
	TotalOrders  int64
	TotalSavings float64
}, error) {

	match := bson.M{
		"appliedDiscount.discountId": bson.M{"$in": discountIDs},
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$appliedDiscount.discountId"},
			{Key: "totalOrders", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "totalSavings", Value: bson.D{{Key: "$sum", Value: "$appliedDiscount.discountAmt"}}},
		}}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := map[string]struct {
		TotalOrders  int64
		TotalSavings float64
	}{}

	for cursor.Next(ctx) {
		var row struct {
			ID           string  `bson:"_id"`
			TotalOrders  int64   `bson:"totalOrders"`
			TotalSavings float64 `bson:"totalSavings"`
		}
		if err := cursor.Decode(&row); err != nil {
			return nil, err
		}

		result[row.ID] = struct {
			TotalOrders  int64
			TotalSavings float64
		}{
			TotalOrders:  row.TotalOrders,
			TotalSavings: row.TotalSavings,
		}
	}

	return result, nil
}

func (r *AcceptedServiceRepo) GetPromoUsageStats(
	ctx context.Context,
	promoIDs []string,
) (map[string]float64, error) {

	match := bson.M{
		"appliedPromo.code": bson.M{"$in": promoIDs},
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$appliedPromo.code"},
			{Key: "totalDiscount", Value: bson.D{
				{Key: "$sum", Value: "$appliedPromo.discountAmt"},
			}},
		}}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := map[string]float64{}

	for cursor.Next(ctx) {
		var row struct {
			Code          string  `bson:"_id"`
			TotalDiscount float64 `bson:"totalDiscount"`
		}
		if err := cursor.Decode(&row); err != nil {
			return nil, err
		}
		result[row.Code] = row.TotalDiscount
	}

	return result, nil
}

func (r *AcceptedServiceRepo) GetOfferUsageByUser(
	ctx context.Context,
	userID, createdAt string,
	skip, limit int64,
) ([]domain.AcceptedService, int64, error) {

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	
	filter := bson.M{
		"user": userObjID,
		"$or": []bson.M{
			{"appliedPromo": bson.M{"$ne": nil}},
			{"appliedDiscount": bson.M{"$ne": nil}},
		},
	}

	if createdAt != "" {
		t, err := time.Parse("2006-01-02", createdAt)
		if err == nil {
	
			startOfDay := t
			endOfDay := t.Add(24 * time.Hour)
	
			filter["createdAt"] = bson.M{
				"$gte": primitive.NewDateTimeFromTime(startOfDay),
				"$lt":  primitive.NewDateTimeFromTime(endOfDay),
			}
		}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var services []domain.AcceptedService
	if err := cursor.All(ctx, &services); err != nil {
		return nil, 0, err
	}

	return services, total, nil
}