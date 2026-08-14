# mailack Python SDK

Cliente oficial en Python para la API de correo certificado de mailack.

## Instalación

```bash
pip install "git+https://github.com/jaak-ai/mailack-sdk.git#subdirectory=python"
```

O clonando el repositorio:

```bash
git clone https://github.com/jaak-ai/mailack-sdk.git
pip install -e ./mailack-sdk/python
```

Solo usa la biblioteca estándar (`urllib`); sin dependencias de red externas.

## Uso rápido

```python
from mailack import Client, APIError

client = Client("https://api.mailack.com", api_key="mlk_…")

try:
    msg, replay = client.send(
        "order-42",
        from_="noreply@acme.mx",
        to="cliente@example.com",
        subject="Recibo",
        text="Gracias por su compra.",
    )
    print(msg["id"], msg["state"], msg["canonical_hash"], "replay=", replay)
except APIError as e:
    if e.is_code("quota_exceeded"):
        ...
    raise
```

### Batch

```python
result = client.send_batch([
    {"idempotency_key": "a1", "from": "a@acme.mx", "to": "1@x.com", "subject": "Hi", "text": "…"},
    {"idempotency_key": "a2", "from": "a@acme.mx", "to": "2@x.com", "subject": "Hi", "text": "…"},
])
print(result["created"], result["failed"])
```

### Plantilla

```python
msg, _ = client.send(
    "welcome-9",
    from_="hola@acme.mx",
    to="user@x.com",
    template_id="<uuid>",
    variables={"nombre": "Ada"},
)
```

### Dominios, webhooks, rates

```python
d = client.create_domain("mail.acme.mx")
# publicar TXT d["dns"]["challenge_host"] = d["dns"]["challenge_value"]
client.verify_domain(d["id"])

hook, secret = client.create_webhook(
    "https://api.suempresa.mx/hooks/mailack",
    ["email.queued", "email.sent", "email.bounced", "email.sealed"],
)
rates = client.rates(days=14)
print(rates["delivery_rate"], rates["bounce_rate"])
```

### Sellado y verificación

```python
msg, _ = client.send("cert-1", from_="a@acme.mx", to="b@x.com",
                     subject="Hi", text="…", certified=True)  # None = default de la cuenta

seal = client.seal_message(msg["id"])            # POST /v1/messages/{id}/seal → recibo de sellado
evidence = client.message_evidence(msg["id"])    # GET  /v1/messages/{id}/evidence → registro de evidencia
bundle = client.proof_bundle(msg["id"])          # GET  /v1/messages/{id}/proof-bundle → documento crudo
result = client.verify_message(msg["id"])        # POST /v1/verify → {"valid": True, …}
```

Los mensajes plain (`certified=False`) entregan igual pero no entran al árbol Merkle y no se pueden sellar (`422 not_certified`).

## Autenticación

API key Bearer (`messages:send`, `evidence:read`, …). El tenant se resuelve en el servidor.

## Contrato

El contrato canónico es el OpenAPI publicado de Mailack; el SDK de Go en [`../go`](../go/) expone el mismo modelo.
