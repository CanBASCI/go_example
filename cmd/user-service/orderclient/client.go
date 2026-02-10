package orderclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// DefaultHTTPClient is used for outbound calls to order-service (10s timeout).
var DefaultHTTPClient = &http.Client{Timeout: 10 * time.Second}

// OrderSummary is the order payload from order-service GET /orders?userId=.
type OrderSummary struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"createdAt"`
}

// ListByUserID fetches orders for the user from order-service.
func ListByUserID(ctx context.Context, baseURL string, userID uuid.UUID) ([]OrderSummary, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = "/orders"
	u.RawQuery = url.Values{"userId": {userID.String()}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := DefaultHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("order-service returned %d", resp.StatusCode)
	}
	var list []OrderSummary
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []OrderSummary{}
	}
	return list, nil
}
