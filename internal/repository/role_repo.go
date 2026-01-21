package repository

import (
	"time"
    "context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoleRepository struct {
	collection *mongo.Collection
}

func NewRoleRepository(db *mongo.Database) *RoleRepository {
	return &RoleRepository{
		collection: db.Collection("roles"),
	}
}

type ZoneStatsAggResult struct {
	TotalActivationMembers int `bson:"totalActivationMembers"`
	TotalProviders         int `bson:"totalProviders"`
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

func (r *RoleRepository) UpdateStatus( ctx context.Context, id primitive.ObjectID, status domain.RoleStatus) error {
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

func (r *RoleRepository) DeleteByID( ctx context.Context, id primitive.ObjectID ) error {
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
func (r *RoleRepository) UpdateByID( ctx context.Context, id primitive.ObjectID, update bson.M ) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	return err
}

func (r *RoleRepository) FindByIDWithCreator( ctx context.Context, id primitive.ObjectID ) (bson.M, error) {
	
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"_id": id},
					{"_id": id.Hex()},
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
				"zoneName":   1,
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

func (r *RoleRepository) FindDistinctRoleTypes(ctx context.Context) ([]string, error) {
	result, err := r.collection.Distinct(ctx, "roleType", bson.M{})
	if err != nil {
		return nil, err
	}

	var roleTypes []string
	for _, v := range result {
		if s, ok := v.(string); ok {
			roleTypes = append(roleTypes, s)
		}
	}
	return roleTypes, nil
}

func (r *RoleRepository) FindByRoleType(ctx context.Context, roleType string) ([]domain.Role, error) {
	filter := bson.M{"roleType": roleType}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []domain.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepository) FindRoleIDsByZoneName(ctx context.Context, zoneName string) ([]primitive.ObjectID, error) {
	filter := bson.M{
		"zoneName": zoneName,
		"status":   "active",
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roleIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var role struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&role); err != nil {
			continue
		}
		roleIDs = append(roleIDs, role.ID)
	}

	return roleIDs, nil
}

func (r *RoleRepository) AggregateZoneStats(ctx context.Context, zoneName string) ([]ZoneStatsAggResult, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"zoneName": zoneName,
				"status":   "active",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "admins",
				"localField":   "_id",
				"foreignField": "roleId",
				"as":           "admins",
			},
		},
		{
			"$addFields": bson.M{
				"admins": bson.M{
					"$filter": bson.M{
						"input": "$admins",
						"as":    "admin",
						"cond":  bson.M{"$eq": []interface{}{"$$admin.status", "active"}},
					},
				},
			},
		},
		{
			"$addFields": bson.M{
				"adminIds": "$admins._id",
			},
		},
		{
			"$lookup": bson.M{
				"from": "providers",
				"let":  bson.M{"adminIds": "$adminIds"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$in": []interface{}{"$createdBy", "$$adminIds"},
							},
						},
					},
				},
				"as": "providers",
			},
		},
		{
			"$group": bson.M{
				"_id":                    nil,
				"totalActivationMembers": bson.M{"$sum": bson.M{"$size": "$adminIds"}},
				"totalProviders":         bson.M{"$sum": bson.M{"$size": "$providers"}},
			},
		},
		{
			"$project": bson.M{
				"_id":                    0,
				"totalActivationMembers": 1,
				"totalProviders":         1,
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []ZoneStatsAggResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
