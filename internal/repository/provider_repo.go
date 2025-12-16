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