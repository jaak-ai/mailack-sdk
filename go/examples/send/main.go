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

	certified := true
	msg, replay, err := client.Send(ctx, "sdk-example-"+time.Now().UTC().Format("20060102T150405"), mailack.SendRequest{
		From:      env("MAILACK_FROM", "noreply@example.com"),
		To:        env("MAILACK_TO", "you@example.com"),
		Subject:   "mailack SDK example",
		Text:      "Hello from the official Go SDK.",
		Certified: &certified, // omit to use the account default (default_certified)
	})
	if err != nil {
		if mailack.Is(err, "domain_not_verified") {
			log.Fatal("verify your From domain in the portal first: ", err)
		}
		log.Fatal(err)
	}
	fmt.Printf("id=%s state=%s hash=%s certified=%v replay=%v\n", msg.ID, msg.State, msg.CanonicalHash, msg.Certified, replay)

	// Seal the message into the Merkle tree, then fetch its evidence and
	// verify the proof. Plain messages (certified=false) fail with 422
	// not_certified / missing_proof_data.
	seal, err := client.Seal(ctx, msg.ID)
	if err != nil {
		log.Fatal("seal: ", err)
	}
	fmt.Printf("sealed: merkle_root=%s certificate=%s\n", seal.MerkleRoot, seal.CertificateID)

	ev, err := client.Evidence(ctx, msg.ID)
	if err != nil {
		log.Fatal("evidence: ", err)
	}
	fmt.Printf("evidence: leaf=%d batch=%s\n", ev.LeafIndex, ev.BatchID)

	vr, err := client.Verify(ctx, msg.ID)
	if err != nil {
		log.Fatal("verify: ", err)
	}
	fmt.Printf("verify: valid=%v\n", vr.Valid)

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
