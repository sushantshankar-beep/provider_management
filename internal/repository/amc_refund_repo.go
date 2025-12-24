package repository

import (
	"context"
	"provider_management/internal/domain"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AMCRefundRepo struct {
	col *mongo.Collection
}

func NewAMCRefundRepo(db *mongo.Database) *AMCRefundRepo {
	return &AMCRefundRepo{col: db.Collection("amcrefundrequests")}
}

func (r *AMCRefundRepo) FindByID(ctx context.Context, id string) (*domain.AMCRefundRequest, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var refund domain.AMCRefundRequest
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&refund)
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *AMCRefundRepo) FindWithFilter(
	ctx context.Context,
	skip, limit int64,
	search, status, sortBy, sortOrder string,
) ([]domain.AMCRefundRequest, int64, error) {
	filter := bson.M{}

	if status != "" && status != "all" {
		filter["status"] = status
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit)

	sortDirection := -1
	if sortOrder == "asc" {
		sortDirection = 1
	}
	opts.SetSort(bson.M{sortBy: sortDirection})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var refunds []domain.AMCRefundRequest
	if err := cursor.All(ctx, &refunds); err != nil {
		return nil, 0, err
	}

	if search != "" {
		filtered := []domain.AMCRefundRequest{}
		searchLower := strings.ToLower(search)

		for _, ref := range refunds {
			if strings.Contains(ref.ID.Hex(), searchLower) ||
				strings.Contains(strings.ToLower(ref.Reason), searchLower) {
				filtered = append(filtered, ref)
			}
		}
		refunds = filtered
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return refunds, total, nil
}

func (r *AMCRefundRepo) UpdateStatus(
	ctx context.Context,
	id, status, payuRequestID, refundTransactionID string,
	payuResponse map[string]interface{},
	adminID primitive.ObjectID,
	timeline domain.Timeline,
) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":              status,
			"payuRequestId":       payuRequestID,
			"refundTransactionId": refundTransactionID,
			"payuRefundResponse":  payuResponse,
			"processedBy":         adminID,
			"updatedAt":           time.Now(),
		},
		"$push": bson.M{
			"timeline": timeline,
		},
	}

	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *AMCRefundRepo) UpdateStatusRejected(
	ctx context.Context,
	id, status, rejectionReason string,
	adminID primitive.ObjectID,
	timeline domain.Timeline,
) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":          status,
			"rejectedAt":      time.Now(),
			"rejectionReason": rejectionReason,
			"processedBy":     adminID,
			"updatedAt":       time.Now(),
		},
		"$push": bson.M{
			"timeline": timeline,
		},
	}

	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *AMCRefundRepo) UpdateStatusApproved(
	ctx context.Context,
	id string,
	bankRefNum string,
	settlementID string,
	refundMode string,
	bankArn string,
	rawResponse interface{},
	timeline domain.Timeline,
) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now()

	update := bson.M{
		"$set": bson.M{
			"status":                        "approved",
			"approvedAt":                    now,
			"completedAt":                   now,
			"bankRefNum":                    bankRefNum,
			"settlementId":                  settlementID,
			"refundMode":                    refundMode,
			"bankArn":                       bankArn,
			"payuRefundStatusCheckResponse": rawResponse,
			"lastStatusCheckAt":             now,
			"updatedAt":                     now,
		},
		"$push": bson.M{
			"timeline": timeline,
		},
	}

	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *AMCRefundRepo) UpdateStatusRejectedPayU(
	ctx context.Context,
	id string,
	timeline domain.Timeline,
) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     "rejected_payu",
			"rejectedAt": time.Now(),
			"updatedAt":  time.Now(),
		},
		"$push": bson.M{
			"timeline": timeline,
		},
	}

	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *AMCRefundRepo) UpdateLastStatusCheck(
	ctx context.Context,
	id string,
	rawResponse map[string]interface{},
) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"payuRefundStatusCheckResponse": rawResponse,
			"lastStatusCheckAt":             time.Now(),
			"updatedAt":                     time.Now(),
		},
	}

	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

type CheckRefundStatusResult struct {
	RefundStatus string
	Amount       float64
	BankRefNum   string
	Mode         string
	SettlementID string
	BankArn      string
	ErrorMsg     string
	RawResponse  map[string]interface{}
}
