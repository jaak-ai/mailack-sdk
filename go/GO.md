# mailack Go SDK

Paquete: `github.com/jaak-ai/mailack-sdk/go` (directorio `go/` de este repositorio).

```go
import mailack "github.com/jaak-ai/mailack-sdk/go"

client := mailack.NewClient(baseURL, mailack.WithAPIKey(key))
msg, replay, err := client.Send(ctx, "idem-key", mailack.SendRequest{
    From: "noreply@acme.mx", To: "cliente@x.com",
    Subject: "Hi", Text: "Hello",
})
rates, err := client.Rates(ctx, 14)
```

```bash
go run ./sdk/examples/send
go test ./sdk/
```

Ver también el índice general: [README.md](README.md).
