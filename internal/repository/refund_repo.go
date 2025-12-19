package repository

import (
	"context"
	"fmt"
	"provider_management/internal/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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