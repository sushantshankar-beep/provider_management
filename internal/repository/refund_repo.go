package repository

import (
	"context"
	"fmt"
	"provider_management/internal/domain"
	"time"
    "strconv"
	"strings"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RefundRepository struct {
	collection *mongo.Collection
	counterCol *mongo.Collection
	userCollection *mongo.Collection
}

func NewRefundRepository(db *mongo.Database) *RefundRepository {
	return &RefundRepository{
		collection: db.Collection("Refunds"),
		counterCol: db.Collection("counters"),
		userCollection: db.Collection("users"),
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

func (r *RefundRepository) UpdateStatus(ctx context.Context, refundID string, status domain.RefundStatus, failureReason string) error {
	update := bson.M{
		"status":    status,
		"updatedAt": time.Now(),
	}

	if status == domain.RefundStatusSuccess || status == domain.RefundStatusFailed {
		update["processedAt"] = time.Now()
	}

	if failureReason != "" {
		update["failureReason"] = failureReason
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"refundId": refundID},
		bson.M{"$set": update},
	)

	if err != nil {
		return fmt.Errorf("failed to update refund status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("refund not found with ID: %s", refundID)
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

	if filter.UserID != "" {
		userObjID, err := primitive.ObjectIDFromHex(filter.UserID)
		if err != nil {
			return nil, 0, err
		}
		query["userId"] = userObjID
	}

	if filter.Search != "" {
		orFilters := []bson.M{}

		searchUpper := strings.ToUpper(filter.Search)

		if strings.HasPrefix(searchUpper, "RF") {
			orFilters = append(orFilters, bson.M{"refundId": filter.Search})
			orFilters = append(orFilters, bson.M{"refundId": bson.M{"$regex": filter.Search, "$options": "i"}})
		}

		if strings.HasPrefix(searchUpper, "VW") {
			if internalID, err := strconv.ParseInt(filter.Search[2:], 10, 64); err == nil {
				var user domain.User
				err := r.userCollection.FindOne(ctx, bson.M{"id": internalID}).Decode(&user)
				if err == nil && user.ID != "" {
					orFilters = append(orFilters, bson.M{"userId": user.ID})
				}
			}
		}

		if strings.HasPrefix(searchUpper, "BK") {
			if bookingNo, err := strconv.ParseInt(filter.Search[2:], 10, 64); err == nil {
				orFilters = append(orFilters, bson.M{"bookingNo": bookingNo})
			}
		}

		if strings.HasPrefix(searchUpper, "CMP") {
			if complaintNo, err := strconv.ParseInt(filter.Search[3:], 10, 64); err == nil {
				orFilters = append(orFilters, bson.M{"complaintNo": complaintNo})
			}
		}

		if numericSearch, err := strconv.ParseInt(filter.Search, 10, 64); err == nil {
			orFilters = append(orFilters,
				bson.M{"bookingNo": numericSearch},
				bson.M{"complaintNo": numericSearch},
			)
		}

		orFilters = append(orFilters, 
			bson.M{"transactionId": bson.M{"$regex": filter.Search, "$options": "i"}},
			bson.M{"refundId": bson.M{"$regex": filter.Search, "$options": "i"}},
		)

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
	err := r.collection.FindOne(ctx, bson.M{"refundId": refundID}).Decode(&refund)
	if err != nil {
		return nil, err
	}
	return &refund, nil
}