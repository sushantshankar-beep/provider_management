package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
)


type PlanRepo struct {
	col *mongo.Collection
}

func NewPlanRepo(db *mongo.Database) *PlanRepo {
	return &PlanRepo{col: db.Collection("amcplans")}
}

func (r *PlanRepo) FindByID(ctx context.Context, id string) (*domain.AMCPlan, error) {
	var plan domain.AMCPlan

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&plan)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}