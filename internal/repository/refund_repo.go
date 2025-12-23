package repository

import (
	"context"
	"fmt"
	"log"
	"provider_management/internal/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RefundRepository struct {
	collection *mongo.Collection
	counterCol *mongo.Collection
}

func NewRefundRepository(db *mongo.Database) *RefundRepository {
	return &RefundRepository{
		collection: db.Collection("Refunds"),
		counterCol: db.Collection("counters"),
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

	// userId stored as ObjectId ✅
	if filter.UserID != "" {
		userObjID, err := primitive.ObjectIDFromHex(filter.UserID)
		if err != nil {
			return nil, 0, err
		}
		query["userId"] = userObjID
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