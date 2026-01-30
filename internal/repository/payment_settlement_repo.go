package repository

import (
	"fmt"
	"time"
	"context"
	"provider_management/internal/dto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/utils"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (r *ProviderSettlementRepo) GetSettlements(ctx context.Context, filter bson.M, skip, limit int64, sortField string, sortOrder int) ([]domain.ProviderSettlement, int64, error) {

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

func (r *ProviderSettlementRepo) FindByID( ctx context.Context, id primitive.ObjectID ) (*domain.ProviderSettlement, error) {

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

func (r *ProviderSettlementRepo) UpdateSettlementPostData( ctx context.Context, settlementID primitive.ObjectID, data *domain.SettledPostData, status domain.SettlementStatus, settledAt time.Time ) error {

	update := bson.M{
		"$set": bson.M{
			"settledPostData": data,
			"status":          status,
			"settledAt":       settledAt,
			"updatedAt":       time.Now(),
		},
	}

	_, err := r.coll.UpdateOne(
		ctx,
		bson.M{"_id": settlementID},
		update,
	)

	return err
}

func (r *ProviderSettlementRepo) FindBySettlementIDs( ctx context.Context,ids []string) ([]domain.ProviderSettlement, error) {
	objectIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		objectIDs = append(objectIDs, objID)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}

	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var settlements []domain.ProviderSettlement
	for cursor.Next(ctx) {
		var settlement domain.ProviderSettlement
		if err := cursor.Decode(&settlement); err != nil {
			return nil, err
		}
		settlements = append(settlements, settlement)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return settlements, nil
}

func (r *ProviderSettlementRepo) GetStats(ctx context.Context, days string) (dto.SettlementStats, error) {
	startDay := utils.GetStartDateFromDays(days)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"createdAt": bson.M{"$gte": startDay},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"settledAmount": bson.M{"$sum": bson.M{"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$status", "settled"}}, "$totalAmount", 0,
				}}},
			},
		},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return dto.SettlementStats{}, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		SettledAmount float64 `bson:"settledAmount"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return dto.SettlementStats{}, err
	}

	stats := dto.SettlementStats{}
	if len(result) > 0 {
		stats.SettledAmount = utils.RoundTo2(result[0].SettledAmount)
	}

	return stats, nil
}

func (r *ProviderSettlementRepo) Update( ctx context.Context, id primitive.ObjectID, updateData map[string]interface{} ) error {

	updateData["updatedAt"] = time.Now()

	update := bson.M{
		"$set": updateData,
	}

	_, err := r.coll.UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
	)

	return err
}
