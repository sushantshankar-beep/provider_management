package repository

import (
	"fmt"
	"time"
	"strconv"
	"strings"
	"context"
	"provider_management/internal/dto"
	"provider_management/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (r *TransactionRepo) FindWithFilter(
	ctx context.Context,
	filters dto.TransactionFilters,
	pagination dto.PaginationParams,
) ([]domain.Transaction, int64, error) {

	andFilters := bson.A{
		bson.M{
			"$or": bson.A{
				bson.M{"AMCPurchaseId": bson.M{"$exists": false}},
				bson.M{"AMCPurchaseId": primitive.NilObjectID},
			},
		},
	}

	if filters.Status != "" {
		andFilters = append(andFilters, bson.M{"status": filters.Status})
	}

	if filters.Method != "" {
		andFilters = append(andFilters, bson.M{"method": filters.Method})
	}

	if filters.CreatedAt != "" {
		const layout = "2006-01-02"
		if date, err := time.Parse(layout, filters.CreatedAt); err == nil {
			start := date
			end := date.Add(24 * time.Hour)
			andFilters = append(andFilters, bson.M{
				"createdAt": bson.M{
					"$gte": start,
					"$lt":  end,
				},
			})
		}
	}

	if filters.Search != "" {
		searchUpper := strings.ToUpper(filters.Search)

		or := bson.A{
			bson.M{"txnid": bson.M{"$regex": filters.Search, "$options": "i"}},
			bson.M{"status": bson.M{"$regex": filters.Search, "$options": "i"}},
			bson.M{"paymentSource": bson.M{"$regex": filters.Search, "$options": "i"}},
		}

		if strings.HasPrefix(searchUpper, "VW") {
			idStr := strings.TrimPrefix(searchUpper, "VW")
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				var user domain.User
				err := r.col.FindOne(ctx, bson.M{"id": id}).Decode(&user)
				if err == nil {
					or = append(or, bson.M{"userId": user.ID})
				}
			}
		}

		if strings.HasPrefix(searchUpper, "BK") {
			idStr := strings.TrimPrefix(searchUpper, "BK")
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				var service domain.AcceptedService
				err := r.col.FindOne(ctx, bson.M{"id": id}).Decode(&service)
				if err == nil {
					or = append(or, bson.M{"serviceId": service.ID})
				}
			}
		}

		andFilters = append(andFilters, bson.M{"$or": or})
	}

	filter := bson.M{"$and": andFilters}

	opts := options.Find().
		SetSkip(pagination.Skip).
		SetLimit(pagination.Limit).
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

func (r *TransactionRepo) FindByServiceID(ctx context.Context, serviceID primitive.ObjectID) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.col.FindOne(ctx, bson.M{"serviceId": serviceID}).Decode(&transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("transaction not found for service ID: %s", serviceID.Hex())
		}
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}
	return &transaction, nil
}

func (r *TransactionRepo) FindByTxnID(ctx context.Context, txnID string) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.col.FindOne(ctx, bson.M{"txnid": txnID}).Decode(&transaction)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepo) UpdateRefundID(ctx context.Context, transactionID primitive.ObjectID, refundID primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"refundId":  refundID,
			"updatedAt": time.Now(),
		},
	}
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": transactionID}, update)
	return err
}

func (r *TransactionRepo) GetAMCRevenueStats(ctx context.Context, period string) (dto.RevenueStats, error) {
	startDate := utils.GetStartDateForPeriod(period)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"createdAt": bson.M{"$gte": startDate},
				"status":    "paid",
				"AMCPurchaseId": bson.M{
					"$exists": true,
					"$ne":     primitive.NilObjectID,
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "amcpurchaseschemas",
				"localField":   "AMCPurchaseId",
				"foreignField": "_id",
				"as":           "amcPurchase",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$amcPurchase",
				"preserveNullAndEmptyArrays": false,
			},
		},
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"amcPurchase.refundStatus": ""},
					{"amcPurchase.refundStatus": bson.M{"$exists": false}},
					{"amcPurchase.refundStatus": bson.M{
						"$nin": []string{"approved", "completed", "refunded"},
					}},
				},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": "%Y-%m-%d",
						"date":   "$createdAt",
					},
				},
				"totalAmount": bson.M{"$sum": "$amount"},
				"count":       bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return dto.RevenueStats{}, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Day    string  `bson:"_id"`
		Amount float64 `bson:"totalAmount"`
		Count  int64   `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return dto.RevenueStats{}, err
	}
	
	dataPoints := make([]dto.RevenueDataPoint, len(results))
	totalAmount := 0.0
	for i, r := range results {
		dataPoints[i] = dto.RevenueDataPoint{
			Day:    r.Day,
			Amount: r.Amount,
		}
		totalAmount += r.Amount
	}

	return dto.RevenueStats{
		Period:      period,
		Data:        dataPoints,
		TotalAmount: totalAmount,
	}, nil
}

func (r *TransactionRepo) GetTransactionStats(
	ctx context.Context,
	days string,
) (dto.TransactionStats, error) {

	startDay := utils.GetStartDateFromDays(days)

	now := time.Now().UTC()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"createdAt": bson.M{"$gte": startDay, "$lte": now,},
				"serviceId": bson.M{"$exists": true, "$ne": nil},
				"status":    "paid",
			},
		},
		{
			"$group": bson.M{
				"_id":         nil,
				"totalAmount": bson.M{"$sum": "$amount"},
				"count":       bson.M{"$sum": 1},
				"totalWithGST": bson.M{
					"$sum": bson.M{
						"$multiply": bson.A{"$amount", 1.18},
					},
				},
			},
		},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return dto.TransactionStats{}, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		TotalAmount  float64 `bson:"totalAmount"`
		TotalWithGST float64 `bson:"totalWithGST"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return dto.TransactionStats{}, err
	}

	var stats dto.TransactionStats
	if len(result) > 0 {
		stats.TotalAmount = utils.RoundTo2(result[0].TotalAmount)
		stats.GSTAmount = utils.RoundTo2(result[0].TotalWithGST - result[0].TotalAmount)
	}

	return stats, nil
}
