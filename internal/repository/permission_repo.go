package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
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