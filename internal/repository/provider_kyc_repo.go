package repository

import (
	"fmt"
	"time"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)


type ProviderKYCRepository struct {
	collection *mongo.Collection
}

func NewProviderKYCRepo(db *mongo.Database) *ProviderKYCRepository {
	return &ProviderKYCRepository{
		collection: db.Collection("provider_kyc"),
	}
}

func (r *ProviderKYCRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.ProviderKYC, error) {
	var kyc domain.ProviderKYC
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&kyc)
	if err != nil {
		return nil, err
	}
	return &kyc, nil
}

func (r *ProviderKYCRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]domain.ProviderKYC, error) {
	if len(ids) == 0 {
		return []domain.ProviderKYC{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"_id": bson.M{"$in": ids},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var kycs []domain.ProviderKYC
	if err := cursor.All(ctx, &kycs); err != nil {
		return nil, err
	}
	return kycs, nil
}

func (r *ProviderKYCRepository) UpdateDocumentVerification( ctx context.Context,kycID string, documentID string, status domain.VerificationStatus, rejectedNote string ) (*domain.ProviderKYC, error) {

	kycObjID, err := primitive.ObjectIDFromHex(kycID)
	if err != nil {
		return nil, fmt.Errorf("invalid kyc id")
	}

	docObjID, err := primitive.ObjectIDFromHex(documentID)
	if err != nil {
		return nil, fmt.Errorf("invalid document id")
	}

	filter := bson.M{
		"_id":           kycObjID,
		"documents._id": docObjID,
	}

	update := bson.M{
		"$set": bson.M{
			"documents.$.verified": status,
			"documents.$.rejectionNote": rejectedNote,
			"updatedAt":            time.Now(),
		},
	}

	var updatedKYC domain.ProviderKYC
	err = r.collection.FindOneAndUpdate( ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedKYC)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("kyc document not found")
		}
		return nil, fmt.Errorf("failed to update document: %v", err)
	}

	return &updatedKYC, nil
}


func (r *ProviderKYCRepository) UpdateKYCStatus( ctx context.Context, kycID string, status domain.KYCStatus,
) (*domain.ProviderKYC, error) {

	objID, err := primitive.ObjectIDFromHex(kycID)
	if err != nil {
		return nil, fmt.Errorf("invalid kyc id")
	}

	updateFields := bson.M{
		"status":    status,
		"updatedAt": time.Now(),
	}

	if status == domain.KYC_APPROVED {
		updateFields["approvedAt"] = time.Now()
	}

	var kyc domain.ProviderKYC
	err = r.collection.FindOneAndUpdate( ctx, bson.M{"_id": objID}, bson.M{"$set": updateFields}, options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&kyc)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("kyc document not found")
		}
		return nil, fmt.Errorf("failed to update kyc status: %v", err)
	}

	return &kyc, nil
}

func (r *ProviderKYCRepository) Create(ctx context.Context,kyc *domain.ProviderKYC) error {
	_, err := r.collection.InsertOne(ctx, kyc)
	return err
}

func (r *ProviderKYCRepository) UpdateByProviderID(
	ctx context.Context,
	providerID primitive.ObjectID,
	updateData bson.M,
) error {

	update := bson.M{
		"$set": updateData,
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"providerId": providerID},
		update,
	)

	return err
}

func (r *ProviderKYCRepository) Count( ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

func (r *ProviderKYCRepository) Find(ctx context.Context,filter bson.M) ([]domain.ProviderKYC, error) {
    var results []domain.ProviderKYC
    cursor, err := r.collection.Find(context.Background(), filter)
    if err != nil {
        return nil, err
    }
    if err := cursor.All(context.Background(), &results); err != nil {
        return nil, err
    }
    return results, nil
}

func (r *ProviderKYCRepository) FindByProviderID(ctx context.Context, providerID primitive.ObjectID) (*domain.ProviderKYC, error) {
    var kyc domain.ProviderKYC
    err := r.collection.FindOne(ctx, bson.M{"providerId": providerID}).Decode(&kyc)
    if err != nil {
        return nil, err
    }
    return &kyc, nil
}

func (r *ProviderKYCRepository) FindGSTByProviderID(ctx context.Context, providerID string) (*domain.ProviderKYC, error) {
    var kyc domain.ProviderKYC
    err := r.collection.FindOne(ctx, bson.M{"providerId": providerID}).Decode(&kyc)
    if err != nil {
        return nil, err
    }
    return &kyc, nil
}
