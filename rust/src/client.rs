use crate::error::{Error, Result};
use crate::types::*;
use reqwest::header::{HeaderMap, HeaderValue, AUTHORIZATION, CONTENT_TYPE};
use reqwest::{Client as HttpClient, Method, StatusCode};
use serde::de::DeserializeOwned;
use serde::Serialize;
use std::time::Duration;

/// HTTP client for one mailack API deployment.
#[derive(Clone)]
pub struct Client {
    base_url: String,
    api_key: String,
    http: HttpClient,
}

impl Client {
    /// Build a client for `base_url` (e.g. `https://api.mailack.com`).
    pub fn new(base_url: impl Into<String>) -> Self {
        let http = HttpClient::builder()
            .timeout(Duration::from_secs(30))
            .user_agent("mailack-rust/0.1.0")
            .build()
            .expect("reqwest client");
        Self {
            base_url: base_url.into().trim_end_matches('/').to_string(),
            api_key: String::new(),
            http,
        }
    }

    /// Set the Bearer API key (`mlk_…`).
    pub fn with_api_key(mut self, key: impl Into<String>) -> Self {
        self.api_key = key.into().trim().to_string();
        self
    }

    /// POST /v1/messages — returns `(message, replay)`.
    pub async fn send(&self, idempotency_key: &str, req: &SendRequest) -> Result<(Message, bool)> {
        let (status, headers, body) = self
            .request(Method::POST, "/v1/messages", Some(req), Some(idempotency_key), None)
            .await?;
        ensure_ok(status, &body)?;
        let msg: Message = serde_json::from_str(&body).map_err(|e| Error::Decode(e.to_string()))?;
        let replay = headers
            .get("idempotent-replay")
            .and_then(|v| v.to_str().ok())
            .map(|s| s.eq_ignore_ascii_case("true"))
            .unwrap_or(false);
        Ok((msg, replay))
    }

