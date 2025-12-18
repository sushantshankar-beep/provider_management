package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"provider_management/internal/domain"
)

type ComplaintRepository interface {
	Create(ctx context.Context, complaint *domain.Complaint) error
	GetByID(ctx context.Context, id string) (*domain.Complaint, error)
	GetByInternalID(ctx context.Context, internalID int64) (*domain.Complaint, error)
	List(ctx context.Context, filter domain.ComplaintFilter) ([]*domain.Complaint, int64, error)
	Update(ctx context.Context, id string, update interface{}) error
	UpdateStatus(ctx context.Context, id string, status string) error
	AddNote(ctx context.Context, id string, note domain.ComplaintNote) error
	SaveAssessment(ctx context.Context, id string, assessment domain.ComplaintAssessment) error
	GetStats(ctx context.Context) (*domain.ComplaintStats, error)
	GenerateInternalID(ctx context.Context) (int64, error)
}

type complaintRepository struct {
	collection *mongo.Collection
	counterCol *mongo.Collection
}

func NewComplaintRepository(db *mongo.Database) ComplaintRepository {
	return &complaintRepository{
		collection: db.Collection("complaints"),
		counterCol: db.Collection("counters"),
	}
}

func (r *complaintRepository) Create(ctx context.Context, complaint *domain.Complaint) error {
	complaint.CreatedAt = time.Now()
	complaint.UpdatedAt = time.Now()

	// Generate internal ID if not set
	if complaint.InternalID == 0 {
		internalID, err := r.GenerateInternalID(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate internal ID: %w", err)
		}
		complaint.InternalID = internalID
	}

	// If ID is not set, MongoDB will generate it
	_, err := r.collection.InsertOne(ctx, complaint)
	if err != nil {
		return fmt.Errorf("failed to create complaint: %w", err)
	}

	return nil
}

func (r *complaintRepository) GetByID(ctx context.Context, id string) (*domain.Complaint, error) {
	var complaint domain.Complaint

	// Try to find by _id (MongoDB ObjectID as string)
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&complaint)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("complaint not found")
		}
		return nil, fmt.Errorf("failed to get complaint: %w", err)
	}
	return &complaint, nil
}

func (r *complaintRepository) GetByInternalID(ctx context.Context, internalID int64) (*domain.Complaint, error) {
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

func (r *complaintRepository) List(ctx context.Context, filter domain.ComplaintFilter) ([]*domain.Complaint, int64, error) {
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

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count complaints: %w", err)
	}

	// Pagination
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

func (r *complaintRepository) Update(ctx context.Context, id string, update interface{}) error {
	updateDoc := bson.M{
		"$set": update,
	}
	updateDoc["$set"].(bson.M)["updatedAt"] = time.Now()

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update complaint: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("complaint not found")
	}

	return nil
}

func (r *complaintRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	now := time.Now()
	update := bson.M{
		"status":    status,
		"updatedAt": now,
	}

	// Update timeline based on status
	switch status {
	case "in_review":
		update["timeline.inReview"] = now
	case "resolved":
		update["timeline.resolved"] = now
	}

	return r.Update(ctx, id, update)
}

func (r *complaintRepository) AddNote(ctx context.Context, id string, note domain.ComplaintNote) error {
	note.CreatedAt = time.Now()

	update := bson.M{
		"$push": bson.M{"notes": note},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to add note: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("complaint not found")
	}

	return nil
}

func (r *complaintRepository) SaveAssessment(ctx context.Context, id string, assessment domain.ComplaintAssessment) error {
	assessment.AssessedAt = time.Now()

	update := bson.M{
		"assessment": assessment,
		"updatedAt":  time.Now(),
	}

	return r.Update(ctx, id, update)
}

func (r *complaintRepository) GetStats(ctx context.Context) (*domain.ComplaintStats, error) {
	pipeline := []bson.M{
		{
			"$facet": bson.M{
				"total": []bson.M{
					{"$count": "count"},
				},
				"user": []bson.M{
					{"$match": bson.M{"raisedBy": "user"}},
					{"$count": "count"},
				},
				"provider": []bson.M{
					{"$match": bson.M{"raisedBy": "provider"}},
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

func (r *complaintRepository) GenerateInternalID(ctx context.Context) (int64, error) {
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
