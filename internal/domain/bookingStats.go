package domain


type BookingStats struct {
    TotalBookings      int64   `bson:"totalBookings" json:"totalBookings"`
    InProgressBookings int64   `bson:"inProgressBookings" json:"inProgressBookings"`
    PendingBookings    int64   `bson:"pendingBookings" json:"pendingBookings"`
    CompletedBookings  int64   `bson:"completedBookings" json:"completedBookings"`
    CancelledBookings  int64   `bson:"cancelledBookings" json:"cancelledBookings"`
    TotalRevenue       float64 `bson:"totalRevenue" json:"totalRevenue"`
}