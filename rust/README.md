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
