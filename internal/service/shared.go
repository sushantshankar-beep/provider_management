package service


type Pagination struct {
    CurrentPage int   `json:"current_page"`
    TotalPages  int   `json:"total_pages"`
    Total       int64 `json:"total"`
	TotalUsers  int64 `json:"total_users"`
    Limit       int   `json:"limit"`
    HasNext     bool  `json:"has_next,omitempty"`
    HasPrev     bool  `json:"has_prev,omitempty"`
}

func defaultStr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
