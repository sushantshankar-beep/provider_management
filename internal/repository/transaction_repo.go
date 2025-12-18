package repository

import (
	"context"

	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"provider_management/internal/domain"
)

type TransactionRepo struct {
	col *mongo.Collection
}

func NewTransactionRepo(db *mongo.Database) *TransactionRepo {
	return &TransactionRepo{col: db.Collection("transactions")}
}

func (r *TransactionRepo) FindAll(ctx context.Context, skip, limit int64) ([]domain.Transaction, error) {
	opts := options.Find().SetSkip(skip).SetLimit(limit)
	cur, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []domain.Transaction
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *TransactionRepo) FindByID(ctx context.Context, id string) (*domain.Transaction, error) {
	var res domain.Transaction

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

func (r *TransactionRepo) Create(ctx context.Context, transaction map[string]interface{}) error {
	_, err := r.col.InsertOne(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}
