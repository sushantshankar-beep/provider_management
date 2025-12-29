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
func (r *RoleRepository) DeleteByID(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
func (r *RoleRepository) FindByID(ctx context.Context,id primitive.ObjectID) (*domain.Role, error) {

	var role domain.Role

	filter := bson.M{
		"$or": []bson.M{
			{"_id": id},
			{"_id": id.Hex()},
		},
	}

	err := r.collection.FindOne(ctx, filter).Decode(&role)
	if err != nil {
		return nil, err
	}

	return &role, nil
}


func (r *RoleRepository) ListWithCreator(ctx context.Context, filter bson.M) ([]bson.M, error) {
	pipeline := []bson.M{
		{"$match": filter},

		{
			"$lookup": bson.M{
				"from": "admins",
				"localField": "createdBy",
				"foreignField": "_id",
				"as": "creator",
			},
		},
		{
			"$unwind": bson.M{
				"path": "$creator",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$project": bson.M{
				"name":        1,
				"roleType":    1,
				"status":      1,
				"zoneScope":   1,
				"description": 1,
				"permissions": 1,
				"createdAt":   1,
				"updatedAt":   1,
				"createdBy": bson.M{
					"id":   "$creator._id",
					"name": "$creator.name",
				},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []bson.M
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}
func (r *RoleRepository) UpdateByID(
	ctx context.Context,
	id primitive.ObjectID,
	update bson.M,
) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	return err
}

func (r *RoleRepository) FindByIDWithCreator(
	ctx context.Context,
	id primitive.ObjectID,
) (bson.M, error) {

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"_id": id},
					{"_id": id.Hex()}, // safety for string _id
				},
			},
		},
		{
			"$lookup": bson.M{
				"from": "admins",
				"localField": "createdBy",
				"foreignField": "_id",
				"as": "creator",
			},
		},
		{
			"$unwind": bson.M{
				"path": "$creator",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$project": bson.M{
				"name":        1,
				"roleType":    1,
				"status":      1,
				"zoneScope":   1,
				"description": 1,
				"permissions": 1,
				"createdAt":   1,
				"updatedAt":   1,
				"createdBy": bson.M{
					"id":   "$creator._id",
					"name": "$creator.name",
				},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, mongo.ErrNoDocuments
	}

	var role bson.M
	if err := cursor.Decode(&role); err != nil {
		return nil, err
	}

	return role, nil
}



