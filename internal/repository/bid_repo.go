package repository

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
)

type BidRepo struct {
	col *mongo.Collection
}

func NewBidRepo(db *mongo.Database) *BidRepo {
	return &BidRepo{col: db.Collection("bids")}
}

func (r *BidRepo) GetBiddingStats(ctx context.Context) (dto.BiddingStats, error) {
	pipeline := []bson.M{
		{
			"$facet": bson.M{
				"total": []bson.M{
					{"$count": "count"},
				},
				"accepted": []bson.M{
					{
						"$match": bson.M{
							"status": domain.BidStatusAccepted,
						},
					},
					{"$count": "count"},
				},
				"rejected": []bson.M{
					{
						"$match": bson.M{
							"status": domain.BidStatusRejected,
						},
					},
					{"$count": "count"},
				},
				"others": []bson.M{
					{
						"$match": bson.M{
							"status": bson.M{
								"$in": []string{
									domain.BidStatusPending,
									domain.BidStatusExpired,
									domain.BidStatusWithdrawn,
								},
							},
						},
					},
					{"$count": "count"},
				},
			},
		},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Error aggregating bidding stats: %v", err)
		return dto.BiddingStats{}, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Total    []struct{ Count int64 } `bson:"total"`
		Accepted []struct{ Count int64 } `bson:"accepted"`
		Rejected []struct{ Count int64 } `bson:"rejected"`
		Others   []struct{ Count int64 } `bson:"others"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		log.Printf("Error decoding bidding stats: %v", err)
		return dto.BiddingStats{}, err
	}

	stats := dto.BiddingStats{}
	
	if len(results) > 0 {
		if len(results[0].Total) > 0 {
			stats.Total = results[0].Total[0].Count
		}
		if len(results[0].Accepted) > 0 {
			stats.AcceptedBidding = results[0].Accepted[0].Count
		}
		if len(results[0].Rejected) > 0 {
			stats.RejectedBidding = results[0].Rejected[0].Count
		}
		if len(results[0].Others) > 0 {
			stats.Others = results[0].Others[0].Count
		}
	}

	log.Printf("Bidding Stats - Total: %d, Accepted: %d, Rejected: %d, Others: %d",
		stats.Total, stats.AcceptedBidding, stats.RejectedBidding, stats.Others)

	return stats, nil
}