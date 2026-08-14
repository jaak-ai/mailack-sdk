use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// JSON body for `POST /v1/messages`.
#[derive(Debug, Clone, Default, Serialize)]
pub struct SendRequest {
    pub from: String,
    pub to: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub subject: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub text: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub html: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub headers: Option<HashMap<String, String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub template_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub variables: Option<HashMap<String, String>>,
    /// Omit to use the account default (default_certified); plain messages
    /// (certified=false) cannot be sealed.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub certified: Option<bool>,
}

/// Certified message returned by the API.
#[derive(Debug, Clone, Deserialize)]
pub struct Message {
    pub id: String,
    #[serde(default)]
    pub tenant_id: String,
    #[serde(default)]
    pub idempotency_key: String,
    #[serde(default)]
    pub from_address: String,
    #[serde(default)]
    pub to_address: String,
    #[serde(default)]
    pub subject: String,
    #[serde(default)]
    pub canonical_hash: String,
    #[serde(default)]
    pub state: String,
    #[serde(default)]
    pub batch_id: Option<String>,
    /// True when the message enters the Merkle tree; plain messages deliver
    /// normally but cannot be sealed.
    #[serde(default)]
    pub certified: bool,
}

/// One entry of `SendBatch`.
#[derive(Debug, Clone, Serialize)]
pub struct BatchItem {
    pub idempotency_key: String,
    pub from: String,
    pub to: String,
    pub subject: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub text: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub html: Option<String>,
    /// Omit to use the account default (default_certified); plain messages
    /// (certified=false) cannot be sealed.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub certified: Option<bool>,
}

