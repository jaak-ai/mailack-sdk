# mailack (Rust)

Cliente oficial en Rust para la API de correo certificado de mailack.

## Instalación

```toml
[dependencies]
mailack = { git = "https://github.com/jaak-ai/mailack-sdk.git", branch = "main" }
tokio = { version = "1", features = ["macros", "rt-multi-thread"] }
```

Cargo localiza el paquete `mailack` dentro del repositorio; no hace falta indicar el subdirectorio.

## Uso

```rust
use mailack::{Client, Error, SendRequest};

#[tokio::main]
async fn main() -> Result<(), Error> {
    let client = Client::new("https://api.mailack.com")
        .with_api_key(std::env::var("MAILACK_API_KEY").expect("key"));

    let (msg, replay) = client
        .send(
            "order-42",
            &SendRequest {
                from: "noreply@acme.mx".into(),
                to: "cliente@example.com".into(),
                subject: Some("Recibo".into()),
                text: Some("Gracias por su compra.".into()),
                ..Default::default()
            },
        )
        .await?;

    println!("{} {} replay={replay}", msg.id, msg.state);

    if let Err(e) = client.send("x", &SendRequest::default()).await {
        if e.is_code("quota_exceeded") {
            // cuota mensual
        }
    }
    Ok(())
}
```

### Plantilla

```rust
let (msg, _) = client
    .send(
        "welcome-9",
        &SendRequest {
            from: "hola@acme.mx".into(),
            to: "user@x.com".into(),
            template_id: Some(template_uuid.into()),
            variables: Some([("nombre".into(), "Ada".into())].into()),
            ..Default::default()
        },
    )
    .await?;
```

### Batch

```rust
use mailack::BatchItem;

let batch = client
    .send_batch(&[BatchItem {
        idempotency_key: "a1".into(),
        from: "a@acme.mx".into(),
        to: "1@x.com".into(),
        subject: "Hi".into(),
        text: Some("…".into()),
        html: None,
        certified: None, // omitido: aplica el default de la cuenta (default_certified)
    }])
    .await?;
println!("created={} failed={}", batch.created, batch.failed);
```

### Dominios, webhooks, rates

```rust
let d = client.create_domain("mail.acme.mx").await?;
// publicar TXT d.dns.challenge_host = d.dns.challenge_value
let ok = client.verify_domain(&d.id).await?;

let (hook, secret) = client
    .create_webhook(
        "https://api.suempresa.mx/hooks/mailack",
        &["email.queued", "email.sent", "email.bounced", "email.sealed"],
        Some("ERP"),
    )
    .await?;
// guardar secret (Mailack-Signature)

let rates = client.rates(14).await?;
println!("delivery={:.1}% bounce={:.1}%", rates.delivery_rate, rates.bounce_rate);
```

### Sello, evidencia y verificación

```rust
// Sellar un mensaje certificado (422 not_certified si certified=false).
let seal = client.seal_message(&msg.id).await?;

// Evidencia criptográfica del mensaje (hashes, leaf_index, certificado).
let ev = client.evidence(&msg.id).await?;

// Bundle de prueba Merkle completo como JSON crudo (422 missing_proof_data si no está sellado).
let bundle = client.proof_bundle(&msg.id).await?;

// Verificación por message_id (404 not_found si no existe).
let res = client.verify(&msg.id).await?;
println!("valid={}", res.valid);
```

`SendRequest.certified` y `BatchItem.certified` son `Option<bool>`: omítelos para usar el default de la cuenta (`default_certified`); los mensajes plain (`certified=false`) se entregan igual pero no entran al árbol Merkle y no se pueden sellar.

## Ejemplo

```bash
export MAILACK_API_URL=http://localhost:8080
export MAILACK_API_KEY=mlk_…
cargo run --example send --manifest-path rust/Cargo.toml
```

## Autenticación

API key Bearer (`messages:send`, `evidence:read`, …). El tenant se resuelve en el servidor.

## Dependencias

- `reqwest` (rustls) + `serde` / `serde_json` + `thiserror`
