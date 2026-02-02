package repository

import (
	"context"
	"log"
	"provider_management/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func (r *SavedVehiclesRepo) FindByIDs(
	ctx context.Context,
	ids []primitive.ObjectID,
) ([]domain.Vehicle, error) {

	log.Println("vehiclesss",ids)
	if len(ids) == 0 {
		return []domain.Vehicle{}, nil
	}

	filter := bson.M{
		"_id": bson.M{"$in": ids},
	}

	cursor, err := r.col.Find(ctx, filter)
	log.Println("cursor",cursor)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
    
	var vehicles []domain.Vehicle
	if err := cursor.All(ctx, &vehicles); err != nil {
		return nil, err
	}

	return vehicles, nil
}
