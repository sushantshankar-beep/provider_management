package repository

import (
	"time"
	"context"     
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminRepository struct {
	collection *mongo.Collection
}

func NewAdminRepository(db *mongo.Database) *AdminRepository {
	return &AdminRepository{
		collection: db.Collection("admins"),
	}
}

func (r *AdminRepository) Create(ctx context.Context, admin *domain.Admin) error {
	admin.CreatedAt = time.Now()
	admin.UpdatedAt = time.Now()
	admin.AdminID = time.Now().UnixNano() / 1000000
	admin.PowerLevel = domain.GetPowerLevel(admin.Role)
	if admin.Status == "" {
		admin.Status = domain.StatusActive
	}
	result, err := r.collection.InsertOne(ctx, admin)
	if err != nil {
		return err
	}
	admin.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *AdminRepository) FindByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) FindByIDAndToken(ctx context.Context, id primitive.ObjectID, token string) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.collection.FindOne(ctx, bson.M{
		"_id":          id,
		"tokens.token": token,
	}).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) FindAll(ctx context.Context, query bson.M, limit, offset int64) ([]domain.Admin, int64, error) {
	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetLimit(limit).
		SetSkip(offset).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var admins []domain.Admin
	if err := cursor.All(ctx, &admins); err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

func (r *AdminRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
    _, err := r.collection.UpdateOne(ctx,
        bson.M{"_id": id},
        update,
    )
    return err
}

func (r *AdminRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *AdminRepository) CountDocuments(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

func (r *AdminRepository) FindOne(ctx context.Context, filter bson.M) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.collection.FindOne(ctx, filter).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (a *AdminRepository) FindAdminIDsByRoleIDs(ctx context.Context, roleIDs []primitive.ObjectID) ([]primitive.ObjectID, error) {
	filter := bson.M{
		"roleId": bson.M{"$in": roleIDs},
		"status": "active",
	}

	cursor, err := a.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var adminIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var admin struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&admin); err != nil {
			continue
		}
		adminIDs = append(adminIDs, admin.ID)
	}

	return adminIDs, nil
}

func (r *AdminRepository) AggregateActivationTeam(ctx context.Context, pipeline []bson.M) ([]domain.ActivationTeamMember, error) {
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var team []domain.ActivationTeamMember
	if err := cursor.All(ctx, &team); err != nil {
		return nil, err
	}

	return team, nil
}