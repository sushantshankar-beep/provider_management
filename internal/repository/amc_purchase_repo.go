package repository

import (
	"context"
	"log"
	"provider_management/internal/domain"
	"time"
    "fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AMCPurchaseRepo struct {
	col *mongo.Collection
}

func NewAMCPurchaseRepo(db *mongo.Database) *AMCPurchaseRepo {
	return &AMCPurchaseRepo{col: db.Collection("amcpurchaseschemas")}
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

func (r *AMCPurchaseRepo) FindActiveAMCs(ctx context.Context) ([]domain.AMCPurchase, error) {
	now := time.Now()
	filter := bson.M{
		"planEndDate":   bson.M{"$gt": now},
		"paymentStatus": "paid",
	}

	var amcs []domain.AMCPurchase
	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &amcs); err != nil {
		return nil, err
	}

	return amcs, nil
}

func (r *AMCPurchaseRepo) FindByID(
	ctx context.Context,
	id string,
) (*domain.AMCPurchase, error) {

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var res domain.AMCPurchase

	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("amc purchase not found")
		}
		return nil, err
	}

	log.Println("decoded purchase:", res)
	return &res, nil
}

func (r *AMCPurchaseRepo) UpdateRefundStatus(ctx context.Context, id, status string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"refundStatus": status,
			"updatedAt":    time.Now(),
		},
	}

	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *AMCPurchaseRepo) UpdateRefundStatusAndPlanStatus(ctx context.Context, id, refundStatus, planStatus string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"refundStatus": refundStatus,
			"planStatus":   planStatus,
			"updatedAt":    time.Now(),
		},
	}

	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}