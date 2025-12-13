package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"provider_management/internal/domain"
)

type UserRepo struct {
	col *mongo.Collection
}

func NewUserRepo(db *mongo.Database) *UserRepo {
	return &UserRepo{col: db.Collection("users")}
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var res domain.User
	
	objID, err := primitive.ObjectIDFromHex(id)
	if err == nil {
		err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&res)
		if err == nil {
			return &res, nil
		}
	}
	
	err = r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}