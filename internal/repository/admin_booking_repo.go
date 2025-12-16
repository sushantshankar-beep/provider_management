package repository

import (
	"context"
	"fmt"
	"provider_management/internal/domain"
  	"log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AdminBookingRepo struct {
	acceptedServiceColl *mongo.Collection
	serviceRequestColl  *mongo.Collection
	userColl            *mongo.Collection
	providerColl        *mongo.Collection
	ratingColl          *mongo.Collection
	transactionColl     *mongo.Collection
	complaintColl       *mongo.Collection
}

func NewAdminBookingRepo(db *mongo.Database) *AdminBookingRepo {
	return &AdminBookingRepo{
		acceptedServiceColl: db.Collection("acceptedservices"),
		serviceRequestColl:  db.Collection("servicerequests"),
		userColl:            db.Collection("users"),
		providerColl:        db.Collection("providerschema"),
		ratingColl:          db.Collection("ratings"),
		transactionColl:     db.Collection("transactions"),
		complaintColl:       db.Collection("complaints"),
	}
}

func (r *AdminBookingRepo) FindAcceptedServices(
	ctx context.Context,
	query bson.M,
	skip, limit int64,
	sort string,
) ([]domain.AcceptedService, int64, error) {

	log.Printf("Query: %v", query)
    log.Printf("Skip: %d, Limit: %d", skip, limit)
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

	log.Printf("Fetched %d services", len(services))
    for i, s := range services {
        log.Printf("Service %d: ID=%s, UserID=%s, ProviderID=%s", 
            i, s.ID, s.UserID, s.ProviderID)
    }
    
	return services, total, nil
}

func (r *AdminBookingRepo) CountAcceptedServices(ctx context.Context, filter bson.M) (int64, error) {
	return r.acceptedServiceColl.CountDocuments(ctx, filter)
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

func (r *AdminBookingRepo) FindAcceptedServiceByServiceRequestID(ctx context.Context, serviceRequestID string) (*domain.AcceptedService, error) {
	var service domain.AcceptedService

	objID, err := primitive.ObjectIDFromHex(serviceRequestID)
	if err == nil {
		err = r.acceptedServiceColl.FindOne(ctx, bson.M{"serviceRequest": objID}).Decode(&service)
		if err == nil {
			return &service, nil
		}
	}

	err = r.acceptedServiceColl.FindOne(ctx, bson.M{"serviceRequest": serviceRequestID}).Decode(&service)
	return &service, err
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

func (r *AdminBookingRepo) FindServiceRequestByInternalID(ctx context.Context, internalID int64) (*domain.ServiceRequest, error) {
	var request domain.ServiceRequest
	filter := bson.M{"id": internalID}
	err := r.serviceRequestColl.FindOne(ctx, filter).Decode(&request)
	return &request, err
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
	if err == nil {
		err = r.providerColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&provider)
		if err == nil {
			return &provider, nil
		}
	}

	err = r.providerColl.FindOne(ctx, bson.M{"_id": id}).Decode(&provider)
	if err != nil {
		var internalID int64
		if _, err := fmt.Sscanf(id, "%d", &internalID); err == nil {
			err = r.providerColl.FindOne(ctx, bson.M{"id": internalID}).Decode(&provider)
			if err == nil {
				return &provider, nil
			}
		}
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