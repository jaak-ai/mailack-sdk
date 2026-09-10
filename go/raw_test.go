package mailack

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRawDownloads(t *testing.T) {
	payload := []byte{0, 255, 13, 10, 128}
	for _, status := range []int{200, 404} {
		for _, event := range []bool{false, true} {
			path, header := "/v1/messages/m/raw", "X-Mailack-Canonical-Hash"
			if event {
				path, header = "/v1/messages/m/events/e/raw", "X-Mailack-Raw-SHA256"
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != path || r.Header.Get("Authorization") != "Bearer test" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set(header, "digest")
				w.WriteHeader(status)
				if status == 404 {
					w.Write([]byte(`{"error":{"code":"not_found"}}`))
					return
				}
				w.Write(payload)
			}))
			client := NewClient(server.URL, WithAPIKey("test"))
			var data []byte
			var hash string
			var err error
			if event {
				result, e := client.GetEventRaw(context.Background(), "m", "e")
				err = e
				if result != nil {
					data, hash = result.Data, result.RawSHA256
				}
			} else {
				result, e := client.GetMessageRaw(context.Background(), "m")
				err = e
				if result != nil {
					data, hash = result.Data, result.CanonicalHash
				}
			}
			server.Close()
			if status == 404 {
				var apiErr *APIError
				if !errors.As(err, &apiErr) || apiErr.Code != "not_found" {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if err != nil || !bytes.Equal(data, payload) || hash != "digest" {
				t.Fatalf("data=%v hash=%s err=%v", data, hash, err)
			}
		}
	}
}
