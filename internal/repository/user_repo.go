package repository

import (
	"context"

	"time"

	"fmt"
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

	var user domain.User
	err := r.col.FindOneAndUpdate(
		ctx,
		query,
		bson.M{"$set": bson.M{"isActive": status}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&user)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Count(ctx context.Context, query bson.M) (int64, error) {
	return r.col.CountDocuments(ctx, query)
}

func (r *UserRepo) GetStatistics(ctx context.Context, days int) (*UserStatistics, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)
	pipeline := []bson.M{
		{
			"$facet": bson.M{
				"total_users": []bson.M{
					{"$count": "count"},
				},
				"active_users": []bson.M{
					{"$match": bson.M{"isActive": domain.AccountStatusActive}},
					{"$count": "count"},
				},
			"inactive_users": []bson.M{
					{"$match": bson.M{
						"isActive": bson.M{
							"$in": []string{
								domain.AccountStatusSuspended,
								domain.AccountStatusBlacklisted,
								domain.AccountStatusDeactivated,
							},
						},
					}},
					{"$count": "count"},
				},
				"new_users": []bson.M{
					{"$match": bson.M{
						"createdAt": bson.M{"$gte": startDate, "$lte": endDate},
					}},
					{"$count": "count"},
				},
			},
		},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalUsers    []struct{ Count int64 } `bson:"total_users"`
		ActiveUsers   []struct{ Count int64 } `bson:"active_users"`
		InactiveUsers []struct{ Count int64 } `bson:"inactive_users"`
		NewUsers      []struct{ Count int64 } `bson:"new_users"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &UserStatistics{}, nil
	}

	stats := &UserStatistics{}

	if len(results[0].TotalUsers) > 0 {
		stats.TotalUsers = results[0].TotalUsers[0].Count
	}
	if len(results[0].ActiveUsers) > 0 {
		stats.ActiveUsers = results[0].ActiveUsers[0].Count
	}
	if len(results[0].InactiveUsers) > 0 {
		stats.InactiveUsers = results[0].InactiveUsers[0].Count
	}
	if len(results[0].NewUsers) > 0 {
		stats.NewUsersLastNDays = results[0].NewUsers[0].Count
	}

	return stats, nil
}

type UserStatistics struct {
	TotalUsers        int64 `json:"total_users"`
	ActiveUsers       int64 `json:"active_users"`
	InactiveUsers     int64 `json:"inactive_users"`
	NewUsersLastNDays int64 `json:"new_users_last_n_days"`
}

func (r *UserRepo) UpdateWallet(ctx context.Context, userID string, newBalance float64) error {
	update := bson.M{
		"$set": bson.M{
			"walletBalance": newBalance,
		},
	}

	result, err := r.col.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) FindByFilter(ctx context.Context, filter bson.M) ([]domain.User, error) {
	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepo) FindByInternalID(
	ctx context.Context,
	internalID int64,
) (*domain.User, error) {

	var user domain.User
	err := r.col.FindOne(ctx, bson.M{
		"internalId": internalID,
	}).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}


func (r *UserRepo) SearchByName(
	ctx context.Context,
	name string,
) ([]domain.User, error) {

	cursor, err := r.col.Find(ctx, bson.M{
		"name": bson.M{
			"$regex":   name,
			"$options": "i",
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepo) FindByIDs(ctx context.Context, userIDs []string) ([]*domain.User, error) {
	if len(userIDs) == 0 {
		return []*domain.User{}, nil
	}

	objectIDs := make([]primitive.ObjectID, 0, len(userIDs))
	for _, id := range userIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, objID)
	}

	if len(objectIDs) == 0 {
		return []*domain.User{}, nil
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}

	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find users by IDs: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}