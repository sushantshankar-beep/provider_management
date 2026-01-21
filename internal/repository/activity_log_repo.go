package repository

import (
	"time"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityLogRepository struct {
	collection *mongo.Collection
}

func NewActivityLogRepository(db *mongo.Database) *ActivityLogRepository {
	return &ActivityLogRepository{
		collection: db.Collection("activity_logs"),
	}
}

func (r *ActivityLogRepository) Create(ctx context.Context, log *domain.ActivityLog) error {
	log.CreatedAt = time.Now()
	if log.ID.IsZero() {
		log.ID = primitive.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

func (r *ActivityLogRepository) GetAll( ctx context.Context, page, limit int, adminID, entityType, action string ) ([]domain.ActivityLogResponse, int64, error) {

	filter := bson.M{}

	if adminID != "" {
		if objID, err := primitive.ObjectIDFromHex(adminID); err == nil {
			filter["admin_id"] = objID
		}
	}
	if entityType != "" {
		filter["entity_type"] = entityType
	}
	if action != "" {
		filter["action"] = action
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := (page - 1) * limit

	projection := bson.M{
		"_id":         1,
		"admin_name":  1,
		"entity_type": 1,
		"entity_id":   1,
		"action":      1,
		"admin_id":    1,
		"created_at":  1,
	}

	opts := options.Find().
		SetProjection(projection).
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []domain.ActivityLogResponse
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}


func (r *ActivityLogRepository) GetByEntityID(ctx context.Context, entityType, entityID string, page, limit int ) ([]domain.ActivityLogResponse, int64, error) {

	filter := bson.M{
		"entity_type": entityType,
		"entity_id":   entityID,
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := (page - 1) * limit

	projection := bson.M{
		"_id":         1,
		"admin_name":  1,
		"entity_type": 1,
		"entity_id":   1,
		"action":      1,
		"admin_id":    1,
		"created_at":  1,
	}

	opts := options.Find().
		SetProjection(projection).
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []domain.ActivityLogResponse
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
