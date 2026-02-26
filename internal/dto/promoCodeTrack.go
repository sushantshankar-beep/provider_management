package dto

type PromoCodeTrack struct {
	ServiceID      string  `json:"serviceId"`
	ServiceNumber  string  `json:"serviceNumber"`
	UserID         string  `json:"userId"`
	UserName       string  `json:"userName,omitempty"`
	UserPhone      string  `json:"userPhone,omitempty"`
	PromoCode      string  `json:"promoCode,omitempty"`
	PromoAmount    float64 `json:"promoAmount"`
	DiscountCode   string  `json:"discountCode,omitempty"`
	DiscountAmount float64 `json:"discountAmount"`
	TotalDiscount  float64 `json:"totalDiscount"`
	AmountPaid     float64 `json:"amountPaid"`
	CreatedAt      string  `json:"createdAt"`
}
