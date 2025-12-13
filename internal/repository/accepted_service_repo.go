package repository

import (
	"context"
     "time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
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


func (r *AcceptedServiceRepo) CountByUserID(ctx context.Context, userID string, filter bson.M) (int64, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	filter["user"] = userObjID
	return r.col.CountDocuments(ctx, filter)
}

func (r *AcceptedServiceRepo) SumExpensesByUserID(ctx context.Context, userID string) (int64, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"user":          userObjID,
			"status":        "completed",
			"paymentStatus": "paid",
		}},
		{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$finalPrice"},
		}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Total int64 `bson:"total"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.Total, nil
	}

	return 0, nil
}

func (r *AcceptedServiceRepo) CountCompletedInDateRange(ctx context.Context, userID primitive.ObjectID, start, end time.Time) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{
		"user":          userID,
		"status":        "completed",
		"paymentStatus": "paid",
		"createdAt": bson.M{
			"$gte": start,
			"$lte": end,
		},
	})
}

func (r *AcceptedServiceRepo) Aggregate(ctx context.Context, pipeline []bson.M) (*mongo.Cursor, error) {
	return r.col.Aggregate(ctx, pipeline)
}
