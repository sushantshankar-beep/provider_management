package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"provider_management/internal/domain"
)

type AssessmentRepo struct {
	col *mongo.Collection
}

func NewAssessmentRepo(db *mongo.Database) *AssessmentRepo {
	return &AssessmentRepo{col: db.Collection("complaint_assessment")}
}

func (r *AssessmentRepo) Save(ctx context.Context, a *domain.Assessment) error {
	_, err := r.col.InsertOne(ctx, a)
	return err
}