/// Per-item result of batch ingest.
#[derive(Debug, Clone, Deserialize)]
pub struct BatchItemResult {
    pub index: i32,
    #[serde(default)]
    pub id: Option<String>,
    #[serde(default)]
    pub state: Option<String>,
    #[serde(default)]
    pub replay: bool,
    #[serde(default)]
    pub error: Option<BatchItemError>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct BatchItemError {
    pub code: String,
    pub message: String,
}

/// Response of `POST /v1/messages/{id}/seal` (201).
#[derive(Debug, Clone, Deserialize)]
pub struct SealResult {
    pub message_id: String,
    #[serde(default)]
    pub batch_id: Option<String>,
    #[serde(default)]
    pub seal_type: String,
    #[serde(default)]
    pub canonical_hash: String,
    #[serde(default)]
    pub merkle_root: String,
    #[serde(default)]
    pub certificate_id: String,
    #[serde(default)]
    pub serial_number: String,
    #[serde(default)]
    pub policy_oid: String,
    #[serde(default)]
    pub algorithm_oid: String,
    #[serde(default)]
    pub sealed_at: String,
}

/// Cryptographic evidence of a sealed message (`GET /v1/messages/{id}/evidence`).
#[derive(Debug, Clone, Deserialize)]
pub struct Evidence {
    pub message_id: String,
    #[serde(default)]
    pub canonical_hash: String,
    #[serde(default)]
    pub mime_sha256: String,
    #[serde(default)]
    pub message_id_header: String,
    #[serde(default)]
    pub date_header: String,
    #[serde(default)]
    pub batch_id: Option<String>,
    #[serde(default)]
    pub merkle_root: String,
    #[serde(default)]
    pub sealed_at: String,
    #[serde(default)]
    pub certificate_id: String,
    #[serde(default)]
    pub leaf_index: i64,
}

/// Response of `POST /v1/verify`.
#[derive(Debug, Clone, Deserialize)]
pub struct VerifyResult {
    #[serde(default)]
    pub valid: bool,
    #[serde(default)]
    pub merkle_root: String,
    #[serde(default)]
    pub certificate_id: String,
    #[serde(default)]
    pub sealed_at: String,
}

/// Response of `POST /v1/messages/batch`.
#[derive(Debug, Clone, Deserialize)]
pub struct BatchResult {
    #[serde(default)]
    pub items: Vec<BatchItemResult>,
    #[serde(default)]
    pub created: i32,
    #[serde(default)]
    pub failed: i32,
    #[serde(default)]
    pub replayed: i32,
}

/// Deliverability rates (`GET /v1/rates`).
#[derive(Debug, Clone, Deserialize)]
pub struct Rates {
    pub window_days: i32,
    #[serde(default)]
    pub ingested: i64,
    #[serde(default)]
    pub sent: i64,
    #[serde(default)]
    pub deferred: i64,
    #[serde(default)]
    pub bounced: i64,
    #[serde(default)]
    pub complained: i64,
    #[serde(default)]
    pub sealed: i64,
    #[serde(default)]
    pub delivery_rate: f64,
    #[serde(default)]
    pub bounce_rate: f64,
    #[serde(default)]
    pub complaint_rate: f64,
    #[serde(default)]
    pub days: Vec<DayRates>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct DayRates {
    pub day: String,
    #[serde(default)]
    pub ingested: i64,
    #[serde(default)]
    pub sent: i64,
    #[serde(default)]
    pub deferred: i64,
    #[serde(default)]
    pub bounced: i64,
    #[serde(default)]
    pub complained: i64,
    #[serde(default)]
    pub sealed: i64,
}

/// Sending domain.
#[derive(Debug, Clone, Default, Deserialize)]
pub struct Domain {
    #[serde(default)]
    pub id: String,
    #[serde(default)]
    pub domain: String,
    #[serde(default)]
    pub status: String,
    #[serde(default)]
    pub verification_token: String,
    #[serde(default)]
    pub dns: DomainDns,
}

#[derive(Debug, Clone, Default, Deserialize)]
pub struct DomainDns {
    #[serde(default)]
    pub challenge_host: String,
    #[serde(default)]
    pub challenge_value: String,
    #[serde(default)]
    pub dkim_host: String,
    #[serde(default)]
    pub spf_hint: String,
}

/// Lifecycle webhook endpoint.
#[derive(Debug, Clone, Deserialize)]
pub struct Webhook {
    pub id: String,
    pub url: String,
    #[serde(default)]
    pub description: String,
    #[serde(default)]
    pub secret_suffix: String,
    #[serde(default)]
    pub events: Vec<String>,
    pub status: String,
}

/// Message template with `{{variables}}`.
#[derive(Debug, Clone, Default, Deserialize)]
pub struct Template {
    #[serde(default)]
    pub id: String,
    #[serde(default)]
    pub name: String,
    #[serde(default)]
    pub subject: String,
    #[serde(default)]
    pub text: String,
    #[serde(default)]
    pub html: String,
    #[serde(default)]
    pub status: String,
    #[serde(default)]
    pub variables: Vec<String>,
}

#[derive(Debug, Deserialize)]
pub(crate) struct Items<T> {
    #[serde(default)]
    pub items: Vec<T>,
}

#[derive(Debug, Deserialize)]
pub(crate) struct WebhooksResponse {
    #[serde(default)]
    pub items: Vec<Webhook>,
    #[serde(default)]
    pub available_events: Vec<String>,
}

#[derive(Debug, Deserialize)]
pub(crate) struct CreateWebhookResponse {
    pub webhook: Webhook,
    #[serde(default)]
    pub secret: String,
}

#[derive(Debug, Deserialize)]
pub(crate) struct VerifyDomainResponse {
    #[serde(default)]
    pub verified: bool,
}

#[derive(Debug, Deserialize)]
pub(crate) struct MessageWrap {
    pub message: Message,
}

#[derive(Debug, Deserialize)]
pub(crate) struct ApiErrorBody {
    pub error: ApiErrorInner,
}

#[derive(Debug, Deserialize)]
pub(crate) struct ApiErrorInner {
    #[serde(default)]
    pub code: String,
    #[serde(default)]
    pub message: String,
}
