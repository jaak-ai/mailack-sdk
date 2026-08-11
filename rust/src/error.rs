use thiserror::Error;

/// Result alias for the SDK.
pub type Result<T> = std::result::Result<T, Error>;

/// Errors returned by the SDK.
#[derive(Debug, Error)]
pub enum Error {
    /// HTTP transport or I/O failure.
    #[error("http: {0}")]
    Http(#[from] reqwest::Error),

    /// Non-2xx API response with a decoded envelope.
    #[error("mailack: HTTP {status} {code}: {message}")]
    Api {
        status: u16,
        code: String,
        message: String,
    },

    /// Response body was not valid JSON / unexpected shape.
    #[error("decode: {0}")]
    Decode(String),
}

impl Error {
    /// True when this is an API error with the given machine code.
    pub fn is_code(&self, code: &str) -> bool {
        matches!(self, Error::Api { code: c, .. } if c == code)
    }
}
