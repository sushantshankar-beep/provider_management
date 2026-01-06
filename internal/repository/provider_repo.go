package repository

import (
	"context"
	"fmt"
	"provider_management/internal/domain"
	"provider_management/internal/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProviderRepo struct {
	col *mongo.Collection
}

func NewProviderRepo(db *mongo.Database) *ProviderRepo {
	return &ProviderRepo{col: db.Collection("providerschemas")}
}

func (r *ProviderRepo) FindAll(
	ctx context.Context,
	query bson.M,
	skip, limit int64,
	sort string,
) ([]domain.Provider, int64, error) {
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

func (r *ProviderRepo) UpdateDocumentVerification(
	ctx context.Context,
	providerID string,
	documentType string,
	documentID string,
	status string,
) (*domain.Provider, error) {

	providerObjID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		return nil, fmt.Errorf("invalid provider id")
	}

	var docObjID primitive.ObjectID
	if documentType == "identityProof" || documentType == "addressProof" {
		docObjID, err = primitive.ObjectIDFromHex(documentID)
		if err != nil {
			return nil, fmt.Errorf("invalid document id")
		}
	}

	filter := bson.M{"_id": providerObjID}
	update := bson.M{}

	switch documentType {
	case "identityProof":
		filter["identityProof._id"] = docObjID
		update = bson.M{"$set": bson.M{"identityProof.$.verified": status}}
	case "addressProof":
		filter["addressProof._id"] = docObjID
		update = bson.M{"$set": bson.M{"addressProof.$.verified": status}}
	case "cancelCheque":
		update = bson.M{"$set": bson.M{"cancelCheque.verified": status}}
	default:
		return nil, fmt.Errorf("invalid document type")
	}

	var updatedProvider domain.Provider
	err = r.col.FindOneAndUpdate(
		ctx,
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %v", err)
	}

	return &updatedProvider, nil
}

func (r *ProviderRepo) UpdateAccountStatus(ctx context.Context, id string, status string) (*domain.Provider, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var provider domain.Provider
	err = r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"isActive": status}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&provider)

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
	err = r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"commissionPercentage": commission}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&provider)

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
					{"$match": bson.M{"isActive": bson.M{"$ne": "active"}}},
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
	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": providerID},
		bson.M{"$push": bson.M{"notes": note}},
	)
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

func (r *ProviderRepo) AggregateZoneStats(
	ctx context.Context,
	pipeline []bson.M,
) ([]domain.ZoneStats, error) {

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []domain.ZoneStats
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ProviderRepo) AggregateActivationTeam(
	ctx context.Context,
	pipeline []bson.M,
) ([]domain.ActivationTeamMember, error) {

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []domain.ActivationTeamMember
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