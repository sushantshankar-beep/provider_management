package repository

import (
	"context"
	"log"
	"provider_management/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type VehiclesRepo struct {
	col *mongo.Collection
}

func NewVehiclesRepo(db *mongo.Database) *VehiclesRepo {
	return &VehiclesRepo{col: db.Collection("vehicles")}
}

func (r *VehiclesRepo) FindByIDs(
	ctx context.Context,
	ids []primitive.ObjectID,
) ([]domain.Vehicle, error) {

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
