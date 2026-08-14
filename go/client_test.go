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
				"id":             "11111111-1111-1111-1111-111111111111",
				"state":          "queued",
				"canonical_hash": "aabb",
				"subject":        "hi",
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

func TestSealEvidenceVerify(t *testing.T) {
	const msgID = "11111111-1111-1111-1111-111111111111"
	var sawCertified *bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/messages":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if v, ok := body["certified"].(bool); ok {
				sawCertified = &v
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": msgID, "state": "queued", "certified": true,
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/messages/"+msgID+"/seal":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message_id": msgID, "batch_id": "b1", "seal_type": "merkle",
				"canonical_hash": "aabb", "merkle_root": "ccdd",
				"certificate_id": "cert-1", "serial_number": "42",
				"policy_oid": "1.2.3", "algorithm_oid": "2.3.4",
				"sealed_at": "2026-08-13T00:00:00Z",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/messages/"+msgID+"/evidence":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message_id": msgID, "canonical_hash": "aabb", "mime_sha256": "eeff",
				"message_id_header": "<m@x>", "date_header": "2026-08-13T00:00:00Z",
				"batch_id": "b1", "merkle_root": "ccdd",
				"sealed_at": "2026-08-13T00:00:00Z", "certificate_id": "cert-1",
				"leaf_index": 7,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/messages/"+msgID+"/proof-bundle":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"version": 1, "message_id": msgID, "canonical_hash": "aabb",
				"leaf_index": 7, "proof_path": []any{}, "merkle_root": "ccdd",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/verify":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid": true, "merkle_root": "ccdd",
				"certificate_id": "cert-1", "sealed_at": "2026-08-13T00:00:00Z",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := mailack.NewClient(srv.URL, mailack.WithAPIKey("mlk_test"))
	ctx := context.Background()

	certified := true
	msg, _, err := c.Send(ctx, "k1", mailack.SendRequest{From: "a@x.com", To: "b@y.com", Text: "t", Certified: &certified})
	require.NoError(t, err)
	require.True(t, msg.Certified)
	require.NotNil(t, sawCertified)
	require.True(t, *sawCertified)

	seal, err := c.Seal(ctx, msg.ID)
	require.NoError(t, err)
	require.Equal(t, "ccdd", seal.MerkleRoot)
	require.Equal(t, "cert-1", seal.CertificateID)

	ev, err := c.Evidence(ctx, msg.ID)
	require.NoError(t, err)
	require.Equal(t, int64(7), ev.LeafIndex)
	require.Equal(t, "eeff", ev.MIMESHA256)

	bundle, err := c.ProofBundle(ctx, msg.ID)
	require.NoError(t, err)
	require.Contains(t, string(bundle), `"merkle_root":"ccdd"`)

	vr, err := c.Verify(ctx, msg.ID)
	require.NoError(t, err)
	require.True(t, vr.Valid)
}
