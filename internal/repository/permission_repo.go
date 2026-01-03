package repository

import (
	"context"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PermissionRepo struct {
	col *mongo.Collection
}

func NewPermissionRepo(db *mongo.Database) *PermissionRepo {
	return &PermissionRepo{col: db.Collection("panelPermissions")}
}

func (r *PermissionRepo) FindAll(ctx context.Context) ([]domain.Permission, error) {
	opts := options.Find().SetSort(bson.M{"order": 1})
	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []domain.Permission
	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *PermissionRepo) FindByID(ctx context.Context, id string) (*domain.Permission, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var permission domain.Permission
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&permission)
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *PermissionRepo) Create(ctx context.Context, permission *domain.Permission) error {
	result, err := r.col.InsertOne(ctx, permission)
	if err != nil {
		return err
	}
	permission.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *PermissionRepo) Update(ctx context.Context, id string, permission *domain.Permission) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{"$set": permission}
	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *PermissionRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.col.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

type RoleRepo struct {
	col *mongo.Collection
}

func NewRoleRepo(db *mongo.Database) *RoleRepo {
	return &RoleRepo{col: db.Collection("roles")}
}

func (r *RoleRepo) FindAll(ctx context.Context, skip, limit int64) ([]domain.Role, int64, error) {
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"createdAt": -1})
	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var roles []domain.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, 0, err
	}

	total, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *RoleRepo) FindByID(ctx context.Context, id string) (*domain.Role, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var role domain.Role
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepo) Create(ctx context.Context, role *domain.Role) error {
	result, err := r.col.InsertOne(ctx, role)
	if err != nil {
		return err
	}
	role.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *RoleRepo) Update(ctx context.Context, id string, role *domain.Role) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{"$set": role}
	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *RoleRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.col.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}