package domain

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProviderFinancialSummary struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	ProviderID         primitive.ObjectID `bson:"providerId" json:"provider_id"`

	TotalEarnings      float64            `bson:"totalEarnings" json:"total_earnings"`
	TotalCommission    float64            `bson:"totalCommission" json:"total_commission"`
	TotalGST           float64            `bson:"totalGST" json:"total_gst"`

	PendingSettlement  float64            `bson:"pendingSettlement" json:"pending_settlement"`
	TotalSettled       float64            `bson:"totalSettled" json:"total_settled"`

	UpdatedAt          time.Time          `bson:"updatedAt" json:"updated_at"`
}
