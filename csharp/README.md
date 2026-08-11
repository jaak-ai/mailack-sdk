# Mailack.Sdk (C# / .NET 8)

Cliente oficial en C# para la API de correo certificado de mailack.

## Instalación

```bash
dotnet add reference path/to/mailack-sdk/csharp/Mailack.Sdk.csproj
# o empaquetar:
dotnet pack csharp/Mailack.Sdk.csproj -c Release
```

Solo usa `HttpClient` y `System.Text.Json` (BCL).

## Uso

```csharp
using Mailack;

using var client = new Client(
    Environment.GetEnvironmentVariable("MAILACK_API_URL") ?? "https://api.mailack.com",
    Environment.GetEnvironmentVariable("MAILACK_API_KEY")!);

try
{
    var result = await client.SendAsync("order-42", new SendRequest
    {
        From = "noreply@acme.mx",
        To = "cliente@example.com",
        Subject = "Recibo",
        Text = "Gracias por su compra.",
    });
    Console.WriteLine($"{result.Id} {result.State} replay={result.Replay}");
}
catch (ApiError e) when (e.Is("quota_exceeded"))
{
    // …
}
```

### Plantilla

```csharp
await client.SendAsync("welcome-9", new SendRequest
{
    From = "hola@acme.mx",
    To = "user@x.com",
    TemplateId = templateUuid,
    Variables = new Dictionary<string, string> { ["nombre"] = "Ada" },
});
```

### Batch / domains / webhooks / rates

```csharp
await client.SendBatchAsync(new[]
{
    new BatchItem { IdempotencyKey = "a1", From = "a@acme.mx", To = "1@x.com", Subject = "Hi", Text = "…" },
});

var domain = await client.CreateDomainAsync("mail.acme.mx");
await client.VerifyDomainAsync(domain["id"]!.ToString()!);

var hook = await client.CreateWebhookAsync(
    "https://api.suempresa.mx/hooks/mailack",
    new[] { "email.queued", "email.sent", "email.bounced", "email.sealed" });
// hook["secret"] una sola vez

var rates = await client.RatesAsync(14);
```

## Ejemplo

```bash
export MAILACK_API_URL=http://localhost:8080
export MAILACK_API_KEY=mlk_…
dotnet run --project csharp/examples/SendExample
```
