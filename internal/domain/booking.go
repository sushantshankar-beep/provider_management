package domain

import "time"

type Booking struct {
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"user_id" bson:"user_id"`
	ServiceID string    `json:"service_id" bson:"service_id"`
	Date      time.Time `json:"date" bson:"date"`
}
