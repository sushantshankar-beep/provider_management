package domain

type Complaint struct {
	ID        string `json:"id" bson:"_id"`
	UserID    string `json:"user_id" bson:"user_id"`
	Category  string `json:"category" bson:"category"`
	Status    string `json:"status" bson:"status"`
	Details   string `json:"details" bson:"details"`
	CreatedAt int64  `json:"created_at" bson:"created_at"`
}
