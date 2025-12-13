package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"provider_management/internal/domain"
)

type AcceptedServiceRepo struct {
	col *mongo.Collection
}

func NewAcceptedServiceRepo(db *mongo.Database) *AcceptedServiceRepo {
	return &AcceptedServiceRepo{col: db.Collection("acceptedservices")}
}

func (r *AcceptedServiceRepo) FindByID(ctx context.Context, id string) (*domain.AcceptedService, error) {
	var res domain.AcceptedService
	
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}