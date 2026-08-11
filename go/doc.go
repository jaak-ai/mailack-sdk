// Package mailack is the official Go SDK for the mailack certified email API.
//
// It is a thin HTTP client over the public REST surface. The canonical
// contract is the published OpenAPI document: where the two disagree, the
// OpenAPI wins and the SDK has a bug.
//
// # Quick start
//
//	client := mailack.NewClient("https://api.mailack.com", mailack.WithAPIKey(os.Getenv("MAILACK_API_KEY")))
//	msg, replay, err := client.Send(ctx, "idem-key-1", mailack.SendRequest{
//	    From: "noreply@acme.mx",
//	    To:   "cliente@example.com",
//	    Subject: "Estado de cuenta",
//	    Text: "Tu estado de cuenta está listo.",
//	})
//
// Authentication: set WithAPIKey (Bearer). The X-Tenant-ID header is derived
// from the key on the server; do not pass tenant IDs from the client.
//
// See sdk/README.md and the examples/ directory for fuller walkthroughs.
package mailack
