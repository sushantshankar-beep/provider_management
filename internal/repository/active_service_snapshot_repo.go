package repository

import (
    "context"
    "provider_management/internal/domain"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

type ActiveServiceSnapshotRepo struct {
    collection *mongo.Collection
}

func NewActiveServiceSnapshotRepo(db *mongo.Database) *ActiveServiceSnapshotRepo {
    return &ActiveServiceSnapshotRepo{
        collection: db.Collection("accepted_service_snapshots"),
    }
}

func (r *ActiveServiceSnapshotRepo) FindByServiceID(ctx context.Context, serviceID primitive.ObjectID) (*domain.ActiveServiceSnapshot, error) {
    var snapshot domain.ActiveServiceSnapshot
    err := r.collection.FindOne(ctx, bson.M{"serviceId": serviceID}).Decode(&snapshot)
    if err != nil {
        return nil, err
    }
    return &snapshot, nil
}