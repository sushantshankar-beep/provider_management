package domain

type Assessment struct {
	ComplaintID string `json:"complaint_id" bson:"complaint_id" validate:"required"`
	Remarks     string `json:"remarks" bson:"remarks" validate:"required"`
	Status      string `json:"status" bson:"status" validate:"required"`
	CreatedAt   int64  `json:"created_at" bson:"created_at"`
}
