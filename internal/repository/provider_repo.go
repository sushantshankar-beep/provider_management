package repository

import (
	"fmt"
	"time"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/dto"
	"provider_management/internal/utils"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProviderRepo struct {
	col *mongo.Collection
}

func NewProviderRepo(db *mongo.Database) *ProviderRepo {
	return &ProviderRepo{col: db.Collection("providerschemas")}
}

func (r *ProviderRepo) FindAll( ctx context.Context, query bson.M, skip, limit int64, sort string ) ([]domain.Provider, int64, error) {
	sortOpts := bson.M{}
	if sort != "" {
		if sort[0] == '-' {
			sortOpts[sort[1:]] = -1
		} else {
			sortOpts[sort] = 1
		}
	} else {
		sortOpts["createdAt"] = -1
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(sortOpts)

	cur, err := r.col.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var providers []domain.Provider
	if err := cur.All(ctx, &providers); err != nil {
		return nil, 0, err
	}

	total, err := r.col.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	return providers, total, nil
}

func (r *ProviderRepo) FindByID(ctx context.Context, id string) (*domain.Provider, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid provider id")
	}

	var provider domain.Provider
	err = r.col.FindOne(ctx, bson.M{"_id": objectID}).Decode(&provider)
	if err != nil {
		return nil, err
	}

	return &provider, nil
}

func (r *ProviderRepo) FindOne(ctx context.Context, query bson.M) (*domain.Provider, error) {
	var provider domain.Provider
	if err := r.col.FindOne(ctx, query).Decode(&provider); err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ProviderRepo) UpdateStatus(ctx context.Context, id string, status string) (*domain.Provider, error) {

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var provider domain.Provider

	err = r.col.FindOneAndUpdate( ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"status": status}}, options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&provider)

	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ProviderRepo) UpdateKYCStatus(ctx context.Context, id string, status string) (*domain.Provider, error) {

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var provider domain.Provider
	err = r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"status": status}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&provider)

	if err != nil {
		return nil, err
	}
	return &provider, nil
}



func (r *ProviderRepo) UpdateAccountStatus(ctx context.Context, id string, status string) (*domain.Provider, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var provider domain.Provider
	err = r.col.FindOneAndUpdate( ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"isActive": status}}, options.FindOneAndUpdate().SetReturnDocument(options.After), ).Decode(&provider)

	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ProviderRepo) UpdateCommission(ctx context.Context, id string, commission float64) (*domain.Provider, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var provider domain.Provider
	err = r.col.FindOneAndUpdate(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"commissionPercentage": utils.RoundTo2(commission)}}, options.FindOneAndUpdate().SetReturnDocument(options.After), ).Decode(&provider)

	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ProviderRepo) Count(ctx context.Context, query bson.M) (int64, error) {
	return r.col.CountDocuments(ctx, query)
}

func (r *ProviderRepo) CountByStatus(ctx context.Context, query bson.M, statusField, statusValue string) (int64, error) {
	finalQuery := bson.M{}
	for k, v := range query {
		finalQuery[k] = v
	}
	finalQuery[statusField] = statusValue
	return r.col.CountDocuments(ctx, finalQuery)
}

func (r *ProviderRepo) GetProviderByObjectID(ctx context.Context, id primitive.ObjectID) (*domain.Provider, error) {
	var provider domain.Provider
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&provider)
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ProviderRepo) CountInactive(ctx context.Context, query bson.M) (int64, error) {

	finalQuery := bson.M{}
	for k, v := range query {
		finalQuery[k] = v
	}

	finalQuery["$and"] = []bson.M{
		query,
		{
			"isActive": bson.M{
				"$nin": []interface{}{
					domain.AccountStatusActive,
					true,
					domain.AccountStatusSuspended,
					domain.AccountStatusBlacklisted,
				},
			},
		},
	}

	count, err := r.col.CountDocuments(ctx, finalQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to count inactive providers: %v", err)
	}
	return count, nil
}

