package domain

import "time"

type Preferences struct {
	Email struct {
		AllowInvoice      bool `bson:"allowInvoice" json:"allow_invoice"`
		AllowPromotional  bool `bson:"allowPromotional" json:"allow_promotional"`
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
	ID                 string       `bson:"_id,omitempty" json:"_id"`
	InternalID         int64        `bson:"id" json:"id"`
	Name               string       `bson:"name" json:"name"`
	Email              string       `bson:"email,omitempty" json:"email,omitempty"`
	ProfileURL         string       `bson:"profileUrl,omitempty" json:"profile_url,omitempty"`
	AppStateStatus     string       `bson:"appStateStatus" json:"app_state_status"`
	Phone              string       `bson:"phone" json:"phone"`
	IsSocketConnected  bool         `bson:"isSocketConnected" json:"is_socket_connected"`
	OTP                OTP          `bson:"otp,omitempty" json:"-"`
	Location           GeoPoint     `bson:"location,omitempty" json:"location,omitempty"`
	Address            string       `bson:"address,omitempty" json:"address,omitempty"`
	Preferences        Preferences  `bson:"preferences,omitempty" json:"preferences,omitempty"`
	SelectedCity       string       `bson:"selectedCity,omitempty" json:"selected_city,omitempty"`
	SelectedCityName   string       `bson:"selectedCityName,omitempty" json:"selected_city_name,omitempty"`
	WalletBalance float64            `bson:"walletBalance" json:"wallet_balance"`
	IsActive           string       `bson:"isActive" json:"is_active"`
	CreatedAt          time.Time    `bson:"createdAt" json:"created_at"`
	UpdatedAt          time.Time    `bson:"updatedAt" json:"updated_at"`
}
