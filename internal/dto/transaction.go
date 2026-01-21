package dto

type TransactionFilters struct {
	Search    string
	Status    string
	Method    string
	CreatedAt string
}

type PaginationParams struct {
	Page  int64
	Limit int64
	Skip  int64
}

type TransactionResponse struct {
	ID            string  `json:"_id"`
	UserID        string  `json:"user_id"`
	UserName      string  `json:"user_name,omitempty"`
	BookingID     string  `json:"booking_id,omitempty"`
	TxnID         string  `json:"txnid"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Status        string  `json:"status"`
	Method        string  `json:"method,omitempty"`
	PaymentSource string  `json:"payment_source,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	BookingNo     string  `json:"booking_no,omitempty"`
}

type PaginatedResponse struct {
	Data       []TransactionResponse `json:"data"`
	Pagination PaginationMeta        `json:"pagination"`
}

type PaginationMeta struct {
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
	Limit       int64 `json:"limit"`
	Page        int64 `json:"page"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int64 `json:"total_pages"`
}