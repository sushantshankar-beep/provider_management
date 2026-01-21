package repository

import (
	"time"
	"context"
	"strings"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
     "provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (r *OrderRepo) FindWithFilter(ctx context.Context, skip, limit int64, search, planStatus, paymentStatus, createdAt, sortBy, sortOrder string ) ([]domain.AMCPurchase, int64, error) {

	andConditions := bson.A{
		bson.M{"paymentStatus": bson.M{"$in": bson.A{"success", "failed"}}},
		bson.M{"planStatus": bson.M{"$in": bson.A{"pending", "cancelled"}}},
	}

	if search != "" {
		searchConditions := bson.A{
			bson.M{"payuTransactionId": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"paymentId": bson.M{"$regex": search, "$options": "i"}},
		}

		if objID, err := primitive.ObjectIDFromHex(search); err == nil {
			searchConditions = append(searchConditions, bson.M{"_id": objID})
		}

		andConditions = append(andConditions, bson.M{
			"$or": searchConditions,
		})
	}

	if planStatus != "" {
		andConditions = append(andConditions, bson.M{
			"planStatus": planStatus,
		})
	}

	if paymentStatus != "" {
		andConditions = append(andConditions, bson.M{
			"paymentStatus": paymentStatus,
		})
	}

	if createdAt != "" {
		day, err := time.Parse("2006-01-02", createdAt)
		if err == nil {
			start := day
			end := day.Add(24*time.Hour - time.Nanosecond)

			andConditions = append(andConditions, bson.M{
				"createdAt": bson.M{
					"$gte": start,
					"$lte": end,
				},
			})
		}
	}

	filter := bson.M{
		"$and": andConditions,
	}

	sortDirection := -1
	if strings.ToLower(sortOrder) == "asc" {
		sortDirection = 1
	}

	if sortBy == "" {
		sortBy = "createdAt"
	}

	opts := options.Find().SetSort(bson.M{sortBy: sortDirection})

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

func (r *OrderRepo) UpdateStatus( ctx context.Context, id, planStatus, paymentStatus string ) (*domain.AMCPurchase, error) {
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