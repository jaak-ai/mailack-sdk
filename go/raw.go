package mailack

import (
	"context"
	"io"
	"net/http"
)

type MessageRaw struct {
	Data          []byte
	CanonicalHash string
}

type EventRaw struct {
	Data      []byte
	RawSHA256 string
}

func (c *Client) GetMessageRaw(ctx context.Context, id string) (*MessageRaw, error) {
	data, hash, err := c.getRaw(ctx, "/v1/messages/"+id+"/raw", "X-Mailack-Canonical-Hash")
	if err != nil {
		return nil, err
	}
	return &MessageRaw{Data: data, CanonicalHash: hash}, nil
}

func (c *Client) GetEventRaw(ctx context.Context, messageID, eventID string) (*EventRaw, error) {
	data, hash, err := c.getRaw(ctx, "/v1/messages/"+messageID+"/events/"+eventID+"/raw", "X-Mailack-Raw-SHA256")
	if err != nil {
		return nil, err
	}
	return &EventRaw{Data: data, RawSHA256: hash}, nil
}

func (c *Client) getRaw(ctx context.Context, path, header string) ([]byte, string, error) {
	resp, err := c.do(ctx, http.MethodGet, path, "", "", nil)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if err := checkStatus(resp); err != nil {
		return nil, "", err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return data, resp.Header.Get(header), nil
}
