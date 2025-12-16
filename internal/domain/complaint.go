package domain

import "time"

type ComplaintTimeline struct {
    Initiated *time.Time `bson:"initiated" json:"initiated"`
    InReview  *time.Time `bson:"inReview" json:"in_review"`
    Resolved  *time.Time `bson:"resolved" json:"resolved"`
}

type Complaint struct {
    ID                string           `bson:"_id,omitempty" json:"_id"`
    InternalID        int64            `bson:"id" json:"id"`
    AcceptedServiceID string           `bson:"acceptedService" json:"accepted_service_id"`
    AcceptedServiceNo int64            `bson:"acceptedServiceId" json:"accepted_service_no"`
    UserID            string           `bson:"userId,omitempty" json:"user_id,omitempty"`
    ProviderID        string           `bson:"providerId,omitempty" json:"provider_id,omitempty"`
    RaisedBy          string           `bson:"raisedBy" json:"raised_by"`
    Problem           string           `bson:"problem" json:"problem"`
    Photos            []string         `bson:"photos" json:"photos"`
    Status            string           `bson:"status" json:"status"`
    Timeline          ComplaintTimeline `bson:"timeline" json:"timeline"`
    UpdatedByAdmin    string           `bson:"updatedByAdmin,omitempty" json:"updated_by_admin,omitempty"`
    AdminUpdatedAt    *time.Time       `bson:"adminUpdatedAt,omitempty" json:"admin_updated_at,omitempty"`
    CreatedAt         time.Time        `bson:"createdAt" json:"created_at"`
    UpdatedAt         time.Time        `bson:"updatedAt" json:"updated_at"`
}