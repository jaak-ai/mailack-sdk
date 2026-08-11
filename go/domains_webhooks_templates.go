package mailack

import (
	"context"
	"net/http"
)

// Domain is a tenant sending domain.
type Domain struct {
	ID                string `json:"id"`
	Domain            string `json:"domain"`
	Status            string `json:"status"`
	VerificationToken string `json:"verification_token"`
	DNS               struct {
		ChallengeHost  string `json:"challenge_host"`
		ChallengeValue string `json:"challenge_value"`
		DKIMHost       string `json:"dkim_host"`
		SPFHint        string `json:"spf_hint"`
	} `json:"dns"`
}

// ListDomains returns registered sending domains.
func (c *Client) ListDomains(ctx context.Context) ([]Domain, error) {
	resp, err := c.do(ctx, http.MethodGet, "/v1/domains", "", "", nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Items []Domain `json:"items"`
	}
	if err := decodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// CreateDomain registers a From FQDN and returns DNS challenge records.
func (c *Client) CreateDomain(ctx context.Context, domain string) (*Domain, error) {
	resp, err := c.do(ctx, http.MethodPost, "/v1/domains", "", "application/json", map[string]string{"domain": domain})
	if err != nil {
		return nil, err
	}
	var d Domain
	if err := decodeJSON(resp, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// VerifyDomain checks the DNS challenge and updates status.
func (c *Client) VerifyDomain(ctx context.Context, id string) (verified bool, err error) {
	resp, err := c.do(ctx, http.MethodPost, "/v1/domains/"+id+"/verify", "", "application/json", map[string]any{})
	if err != nil {
		return false, err
	}
	var out struct {
		Verified bool `json:"verified"`
	}
	if err := decodeJSON(resp, &out); err != nil {
		return false, err
	}
	return out.Verified, nil
}

// Webhook is a lifecycle HTTPS endpoint.
type Webhook struct {
	ID           string   `json:"id"`
	URL          string   `json:"url"`
	Description  string   `json:"description,omitempty"`
	SecretSuffix string   `json:"secret_suffix"`
	Events       []string `json:"events"`
	Status       string   `json:"status"`
}

// ListWebhooks returns endpoints and the available event catalog.
func (c *Client) ListWebhooks(ctx context.Context) (hooks []Webhook, events []string, err error) {
	resp, err := c.do(ctx, http.MethodGet, "/v1/webhooks", "", "", nil)
	if err != nil {
		return nil, nil, err
	}
	var out struct {
		Items           []Webhook `json:"items"`
		AvailableEvents []string  `json:"available_events"`
	}
	if err := decodeJSON(resp, &out); err != nil {
		return nil, nil, err
	}
	return out.Items, out.AvailableEvents, nil
}

// CreateWebhook registers an HTTPS endpoint. secret is shown only once.
func (c *Client) CreateWebhook(ctx context.Context, hookURL string, events []string, description string) (hook *Webhook, secret string, err error) {
	body := map[string]any{"url": hookURL, "events": events}
	if description != "" {
		body["description"] = description
	}
	resp, err := c.do(ctx, http.MethodPost, "/v1/webhooks", "", "application/json", body)
	if err != nil {
		return nil, "", err
	}
	var out struct {
		Webhook Webhook `json:"webhook"`
		Secret  string  `json:"secret"`
	}
	if err := decodeJSON(resp, &out); err != nil {
		return nil, "", err
	}
	return &out.Webhook, out.Secret, nil
}

// DisableWebhook stops deliveries to the endpoint.
func (c *Client) DisableWebhook(ctx context.Context, id string) error {
	resp, err := c.do(ctx, http.MethodDelete, "/v1/webhooks/"+id, "", "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp)
}

// Template is a reusable subject/body with {{variables}}.
type Template struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Subject   string   `json:"subject"`
	Text      string   `json:"text,omitempty"`
	HTML      string   `json:"html,omitempty"`
	Status    string   `json:"status"`
	Variables []string `json:"variables"`
}

// ListTemplates returns active message templates.
func (c *Client) ListTemplates(ctx context.Context) ([]Template, error) {
	resp, err := c.do(ctx, http.MethodGet, "/v1/templates", "", "", nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Items []Template `json:"items"`
	}
	if err := decodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// CreateTemplate stores a new active template.
func (c *Client) CreateTemplate(ctx context.Context, name, subject, text, html string) (*Template, error) {
	resp, err := c.do(ctx, http.MethodPost, "/v1/templates", "", "application/json", map[string]string{
		"name": name, "subject": subject, "text": text, "html": html,
	})
	if err != nil {
		return nil, err
	}
	var t Template
	if err := decodeJSON(resp, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
