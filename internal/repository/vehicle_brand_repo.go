package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"provider_management/internal/domain"
)

type VehicleBrandRepo struct {
	col *mongo.Collection
}

func NewVehicleBrandRepo(db *mongo.Database) *VehicleBrandRepo {
	return &VehicleBrandRepo{col: db.Collection("vehiclebrands")}
}

func (r *VehicleBrandRepo) FindWithFilter(ctx context.Context, vehicleType, brandName string) ([]domain.VehicleBrand, error) {
	filter := bson.M{}

	if vehicleType != "" {
		filter["vehicleType"] = vehicleType
	}

	if brandName != "" {
		filter["brandName"] = brandName
	}

	opts := options.Find().
		SetProjection(bson.M{"brandName": 1, "vehicleType": 1, "modelName": 1}).
		SetSort(bson.M{"brandName": 1})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var brands []domain.VehicleBrand
	if err := cursor.All(ctx, &brands); err != nil {
		return nil, err
	}

	return brands, nil
}

func (r *VehicleBrandRepo) FindByTypeAndName(ctx context.Context, vehicleType, brandName string) (*domain.VehicleBrand, error) {
	var brand domain.VehicleBrand

	opts := options.FindOne().SetProjection(bson.M{"modelName": 1})

	err := r.col.FindOne(ctx, bson.M{"vehicleType": vehicleType, "brandName": brandName}, opts).Decode(&brand)
	if err != nil {
		return nil, err
	}

	return &brand, nil
}