package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RatingRepo struct {
	collection *mongo.Collection
}

func NewRatingRepo(db *mongo.Database) *RatingRepo {
	return &RatingRepo{
		collection: db.Collection("ratings"),
	}
}

func (r *RatingRepo) FindRatingsByBookingID(
	ctx context.Context,
	bookingID primitive.ObjectID,
) ([]*domain.Rating, error) {

	filter := bson.M{"bookingId": bookingID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []*domain.Rating
	if err := cursor.All(ctx, &ratings); err != nil {
		return nil, err
	}

	return ratings, nil
}

func (r *RatingRepo) FindRatingsByBookingIDs(
	ctx context.Context,
	bookingIDs []primitive.ObjectID,
) ([]*domain.Rating, error) {

	filter := bson.M{"bookingId": bson.M{"$in": bookingIDs}}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []*domain.Rating
	if err := cursor.All(ctx, &ratings); err != nil {
		return nil, err
	}

	return ratings, nil
}
