package repository

import (
	"context"
	"fmt"

	"provider_management/internal/domain"

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
	var provider domain.Provider

	objID, err := primitive.ObjectIDFromHex(id)
	if err == nil {
		err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&provider)
		if err == nil {
			return &provider, nil
		}
	}

	err = r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&provider)
	if err != nil {
		var internalID int64
		if _, err := fmt.Sscanf(id, "%d", &internalID); err == nil {
			err = r.col.FindOne(ctx, bson.M{"id": internalID}).Decode(&provider)
			if err == nil {
				return &provider, nil
			}
		}
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
	id string,
	documentType string,
	documentID string,
	action string,
) (*domain.Provider, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	updateField := ""
	var update bson.M

	if documentType == "identityProof" {
		updateField = "identityProof.$.verified"
		update = bson.M{
			"$set": bson.M{updateField: action},
		}
	} else if documentType == "addressProof" {
		updateField = "addressProof.$.verified"
		update = bson.M{
			"$set": bson.M{updateField: action},
		}
	} else if documentType == "cancelCheque" {
		update = bson.M{
			"$set": bson.M{"cancelCheque.verified": action},
		}
	} else {
		return nil, fmt.Errorf("invalid document type")
	}

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