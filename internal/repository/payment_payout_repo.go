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

type PaymentPayoutRepo struct {
	col *mongo.Collection
}

func NewPaymentPayoutRepo(db *mongo.Database) *PaymentPayoutRepo {
	return &PaymentPayoutRepo{col: db.Collection("payment_payouts")}
}

func (r *PaymentPayoutRepo) Create( ctx context.Context, payout *domain.PaymentPayout ) error {
	_, err := r.col.InsertOne(ctx, payout)
	return err
}

func (r *PaymentPayoutRepo) GetPayouts(
	ctx context.Context,
	filter bson.M,
	skip, limit int64,
	sortField string,
	sortOrder int,
) ([]domain.PaymentPayout, int64, error) {

	if sortField == "" {
		sortField = "updatedAt"
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{sortField: sortOrder})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var payouts []domain.PaymentPayout
	if err := cursor.All(ctx, &payouts); err != nil {
		return nil, 0, err
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return payouts, total, nil
}

func (r *PaymentPayoutRepo) FindByID(
	ctx context.Context,
	id primitive.ObjectID,
) (*domain.PaymentPayout, error) {

	if r == nil || r.col == nil {
		return nil, fmt.Errorf("payment payout repository not initialized")
	}

	var payout domain.PaymentPayout
    log.Println("IDVSKDJBJK",id);
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&payout)
	if err != nil {
		return nil, err
	}

	return &payout, nil
}

func (r *PaymentPayoutRepo) MarkSettled(
	ctx context.Context,
	payoutID primitive.ObjectID,
	settlementID primitive.ObjectID,
) error {

	update := bson.M{
		"$set": bson.M{
			"status":       "settled",
			"settlementId": settlementID,
			"updatedAt":    time.Now(),
		},
	}

	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": payoutID},
		update,
	)
	return err
}

func (r *PaymentPayoutRepo) FindByPayoutID(
    ctx context.Context,
    payoutID int64,
) (*domain.PaymentPayout, error) {
    
    if r == nil || r.col == nil {
        return nil, fmt.Errorf("payment payout repository not initialized")
    }

    var payout domain.PaymentPayout
    err := r.col.FindOne(ctx, bson.M{"payoutId": payoutID}).Decode(&payout)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, fmt.Errorf("payout with ID %d not found", payoutID)
        }
        return nil, fmt.Errorf("failed to find payout: %v", err)
    }

    return &payout, nil
}