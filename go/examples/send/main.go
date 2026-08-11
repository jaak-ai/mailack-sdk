// Example: send one certified message with the mailack Go SDK.
//
//	export MAILACK_API_URL=http://localhost:8080
//	export MAILACK_API_KEY=mlk_…
//	go run ./sdk/examples/send
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	mailack "github.com/jaak-ai/mailack-sdk/go"
)

func main() {
	base := env("MAILACK_API_URL", "http://localhost:8080")
	key := os.Getenv("MAILACK_API_KEY")
	if key == "" {
		log.Fatal("MAILACK_API_KEY is required")
	}

	client := mailack.NewClient(base, mailack.WithAPIKey(key))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	msg, replay, err := client.Send(ctx, "sdk-example-"+time.Now().UTC().Format("20060102T150405"), mailack.SendRequest{
		From:    env("MAILACK_FROM", "noreply@example.com"),
		To:      env("MAILACK_TO", "you@example.com"),
		Subject: "mailack SDK example",
		Text:    "Hello from the official Go SDK.",
	})
	if err != nil {
		if mailack.Is(err, "domain_not_verified") {
			log.Fatal("verify your From domain in the portal first: ", err)
		}
		log.Fatal(err)
	}
	fmt.Printf("id=%s state=%s hash=%s replay=%v\n", msg.ID, msg.State, msg.CanonicalHash, replay)

	if rates, err := client.Rates(ctx, 7); err == nil {
		fmt.Printf("rates(7d): delivery=%.1f%% bounce=%.1f%% ingested=%d\n",
			rates.DeliveryRate, rates.BounceRate, rates.Ingested)
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
