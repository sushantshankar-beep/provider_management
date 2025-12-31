package repository

import (
	"context"
	"time"
	"log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"provider_management/internal/domain"
)

type AMCTransactionRepo struct {
	col *mongo.Collection
}

func NewAMCTransactionRepo(db *mongo.Database) *AMCTransactionRepo {
	return &AMCTransactionRepo{col: db.Collection("transactions")}
}

func (r *AMCTransactionRepo) FindAMCTransactions(
	ctx context.Context,
	skip, limit int64,
	status, method, createdAt string,
) ([]domain.Transaction, int64, error) {

	filter := bson.M{
		"AMCPurchaseId": bson.M{"$exists": true, "$ne": nil},
		"status":        bson.M{"$in": bson.A{"paid", "failed"}},
	}

	if status != "" {
		filter["status"] = status
	}

	if method != "" {
		filter["method"] = method
	}

	if createdAt != "" {
		var parsedDate time.Time
		var err error
		
		parsedDate, err = time.Parse("2006/01/02", createdAt)
		if err != nil {
			parsedDate, err = time.Parse("02/01/2006", createdAt)
		}
		
		if err == nil {
			startOfDay := time.Date(
				parsedDate.Year(),
				parsedDate.Month(),
				parsedDate.Day(),
				0, 0, 0, 0,
				time.UTC,
			)
			
			endOfDay := time.Date(
				parsedDate.Year(),
				parsedDate.Month(),
				parsedDate.Day(),
				23, 59, 59, 999999999,
				time.UTC,
			)
			
			filter["createdAt"] = bson.M{
				"$gte": startOfDay,
				"$lte": endOfDay,
			}
			log.Printf("Searching for date range: %s to %s\n", startOfDay.Format(time.RFC3339), endOfDay.Format(time.RFC3339))
		} else {
			log.Printf("Error parsing date %s: %v\n", createdAt, err)
		}
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var txns []domain.Transaction
	if err := cursor.All(ctx, &txns); err != nil {
		return nil, 0, err
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return txns, total, nil
}

func (r *AMCTransactionRepo) FindByInternalID(ctx context.Context, internalID int64) (*domain.Transaction, error) {
	var res domain.Transaction
	err := r.col.FindOne(ctx, bson.M{"id": internalID}).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
