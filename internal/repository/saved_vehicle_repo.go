package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SavedVehiclesRepo struct {
	col *mongo.Collection
}

func NewSavedVehiclesRepo(db *mongo.Database) *SavedVehiclesRepo {
	return &SavedVehiclesRepo{col: db.Collection("savedvehicles")}
}

func (r *SavedVehiclesRepo) FindByUserID(ctx context.Context, userID string) ([]domain.SavedVehicle, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.col.Find(ctx, bson.M{"userId": userObjID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var vehicles []domain.SavedVehicle
	if err := cursor.All(ctx, &vehicles); err != nil {
		return nil, err
	}

	return vehicles, nil
}

func (r *SavedVehiclesRepo) Aggregate(ctx context.Context, pipeline []bson.M) (*mongo.Cursor, error) {
	return r.col.Aggregate(ctx, pipeline)
}


func (r *SavedVehiclesRepo) FindByID(ctx context.Context, id string) (*domain.SavedVehicle, error) {
	var vehicle domain.SavedVehicle

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&vehicle)
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}