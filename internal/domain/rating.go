package domain

import "time"

type Rating struct {
    ID               string    `bson:"_id,omitempty" json:"_id"`
    InternalID       int64     `bson:"id" json:"id"`
    ServiceID        string    `bson:"service" json:"service_id"`
    RatedBy          string    `bson:"ratedBy" json:"rated_by"`
    RatedTo          string    `bson:"ratedTo" json:"rated_to"`
    RaterType        string    `bson:"raterType" json:"rater_type"`
    Stars            int       `bson:"stars" json:"stars"`
    Review           string    `bson:"review" json:"review"`
    RecommendToFriend bool     `bson:"recommendToFriend" json:"recommend_to_friend"`
    CreatedAt        time.Time `bson:"createdAt" json:"created_at"`
    UpdatedAt        time.Time `bson:"updatedAt" json:"updated_at"`
}