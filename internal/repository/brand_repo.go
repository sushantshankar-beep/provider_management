package repository

import (
	"context"
	"time"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProviderVehicleBrandRepo struct {
	col *mongo.Collection
}

func NewProviderVehicleBrandRepo(db *mongo.Database) *ProviderVehicleBrandRepo {
	return &ProviderVehicleBrandRepo{col: db.Collection("vehicleBrands")}
}

func (r *ProviderVehicleBrandRepo) FindByVehicle( ctx context.Context, vehicle string) ([]domain.ProviderVehicleBrand, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cur, err := r.col.Find(ctx, bson.M{
		"vehicleType": vehicle,
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var brands []domain.ProviderVehicleBrand
	if err := cur.All(ctx, &brands); err != nil {
		return nil, err
	}

	return brands, nil
}
