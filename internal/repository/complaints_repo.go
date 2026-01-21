package repository

import (
	"fmt"
	"time"
	"strconv"
	"strings"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"provider_management/internal/dto"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ComplaintRepository struct {
	collection *mongo.Collection
	counterCol *mongo.Collection
}

func NewComplaintRepository(db *mongo.Database) *ComplaintRepository {
	return &ComplaintRepository{
		collection: db.Collection("complaints"),
		counterCol: db.Collection("counters"),
	}
}

func (r *ComplaintRepository) Create(ctx context.Context, complaint *domain.Complaint) error {
	complaint.CreatedAt = time.Now()
	complaint.UpdatedAt = time.Now()

	if complaint.InternalID == 0 {
		internalID, err := r.GenerateInternalID(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate internal ID: %w", err)
		}
		complaint.InternalID = internalID
	}

	_, err := r.collection.InsertOne(ctx, complaint)
	if err != nil {
		return fmt.Errorf("failed to create complaint: %w", err)
	}

	return nil
}

func (r *ComplaintRepository) GetByID(ctx context.Context, id string) (*domain.Complaint, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ObjectID format: %w", err)
	}

	var complaint domain.Complaint
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&complaint)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("complaint not found")
		}

		return nil, fmt.Errorf("failed to get complaint: %w", err)
	}

	return &complaint, nil
}

func (r *ComplaintRepository) GetByInternalID(ctx context.Context, internalID int64) (*domain.Complaint, error) {

	var complaint domain.Complaint
	err := r.collection.FindOne(ctx, bson.M{"id": internalID}).Decode(&complaint)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			return nil, fmt.Errorf("complaint not found")
		}

		return nil, fmt.Errorf("failed to get complaint: %w", err)
	}

	return &complaint, nil
}

func (r *ComplaintRepository) ListComplaints(ctx context.Context, filter dto.ComplaintFilter) ([]*domain.Complaint, int64, *dto.ComplaintStats, error) {
	query := r.buildQuery(filter)

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to count complaints: %w", err)
	}

	stats, err := r.getFilteredStats(ctx, filter)
	if err != nil {
		return nil, 0, nil, err
	}

	skip := (filter.Page - 1) * filter.Limit
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(filter.Limit))

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to list complaints: %w", err)
	}
	defer cursor.Close(ctx)

	var complaints []*domain.Complaint
	if err := cursor.All(ctx, &complaints); err != nil {
		return nil, 0, nil, fmt.Errorf("failed to decode complaints: %w", err)
	}

	return complaints, total, stats, nil
}

func (r *ComplaintRepository) buildQuery(filter dto.ComplaintFilter) bson.M {
	query := bson.M{}

	if filter.Status != nil {
		query["status"] = *filter.Status
	}
	if filter.RaisedBy != nil {
		query["raisedBy"] = *filter.RaisedBy
	}
	if filter.Category != nil {
		query["problem"] = *filter.Category
	}
	if filter.UserID != nil {
		if userObjID, err := primitive.ObjectIDFromHex(*filter.UserID); err == nil {
			query["userId"] = userObjID
		}
	}
	if filter.ProviderID != nil {
		if providerObjID, err := primitive.ObjectIDFromHex(*filter.ProviderID); err == nil {
			query["providerId"] = providerObjID
		}
	}
	if filter.CreatedAtFrom != nil && filter.CreatedAtTo != nil {
		query["createdAt"] = bson.M{
			"$gte": *filter.CreatedAtFrom,
			"$lte": *filter.CreatedAtTo,
		}
	}

	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		search := strings.TrimSpace(*filter.SearchQuery)
		searchUpper := strings.ToUpper(search)

		orConditions := []bson.M{
			{"problem": bson.M{"$regex": search, "$options": "i"}},
			{"status": bson.M{"$regex": search, "$options": "i"}},
			{"raisedBy": bson.M{"$regex": search, "$options": "i"}},
		}

		if strings.HasPrefix(searchUpper, "CMP") {
			id := strings.TrimPrefix(searchUpper, "CMP")
			if num, err := strconv.ParseInt(id, 10, 64); err == nil {
				orConditions = append(orConditions, bson.M{"id": num})
			}
		} else if num, err := strconv.ParseInt(search, 10, 64); err == nil {
			orConditions = append(orConditions, bson.M{"id": num})
		}

		if strings.HasPrefix(searchUpper, "BK") {
			id := strings.TrimPrefix(searchUpper, "BK")
			if num, err := strconv.ParseInt(id, 10, 64); err == nil {
				orConditions = append(orConditions, bson.M{"acceptedServiceId": num})
			}
		}

		if len(query) > 0 {
			query = bson.M{"$and": []bson.M{query, {"$or": orConditions}}}
		} else {
			query["$or"] = orConditions
		}
	}

	return query
}

