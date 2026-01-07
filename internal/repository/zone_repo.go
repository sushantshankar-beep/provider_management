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

type ZoneRepo struct {
	col *mongo.Collection
}

func NewZoneRepo(db *mongo.Database) *ZoneRepo {
	return &ZoneRepo{col: db.Collection("zones")}
}

func (r *ZoneRepo) FindByID(ctx context.Context, id string) (*domain.Zone, error) {
	var zone domain.Zone

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&zone)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

func (r *ZoneRepo) FindAll(ctx context.Context) ([]domain.Zone, error) {
	cursor, err := r.col.Find(ctx, bson.M{"isActive": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var zones []domain.Zone
	if err := cursor.All(ctx, &zones); err != nil {
		return nil, err
	}
	return zones, nil
}

func (r *ZoneRepo) Create(ctx context.Context, zone *domain.Zone) error {
	result, err := r.col.InsertOne(ctx, zone)
	if err != nil {
		return err
	}
	zone.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ZoneRepo) FindByName(ctx context.Context, name string) (*domain.Zone, error) {
	var zone domain.Zone
	err := r.col.FindOne(ctx, bson.M{"zoneName": name}).Decode(&zone)
	if err != nil {
		return nil, err
	}

	return &zone, nil
}

func (r *ZoneRepo) FindWithFilter(ctx context.Context, skip, limit int64, search string, isActive *bool, state string,
	createdAt *time.Time,
	updatedAt *time.Time) ([]domain.Zone, int64, error) {
	filter := bson.M{}

	if search != "" {
		filter["$or"] = bson.A{
			bson.M{"zoneName": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"stateName": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	if isActive != nil {
		filter["isActive"] = *isActive
	}

	if state != "" {
		filter["stateName"] = state
	}

	if createdAt != nil {
		start := time.Date(
			createdAt.Year(),
			createdAt.Month(),
			createdAt.Day(),
			0, 0, 0, 0,
			time.UTC,
		)
		end := start.Add(24 * time.Hour)

		filter["createdAt"] = bson.M{
			"$gte": start,
			"$lt":  end,
		}
	}

	if updatedAt != nil {
		start := time.Date(
			updatedAt.Year(),
			updatedAt.Month(),
			updatedAt.Day(),
			0, 0, 0, 0,
			time.UTC,
		)
		end := start.Add(24 * time.Hour)

		filter["updatedAt"] = bson.M{
			"$gte": start,
			"$lt":  end,
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

	var zones []domain.Zone
	if err := cursor.All(ctx, &zones); err != nil {
		return nil, 0, err
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return zones, total, nil
}

func (r *ZoneRepo) FindActive(ctx context.Context, state string) ([]domain.Zone, error) {
	filter := bson.M{"isActive": true}

	if state != "" {
		filter["stateName"] = state
	}

	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var zones []domain.Zone
	if err = cursor.All(ctx, &zones); err != nil {
		return nil, err
	}

	return zones, nil
}
func (r *ZoneRepo) CountActive(ctx context.Context, search string, isActive bool) (int64, error) {
	filter := bson.M{"isActive": isActive}

	if search != "" {
		filter["$or"] = bson.A{
			bson.M{"zoneName": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"stateName": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	return r.col.CountDocuments(ctx, filter)
}

func (r *ZoneRepo) Update(ctx context.Context, id string, zone *domain.Zone) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"zoneName":  zone.ZoneName,
			"stateName": zone.StateName,
			"isActive":  zone.IsActive,
			"updatedAt": zone.UpdatedAt,
		},
	}

	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *ZoneRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := r.col.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
func (r *ZoneRepo) FindActiveStates(ctx context.Context) ([]string, error) {
	states, err := r.col.Distinct(
		ctx,
		"stateName",
		bson.M{"isActive": true},
	)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, state := range states {
		if s, ok := state.(string); ok {
			result = append(result, s)
		}
	}

	return result, nil
}
