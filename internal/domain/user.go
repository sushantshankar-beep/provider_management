package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Preferences struct {
	Email struct {
		AllowInvoice     bool `bson:"allowInvoice" json:"allow_invoice"`
		AllowPromotional bool `bson:"allowPromotional" json:"allow_promotional"`
	} `bson:"email" json:"email"`

	SMSWhatsapp string `bson:"smsWhatsapp" json:"sms_whatsapp"`

	PushNotifications struct {
		Allowed bool `bson:"allowed" json:"allowed"`
	} `bson:"pushNotifications" json:"push_notifications"`

	PictureInPicture struct {
		Allowed bool `bson:"allowed" json:"allowed"`
	} `bson:"pictureInPicture" json:"picture_in_picture"`
}

type User struct {
	ID                string      `bson:"_id,omitempty" json:"_id"`
	UserCode          string      `bson:"userCode" json:"userCode"`
	InternalID        int64       `bson:"id" json:"id"`
	Name              string      `bson:"name" json:"name"`
	Email             string      `bson:"email,omitempty" json:"email,omitempty"`
	ImageUrl          string      `bson:"image_url" json:"image_url"`
	AppStateStatus    string      `bson:"appStateStatus" json:"app_state_status"`
	Phone             string      `bson:"phone" json:"phone"`
	IsSocketConnected bool        `bson:"isSocketConnected" json:"is_socket_connected"`
	OTP               OTP         `bson:"otp,omitempty" json:"-"`
	Location          GeoPoint    `bson:"location,omitempty" json:"location,omitempty"`
	Address           string      `bson:"address,omitempty" json:"address,omitempty"`
	Preferences       Preferences `bson:"preferences,omitempty" json:"preferences,omitempty"`
	SelectedCity      string      `bson:"selectedCity,omitempty" json:"selected_city,omitempty"`
	SelectedCityName  string      `bson:"selectedCityName,omitempty" json:"selected_city_name,omitempty"`
	WalletBalance     float64     `bson:"walletBalance" json:"wallet_balance"`
	Notes             []UserNote  `bson:"notes,omitempty" json:"notes,omitempty"`
	IsActive          string      `bson:"isActive" json:"is_active"`
	ServiceOTP        string      `bson:"service_otp" json:"service_otp"`
	CreatedAt         time.Time   `bson:"createdAt" json:"created_at"`
	UpdatedAt         time.Time   `bson:"updatedAt" json:"updated_at"`
	VehicleID           *primitive.ObjectID  `bson:"vehicleId,omitempty" json:"vehicleId,omitempty"`
	PrimaryVehicleID    *primitive.ObjectID  `bson:"primaryVehicleId,omitempty" json:"primaryVehicleId,omitempty"`
	FallbackVehicleIDs []primitive.ObjectID  `bson:"fallbackVehicleIds,omitempty" json:"fallbackVehicleIds,omitempty"`
}

type UserNote struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Content   string    `bson:"content" json:"content"`
	AddedBy   string    `bson:"addedBy" json:"addedBy"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
