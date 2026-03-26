package repository

import (
	"fmt"
	"time"
	"strings"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefundRepository struct {
	collection *mongo.Collection
	counterCol *mongo.Collection
	userCollection *mongo.Collection
	complaintCollection *mongo.Collection
	acceptedServiceCollection *mongo.Collection
}

func NewRefundRepository(db *mongo.Database) *RefundRepository {
	return &RefundRepository{
		collection: db.Collection("refund_transactions"),
		counterCol: db.Collection("counters"),
		userCollection: db.Collection("users"),
		complaintCollection: db.Collection("complaints"),
		acceptedServiceCollection: db.Collection("acceptedservices"),
	}
}

func (r *RefundRepository) Create(ctx context.Context, refund *domain.Refund) error {
	refund.CreatedAt = time.Now()
	refund.UpdatedAt = time.Now()

	if refund.RefundID == "" {
		refundID, err := r.GenerateRefundID(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate refund ID: %w", err)
		}
		refund.RefundID = refundID
	}

	_, err := r.collection.InsertOne(ctx, refund)
	if err != nil {
		return fmt.Errorf("failed to create refund: %w", err)
	}

	return nil
}

func (r *RefundRepository) GenerateRefundID(ctx context.Context) (string, error) {
	filter := bson.M{"_id": "refund_id"}
	update := bson.M{"$inc": bson.M{"sequence_value": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		SequenceValue int64 `bson:"sequence_value"`
	}

	err := r.counterCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to generate refund ID: %w", err)
	}

	timestamp := time.Now().Unix()
	return fmt.Sprintf("RF%d%d", timestamp, result.SequenceValue), nil
}

func (r *RefundRepository) UpdateStatus(
	ctx context.Context,
	id primitive.ObjectID,
	status domain.RefundStatus,
	additionalData map[string]interface{},
) error {
	updateData := map[string]interface{}{
		"status":    status,
		"updatedAt": time.Now(),
	}

	for key, value := range additionalData {
		updateData[key] = value
	}

	update := bson.M{"$set": updateData}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("refund not found")
	}

	return nil
}


func (r *RefundRepository) FindAll(
	ctx context.Context,
	filter domain.RefundFilter,
	skip, limit int,
) ([]domain.Refund, int64, error) {

	query := bson.M{}

	if filter.Status != "" {
		query["status"] = filter.Status
	}

	if filter.Mode != "" {
		query["mode"] = filter.Mode
	}

	if filter.Reason != "" {
		query["reason"] = filter.Reason
	}
	
	if filter.UserID != "" {
		userObjID, err := primitive.ObjectIDFromHex(filter.UserID)
		if err != nil {
			return nil, 0, err
		}
		query["userId"] = userObjID
	}

	if filter.Search != "" {
		orFilters := []bson.M{}
		search := strings.TrimSpace(filter.Search)
		searchUpper := strings.ToUpper(search)
		
		orFilters = append(orFilters, bson.M{
		   "_id": bson.M{"$regex": search, "$options": "i"},
		})
		
		orFilters = append(orFilters, bson.M{
			"txnid": bson.M{"$regex": search, "$options": "i"},
		})
	
		if strings.HasPrefix(searchUpper, "VHCR") {
			var user domain.User
			err := r.userCollection.FindOne(
				ctx,
				bson.M{"userCode": search},
			).Decode(&user)
	
			if err == nil {
				orFilters = append(orFilters, bson.M{
					"userId": user.ID,
				})
			}
		}
	
		if strings.HasPrefix(searchUpper, "VHBK") {
			var service domain.AcceptedService
			err := r.acceptedServiceCollection.FindOne(
				ctx,
				bson.M{"serviceNumber": search},
			).Decode(&service)
	
			if err == nil {
				orFilters = append(orFilters, bson.M{
					"serviceId": service.ID.Hex(),
				})
			}
		}
	
		if strings.HasPrefix(searchUpper, "CMP") {
			var complaint domain.Complaint
			err := r.complaintCollection.FindOne(
				ctx,
				bson.M{"complaintNumber": search},
			).Decode(&complaint)
	
			if err == nil {
				orFilters = append(orFilters, bson.M{
					"complaintId": complaint.ID,
				})
			}
		}
	
		if len(orFilters) > 0 {
			query["$or"] = orFilters
		}
	}
	

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	skipInt64 := int64(skip)
	limitInt64 := int64(limit)

	cursor, err := r.collection.Find(
		ctx,
		query,
		&options.FindOptions{
			Skip:  &skipInt64,
			Limit: &limitInt64,
			Sort:  bson.M{"createdAt": -1},
		},
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var refunds []domain.Refund
	if err := cursor.All(ctx, &refunds); err != nil {
		return nil, 0, err
	}

	return refunds, total, nil
}

func (r *RefundRepository) FindByRefundID(ctx context.Context, refundID string) (*domain.Refund, error) {
	var refund domain.Refund
	err := r.collection.FindOne(ctx, bson.M{"_id": refundID}).Decode(&refund)
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *RefundRepository) Update(
	ctx context.Context,
	id primitive.ObjectID,
	update bson.M,
) error {

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
	)

	return err
}

func (r *RefundRepository) FindByID(ctx context.Context, id string) (*domain.Refund, error) {
	var refund domain.Refund

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		filter := bson.M{"id": id}
		err = r.collection.FindOne(ctx, filter).Decode(&refund)
		if err != nil {
			return nil, err
		}
		return &refund, nil
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&refund)
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *RefundRepository) FindByComplaintID(ctx context.Context, complaintID string) (*domain.Refund, error) {
	var refund domain.Refund
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	err := r.collection.FindOne(ctx, bson.M{"complaintId": complaintID}, opts).Decode(&refund)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &refund, nil
}


