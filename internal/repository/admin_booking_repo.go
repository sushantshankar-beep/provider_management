package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"provider_management/internal/domain"
	"time"
)

type AdminBookingRepo struct {
	acceptedServiceColl *mongo.Collection
	serviceRequestColl  *mongo.Collection
	userColl            *mongo.Collection
	providerColl        *mongo.Collection
	ratingColl          *mongo.Collection
	transactionColl     *mongo.Collection
	complaintColl       *mongo.Collection
	collection          *mongo.Collection
}

func NewAdminBookingRepo(db *mongo.Database) *AdminBookingRepo {
	repo := &AdminBookingRepo{
		acceptedServiceColl: db.Collection("acceptedservices"),
		serviceRequestColl:  db.Collection("servicerequests"),
		userColl:            db.Collection("users"),
		providerColl:        db.Collection("providerschemas"),
		ratingColl:          db.Collection("ratings"),
		transactionColl:     db.Collection("transactions"),
		complaintColl:       db.Collection("complaints"),
	}
	repo.ensureIndexes(context.Background())
	return repo
}

func (r *AdminBookingRepo) ensureIndexes(ctx context.Context) {
	r.acceptedServiceColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user", Value: 1}},
	})
	r.acceptedServiceColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "provider", Value: 1}},
	})
	r.acceptedServiceColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	})
	r.acceptedServiceColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "paymentStatus", Value: 1}},
	})
	r.acceptedServiceColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "createdAt", Value: -1}},
	})
	r.acceptedServiceColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "id", Value: 1}},
	})
	r.acceptedServiceColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}, {Key: "paymentStatus", Value: 1}},
	})
	r.userColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "id", Value: 1}},
	})
	r.userColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
	})
	r.userColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "phone", Value: 1}},
	})
	r.userColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "email", Value: 1}},
	})
	r.providerColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "id", Value: 1}},
	})
	r.providerColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "providerId", Value: 1}},
	})
	r.providerColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
	})
	r.providerColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "phone", Value: 1}},
	})
	r.ratingColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "service", Value: 1}},
	})
	r.serviceRequestColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "id", Value: 1}},
	})
}

func (r *AdminBookingRepo) FindAcceptedServices(
	ctx context.Context,
	query bson.M,
	skip, limit int64,
	sort string,
) ([]domain.AcceptedService, int64, error) {
	sortOpts := bson.M{}
	if sort != "" {
		if sort[0] == '-' {
			sortOpts[sort[1:]] = -1
		} else {
			sortOpts[sort] = 1
		}
	} else {
		sortOpts["createdAt"] = -1
	}

	total, err := r.acceptedServiceColl.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(sortOpts)

	cursor, err := r.acceptedServiceColl.Find(ctx, query, opts)
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

func (r *AdminBookingRepo) CountAcceptedServices(ctx context.Context, filter bson.M) (int64, error) {
	return r.acceptedServiceColl.CountDocuments(ctx, filter)
}

func (r *AdminBookingRepo) GetBookingStatsBatch(ctx context.Context, baseFilter bson.M) (*domain.BookingStats, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: baseFilter}},
		bson.D{{Key: "$facet", Value: bson.D{
			{Key: "total", Value: bson.A{bson.D{{Key: "$count", Value: "count"}}}},
			{Key: "pending", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.D{{Key: "status", Value: bson.D{{Key: "$in", Value: []string{"pending", "accepted", "reached"}}}}}}},
				bson.D{{Key: "$count", Value: "count"}},
			}},
			{Key: "inProgress", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.D{{Key: "status", Value: "in_progress"}}}},
				bson.D{{Key: "$count", Value: "count"}},
			}},
			{Key: "completed", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.D{{Key: "status", Value: "completed"}}}},
				bson.D{{Key: "$count", Value: "count"}},
			}},
			{Key: "cancelled", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.D{{Key: "status", Value: "cancelled"}}}},
				bson.D{{Key: "$count", Value: "count"}},
			}},
			{Key: "revenue", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "status", Value: "completed"},
					{Key: "paymentStatus", Value: "success"},
				}}},
				bson.D{{Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "total", Value: bson.D{{Key: "$sum", Value: "$finalPrice"}}},
				}}},
			}},
		}}},
	}

	cursor, err := r.acceptedServiceColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	stats := &domain.BookingStats{}

	if len(result) > 0 {
		facets := result[0]

		if total, ok := facets["total"].(primitive.A); ok && len(total) > 0 {
			if countDoc, ok := total[0].(bson.M); ok {
				if count, ok := countDoc["count"].(int32); ok {
					stats.TotalBookings = int64(count)
				} else if count, ok := countDoc["count"].(int64); ok {
					stats.TotalBookings = count
				}
			}
		}

		if pending, ok := facets["pending"].(primitive.A); ok && len(pending) > 0 {
			if countDoc, ok := pending[0].(bson.M); ok {
				if count, ok := countDoc["count"].(int32); ok {
					stats.PendingBookings = int64(count)
				} else if count, ok := countDoc["count"].(int64); ok {
					stats.PendingBookings = count
				}
			}
		}

		if inProgress, ok := facets["inProgress"].(primitive.A); ok && len(inProgress) > 0 {
			if countDoc, ok := inProgress[0].(bson.M); ok {
				if count, ok := countDoc["count"].(int32); ok {
					stats.InProgressBookings = int64(count)
				} else if count, ok := countDoc["count"].(int64); ok {
					stats.InProgressBookings = count
				}
			}
		}

		if completed, ok := facets["completed"].(primitive.A); ok && len(completed) > 0 {
			if countDoc, ok := completed[0].(bson.M); ok {
				if count, ok := countDoc["count"].(int32); ok {
					stats.CompletedBookings = int64(count)
				} else if count, ok := countDoc["count"].(int64); ok {
					stats.CompletedBookings = count
				}
			}
		}

		if cancelled, ok := facets["cancelled"].(primitive.A); ok && len(cancelled) > 0 {
			if countDoc, ok := cancelled[0].(bson.M); ok {
				if count, ok := countDoc["count"].(int32); ok {
					stats.CancelledBookings = int64(count)
				} else if count, ok := countDoc["count"].(int64); ok {
					stats.CancelledBookings = count
				}
			}
		}

		if revenue, ok := facets["revenue"].(primitive.A); ok && len(revenue) > 0 {
			if revenueDoc, ok := revenue[0].(bson.M); ok {
				if total, ok := revenueDoc["total"].(float64); ok {
					stats.TotalRevenue = total
				} else if total, ok := revenueDoc["total"].(int32); ok {
					stats.TotalRevenue = float64(total)
				} else if total, ok := revenueDoc["total"].(int64); ok {
					stats.TotalRevenue = float64(total)
				}
			}
		}
	}

	return stats, nil
}

