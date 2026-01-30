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

type AMCPlanRepo struct {
	col *mongo.Collection
}

func NewAMCPlanRepo(db *mongo.Database) *AMCPlanRepo {
	return &AMCPlanRepo{col: db.Collection("amcplans")}
}

func (r *AMCPlanRepo) Create(ctx context.Context, plan *domain.AMCPlan) error {
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()
	result, err := r.col.InsertOne(ctx, plan)
	if err != nil {
		return err
	}
	plan.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *AMCPlanRepo) FindBySlug(ctx context.Context, slug string) (*domain.AMCPlan, error) {
	var plan domain.AMCPlan
	err := r.col.FindOne(ctx, bson.M{"planSlug": slug}).Decode(&plan)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *AMCPlanRepo) GetPlans( ctx context.Context, filter bson.M, skip, limit int64, sortField string, sortOrder int ) ([]domain.AMCPlan, int64, error) {
	if sortField == "" {
		sortField = "createdAt"
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

	var plans []domain.AMCPlan
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, 0, err
	}

	total, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	return plans, total, nil
}

func (r *AMCPlanRepo) FindByID(ctx context.Context, id string) (*domain.AMCPlan, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var plan domain.AMCPlan
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&plan)
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

func (r *AMCPlanRepo) Update(ctx context.Context, id string, update bson.M) error {
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

func (r *AMCPlanRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.col.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *AMCPlanRepo) CountDocuments(ctx context.Context, filter bson.M) (int64, error) {
	return r.col.CountDocuments(ctx, filter)
}

func (r *AMCPlanRepo) FindByCategory(
	ctx context.Context,
	vehicleType, category, cityName string,
) ([]domain.AMCPlan, error) {
	filter := bson.M{
		"planVehicleType": vehicleType,
		"planCategory":    category,
		"planCity":        bson.M{"$in": []string{cityName}},
		"isActive":        true,
	}

	opts := options.Find().
		SetSort(bson.M{"sorting": -1, "planTotalAmount": -1})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plans []domain.AMCPlan
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, err
	}

	return plans, nil
}