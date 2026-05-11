package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ap2/order-service/internal/domain"
)

type PaymentClient struct {
	baseURL string
	client  *http.Client
}

func NewPaymentClient(baseURL string, client *http.Client) *PaymentClient {
	return &PaymentClient{baseURL: strings.TrimRight(baseURL, "/"), client: client}
}

func (p *PaymentClient) Authorize(ctx context.Context, req domain.PaymentRequest) (*domain.PaymentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/payments", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("payment service server error: %s", string(respBody))
	}

	var paymentResp domain.PaymentResponse
	if err := json.Unmarshal(respBody, &paymentResp); err != nil {
		return nil, err
	}
	return &paymentResp, nil
}
