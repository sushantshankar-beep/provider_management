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
	log.Printf("GetByID - Looking up complaint by MongoDB _id: %s", id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("GetByID - Invalid ObjectID format: %v", err)
		return nil, fmt.Errorf("invalid ObjectID format: %w", err)
	}

	var complaint domain.Complaint
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&complaint)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("GetByID - Complaint not found with _id: %s", id)
			return nil, fmt.Errorf("complaint not found")
		}
		log.Printf("GetByID - Database error: %v", err)
		return nil, fmt.Errorf("failed to get complaint: %w", err)
	}

	log.Printf("GetByID - Found complaint: _id=%s, internal_id=%d", complaint.ID, complaint.InternalID)
	return &complaint, nil
}

func (r *ComplaintRepository) GetByInternalID(ctx context.Context, internalID int64) (*domain.Complaint, error) {
	log.Printf("GetByInternalID - Looking up complaint by internal ID: %d", internalID)

	var complaint domain.Complaint
	err := r.collection.FindOne(ctx, bson.M{"id": internalID}).Decode(&complaint)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("GetByInternalID - Complaint not found with internal ID: %d", internalID)
			return nil, fmt.Errorf("complaint not found")
		}
		log.Printf("GetByInternalID - Database error: %v", err)
		return nil, fmt.Errorf("failed to get complaint: %w", err)
	}

	log.Printf("GetByInternalID - Found complaint: _id=%s, internal_id=%d", complaint.ID, complaint.InternalID)
	return &complaint, nil
}

func (r *ComplaintRepository) List(ctx context.Context, filter domain.ComplaintFilter) ([]*domain.Complaint, int64, error) {
	query := bson.M{}

	if filter.Status != nil {
		query["status"] = *filter.Status
	}

	if filter.RaisedBy != nil {
		query["raisedBy"] = *filter.RaisedBy
	}

	if filter.Category != nil {
		query["category"] = *filter.Category
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
		query["$or"] = []bson.M{
			{"problem": bson.M{"$regex": *filter.SearchQuery, "$options": "i"}},
			{"bookingNumber": bson.M{"$regex": *filter.SearchQuery, "$options": "i"}},
			{"userName": bson.M{"$regex": *filter.SearchQuery, "$options": "i"}},
			{"providerName": bson.M{"$regex": *filter.SearchQuery, "$options": "i"}},
		}
	}

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count complaints: %w", err)
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

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list complaints: %w", err)
	}
	defer cursor.Close(ctx)

	var complaints []*domain.Complaint
	if err := cursor.All(ctx, &complaints); err != nil {
		return nil, 0, fmt.Errorf("failed to decode complaints: %w", err)
	}

	return complaints, total, nil
}

func (r *ComplaintRepository) Update(ctx context.Context, id string, update interface{}) error {

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Update - Invalid ObjectID format: %v", err)
		return fmt.Errorf("invalid ObjectID format: %w", err)
	}

	updateDoc := bson.M{
		"$set": update,
	}
	updateDoc["$set"].(bson.M)["updatedAt"] = time.Now()

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateDoc)
	if err != nil {
		log.Printf("Update - Database error: %v", err)
		return fmt.Errorf("failed to update complaint: %w", err)
	}

	if result.MatchedCount == 0 {
		log.Printf("Update - Complaint not found with _id: %s", id)
		return fmt.Errorf("complaint not found")
	}

	log.Printf("Update - Successfully updated complaint. Matched: %d, Modified: %d",
		result.MatchedCount, result.ModifiedCount)
	return nil
}

func (r *ComplaintRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	log.Printf("UpdateStatus - Updating status to '%s' for complaint _id: %s", status, id)

	now := time.Now()
	update := bson.M{
		"status":    status,
		"updatedAt": now,
	}

	switch status {
	case "in_review":
		update["timeline.inReview"] = now
		log.Printf("UpdateStatus - Setting timeline.inReview")
	case "resolved":
		update["timeline.resolved"] = now
		log.Printf("UpdateStatus - Setting timeline.resolved")
	}

	err := r.Update(ctx, id, update)
	if err != nil {
		log.Printf("UpdateStatus - Failed to update: %v", err)
	} else {
		log.Printf("UpdateStatus - Successfully updated status")
	}
	return err
}

func (r *ComplaintRepository) AddNote(ctx context.Context, complaintID string, note domain.ComplaintNote) error {
	log.Printf("AddNote - Adding note to complaint _id: %s", complaintID)

	objectID, err := primitive.ObjectIDFromHex(complaintID)
	if err != nil {
		log.Printf("AddNote - Invalid ObjectID format: %v", err)
		return fmt.Errorf("invalid ObjectID format: %w", err)
	}

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
		bson.M{"_id": objectID},
		update,
	)

	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("complaint not found with ID: %s", complaintID)
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("failed to add note, complaint not modified")
	}

	return nil
}

func (r *ComplaintRepository) SaveAssessment(ctx context.Context, id string, assessment domain.ComplaintAssessment) error {
	log.Printf("SaveAssessment - Saving assessment for complaint _id: %s", id)

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
