package catalog

// CreateBookingJobRequest -- запрос на создание задания бронирования в Catalog.
type CreateBookingJobRequest struct {
	BookingID  int64  `json:"bookingId"`
	ResourceID int64  `json:"resourceId"`
	StartDate  string `json:"startDate"`
	EndDate    string `json:"endDate"`
}

// BookingJobResponse -- ответ от Catalog о статусе задания.
type BookingJobResponse struct {
	ID        int64  `json:"id"`
	BookingID int64  `json:"bookingId"`
	Status    string `json:"status"` // "pending", "confirmed", "denied"
}

// CancelBookingJobRequest -- запрос на отмену задания бронирования.
type CancelBookingJobRequest struct {
	BookingID int64 `json:"bookingId"`
}
