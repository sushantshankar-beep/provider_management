package dto

type DashboardResponse struct {
	Providers   ProvidersStats   `json:"providers"`
	Users       UsersStats       `json:"users"`
	Biddings    BiddingStats  `json:"biddings"`
	Bookings    BookingsStats    `json:"bookings"`
	Settlement  CombinedSettlementStats  `json:"settlement"`
	Complaints  ComplaintsStats  `json:"complaints"`
	Revenue     RevenueStats     `json:"revenue"`
	TopServices []TopService     `json:"top_services"`
}

type ProvidersStats struct {
	Total    int64 `json:"total"`
	Active   int64 `json:"active"`
	Inactive int64 `json:"inactive"`
}

type UsersStats struct {
	Total    int64 `json:"total"`
	AMC      int64 `json:"amc"`
	Active   int64 `json:"active"`
	Inactive int64 `json:"inactive"`
}

type BookingsStats struct {
	Total     int64 `json:"total"`
	Completed int64 `json:"completed"`
	Ongoing   int64 `json:"ongoing"`
}

type SettlementStats struct {
	SettledAmount    float64  `json:"settled_amount"`
}

type ComplaintsStats struct {
	Total          int64 `json:"total"`
	Resolved       int64 `json:"resolved"`
	UnderReview    int64 `json:"under_review"`
}

type RevenueStats struct {
	Period string             `json:"period"`
	Data   []RevenueDataPoint `json:"data"`
	TotalAmount float64         `json:totalAmount`
}

type RevenueDataPoint struct {
	Day    string  `json:"day"`
	Amount float64 `json:"amount"`
}

type TopService struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
       Count      int64   `json:"count"`
}

type TransactionStats struct {
    TotalAmount  float64 `json:"total_amount"`
    GSTAmount    float64 `json:"gst_amount"`
}


type CombinedSettlementStats struct {
    TransactionStats TransactionStats `json:"transaction_stats"`
    SettlementStats  SettlementStats  `json:"settlement_stats"`
    PlatformRevenue  float64          `json:"platform_revenue"`
	SettledPercentage   float64       `json:"settled_percentage"`
    PlatformPercentage  float64   `json:"platform_percentage"`
    GSTPercentage       float64  `json:"gst_percentage"`
}

type BiddingStats struct {
	Total            int64 `json:"total"`
	AcceptedBidding  int64 `json:"acceptedBidding"`
	RejectedBidding  int64 `json:"rejectedBidding"`
	Others           int64 `json:"others"` // pending, expired, withdrawn
}