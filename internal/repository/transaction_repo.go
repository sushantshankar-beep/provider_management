package repository

import (
	"context"
    "strings"
	"strconv"
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

func (r *TransactionRepo) FindWithFilter(
	ctx context.Context,
	skip, limit int64,
	search string,
) ([]domain.Transaction, int64, error) {

	filter := bson.M{}

	if search != "" {
		or := bson.A{
			bson.M{"txnid": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"status": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"paymentSource": bson.M{"$regex": search, "$options": "i"}},
		}

		if strings.HasPrefix(strings.ToUpper(search), "VW") {
			if id, err := strconv.ParseInt(strings.TrimPrefix(search, "VW"), 10, 64); err == nil {
				or = append(or, bson.M{"userInternalId": id})
			}
		}

		if strings.HasPrefix(strings.ToUpper(search), "BK") {
			if id, err := strconv.ParseInt(strings.TrimPrefix(search, "BK"), 10, 64); err == nil {
				or = append(or, bson.M{"acceptedServiceInternalId": id})
			}
		}

		filter["$or"] = or
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