package repository

import (
	"fmt"
	"time"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SettlementHistoryRepository struct {
	collection *mongo.Collection
}

func NewSettlementHistoryRepository(db *mongo.Database) *SettlementHistoryRepository {
	return &SettlementHistoryRepository{collection: db.Collection("settlementHistory")}
}

func (r *SettlementHistoryRepository) Create(ctx context.Context, settlement *domain.SettlementRecord) (*domain.SettlementRecord, error) {
	result, err := r.collection.InsertOne(ctx, settlement)
	if err != nil {
		return nil, err
	}

	settlement.ID = result.InsertedID.(primitive.ObjectID)
	return settlement, nil
}

func (r *SettlementHistoryRepository) FindByID(ctx context.Context, id string) (*domain.SettlementRecord, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var settlement domain.SettlementRecord
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&settlement)
	if err != nil {
		return nil, err
	}

	return &settlement, nil
}

func (r *SettlementHistoryRepository) Find(ctx context.Context, filter bson.M) ([]*domain.SettlementRecord, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var settlements []*domain.SettlementRecord
	if err = cursor.All(ctx, &settlements); err != nil {
		return nil, err
	}

	return settlements, nil
}

func (r *SettlementHistoryRepository) FindWithPagination(ctx context.Context, filter bson.M, skip, limit int, sortField string, sortOrder int) ([]*domain.SettlementRecord, error) {
	opts := options.Find()
	opts.SetSkip(int64(skip))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.M{sortField: sortOrder})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var settlements []*domain.SettlementRecord
	if err = cursor.All(ctx, &settlements); err != nil {
		return nil, err
	}

	return settlements, nil
}

func (r *SettlementHistoryRepository) UpdateByID( ctx context.Context, id primitive.ObjectID, updateData map[string]interface{},
) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": updateData}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("settlement record not found")
	}

	return nil
}

func (r *SettlementHistoryRepository) UpdateMany(ctx context.Context, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	return r.collection.UpdateMany(ctx, filter, update)
}

func (r *SettlementHistoryRepository) Count(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

func (r *SettlementHistoryRepository) Aggregate(ctx context.Context, pipeline []bson.M) (*mongo.Cursor, error) {
	return r.collection.Aggregate(ctx, pipeline)
}

func (r *SettlementHistoryRepository) FindByServiceID(ctx context.Context, serviceID primitive.ObjectID) (*domain.SettlementRecord, error) {
	var record domain.SettlementRecord
	filter := bson.M{"serviceId": serviceID}

	err := r.collection.FindOne(ctx, filter).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil 
		}
		return nil, err
	}

	return &record, nil
}

func (r *SettlementHistoryRepository) UpdateStatusBySettlementID(ctx context.Context, settlementID primitive.ObjectID, status domain.SettlementStatus,
	settledAt *time.Time,
) error {
	filter := bson.M{"settlementId": settlementID}
	update := bson.M{
		"$set": bson.M{
			"settlementStatus": status,
			"settledAt":        settledAt,
			"updatedAt":        time.Now(),
		},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}


func (r *SettlementHistoryRepository) GetSettlementRecords(
	ctx context.Context,
	filter bson.M,
	skip int,
	limit int,
	sortField string,
	sortOrder int,
) ([]*domain.SettlementRecord, int64, error) {

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find()
	opts.SetSkip(int64(skip))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.M{sortField: sortOrder})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var records []*domain.SettlementRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}

	return records, total, nil
}