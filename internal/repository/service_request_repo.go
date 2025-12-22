package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
)


type ServiceRequestRepo struct {
	col *mongo.Collection
}

func NewServiceRequestRepo(db *mongo.Database) *ServiceRequestRepo {
	return &ServiceRequestRepo{col: db.Collection("servicerequests")}
}

func (r *ServiceRequestRepo) FindByID(ctx context.Context, id string) (*domain.ServiceRequest, error) {
	var serviceRequest domain.ServiceRequest

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&serviceRequest)
	if err != nil {
		return nil, err
	}
	return &serviceRequest, nil
}
