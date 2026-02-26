package repository

import (
	"context"
	"provider_management/internal/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DiscountRepo struct {
	collection *mongo.Collection
}

func NewDiscountRepo(db *mongo.Database) *DiscountRepo {
	return &DiscountRepo{collection: db.Collection("discounts")}
}

func (r *DiscountRepo) Create(ctx context.Context, d *domain.Discount) error {
	now := time.Now()
	d.ID = primitive.NewObjectID()
	d.CreatedAt = now
	d.UpdatedAt = now
	d.TotalSavings = 0
	d.TotalOrders = 0

	_, err := r.collection.InsertOne(ctx, d)
	return err
}

func (r *DiscountRepo) GetDiscounts(ctx context.Context, filter bson.M, skip, limit int64, sortBy string, sortOrder int) ([]domain.Discount, int64, error) {
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: sortBy, Value: sortOrder}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var discounts []domain.Discount
	if err := cursor.All(ctx, &discounts); err != nil {
		return nil, 0, err
	}

	return discounts, total, nil
}

func (r *DiscountRepo) FindByID(ctx context.Context, id string) (*domain.Discount, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var d domain.Discount
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (r *DiscountRepo) Update(ctx context.Context, id string, update bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update["updatedAt"] = time.Now()
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *DiscountRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"status": status, "updatedAt": time.Now()}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *DiscountRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *DiscountRepo) GetStats(ctx context.Context) (map[string]int64, error) {
	statuses := []domain.DiscountStatus{
		domain.DiscountStatusActive,
		domain.DiscountStatusScheduled,
		domain.DiscountStatusExpired,
		domain.DiscountStatusDraft,
		domain.DiscountStatusPaused,
	}

	result := map[string]int64{}

	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	result["total"] = total

	for _, s := range statuses {
		count, err := r.collection.CountDocuments(ctx, bson.M{"status": s})
		if err != nil {
			return nil, err
		}
		result[string(s)] = count
	}

	return result, nil
}

func (r *DiscountRepo) BulkUpdateExpired(ctx context.Context, now time.Time) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{
			"endAt":  bson.M{"$ne": nil, "$lt": now},
			"status": bson.M{"$nin": bson.A{domain.DiscountStatusExpired, domain.DiscountStatusDraft, domain.DiscountStatusPaused,domain.DiscountStatusInActive }},
		},
		bson.M{"$set": bson.M{"status": domain.DiscountStatusExpired, "updatedAt": now}},
	)
	return err
}

func (r *DiscountRepo) BulkActivateScheduled(ctx context.Context, now time.Time) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{
			"status":  domain.DiscountStatusScheduled,
			"startAt": bson.M{"$lte": now},
			"$or": bson.A{
				bson.M{"endAt": nil},
				bson.M{"endAt": bson.M{"$gt": now}},
			},
		},
		bson.M{"$set": bson.M{"status": domain.DiscountStatusActive, "updatedAt": now}},
	)
	return err
}

func (r *DiscountRepo) GetDiscountsForUsage(
	ctx context.Context,
	skip, limit int64,
	search, status string,
) ([]domain.Discount, int64, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}
	if search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"code": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"updatedAt": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var discounts []domain.Discount
	if err := cursor.All(ctx, &discounts); err != nil {
		return nil, 0, err
	}
	return discounts, total, nil
}