func (r *AdminBookingRepo) FindUsersByIDs(ctx context.Context, ids []string) ([]domain.User, error) {
	if len(ids) == 0 {
		return []domain.User{}, nil
	}

	var objIDs []primitive.ObjectID
	for _, id := range ids {
		if objID, err := primitive.ObjectIDFromHex(id); err == nil {
			objIDs = append(objIDs, objID)
		}
	}

	if len(objIDs) == 0 {
		return []domain.User{}, nil
	}

	filter := bson.M{"_id": bson.M{"$in": objIDs}}
	cursor, err := r.userColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *AdminBookingRepo) FindProvidersByIDs(ctx context.Context, ids []string) ([]domain.Provider, error) {
	if len(ids) == 0 {
		return []domain.Provider{}, nil
	}

	var objIDs []primitive.ObjectID
	for _, id := range ids {
		if objID, err := primitive.ObjectIDFromHex(id); err == nil {
			objIDs = append(objIDs, objID)
		}
	}

	if len(objIDs) == 0 {
		return []domain.Provider{}, nil
	}

	filter := bson.M{"_id": bson.M{"$in": objIDs}}
	cursor, err := r.providerColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var providers []domain.Provider
	if err := cursor.All(ctx, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}

func (r *AdminBookingRepo) FindServiceRequestsByIDs(ctx context.Context, ids []string) ([]domain.ServiceRequest, error) {
	if len(ids) == 0 {
		return []domain.ServiceRequest{}, nil
	}

	var objIDs []primitive.ObjectID
	for _, id := range ids {
		if objID, err := primitive.ObjectIDFromHex(id); err == nil {
			objIDs = append(objIDs, objID)
		}
	}

	if len(objIDs) == 0 {
		return []domain.ServiceRequest{}, nil
	}

	filter := bson.M{"_id": bson.M{"$in": objIDs}}
	cursor, err := r.serviceRequestColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []domain.ServiceRequest
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *AdminBookingRepo) FindAcceptedServiceByID(ctx context.Context, id string) (*domain.AcceptedService, error) {
	var service domain.AcceptedService

	objID, err := primitive.ObjectIDFromHex(id)
	if err == nil {
		err = r.acceptedServiceColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&service)
		if err == nil {
			return &service, nil
		}
	}

	err = r.acceptedServiceColl.FindOne(ctx, bson.M{"_id": id}).Decode(&service)
	if err != nil {
		var internalID int64
		if _, err := fmt.Sscanf(id, "%d", &internalID); err == nil {
			err = r.acceptedServiceColl.FindOne(ctx, bson.M{"id": internalID}).Decode(&service)
			if err == nil {
				return &service, nil
			}
		}
		return nil, err
	}
	return &service, nil
}

