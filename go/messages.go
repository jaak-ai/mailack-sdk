package mailack

import (
	"context"
	"net/http"
	"time"
)

// SendRequest is the JSON form of POST /v1/messages.
type SendRequest struct {
	From       string            `json:"from"`
	To         string            `json:"to"`
	Subject    string            `json:"subject,omitempty"`
	Text       string            `json:"text,omitempty"`
	HTML       string            `json:"html,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	TemplateID string            `json:"template_id,omitempty"`
	Variables  map[string]string `json:"variables,omitempty"`
}

// Message is a certified outbound message as returned by the API.
type Message struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	IdempotencyKey  string     `json:"idempotency_key"`
	FromAddress     string     `json:"from_address"`
	ToAddress       string     `json:"to_address"`
	Subject         string     `json:"subject"`
	CanonicalHash   string     `json:"canonical_hash"`
	MessageIDHeader string     `json:"message_id_header"`
	DateHeader      time.Time  `json:"date_header"`
	State           string     `json:"state"`
	BatchID         *string    `json:"batch_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Send ingests one certified message. idempotencyKey is required: replaying
// the same key returns the original message with replay=true.
func (c *Client) Send(ctx context.Context, idempotencyKey string, req SendRequest) (msg *Message, replay bool, err error) {
	resp, err := c.do(ctx, http.MethodPost, "/v1/messages", idempotencyKey, "application/json", req)
	if err != nil {
		return nil, false, err
	}
	var m Message
	if err := decodeJSON(resp, &m); err != nil {
		return nil, false, err
	}
	return &m, resp.Header.Get("Idempotent-Replay") == "true", nil
}

// BatchItem is one entry of SendBatch.
type BatchItem struct {
	IdempotencyKey string            `json:"idempotency_key"`
	From           string            `json:"from"`
	To             string            `json:"to"`
	Subject        string            `json:"subject"`
	Text           string            `json:"text,omitempty"`
	HTML           string            `json:"html,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
}

// BatchItemResult is the per-message outcome of SendBatch.
type BatchItemResult struct {
	Index  int    `json:"index"`
	ID     string `json:"id,omitempty"`
	State  string `json:"state,omitempty"`
	Replay bool   `json:"replay,omitempty"`
	Error  *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// BatchResult is the response of POST /v1/messages/batch.
type BatchResult struct {
	Items    []BatchItemResult `json:"items"`
	Created  int               `json:"created"`
	Failed   int               `json:"failed"`
	Replayed int               `json:"replayed"`
}

// SendBatch ingests up to 100 messages. Partial success is allowed.
func (c *Client) SendBatch(ctx context.Context, items []BatchItem) (*BatchResult, error) {
	resp, err := c.do(ctx, http.MethodPost, "/v1/messages/batch", "", "application/json", map[string]any{
		"messages": items,
	})
	if err != nil {
		return nil, err
	}
	var out BatchResult
	if err := decodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMessage returns one message and its evidence events.
func (c *Client) GetMessage(ctx context.Context, id string) (*Message, error) {
	resp, err := c.do(ctx, http.MethodGet, "/v1/messages/"+id, "", "", nil)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		Message Message `json:"message"`
	}
	if err := decodeJSON(resp, &wrap); err != nil {
		return nil, err
	}
	return &wrap.Message, nil
}

// Rates is the deliverability dashboard payload (GET /v1/rates).
type Rates struct {
	WindowDays    int     `json:"window_days"`
	Ingested      int64   `json:"ingested"`
	Sent          int64   `json:"sent"`
	Deferred      int64   `json:"deferred"`
	Bounced       int64   `json:"bounced"`
	Complained    int64   `json:"complained"`
	Sealed        int64   `json:"sealed"`
	DeliveryRate  float64 `json:"delivery_rate"`
	BounceRate    float64 `json:"bounce_rate"`
	ComplaintRate float64 `json:"complaint_rate"`
	Days          []struct {
		Day        string `json:"day"`
		Ingested   int64  `json:"ingested"`
		Sent       int64  `json:"sent"`
		Deferred   int64  `json:"deferred"`
		Bounced    int64  `json:"bounced"`
		Complained int64  `json:"complained"`
		Sealed     int64  `json:"sealed"`
	} `json:"days"`
}

// Rates returns tenant deliverability counters for the last days (1–90, default 14).
func (c *Client) Rates(ctx context.Context, days int) (*Rates, error) {
	path := "/v1/rates"
	if days > 0 {
		path = query(path, map[string]string{"days": itoa(days)})
	}
	resp, err := c.do(ctx, http.MethodGet, path, "", "", nil)
	if err != nil {
		return nil, err
	}
	var out Rates
	if err := decodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
