package repository

import (
	"context"
	"provider_management/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AgreementRepo struct {
	col *mongo.Collection
}

func NewAgreementRepo(db *mongo.Database) *AgreementRepo {
	return &AgreementRepo{col: db.Collection("providerAgreement")}
}

func (r *AgreementRepo) FindByID(ctx context.Context, id string) (*domain.Agreement, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var agreement domain.Agreement
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&agreement)
	if err != nil {
		return nil, err
	}
	return &agreement, nil
}

func (r *AgreementRepo) Create(ctx context.Context, agreement *domain.Agreement) error {
	result, err := r.col.InsertOne(ctx, agreement)
	if err != nil {
		return err
	}
	agreement.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *AgreementRepo) Update(ctx context.Context, id string, agreement *domain.Agreement) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{"$set": agreement}
	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *AgreementRepo) FindDefault(ctx context.Context) (*domain.Agreement, error) {
	var agreement domain.Agreement
	err := r.col.FindOne(ctx, bson.M{}).Decode(&agreement)
	if err != nil {
		return nil, err
	}
	return &agreement, nil
}