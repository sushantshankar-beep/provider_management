package repository

import (
	"context"

	"provider_management/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AMCPurchaseRepo struct {
	col *mongo.Collection
}

func NewAMCPurchaseRepo(db *mongo.Database) *AMCPurchaseRepo {
	return &AMCPurchaseRepo{col: db.Collection("amcpurchases")}
}

func (r *AMCPurchaseRepo) FindActiveByUserID(ctx context.Context, userID primitive.ObjectID) (*domain.AMCPurchase, error) {
	var amc domain.AMCPurchase
	err := r.col.FindOne(ctx, bson.M{
		"user":          userID,
		"planStatus":    "active",
		"paymentStatus": "success",
	}).Decode(&amc)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &amc, nil
}

func (r *AMCPurchaseRepo) FindActiveByUserIDs(ctx context.Context, userIDs []primitive.ObjectID) ([]domain.AMCPurchase, error) {
	cursor, err := r.col.Find(ctx, bson.M{
		"user":          bson.M{"$in": userIDs},
		"paymentStatus": "success",
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var amcList []domain.AMCPurchase
	if err := cursor.All(ctx, &amcList); err != nil {
		return nil, err
	}

	return amcList, nil
}