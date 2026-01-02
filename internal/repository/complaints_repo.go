package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"strconv"
	"strings"
	"time"
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

func (r *ComplaintRepository) List(ctx context.Context, filter domain.ComplaintFilter) ([]*domain.Complaint, int64, *domain.ComplaintStats, error) {
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
		userObjID, err := primitive.ObjectIDFromHex(*filter.UserID)
		if err == nil {
			query["userId"] = userObjID
		}
	}

	if filter.ProviderID != nil {
		providerObjID, err := primitive.ObjectIDFromHex(*filter.ProviderID)
		if err == nil {
			query["providerId"] = providerObjID
		}
	}

	if filter.DateFrom != nil || filter.DateTo != nil {
		dateQuery := bson.M{}
		if filter.DateFrom != nil {
			dateQuery["$gte"] = *filter.DateFrom
		}
		if filter.DateTo != nil {
			dateQuery["$lte"] = *filter.DateTo
		}
		query["createdAt"] = dateQuery
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

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to count complaints: %w", err)
	}

	statsQuery := bson.M{}
	if filter.UserID != nil {
		userObjID, err := primitive.ObjectIDFromHex(*filter.UserID)
		if err == nil {
			statsQuery["userId"] = userObjID
		}
	}
	if filter.ProviderID != nil {
		providerObjID, err := primitive.ObjectIDFromHex(*filter.ProviderID)
		if err == nil {
			statsQuery["providerId"] = providerObjID
		}
	}

	stats := &domain.ComplaintStats{}
	
	pipeline := []bson.M{
		{"$match": statsQuery},
		{"$facet": bson.M{
			"total": []bson.M{
				{"$count": "count"},
			},
			"resolved": []bson.M{
				{"$match": bson.M{"status": "resolved"}},
				{"$count": "count"},
			},
			"unresolved": []bson.M{
				{"$match": bson.M{"status": bson.M{"$ne": "resolved"}}},
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
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to aggregate stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	if len(results) > 0 {
		result := results[0]
		if totalArr, ok := result["total"].(bson.A); ok && len(totalArr) > 0 {
			if totalDoc, ok := totalArr[0].(bson.M); ok {
				if count, ok := totalDoc["count"].(int32); ok {
					stats.TotalComplaints = int64(count)
				}
			}
		}
		if resolvedArr, ok := result["resolved"].(bson.A); ok && len(resolvedArr) > 0 {
			if resolvedDoc, ok := resolvedArr[0].(bson.M); ok {
				if count, ok := resolvedDoc["count"].(int32); ok {
					stats.StatusResolved = int64(count)
				}
			}
		}
		if unresolvedArr, ok := result["unresolved"].(bson.A); ok && len(unresolvedArr) > 0 {
			if unresolvedDoc, ok := unresolvedArr[0].(bson.M); ok {
				if count, ok := unresolvedDoc["count"].(int32); ok {
					stats.StatusUnresolved = int64(count)
				}
			}
		}
		if userArr, ok := result["raisedByUser"].(bson.A); ok && len(userArr) > 0 {
			if userDoc, ok := userArr[0].(bson.M); ok {
				if count, ok := userDoc["count"].(int32); ok {
					stats.RaisedByYou = int64(count)
				}
			}
		}
		if providerArr, ok := result["raisedByProvider"].(bson.A); ok && len(providerArr) > 0 {
			if providerDoc, ok := providerArr[0].(bson.M); ok {
				if count, ok := providerDoc["count"].(int32); ok {
					stats.RaisedByProviders = int64(count)
				}
			}
		}
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 10
	}
	skip := (page - 1) * limit

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err = r.collection.Find(ctx, query, opts)
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
func (r *ComplaintRepository) Update(ctx context.Context, id string, update interface{}) error {

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid ObjectID format: %w", err)
	}

	updateDoc := bson.M{
		"$set": update,
	}
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

func (r *ComplaintRepository) UpdateStatus(ctx context.Context, id string, status string) error {

	now := time.Now()
	update := bson.M{
		"status":    status,
		"updatedAt": now,
	}

	switch status {
	case "in_review":
		update["timeline.inReview"] = now

	case "resolved":
		update["timeline.resolved"] = now

	}

	err := r.Update(ctx, id, update)
	if err != nil {
		log.Printf("UpdateStatus - Failed to update: %v", err)
	} else {
		log.Printf("UpdateStatus - Successfully updated status")
	}
	return err
}

func (r *ComplaintRepository) AddNote(
	ctx context.Context,
	internalID int64,
	note domain.ComplaintNote,
) error {

	log.Println("addNote internalID:", internalID)

	update := bson.M{
		"$push": bson.M{
			"notes": note,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"id": internalID},
		update,
	)

	log.Println("Result:", result)

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

	err := r.Update(ctx, id, update)
	if err != nil {
		log.Printf("SaveAssessment - Failed: %v", err)
	} else {
		log.Printf("SaveAssessment - Successfully saved assessment")
	}
	return err
}

func (r *ComplaintRepository) GetStats(ctx context.Context) (*domain.ComplaintStats, error) {
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

	stats := &domain.ComplaintStats{}
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