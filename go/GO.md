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

## Attachments (Send only)

`SendRequest.Attachments` mirrors the API `attachments` field on `POST /v1/messages`:

```go
msg, _, err := client.Send(ctx, "idem-key", mailack.SendRequest{
    From: "noreply@acme.mx", To: "cliente@x.com",
    Subject: "Documento firmado", Text: "Adjunto el PDF.",
    Attachments: []mailack.Attachment{
        mailack.NewAttachment("documento.pdf", "application/pdf", pdfBytes),
    },
})
```

- Each item: `filename` (required), `content_type` (optional; API default `application/octet-stream`), `content` (standard base64 of the file bytes).
- Limits: max **25** attachments; the hard bound is the assembled message size (**25 MiB**), which includes base64 overhead.
- The API has **no Cc/Bcc** (`to` is a single address string).
- `SendBatch` / `BatchItem` do **not** accept attachments (server `batchMessageItem` has no field); use `Send` for messages with files.
- Omit the field (or leave it nil/empty) to keep wire-compatible payloads with older clients.

```bash
go run ./sdk/examples/send
go test ./sdk/
```

Ver también el índice general: [README.md](README.md).

## Descarga RAW

```go
raw, err := client.GetMessageRaw(ctx, "message-id")
if err != nil { return err }
fmt.Println(raw.CanonicalHash, len(raw.Data))
event, err := client.GetEventRaw(ctx, "message-id", "event-id")
if err != nil { return err }
fmt.Println(event.RawSHA256, len(event.Data))
```

El cuerpo se conserva como bytes. El hash procede de `X-Mailack-Canonical-Hash`
para mensajes y `X-Mailack-Raw-SHA256` para eventos; si falta, se devuelve una cadena vacía.
