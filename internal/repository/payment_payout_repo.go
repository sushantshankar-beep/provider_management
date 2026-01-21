package repository

import (
	"fmt"
	"time"
	"context"
	"strings"
	"strconv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentPayoutRepo struct {
	col *mongo.Collection
}

func NewPaymentPayoutRepo(db *mongo.Database) *PaymentPayoutRepo {
	return &PaymentPayoutRepo{col: db.Collection("paymentPayouts")}
}

func (r *PaymentPayoutRepo) Create(ctx context.Context, payout *domain.PaymentPayout) error {
	_, err := r.col.InsertOne(ctx, payout)
	return err
}

func (r *PaymentPayoutRepo) GetProviderPayouts(ctx context.Context, filters dto.PayoutFilters, sort dto.PayoutSort, pagination dto.PaginationParams) ([]domain.PaymentPayout, int64, error) {

	filter := r.buildFilter(filters)
	sortField, sortOrder := r.buildSort(sort)

	skip := (pagination.Page - 1) * pagination.Limit

	opts := options.Find().
		SetSkip(skip).
		SetLimit(pagination.Limit).
		SetSort(bson.M{sortField: sortOrder})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var payouts []domain.PaymentPayout
	if err := cursor.All(ctx, &payouts); err != nil {
		return nil, 0, err
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return payouts, total, nil
}

func (r *PaymentPayoutRepo) buildFilter(filters dto.PayoutFilters) bson.M {
	filter := bson.M{}

	if filters.ProviderID != "" {
		if objID, err := primitive.ObjectIDFromHex(filters.ProviderID); err == nil {
			filter["providerId"] = objID
		}
	}

	if filters.Status != "" {
		switch strings.ToLower(filters.Status) {
		case "pending_settlement":
			filter["status"] = bson.M{
				"$in": []string{
					string(domain.PayoutStatusPending),
					string(domain.PayoutStatusPartiallySettled),
				},
			}
		default:
			filter["status"] = filters.Status
		}
	}

	if filters.PeriodFrom != "" && filters.PeriodTo != "" {
		from, err1 := time.Parse(time.RFC3339, filters.PeriodFrom)
		to, err2 := time.Parse(time.RFC3339, filters.PeriodTo)
		if err1 == nil && err2 == nil {
			filter["createdAt"] = bson.M{"$gte": from, "$lte": to}
		}
	}

	if filters.Search != "" {
		orFilters := []bson.M{}

		if objID, err := primitive.ObjectIDFromHex(filters.Search); err == nil {
			orFilters = append(orFilters, bson.M{"providerId": objID})
		}

		if len(filters.Search) > 3 && strings.ToUpper(filters.Search[:3]) == "PAY" {
			if payoutID, err := strconv.ParseInt(filters.Search[3:], 10, 64); err == nil {
				orFilters = append(orFilters, bson.M{"payoutId": payoutID})
			}
		}

		if payoutID, err := strconv.ParseInt(filters.Search, 10, 64); err == nil {
			orFilters = append(orFilters, bson.M{"payoutId": payoutID})
		}

		orFilters = append(orFilters, bson.M{"status": bson.M{"$regex": filters.Search, "$options": "i"}})

		if len(orFilters) > 0 {
			filter["$or"] = orFilters
		}
	}

	return filter
}

func (r *PaymentPayoutRepo) buildSort(sort dto.PayoutSort) (string, int) {
	order := -1
	if sort.SortOrder == "asc" {
		order = 1
	}

	if sort.SortBy == "" {
		return "createdAt", order
	}

	switch sort.SortBy {
	case "payout_id":
		return "payoutId", order
	case "provider_id":
		return "providerId", order
	case "base_amount":
		return "baseAmount", order
	case "net_payable":
		return "netPayable", order
	case "status":
		return "status", order
	case "created_at":
		return "createdAt", order
	case "updated_at":
		return "updatedAt", order
	default:
		return sort.SortBy, order
	}
}

func (r *PaymentPayoutRepo) FindByID( ctx context.Context, id primitive.ObjectID ) (*domain.PaymentPayout, error) {

	if r == nil || r.col == nil {
		return nil, fmt.Errorf("payment payout repository not initialized")
	}

	var payout domain.PaymentPayout

	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&payout)
	if err != nil {
		return nil, err
	}

	return &payout, nil
}

func (r *PaymentPayoutRepo) MarkSettled( ctx context.Context, payoutID primitive.ObjectID, settlementID primitive.ObjectID ) error {

	update := bson.M{
		"$set": bson.M{
			"status":       "settled",
			"settlementId": settlementID,
			"updatedAt":    time.Now(),
		},
	}

	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": payoutID},
		update,
	)
	return err
}