    /// POST /v1/messages/batch (max 100).
    pub async fn send_batch(&self, messages: &[BatchItem]) -> Result<BatchResult> {
        #[derive(Serialize)]
        struct Body<'a> {
            messages: &'a [BatchItem],
        }
        self.get_json(
            Method::POST,
            "/v1/messages/batch",
            Some(&Body { messages }),
            None,
            None,
        )
        .await
    }

    /// GET /v1/messages/{id}
    pub async fn get_message(&self, id: &str) -> Result<Message> {
        let path = format!("/v1/messages/{id}");
        let wrap: MessageWrap = self.get_json(Method::GET, &path, None::<&()>, None, None).await?;
        Ok(wrap.message)
    }

    /// GET /v1/rates?days=
    pub async fn rates(&self, days: u32) -> Result<Rates> {
        let q = [("days", days.to_string())];
        self.get_json(Method::GET, "/v1/rates", None::<&()>, None, Some(&q))
            .await
    }

    /// GET /v1/domains
    pub async fn list_domains(&self) -> Result<Vec<Domain>> {
        let items: Items<Domain> = self
            .get_json(Method::GET, "/v1/domains", None::<&()>, None, None)
            .await?;
        Ok(items.items)
    }

    /// POST /v1/domains
    pub async fn create_domain(&self, domain: &str) -> Result<Domain> {
        #[derive(Serialize)]
        struct Body<'a> {
            domain: &'a str,
        }
        self.get_json(
            Method::POST,
            "/v1/domains",
            Some(&Body { domain }),
            None,
            None,
        )
        .await
    }

    /// POST /v1/domains/{id}/verify
    pub async fn verify_domain(&self, id: &str) -> Result<bool> {
        let path = format!("/v1/domains/{id}/verify");
        let res: VerifyDomainResponse = self
            .get_json(Method::POST, &path, Some(&serde_json::json!({})), None, None)
            .await?;
        Ok(res.verified)
    }

    /// GET /v1/webhooks — returns `(hooks, available_events)`.
    pub async fn list_webhooks(&self) -> Result<(Vec<Webhook>, Vec<String>)> {
        let res: WebhooksResponse = self
            .get_json(Method::GET, "/v1/webhooks", None::<&()>, None, None)
            .await?;
        Ok((res.items, res.available_events))
    }

    /// POST /v1/webhooks — returns `(webhook, secret)` (secret once).
    pub async fn create_webhook(
        &self,
        url: &str,
        events: &[&str],
        description: Option<&str>,
    ) -> Result<(Webhook, String)> {
        #[derive(Serialize)]
        struct Body<'a> {
            url: &'a str,
            events: &'a [&'a str],
            #[serde(skip_serializing_if = "Option::is_none")]
            description: Option<&'a str>,
        }
        let res: CreateWebhookResponse = self
            .get_json(
                Method::POST,
                "/v1/webhooks",
                Some(&Body {
                    url,
                    events,
                    description,
                }),
                None,
                None,
            )
            .await?;
        Ok((res.webhook, res.secret))
    }

    /// DELETE /v1/webhooks/{id}
    pub async fn disable_webhook(&self, id: &str) -> Result<()> {
        let path = format!("/v1/webhooks/{id}");
        let (status, _, body) = self
            .request(Method::DELETE, &path, None::<&()>, None, None)
            .await?;
        ensure_ok(status, &body)?;
        Ok(())
    }

    /// GET /v1/templates
    pub async fn list_templates(&self) -> Result<Vec<Template>> {
        let items: Items<Template> = self
            .get_json(Method::GET, "/v1/templates", None::<&()>, None, None)
            .await?;
        Ok(items.items)
    }

    /// POST /v1/templates
    pub async fn create_template(
        &self,
        name: &str,
        subject: &str,
        text: &str,
        html: &str,
    ) -> Result<Template> {
        #[derive(Serialize)]
        struct Body<'a> {
            name: &'a str,
            subject: &'a str,
            text: &'a str,
            html: &'a str,
        }
        self.get_json(
            Method::POST,
            "/v1/templates",
            Some(&Body {
                name,
                subject,
                text,
                html,
            }),
            None,
            None,
        )
        .await
    }

    async fn get_json<T: DeserializeOwned, B: Serialize>(
        &self,
        method: Method,
        path: &str,
        body: Option<&B>,
        idempotency_key: Option<&str>,
        query: Option<&[(&str, String)]>,
    ) -> Result<T> {
        let (status, _, text) = self
            .request(method, path, body, idempotency_key, query)
            .await?;
        ensure_ok(status, &text)?;
        serde_json::from_str(&text).map_err(|e| Error::Decode(e.to_string()))
    }

    async fn request<B: Serialize>(
        &self,
        method: Method,
        path: &str,
        body: Option<&B>,
        idempotency_key: Option<&str>,
        query: Option<&[(&str, String)]>,
    ) -> Result<(StatusCode, HeaderMap, String)> {
        let mut url = format!("{}{}", self.base_url, path);
        if let Some(q) = query {
            let mut first = true;
            for (k, v) in q {
                url.push(if first { '?' } else { '&' });
                first = false;
                url.push_str(&urlencoding_pair(k, v));
            }
        }

        let mut req = self.http.request(method, &url);
        if let Some(key) = idempotency_key {
            req = req.header("Idempotency-Key", key);
        }
        if !self.api_key.is_empty() {
            let val = format!("Bearer {}", self.api_key);
            req = req.header(AUTHORIZATION, HeaderValue::from_str(&val).unwrap());
        }
        if let Some(b) = body {
            req = req.header(CONTENT_TYPE, "application/json").json(b);
        }

        let resp = req.send().await?;
        let status = resp.status();
        let headers = resp.headers().clone();
        let text = resp.text().await?;
        Ok((status, headers, text))
    }
}

fn ensure_ok(status: StatusCode, body: &str) -> Result<()> {
    if status.is_success() {
        return Ok(());
    }
    if let Ok(env) = serde_json::from_str::<ApiErrorBody>(body) {
        return Err(Error::Api {
            status: status.as_u16(),
            code: env.error.code,
            message: env.error.message,
        });
    }
    Err(Error::Api {
        status: status.as_u16(),
        code: "http_error".into(),
        message: body.to_string(),
    })
}

fn urlencoding_pair(k: &str, v: &str) -> String {
    // minimal encoding for query values (days=N etc.)
    format!(
        "{}={}",
        k,
        v.replace(' ', "%20")
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn error_is_code() {
        let e = Error::Api {
            status: 429,
            code: "quota_exceeded".into(),
            message: "quota".into(),
        };
        assert!(e.is_code("quota_exceeded"));
        assert!(!e.is_code("other"));
    }

    #[test]
    fn client_trims_base_url() {
        let c = Client::new("http://localhost:8080/");
        assert_eq!(c.base_url, "http://localhost:8080");
    }
}