func (r *ComplaintRepository) Update(ctx context.Context, id string, update interface{}) error {

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return fmt.Errorf("invalid ObjectID format: %w", err)
	}

	updateDoc := bson.M{ "$set": update	}
	updateDoc["$set"].(bson.M)["updatedAt"] = time.Now()

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateDoc)

	if err != nil {
		return fmt.Errorf("failed to update complaint: %w", err)
	}

	if result.MatchedCount == 0 {
       return fmt.Errorf("complaint not found")
	}

	return nil
}

func (r *ComplaintRepository) getFilteredStats(ctx context.Context, filter dto.ComplaintFilter) (*dto.ComplaintStats, error) {
	statsQuery := bson.M{}

	if filter.UserID != nil {
		if userObjID, err := primitive.ObjectIDFromHex(*filter.UserID); err == nil {
			statsQuery["userId"] = userObjID
		}
	}
	if filter.ProviderID != nil {
		if providerObjID, err := primitive.ObjectIDFromHex(*filter.ProviderID); err == nil {
			statsQuery["providerId"] = providerObjID
		}
	}

	pipeline := []bson.M{
		{"$match": statsQuery},
		{
			"$facet": bson.M{
				"total": []bson.M{
					{"$count": "count"},
				},
				"resolved": []bson.M{
					{"$match": bson.M{"status": domain.ComplaintStatusResolved}},
					{"$count": "count"},
				},
				"unresolved": []bson.M{
					{"$match": bson.M{"status": domain.ComplaintStatusInReview}},
					{"$count": "count"},
				},
				"initiated": []bson.M{
					{"$match": bson.M{"status": domain.ComplaintStatusInitiated}},
					{"$count": "count"},
				},
				"raisedByUser": []bson.M{
					{"$match": bson.M{"raisedBy": "User"}},
					{"$count": "count"},
				},
				"raisedByProvider": []bson.M{
					{"$match": bson.M{"raisedBy": "Provider"}},
					{"$count": "count"},
				},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	stats := &dto.ComplaintStats{}
	if len(results) > 0 {
		result := results[0]
		stats.TotalComplaints = extractCount(result, "total")
		stats.StatusResolved = extractCount(result, "resolved")
		stats.StatusUnresolved = extractCount(result, "unresolved")
		stats.StatusInitiated = extractCount(result, "initiated")
		stats.RaisedByYou = extractCount(result, "raisedByUser")
		stats.RaisedByProviders = extractCount(result, "raisedByProvider")
	}

	return stats, nil
}

func extractCount(result bson.M, key string) int64 {
	if arr, ok := result[key].(bson.A); ok && len(arr) > 0 {
		if doc, ok := arr[0].(bson.M); ok {
			if count, ok := doc["count"].(int32); ok {
				return int64(count)
			}
		}
	}
	return 0
}

func (r *ComplaintRepository) UpdateStatus(ctx context.Context, id string, status domain.ComplaintStatus) error {
	now := time.Now()

	update := bson.M{
		"status":    status,
		"updatedAt": now,
	}

	switch status {
	case domain.ComplaintStatusInReview:
		update["timeline.inReview"] = now
	case domain.ComplaintStatusResolved:
		update["timeline.resolved"] = now

	}

	return r.Update(ctx, id, update)
}

func (r *ComplaintRepository) AddNote( ctx context.Context, internalID int64, note domain.ComplaintNote ) error {

	update := bson.M{
		"$push": bson.M{
			"notes": note,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne( ctx, bson.M{"id": internalID}, update )

	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("complaint not found with id: %d", internalID)
	}

	return nil
}

func (r *ComplaintRepository) SaveAssessment(ctx context.Context, id string, assessment domain.ComplaintAssessment) error {
	assessment.AssessedAt = time.Now()

	update := bson.M{
		"assessment": assessment,
		"updatedAt":  time.Now(),
	}

	return r.Update(ctx, id, update)
}

func (r *ComplaintRepository) GetStats(ctx context.Context) (*dto.ComplaintStats, error) {
	pipeline := []bson.M{
		{
			"$facet": bson.M{
				"total": []bson.M{
					{"$count": "count"},
				},
				"user": []bson.M{
					{"$match": bson.M{"raisedBy": "User"}},
					{"$count": "count"},
				},
				"provider": []bson.M{
					{"$match": bson.M{"raisedBy": "Provider"}},
					{"$count": "count"},
				},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer cursor.Close(ctx)

	var result []struct {
		Total    []struct{ Count int64 } `bson:"total"`
		User     []struct{ Count int64 } `bson:"user"`
		Provider []struct{ Count int64 } `bson:"provider"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	stats := &dto.ComplaintStats{}
	if len(result) > 0 {
		if len(result[0].Total) > 0 {
			stats.TotalComplaints = result[0].Total[0].Count
		}
		if len(result[0].User) > 0 {
			stats.UserComplaints = result[0].User[0].Count
		}
		if len(result[0].Provider) > 0 {
			stats.ProviderComplaints = result[0].Provider[0].Count
		}
	}

	return stats, nil
}

func (r *ComplaintRepository) GenerateInternalID(ctx context.Context) (int64, error) {

	filter := bson.M{"_id": "complaint_internal_id"}
	update := bson.M{"$inc": bson.M{"sequence_value": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		SequenceValue int64 `bson:"sequence_value"`
	}

	err := r.counterCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to generate internal ID: %w", err)
	}

	return result.SequenceValue, nil
}

func (r *ComplaintRepository) GetDashboardStats(ctx context.Context) (dto.ComplaintsStats, error) {
	pipeline := []bson.M{
		{
			"$facet": bson.M{
				"total": []bson.M{
					{"$count": "count"},
				},
				"resolved": []bson.M{
					{"$match": bson.M{"status": "resolved"}},
					{"$count": "count"},
				},
				"underReview": []bson.M{
					{"$match": bson.M{"status": "in_review"}},
					{"$count": "count"},
				},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return dto.ComplaintsStats{}, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		Total          []struct{ Count int64 } `bson:"total"`
		Resolved       []struct{ Count int64 } `bson:"resolved"`
		UnderReview    []struct{ Count int64 } `bson:"underReview"`
		UserRaised     []struct{ Count int64 } `bson:"userRaised"`
		ProviderRaised []struct{ Count int64 } `bson:"providerRaised"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return dto.ComplaintsStats{}, err
	}

	stats := dto.ComplaintsStats{}
	if len(result) > 0 {
		if len(result[0].Total) > 0 {
			stats.Total = result[0].Total[0].Count
		}
		if len(result[0].Resolved) > 0 {
			stats.Resolved = result[0].Resolved[0].Count
		}
		if len(result[0].UnderReview) > 0 {
			stats.UnderReview = result[0].UnderReview[0].Count
		}
	}

	return stats, nil
}