//! Official Rust SDK for the mailack certified email API.
//!
//! # Quick start
//!
//! ```no_run
//! use mailack::{Client, SendRequest};
//!
//! # async fn run() -> Result<(), mailack::Error> {
//! let client = Client::new("https://api.mailack.com")
//!     .with_api_key(std::env::var("MAILACK_API_KEY").unwrap());
//! let (msg, replay) = client
//!     .send(
//!         "idem-1",
//!         &SendRequest {
//!             from: "noreply@acme.mx".into(),
//!             to: "cliente@example.com".into(),
//!             subject: Some("Recibo".into()),
//!             text: Some("Gracias.".into()),
//!             ..Default::default()
//!         },
//!     )
//!     .await?;
//! println!("{} {} replay={replay}", msg.id, msg.state);
//! # Ok(())
//! # }
//! ```
//!
//! Authentication uses a Bearer API key (`mlk_…`). The tenant is resolved
//! server-side from the key.

mod client;
mod error;
mod types;

pub use client::Client;
pub use error::{Error, Result};
pub use types::*;
