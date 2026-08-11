package mailack_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	mailack "github.com/jaak-ai/mailack-sdk/go"
)

func TestSendAndRates(t *testing.T) {
	var sawAuth, sawIdem string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/messages":
			sawIdem = r.Header.Get("Idempotency-Key")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "11111111-1111-1111-1111-111111111111",
				"state": "queued",
				"canonical_hash": "aabb",
				"subject": "hi",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/rates":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"window_days": 7, "ingested": 10, "sent": 8, "bounced": 1,
				"delivery_rate": 80.0, "bounce_rate": 10.0, "days": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := mailack.NewClient(srv.URL, mailack.WithAPIKey("mlk_test"))
	msg, replay, err := c.Send(context.Background(), "k1", mailack.SendRequest{
		From: "a@x.com", To: "b@y.com", Subject: "hi", Text: "body",
	})
	require.NoError(t, err)
	require.False(t, replay)
	require.Equal(t, "queued", msg.State)
	require.Equal(t, "Bearer mlk_test", sawAuth)
	require.Equal(t, "k1", sawIdem)

	rates, err := c.Rates(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, float64(80), rates.DeliveryRate)
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"code": "quota_exceeded", "message": "quota hit"},
		})
	}))
	t.Cleanup(srv.Close)
	c := mailack.NewClient(srv.URL, mailack.WithAPIKey("x"))
	_, _, err := c.Send(context.Background(), "k", mailack.SendRequest{From: "a", To: "b", Text: "t"})
	require.Error(t, err)
	require.True(t, mailack.Is(err, "quota_exceeded"))
}
