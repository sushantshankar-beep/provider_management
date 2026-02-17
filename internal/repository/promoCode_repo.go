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

type PromoCodeRepo struct {
	collection *mongo.Collection
}

func NewPromoCodeRepo(db *mongo.Database) *PromoCodeRepo {
	return &PromoCodeRepo{collection: db.Collection("promoCodes")}
}

func (r *PromoCodeRepo) Create(ctx context.Context, promo *domain.PromoCode) error {
	now := time.Now()
	promo.ID = primitive.NewObjectID()
	promo.CreatedAt = now
	promo.UpdatedAt = now
	promo.UsageCount = 0
	promo.TotalDiscount = 0

	_, err := r.collection.InsertOne(ctx, promo)
	return err
}

func (r *PromoCodeRepo) GetPromoCodes(ctx context.Context, filter bson.M, skip, limit int64, sortBy string, sortOrder int) ([]domain.PromoCode, int64, error) {
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

	var promos []domain.PromoCode
	if err := cursor.All(ctx, &promos); err != nil {
		return nil, 0, err
	}

	return promos, total, nil
}

func (r *PromoCodeRepo) FindByID(ctx context.Context, id string) (*domain.PromoCode, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var promo domain.PromoCode
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&promo)
	if err != nil {
		return nil, err
	}

	return &promo, nil
}

func (r *PromoCodeRepo) FindByCode(ctx context.Context, code string) (*domain.PromoCode, error) {
	var promo domain.PromoCode
	err := r.collection.FindOne(ctx, bson.M{"code": code}).Decode(&promo)
	if err != nil {
		return nil, err
	}
	return &promo, nil
}

func (r *PromoCodeRepo) Update(ctx context.Context, id string, update bson.M) error {
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

func (r *PromoCodeRepo) UpdateStatus(ctx context.Context, id string, status string) error {
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

func (r *PromoCodeRepo) Delete(ctx context.Context, id string) error {
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

func (r *PromoCodeRepo) GetStats(ctx context.Context) (map[string]int64, error) {
	statuses := []domain.PromoStatus{
		domain.PromoStatusActive,
		domain.PromoStatusScheduled,
		domain.PromoStatusExpired,
		domain.PromoStatusDraft,
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