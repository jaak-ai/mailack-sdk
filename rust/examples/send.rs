//! Example: send one certified message, then seal → evidence → verify.
//!
//! ```bash
//! export MAILACK_API_URL=http://localhost:8080
//! export MAILACK_API_KEY=mlk_…
//! cargo run --example send --manifest-path rust/Cargo.toml
//! ```

use mailack::{Client, SendRequest};
use std::env;
use std::time::{SystemTime, UNIX_EPOCH};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let base = env::var("MAILACK_API_URL").unwrap_or_else(|_| "http://localhost:8080".into());
    let key = env::var("MAILACK_API_KEY").expect("MAILACK_API_KEY is required");

    let client = Client::new(base).with_api_key(key);
    let ts = SystemTime::now()
        .duration_since(UNIX_EPOCH)?
        .as_secs();
    let idem = format!("rust-sdk-{ts}");

    let (msg, replay) = client
        .send(
            &idem,
            &SendRequest {
                from: env::var("MAILACK_FROM").unwrap_or_else(|_| "noreply@example.com".into()),
                to: env::var("MAILACK_TO").unwrap_or_else(|_| "you@example.com".into()),
                subject: Some("mailack Rust SDK example".into()),
                text: Some("Hello from the Rust SDK.".into()),
                certified: Some(true),
                ..Default::default()
            },
        )
        .await
        .map_err(|e| {
            eprintln!("{e}");
            e
        })?;

    println!(
        "id={} state={} hash={} certified={} replay={replay}",
        msg.id, msg.state, msg.canonical_hash, msg.certified
    );

    if msg.certified {
        let seal = client.seal_message(&msg.id).await?;
        println!(
            "sealed: type={} root={} cert={}",
            seal.seal_type, seal.merkle_root, seal.certificate_id
        );

        let ev = client.evidence(&msg.id).await?;
        println!(
            "evidence: mime_sha256={} leaf_index={}",
            ev.mime_sha256, ev.leaf_index
        );

        let res = client.verify(&msg.id).await?;
        println!("verify: valid={}", res.valid);
    }

    if let Ok(rates) = client.rates(7).await {
        println!(
            "rates(7d): delivery={:.1}% bounce={:.1}% ingested={}",
            rates.delivery_rate, rates.bounce_rate, rates.ingested
        );
    }
    Ok(())
}
