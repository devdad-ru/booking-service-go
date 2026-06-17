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
	repo        models.BookingRepository
	queriesRepo models.BookingQueriesRepository
	logger      *zap.Logger
}

// NewBookingsQueries создаёт новый BookingsQueries.
func NewBookingsQueries(repo models.BookingRepository, logger *zap.Logger) *BookingsQueries {
	qRepo, _ := repo.(models.BookingQueriesRepository)
	return &BookingsQueries{
		repo:        repo,
		queriesRepo: qRepo,
		logger:      logger,
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

func (q *BookingsQueries) GetStatistics(ctx context.Context, req dto.BookingStatisticsRequest) (dto.BookingStatisticsResponse, error) {
	if q.queriesRepo == nil {
		return dto.BookingStatisticsResponse{}, fmt.Errorf("аналитический репозиторий не инициализирован")
	}

	const layout = "2006-01-02"
	dateFrom, err := time.Parse(layout, req.DateFrom)
	if err != nil {
		return dto.BookingStatisticsResponse{}, fmt.Errorf("некорректный формат dateFrom: %w", err)
	}

	dateTo, err := time.Parse(layout, req.DateTo)
	if err != nil {
		return dto.BookingStatisticsResponse{}, fmt.Errorf("некорректный формат dateTo: %w", err)
	}

	dateTo = dateTo.Add(24*time.Hour - time.Nanosecond)

	if dateTo.Before(dateFrom) {
		return dto.BookingStatisticsResponse{}, fmt.Errorf("dateTo не может быть раньше dateFrom")
	}

	stats, err := q.queriesRepo.GetStatistics(ctx, dateFrom, dateTo)
	if err != nil {
		return dto.BookingStatisticsResponse{}, fmt.Errorf("сервис статистики: %w", err)
	}

	statusCounts := make(map[string]int64)
	for status, count := range stats.StatusCounts {
		statusCounts[string(status)] = count
	}

	topResources := make([]dto.TopResourceDTO, 0, len(stats.TopResources))
	for _, res := range stats.TopResources {
		topResources = append(topResources, dto.TopResourceDTO{
			ResourceID:   res.ResourceID,
			BookingCount: res.BookingCount,
		})
	}

	return dto.BookingStatisticsResponse{
		TotalCount:   stats.TotalCount,
		StatusCounts: statusCounts,
		TopResources: topResources,
	}, nil
}
