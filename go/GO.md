# mailack Go SDK

Paquete: `github.com/jaak-ai/mailack-sdk/go` (directorio `go/` de este repositorio).

```go
import mailack "github.com/jaak-ai/mailack-sdk/go"

client := mailack.NewClient(baseURL, mailack.WithAPIKey(key))
msg, replay, err := client.Send(ctx, "idem-key", mailack.SendRequest{
    From: "noreply@acme.mx", To: "cliente@x.com",
    Subject: "Hi", Text: "Hello",
    Certified: &certified, // *bool opcional: omite para usar default_certified de la cuenta
})
rates, err := client.Rates(ctx, 14)
```

Capacidades del cliente:

- `Send` / `SendBatch` — ingesta de mensajes (flag opcional `certified` por mensaje; los plain no se sellan).
- `GetMessage` — un mensaje por ID.
- `Seal` — sella un mensaje certificado en el árbol Merkle (`POST /v1/messages/{id}/seal`).
- `Evidence` — expediente de evidencia de un mensaje sellado (`GET /v1/messages/{id}/evidence`).
- `ProofBundle` — bundle de prueba Merkle en JSON crudo (`GET /v1/messages/{id}/proof-bundle`).
- `Verify` — verifica la prueba Merkle por `message_id` (`POST /v1/verify`).
- `Rates` — contadores de entregabilidad.

```bash
go run ./sdk/examples/send
go test ./sdk/
```

Ver también el índice general: [README.md](README.md).
