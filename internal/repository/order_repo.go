package repository

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"provider_management/internal/domain"
)

type OrderRepo struct {
	col *mongo.Collection
}

func NewOrderRepo(db *mongo.Database) *OrderRepo {
	return &OrderRepo{col: db.Collection("amcpurchaseschemas")}
}

func (r *OrderRepo) FindByID(ctx context.Context, id string) (*domain.AMCPurchase, error) {
	var order domain.AMCPurchase

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		filter := bson.M{"id": id}
		err = r.col.FindOne(ctx, filter).Decode(&order)
		if err != nil {
			return nil, err
		}
		return &order, nil
	}

	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepo) FindWithFilter(
	ctx context.Context,
	skip, limit int64,
	search, status, paymentStatus, startDate, endDate, sortBy, sortOrder string,
) ([]domain.AMCPurchase, int64, error) {

	filter := bson.M{
		"paymentStatus": bson.M{"$in": bson.A{"success", "failed"}},
	}

	if search != "" {
		userFilter := bson.M{
			"$or": bson.A{
				bson.M{"name": bson.M{"$regex": search, "$options": "i"}},
				bson.M{"email": bson.M{"$regex": search, "$options": "i"}},
				bson.M{"phone": bson.M{"$regex": search, "$options": "i"}},
			},
		}

		userRepo := NewUserRepo(r.col.Database())
		users, err := userRepo.FindByFilter(ctx, userFilter)
		if err == nil && len(users) > 0 {
			userIDs := make([]primitive.ObjectID, len(users))
			for i, u := range users {
				if objID, err := primitive.ObjectIDFromHex(u.ID); err == nil {
					userIDs[i] = objID
				}
			}

			searchConditions := bson.A{
				bson.M{"user": bson.M{"$in": userIDs}},
				bson.M{"payuTransactionId": bson.M{"$regex": search, "$options": "i"}},
				bson.M{"paymentId": bson.M{"$regex": search, "$options": "i"}},
			}

			if objID, err := primitive.ObjectIDFromHex(search); err == nil {
				searchConditions = append(searchConditions, bson.M{"_id": objID})
			}

			filter = bson.M{
				"$and": bson.A{
					bson.M{"paymentStatus": bson.M{"$in": bson.A{"success", "failed"}}},
					bson.M{"$or": searchConditions},
				},
			}
		} else {
			searchConditions := bson.A{
				bson.M{"payuTransactionId": bson.M{"$regex": search, "$options": "i"}},
				bson.M{"paymentId": bson.M{"$regex": search, "$options": "i"}},
			}

			if objID, err := primitive.ObjectIDFromHex(search); err == nil {
				searchConditions = append(searchConditions, bson.M{"_id": objID})
			}

			filter = bson.M{
				"$and": bson.A{
					bson.M{"paymentStatus": bson.M{"$in": bson.A{"success", "failed"}}},
					bson.M{"$or": searchConditions},
				},
			}
		}
	}

	if status != "" {
		filter["planStatus"] = status
	}


	if paymentStatus != "" && (paymentStatus == "success" || paymentStatus == "failed") {
		filter["paymentStatus"] = paymentStatus
	}

	if startDate != "" || endDate != "" {
		dateFilter := bson.M{}
		if startDate != "" {
			if t, err := time.Parse(time.RFC3339, startDate); err == nil {
				dateFilter["$gte"] = t
			}
		}
		if endDate != "" {
			if t, err := time.Parse(time.RFC3339, endDate); err == nil {
				dateFilter["$lte"] = t
			}
		}
		if len(dateFilter) > 0 {
			filter["createdAt"] = dateFilter
		}
	}

	sortDirection := -1
	if strings.ToLower(sortOrder) == "asc" {
		sortDirection = 1
	}
	sortField := sortBy
	if sortField == "" {
		sortField = "updatedAt"
	}

	opts := options.Find().
		SetSort(bson.M{sortField: sortDirection})
	if limit > 0 {
		opts.SetSkip(skip).SetLimit(limit)
	}

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []domain.AMCPurchase
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, 0, err
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *OrderRepo) UpdateStatus(
	ctx context.Context,
	id, planStatus, paymentStatus string,
) (*domain.AMCPurchase, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update := bson.M{}
	if planStatus != "" {
		update["planStatus"] = planStatus
	}
	if paymentStatus != "" {
		update["paymentStatus"] = paymentStatus
	}

	if len(update) == 0 {
		return r.FindByID(ctx, id)
	}

	update["updatedAt"] = time.Now()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var order domain.AMCPurchase

	err = r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": update},
		opts,
	).Decode(&order)

	if err != nil {
		return nil, err
	}

	return &order, nil
}