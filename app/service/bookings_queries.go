package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"booking-service/app/api/dto"
	"booking-service/app/models"
)

// BookingsQueries обрабатывает запросы (чтение данных) для бронирований.
type BookingsQueries struct {
	repo   models.BookingRepository
	logger *zap.Logger
}

// NewBookingsQueries создаёт новый BookingsQueries.
func NewBookingsQueries(repo models.BookingRepository, logger *zap.Logger) *BookingsQueries {
	return &BookingsQueries{
		repo:   repo,
		logger: logger,
	}
}

// GetByID возвращает бронирование по ID.
func (q *BookingsQueries) GetByID(ctx context.Context, id int64) (dto.BookingResponse, error) {
	booking, err := q.repo.GetByID(ctx, id)
	if err != nil {
		return dto.BookingResponse{}, err
	}

	return mapBookingToResponse(booking), nil
}

// GetStatus возвращает статус бронирования по ID.
func (q *BookingsQueries) GetStatus(ctx context.Context, id int64) (models.BookingStatus, error) {
	booking, err := q.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	return booking.Status(), nil
}

// GetByFilter возвращает список бронирований с пагинацией.
func (q *BookingsQueries) GetByFilter(ctx context.Context, req dto.GetBookingsByFilterRequest) (dto.PagedResponse[dto.BookingResponse], error) {
	filter := models.NewDefaultFilter()

	if req.Page > 0 {
		filter.Page = req.Page
	}
	if req.Size > 0 {
		filter.Size = req.Size
	}
	if req.UserID != nil {
		filter.UserID = req.UserID
	}
	if req.ResourceID != nil {
		filter.ResourceID = req.ResourceID
	}
	if req.Status != nil {
		status := models.BookingStatus(*req.Status)
		if !status.IsValid() {
			return dto.PagedResponse[dto.BookingResponse]{}, fmt.Errorf("некорректный статус: %s", *req.Status)
		}
		filter.Status = &status
	}

	bookings, totalCount, err := q.repo.GetByFilter(ctx, filter)
	if err != nil {
		return dto.PagedResponse[dto.BookingResponse]{}, fmt.Errorf("получение бронирований: %w", err)
	}

	items := make([]dto.BookingResponse, 0, len(bookings))
	for i := range bookings {
		items = append(items, mapBookingToResponse(&bookings[i]))
	}

	return dto.PagedResponse[dto.BookingResponse]{
		Items:      items,
		TotalCount: totalCount,
		Page:       filter.Page,
		Size:       filter.Size,
	}, nil
}

// GetStatistics возвращает агрегированную статистику за период.
func (q *BookingsQueries) GetStatistics(ctx context.Context, dateFrom, dateTo time.Time) (dto.BookingStatisticsResponse, error) {
	dateToExclusive := dateTo.AddDate(0, 0, 1)

	data, err := q.repo.GetStatistics(ctx, dateFrom, dateToExclusive)
	if err != nil {
		return dto.BookingStatisticsResponse{}, fmt.Errorf("получение статистики: %w", err)
	}

	return mapStatisticsToResponse(data), nil
}

func mapStatisticsToResponse(data *models.StatisticsData) dto.BookingStatisticsResponse {
	countsByStatus := make(map[models.BookingStatus]int64, len(data.ByStatus))
	for _, item := range data.ByStatus {
		countsByStatus[item.Status] = item.Count
	}

	allStatuses := []models.BookingStatus{
		models.BookingStatusAwaitsConfirmation,
		models.BookingStatusConfirmed,
		models.BookingStatusCancelled,
		models.BookingStatusCancellationPending,
	}

	byStatus := make([]dto.StatusCountItem, 0, len(allStatuses))
	for _, status := range allStatuses {
		byStatus = append(byStatus, dto.StatusCountItem{
			Status: string(status),
			Count:  countsByStatus[status],
		})
	}

	topResources := make([]dto.ResourceCountItem, 0, len(data.TopResources))
	for _, item := range data.TopResources {
		topResources = append(topResources, dto.ResourceCountItem{
			ResourceID: item.ResourceID,
			Count:      item.Count,
		})
	}

	return dto.BookingStatisticsResponse{
		TotalCount:   data.TotalCount,
		ByStatus:     byStatus,
		TopResources: topResources,
	}
}

// mapBookingToResponse конвертирует доменный объект в DTO ответа.
func mapBookingToResponse(b *models.Booking) dto.BookingResponse {
	return dto.BookingResponse{
		ID:         b.ID(),
		Status:     string(b.Status()),
		UserID:     b.UserID(),
		ResourceID: b.ResourceID(),
		StartDate:  b.StartDate().Format(dto.DateFormat),
		EndDate:    b.EndDate().Format(dto.DateFormat),
		CreatedAt:  b.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
