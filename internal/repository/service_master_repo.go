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

type ServiceMasterRepo struct {
	col *mongo.Collection
}

func NewServiceMasterRepo(db *mongo.Database) *ServiceMasterRepo {
	return &ServiceMasterRepo{col: db.Collection("serviceMaster")}
}

func (r *ServiceMasterRepo) Create(ctx context.Context, service *domain.ServiceMaster) error {
	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, service)
	return err
}

func (r *ServiceMasterRepo) GetServices( ctx context.Context, filter bson.M, skip, limit int64, sortField string, sortOrder int ) ([]domain.ServiceMaster, int64, error) {
	if sortField == "" {
		sortField = "updatedAt"
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{sortField: sortOrder})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var services []domain.ServiceMaster
	if err := cursor.All(ctx, &services); err != nil {
		return nil, 0, err
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return services, total, nil
}

func (r *ServiceMasterRepo) FindByID(ctx context.Context, id string) (*domain.ServiceMaster, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var service domain.ServiceMaster
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&service)
	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *ServiceMasterRepo) Update(ctx context.Context, id string, update bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update["updatedAt"] = time.Now()

	_, err = r.col.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": update},
	)
	return err
}

func (r *ServiceMasterRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.col.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *ServiceMasterRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"status":    status,
		"updatedAt": time.Now(),
	}

	_, err = r.col.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": update},
	)
	return err
}