func (r *ProviderRepo) CountByMultipleStatuses(ctx context.Context, filter bson.M, field string, statuses []string) (int64, error) {
	countFilter := bson.M{}
	for k, v := range filter {
		countFilter[k] = v
	}

	countFilter[field] = bson.M{"$in": statuses}

	count, err := r.col.CountDocuments(ctx, countFilter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ProviderRepo) GetStats(ctx context.Context) (dto.ProvidersStats, error) {
	pipeline := []bson.M{
		{
			"$facet": bson.M{
				"total": []bson.M{
					{"$count": "count"},
				},
				"active": []bson.M{
					{"$match": bson.M{"isActive": "active"}},
					{"$count": "count"},
				},
				"inactive": []bson.M{
					{
						"$match": bson.M{
							"isActive": bson.M{
								"$in": []string{"suspended", "blacklisted"},
							},
						},
					},
					{"$count": "count"},
				},
			},
		},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return dto.ProvidersStats{}, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		Total    []struct{ Count int64 } `bson:"total"`
		Active   []struct{ Count int64 } `bson:"active"`
		Inactive []struct{ Count int64 } `bson:"inactive"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return dto.ProvidersStats{}, err
	}

	stats := dto.ProvidersStats{}
	if len(result) > 0 {
		if len(result[0].Total) > 0 {
			stats.Total = result[0].Total[0].Count
		}
		if len(result[0].Active) > 0 {
			stats.Active = result[0].Active[0].Count
		}
		if len(result[0].Inactive) > 0 {
			stats.Inactive = result[0].Inactive[0].Count
		}
	}

	return stats, nil
}

func (r *ProviderRepo) AddProviderNote(ctx context.Context, providerID primitive.ObjectID, note domain.ProviderNote) error {
	_, err := r.col.UpdateOne( ctx, bson.M{"_id": providerID}, bson.M{"$push": bson.M{"notes": note}})
	return err
}

func (r *ProviderRepo) Create(ctx context.Context, provider *domain.Provider) error {
	_, err := r.col.InsertOne(ctx, provider)
	return err
}

func (r *ProviderRepo) Update(ctx context.Context, id string, updateData bson.M) (*domain.Provider, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update := bson.M{"$set": updateData}

	var provider domain.Provider
	err = r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&provider)

	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ProviderRepo) AggregateZoneStats( ctx context.Context, pipeline []bson.M) ([]dto.ProviderZoneStats, error) {

	cursor, err := r.col.Aggregate(ctx, pipeline)
	
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []dto.ProviderZoneStats
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ProviderRepo) AggregateActivationTeam( ctx context.Context, pipeline []bson.M ) ([]dto.ProviderActivationTeamMember, error) {

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []dto.ProviderActivationTeamMember
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ProviderRepo) CountByCreator(ctx context.Context, adminID primitive.ObjectID) (int, error) {
	filter := bson.M{
		"createdBy": adminID,
	}

	count, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *ProviderRepo) CountByCreators(ctx context.Context, adminIDs []primitive.ObjectID) (int, error) {
	filter := bson.M{
		"createdBy": bson.M{"$in": adminIDs},
	}

	count, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *ProviderRepo) CountByCreatedBy(ctx context.Context, adminIDs []primitive.ObjectID) (int, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{
		"createdBy": bson.M{"$in": adminIDs},
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *ProviderRepo) FindByObjectIDs( ctx context.Context, ids []primitive.ObjectID ) ([]domain.Provider, error) {
	filter := bson.M{
		"_id": bson.M{"$in": ids},
	}
	
	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var providers []domain.Provider
	if err := cursor.All(ctx, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}

func (r *ProviderRepo) UpdateKYCID(
	ctx context.Context,
	providerID primitive.ObjectID,
	kycID primitive.ObjectID,
) error {

	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": providerID},
		bson.M{
			"$set": bson.M{
				"kycId":     kycID,
				"updatedAt": time.Now(),
			},
		},
	)

	return err
}

func (r *ProviderRepo) FindByPhone(ctx context.Context, phone string) (*domain.Provider, error) {
	var provider domain.Provider
	filter := bson.M{"phone": phone}
	err := r.col.FindOne(ctx, filter).Decode(&provider)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &provider, nil
}

func (r *ProviderRepo) FindByProviderCode(
	ctx context.Context,
	code string,
) (*domain.Provider, error) {

	var provider domain.Provider

	err := r.col.FindOne(
		ctx,
		bson.M{"providerCode": code},
	).Decode(&provider)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	return &provider, err
}
