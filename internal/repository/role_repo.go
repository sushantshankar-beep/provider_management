package repository

import (
	"context"
	"time"

	"provider_management/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RoleRepository struct {
	collection *mongo.Collection
}

func NewRoleRepository(db *mongo.Database) *RoleRepository {
	return &RoleRepository{
		collection: db.Collection("roles"),
	}
}

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, role)
	return err
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&role)
	return &role, err
}

func (r *RoleRepository) List(ctx context.Context, filter bson.M) ([]domain.Role, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []domain.Role
	for cursor.Next(ctx) {
		var role domain.Role
		if err := cursor.Decode(&role); err == nil {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func (r *RoleRepository) UpdateStatus(
	ctx context.Context,
	id primitive.ObjectID,
	status domain.RoleStatus,
) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status":    status,
				"updatedAt": time.Now(),
			},
		},
	)
	return err
}