func (r *PaymentPayoutRepo) FindByPayoutID( ctx context.Context, payoutID int64 ) (*domain.PaymentPayout, error) {

	if r == nil || r.col == nil {
		return nil, fmt.Errorf("payment payout repository not initialized")
	}

	var payout domain.PaymentPayout
	err := r.col.FindOne(ctx, bson.M{"payoutId": payoutID}).Decode(&payout)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payout with ID %d not found", payoutID)
		}
		return nil, fmt.Errorf("failed to find payout: %v", err)
	}

	return &payout, nil
}

func (r *PaymentPayoutRepo) UpdateStatus( ctx context.Context, payoutID primitive.ObjectID, status domain.PaymentPayoutStatus,
settlementID *primitive.ObjectID,
) error {

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	if settlementID != nil {
		update["$set"].(bson.M)["settlementId"] = settlementID
	}

	_, err := r.col.UpdateOne(ctx, bson.M{"_id": payoutID}, update)
	return err
}

func (r *PaymentPayoutRepo) FindByInternalID(ctx context.Context, id primitive.ObjectID) (*domain.PaymentPayout, error) {
	var payout domain.PaymentPayout
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&payout)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &payout, nil
}

func (r *PaymentPayoutRepo) GetProviderDeductions( ctx context.Context, providerID primitive.ObjectID ) ([]domain.PaymentPayout, error) {
	
	filter := bson.M{
		"providerId":   providerID,
		"isDeduction":  true,                                             
		"status":       bson.M{"$in": []string{"pending", "complaint"}},
	}

	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var payouts []domain.PaymentPayout
	if err := cursor.All(ctx, &payouts); err != nil {
		return nil, err
	}

	return payouts, nil
}

func (r *PaymentPayoutRepo) GetPayoutsForSettlement( ctx context.Context, payoutIDs []primitive.ObjectID ) ([]domain.PaymentPayout, error) {
	filter := bson.M{
		"_id": bson.M{"$in": payoutIDs},
	}

	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var payouts []domain.PaymentPayout
	if err := cursor.All(ctx, &payouts); err != nil {
		return nil, err
	}

	return payouts, nil
}


func (r *PaymentPayoutRepo) UpdateMultipleStatus( ctx context.Context, payoutIDs []primitive.ObjectID, status domain.PaymentPayoutStatus, settlementID *primitive.ObjectID ) error {
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	if settlementID != nil {
		update["$set"].(bson.M)["settlementId"] = settlementID
	}

	_, err := r.col.UpdateMany(
		ctx,
		bson.M{"_id": bson.M{"$in": payoutIDs}},
		update,
	)
	return err
}

func (r *PaymentPayoutRepo) FindPendingByProvider( ctx context.Context, providerID primitive.ObjectID ) (*domain.PaymentPayout, error) {

	filter := bson.M{
		"providerId": providerID,
		"status": bson.M{
			"$in": []domain.PaymentPayoutStatus{
				domain.PayoutStatusPending,
				domain.PayoutStatusPartiallySettled,
			},
		},
	}
	
	var payout domain.PaymentPayout
	err := r.col.FindOne(ctx, filter).Decode(&payout)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &payout, nil
}

func (r *PaymentPayoutRepo) Update(ctx context.Context, payout *domain.PaymentPayout) error {
	_, err := r.col.UpdateByID(
		ctx,
		payout.ID,
		bson.M{"$set": payout},
	)
	return err
}

func (r *PaymentPayoutRepo) FindLatestByProvider(ctx context.Context, providerID primitive.ObjectID ) (*domain.PaymentPayout, error) {

	filter := bson.M{"providerId": providerID}
	opts := options.FindOne().
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	var payout domain.PaymentPayout
	err := r.col.FindOne(ctx, filter, opts).Decode(&payout)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &payout, nil
}


func (r *PaymentPayoutRepo) FindPendingByProviderAndService(ctx context.Context, providerID, serviceID primitive.ObjectID) (*domain.PaymentPayout, error) {
	filter := bson.M{
		"providerId": providerID,
		"serviceIds": serviceID,
		"status":     domain.PayoutStatusPending,
	}

	var payout domain.PaymentPayout
	err := r.col.FindOne(ctx, filter).Decode(&payout)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &payout, nil
}

func (r *PaymentPayoutRepo) UpdateFields(ctx context.Context, payoutID string, update bson.M) error {
	objID, err := primitive.ObjectIDFromHex(payoutID)
	if err != nil {
		return err
	}

	_, err = r.col.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		update,
	)

	return err
}
