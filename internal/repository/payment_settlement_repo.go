package repository

import (
	"context"
	
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"time"
	"fmt"
		"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
)

type ProviderSettlementRepo struct {
	coll *mongo.Collection
}

func NewProviderSettlementRepo(db *mongo.Database) *ProviderSettlementRepo {
	return &ProviderSettlementRepo{
		coll: db.Collection("provider_settlements"),
	}
}

func (r *ProviderSettlementRepo) Create(ctx context.Context, settlement *domain.ProviderSettlement) error {
	settlement.CreatedAt = time.Now()
	result, err := r.coll.InsertOne(ctx, settlement)
	if err != nil {
		return err
	}
	settlement.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ProviderSettlementRepo) GetSettlements(
	ctx context.Context,
	filter bson.M,
	skip, limit int64,
	sortField string,
	sortOrder int,
) ([]domain.ProviderSettlement, int64, error) {

	if sortField == "" {
		sortField = "createdAt"
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{sortField: sortOrder})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find settlements: %v", err)
	}
	defer cursor.Close(ctx)

	var settlements []domain.ProviderSettlement
	if err := cursor.All(ctx, &settlements); err != nil {
		return nil, 0, fmt.Errorf("failed to decode settlements: %v", err)
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count settlements: %v", err)
	}

	return settlements, total, nil
}

func (r *ProviderSettlementRepo) FindByID(
	ctx context.Context,
	id primitive.ObjectID,
) (*domain.ProviderSettlement, error) {

	if r == nil || r.coll == nil {
		return nil, fmt.Errorf("settlement repository not initialized")
	}

	var settlement domain.ProviderSettlement
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&settlement)
	if err != nil {
		return nil, fmt.Errorf("settlement not found: %v", err)
	}

	return &settlement, nil
}