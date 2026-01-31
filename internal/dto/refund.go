package dto

import (
	"provider_management/internal/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefundListItemDTO struct {
	ID              primitive.ObjectID  `json:"id"`
	RefundID        string              `json:"refund_id"`
	UserCode        string              `json:"userCode"`
	TxnID    string             `bson:"txnid"`
	NetRefund         float64            `json:"net_refund"`
	ServiceNumber   string              `json:"serviceNumber"`
	ComplaintNumber string              `json:"complaintNumber"`
	GST             float64             `json:"gst"`
	Mode            string              `json:"mode"`
	Amount          float64             `json:"amount"`
	Reason          string              `json:"reason"`
	Status          domain.RefundStatus `json:"status"`
	SubmittedAt     time.Time           `json:"submitted_at"`
}

type RefundDetailDTO struct {
	ID              primitive.ObjectID  `json:"id"`
	RefundID        string              `json:"refund_id"`
	UserCode        string              `json:"userCode"`
	Amount          float64             `json:"amount"`
	NetRefund         float64            `json:"net_refund"`
	ServiceNumber   string              `json:"serviceNumber"`
	ComplaintNumber string              `json:"complaintNumber"`
    ComplaintID              string                  `json:"complaintId"`
	TransactionID   string              `json:"txnId"`
	GST             float64             `json:"gst"`
	Mode            string              `json:"mode"`
	Reason          string              `json:"reason"`
	Status          domain.RefundStatus `json:"status"`
	Timeline        domain.RefundTimeline     `bson:"timeline" json:"timeline"`
	CreatedAt       time.Time           `json:"created_at"`
}

type RefundListResponse struct {
	Refunds    []RefundListItemDTO `json:"refunds"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int64               `json:"total_pages"`
}