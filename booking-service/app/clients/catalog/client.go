package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client -- HTTP-клиент для взаимодействия с Catalog-сервисом.
type Client struct {
	baseURL    string
	httpClient *http.Client
	retryCfg   RetryConfig
	logger     *zap.Logger
}

// NewClient создаёт новый Catalog-клиент.
func NewClient(baseURL string, timeout time.Duration, maxRetries int, retryBaseDelay time.Duration, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		retryCfg: RetryConfig{
			MaxRetries: maxRetries,
			BaseDelay:  retryBaseDelay,
		},
		logger: logger,
	}
}

// CreateBookingJob отправляет запрос на создание задания бронирования в Catalog.
func (c *Client) CreateBookingJob(ctx context.Context, req CreateBookingJobRequest) (*BookingJobResponse, error) {
	var result BookingJobResponse

	err := withRetry(ctx, c.retryCfg, "CreateBookingJob", func() error {
		resp, err := c.doJSON(ctx, http.MethodPost, "/api/booking-jobs", req)
		if err != nil {
			return err
		}
		return json.Unmarshal(resp, &result)
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBookingJobByBookingID получает статус задания по ID бронирования.
func (c *Client) GetBookingJobByBookingID(ctx context.Context, bookingID int64) (*BookingJobResponse, error) {
	var result BookingJobResponse

	err := withRetry(ctx, c.retryCfg, "GetBookingJobByBookingID", func() error {
		path := fmt.Sprintf("/api/booking-jobs/by-booking/%d", bookingID)
		resp, err := c.doJSON(ctx, http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		return json.Unmarshal(resp, &result)
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CancelBookingJob отменяет задание бронирования.
func (c *Client) CancelBookingJob(ctx context.Context, req CancelBookingJobRequest) error {
	return withRetry(ctx, c.retryCfg, "CancelBookingJob", func() error {
		_, err := c.doJSON(ctx, http.MethodPost, "/api/booking-jobs/cancel", req)
		return err
	})
}

// doJSON выполняет HTTP-запрос с JSON-телом и возвращает тело ответа.
func (c *Client) doJSON(ctx context.Context, method, path string, body any) ([]byte, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("сериализация тела запроса: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("создание запроса: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("выполнение запроса %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("чтение ответа: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d от Catalog (%s %s): %s", resp.StatusCode, method, path, string(respBody))
	}

	return respBody, nil
}
