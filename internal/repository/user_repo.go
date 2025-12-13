package repository

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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

func (r *UserRepo) FindAll(
	ctx context.Context,
	query bson.M,
	skip, limit int64,
) ([]domain.User, int64, error) {

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})
    
	cur, err := r.col.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var users []domain.User
	if err := cur.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	total, err := r.col.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}
    log.Println("Repo found total users:", total)
	return users, total, nil
}

func (r *UserRepo) FindOne(ctx context.Context, query bson.M) (*domain.User, error) {
	var user domain.User
	if err := r.col.FindOne(ctx, query).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) UpdateStatus(
	ctx context.Context,
	query bson.M,
	status string,
) (*domain.User, error) {

	isActive := false
if status == "active" {
    isActive = true
}


	var user domain.User
	err := r.col.FindOneAndUpdate(
		ctx,
		query,
		bson.M{"$set": bson.M{"isActive": isActive}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&user)
    log.Println("Repo updated user status to:", status)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Count(ctx context.Context, query bson.M) (int64, error) {
	return r.col.CountDocuments(ctx, query)
}


