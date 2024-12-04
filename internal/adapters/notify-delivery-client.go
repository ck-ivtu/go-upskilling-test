package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-test/internal/model"
	"go.uber.org/zap"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	clientOnce sync.Once
	client     *http.Client
)

type NotifyDeliveryClient struct {
	basePath  string
	webhookId string
	client    *http.Client
	Logger    *zap.Logger
}

func NewNotifyDeliveryClient(webhookId string, logger *zap.Logger) *NotifyDeliveryClient {
	clientOnce.Do(func() {
		client = initClient()
	})

	return &NotifyDeliveryClient{
		basePath:  "https://webhook.site",
		client:    client,
		webhookId: webhookId,
		Logger:    logger,
	}
}

func (nc *NotifyDeliveryClient) Notify(ctx context.Context, deliveryPackage model.DeliveryPackage) error {
	webhookURL := fmt.Sprintf("%s/%s", nc.basePath, nc.webhookId)

	payload, err := json.Marshal(deliveryPackage)
	if err != nil {
		nc.Logger.Error("Failed to marshal delivery package", zap.Error(err))
		return fmt.Errorf("failed to marshal delivery package: %w", err)
	}

	nc.Logger.Info("Sending request to webhook", zap.String("webhookURL", webhookURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		nc.Logger.Error("Failed to create HTTP request", zap.Error(err))
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := nc.client.Do(req)
	if err != nil {
		nc.Logger.Error("Failed to send request", zap.Error(err))
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	nc.Logger.Info("Received response from webhook", zap.Int("statusCode", resp.StatusCode))

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		nc.Logger.Error("Webhook responded with an error", zap.Int("statusCode", resp.StatusCode))
		return fmt.Errorf("webhook responded with status code: %d", resp.StatusCode)
	}

	nc.Logger.Info("Successfully sent delivery notification")
	return nil
}

func initClient() *http.Client {
	dc := func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := net.Dialer{
			Timeout: time.Minute,
		}
		return dialer.DialContext(ctx, network, addr)
	}

	c := http.Client{
		Transport: &http.Transport{
			DialContext:           dc,
			MaxIdleConnsPerHost:   3,
			MaxIdleConns:          12,
			MaxConnsPerHost:       3,
			IdleConnTimeout:       time.Minute,
			ResponseHeaderTimeout: time.Second,
			WriteBufferSize:       8 << 10,
			ReadBufferSize:        8 << 10,
		},
		CheckRedirect: func(next *http.Request, history []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * 30,
	}

	return &c
}
