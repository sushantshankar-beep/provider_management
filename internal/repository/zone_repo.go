package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
)

type ZoneRepo struct {
	col *mongo.Collection
}

func NewZoneRepo(db *mongo.Database) *ZoneRepo {
	return &ZoneRepo{col: db.Collection("zones")}
}

func (r *ZoneRepo) FindByID(ctx context.Context, id string) (*domain.Zone, error) {
	var zone domain.Zone

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&zone)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

func (r *ZoneRepo) FindAll(ctx context.Context) ([]domain.Zone, error) {
	cursor, err := r.col.Find(ctx, bson.M{"isActive": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var zones []domain.Zone
	if err := cursor.All(ctx, &zones); err != nil {
		return nil, err
	}
	return zones, nil
}