func (r *AdminBookingRepo) FindAcceptedServiceByServiceRequestID(
	ctx context.Context,
	srID primitive.ObjectID,
) (*domain.AcceptedService, error) {
	var svc domain.AcceptedService
	err := r.acceptedServiceColl.FindOne(ctx, bson.M{
		"serviceRequest": srID,
	}).Decode(&svc)
	return &svc, err
}

func (r *AdminBookingRepo) UpdateAcceptedService(ctx context.Context, id string, update bson.M) (*domain.AcceptedService, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var service domain.AcceptedService
	err = r.acceptedServiceColl.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&service)

	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *AdminBookingRepo) AggregateRevenue(ctx context.Context, filter bson.M) (float64, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "totalRevenue", Value: bson.D{{Key: "$sum", Value: "$finalPrice"}}},
		}}},
	}

	cursor, err := r.acceptedServiceColl.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalRevenue float64 `bson:"totalRevenue"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.TotalRevenue, nil
	}

	return 0, nil
}

func (r *AdminBookingRepo) FindServiceRequestByID(ctx context.Context, id string) (*domain.ServiceRequest, error) {
	var request domain.ServiceRequest

	objID, err := primitive.ObjectIDFromHex(id)
	if err == nil {
		err = r.serviceRequestColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&request)
		if err == nil {
			return &request, nil
		}
	}

	err = r.serviceRequestColl.FindOne(ctx, bson.M{"_id": id}).Decode(&request)
	if err != nil {
		var internalID int64
		if _, err := fmt.Sscanf(id, "%d", &internalID); err == nil {
			err = r.serviceRequestColl.FindOne(ctx, bson.M{"id": internalID}).Decode(&request)
			if err == nil {
				return &request, nil
			}
		}
		return nil, err
	}
	return &request, nil
}

func (r *AdminBookingRepo) FindServiceRequestByInternalID(
	ctx context.Context,
	internalID int64,
) (*domain.ServiceRequest, error) {
	var sr domain.ServiceRequest
	err := r.serviceRequestColl.FindOne(ctx, bson.M{"id": internalID}).Decode(&sr)
	return &sr, err
}

func (r *AdminBookingRepo) FindUserByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User

	objID, err := primitive.ObjectIDFromHex(id)
	if err == nil {
		err = r.userColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
		if err == nil {
			return &user, nil
		}
	}

	err = r.userColl.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		var internalID int64
		if _, err := fmt.Sscanf(id, "%d", &internalID); err == nil {
			err = r.userColl.FindOne(ctx, bson.M{"id": internalID}).Decode(&user)
			if err == nil {
				return &user, nil
			}
		}
		return nil, err
	}
	return &user, nil
}

func (r *AdminBookingRepo) FindUserByInternalID(ctx context.Context, internalID int64) (*domain.User, error) {
	var user domain.User
	filter := bson.M{"id": internalID}
	err := r.userColl.FindOne(ctx, filter).Decode(&user)
	return &user, err
}

func (r *AdminBookingRepo) FindUsers(ctx context.Context, filter bson.M) ([]domain.User, error) {
	cursor, err := r.userColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *AdminBookingRepo) FindProviderByID(ctx context.Context, id string) (*domain.Provider, error) {
	var provider domain.Provider

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.providerColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&provider)
	if err != nil {
		return nil, err
	}

	return &provider, nil
}

func (r *AdminBookingRepo) FindProviderByInternalID(ctx context.Context, internalID int64) (*domain.Provider, error) {
	var provider domain.Provider
	filter := bson.M{"id": internalID}
	err := r.providerColl.FindOne(ctx, filter).Decode(&provider)
	return &provider, err
}

func (r *AdminBookingRepo) FindProviders(ctx context.Context, filter bson.M) ([]domain.Provider, error) {
	cursor, err := r.providerColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var providers []domain.Provider
	if err := cursor.All(ctx, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}

func (r *AdminBookingRepo) UpdateProvider(ctx context.Context, id string, update bson.M) (*domain.Provider, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var provider domain.Provider
	err = r.providerColl.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&provider)

	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *AdminBookingRepo) FindRatingsByServiceIDs(ctx context.Context, serviceIDs []string) ([]domain.Rating, error) {
	var objIDs []primitive.ObjectID
	for _, id := range serviceIDs {
		if objID, err := primitive.ObjectIDFromHex(id); err == nil {
			objIDs = append(objIDs, objID)
		}
	}

	if len(objIDs) == 0 {
		return []domain.Rating{}, nil
	}

	filter := bson.M{"service": bson.M{"$in": objIDs}}
	cursor, err := r.ratingColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []domain.Rating
	if err := cursor.All(ctx, &ratings); err != nil {
		return nil, err
	}

	return ratings, nil
}

func (r *AdminBookingRepo) FindTransactionByTxnID(ctx context.Context, txnID string) (*domain.Transaction, error) {
	var transaction domain.Transaction
	filter := bson.M{"txnid": txnID}
	err := r.transactionColl.FindOne(ctx, filter).Decode(&transaction)
	return &transaction, err
}

func (r *AdminBookingRepo) FindTransactionByOrderID(ctx context.Context, orderID string) (*domain.Transaction, error) {
	var transaction domain.Transaction

	filter := bson.M{"$or": []bson.M{
		{"txnid": orderID},
		{"mihpayid": orderID},
		{"_id": orderID},
	}}

	err := r.transactionColl.FindOne(ctx, filter).Decode(&transaction)
	return &transaction, err
}

func (r *AdminBookingRepo) FindComplaintByID(ctx context.Context, id string) (*domain.Complaint, error) {
	var complaint domain.Complaint

	objID, err := primitive.ObjectIDFromHex(id)
	if err == nil {
		err = r.complaintColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&complaint)
		if err == nil {
			return &complaint, nil
		}
	}

	err = r.complaintColl.FindOne(ctx, bson.M{"_id": id}).Decode(&complaint)
	if err != nil {
		var internalID int64
		if _, err := fmt.Sscanf(id, "%d", &internalID); err == nil {
			err = r.complaintColl.FindOne(ctx, bson.M{"id": internalID}).Decode(&complaint)
			if err == nil {
				return &complaint, nil
			}
		}
		return nil, err
	}
	return &complaint, nil
}

func (r *AdminBookingRepo) FindServiceRequestsByInternalID(ctx context.Context, internalIDs []int64) ([]domain.ServiceRequest, error) {
	if len(internalIDs) == 0 {
		return []domain.ServiceRequest{}, nil
	}

	filter := bson.M{"id": bson.M{"$in": internalIDs}}
	cursor, err := r.serviceRequestColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []domain.ServiceRequest
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *AdminBookingRepo) FindProviderByProviderID(ctx context.Context, providerID string) (*domain.Provider, error) {
	var provider domain.Provider

	filter := bson.M{"providerId": providerID}
	err := r.providerColl.FindOne(ctx, filter).Decode(&provider)
	if err == nil {
		return &provider, nil
	}

	if primitive.IsValidObjectID(providerID) {
		objID, _ := primitive.ObjectIDFromHex(providerID)
		filter = bson.M{"_id": objID}
		err = r.providerColl.FindOne(ctx, filter).Decode(&provider)
		if err == nil {
			return &provider, nil
		}
	}

	return nil, err
}

func (r *AdminBookingRepo) FindUsersByInternalID(ctx context.Context, internalID int64) ([]domain.User, error) {
	filter := bson.M{"id": internalID}
	cursor, err := r.userColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *AdminBookingRepo) FindUsersBySearch(ctx context.Context, search string) ([]domain.User, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"phone": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
		},
	}

	opts := options.Find().SetLimit(50)
	cursor, err := r.userColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *AdminBookingRepo) FindProvidersBySearch(ctx context.Context, search string) ([]domain.Provider, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"phone": bson.M{"$regex": search, "$options": "i"}},
			{"providerId": bson.M{"$regex": search, "$options": "i"}},
			{"_id": search},
		},
	}

	opts := options.Find().SetLimit(50)
	cursor, err := r.providerColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var providers []domain.Provider
	if err := cursor.All(ctx, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}

func (r *AdminBookingRepo) FindServiceRequestByInternalIDs(ctx context.Context, internalID int64) (*domain.ServiceRequest, error) {
	var sr domain.ServiceRequest
	err := r.serviceRequestColl.FindOne(ctx, bson.M{"id": internalID}).Decode(&sr)
	if err != nil {
		return nil, err
	}
	return &sr, nil
}

func (r *AdminBookingRepo) UpdateProviderIsAssigned(ctx context.Context, providerID string, isAssigned bool) error {
	objID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		return err
	}

	_, err = r.providerColl.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"isAssigned": isAssigned}},
	)
	return err
}

func (r *AdminBookingRepo) AddBookingNote(
	ctx context.Context,
	serviceID primitive.ObjectID,
	note domain.BookingNote,
) error {
	_, err := r.acceptedServiceColl.UpdateOne(
		ctx,
		bson.M{"_id": serviceID},
		bson.M{
			"$push": bson.M{
				"notes": note,
			},
			"$set": bson.M{
				"updatedAt": time.Now(),
			},
		},
	)
	return err
}