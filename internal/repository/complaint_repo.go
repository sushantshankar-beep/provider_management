package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"provider_management/internal/domain"
)

type ComplaintRepo struct {
	col *mongo.Collection
}

func NewComplaintRepo(db *mongo.Database) *ComplaintRepo {
	return &ComplaintRepo{col: db.Collection("complaints")}
}

func (r *ComplaintRepo) Save(ctx context.Context, c *domain.Complaint) error {
	_, err := r.col.InsertOne(ctx, c)
	return err
}

func (r *ComplaintRepo) FindAll(ctx context.Context, skip, limit int64) ([]domain.Complaint, error) {
	cur, err := r.col.Find(ctx, bson.M{}, nil)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []domain.Complaint
	for cur.Next(ctx) {
		var c domain.Complaint
		cur.Decode(&c)
		list = append(list, c)
	}
	return list, nil
}

func (r *ComplaintRepo) FindByID(ctx context.Context, id string) (*domain.Complaint, error) {
	var res domain.Complaint
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
