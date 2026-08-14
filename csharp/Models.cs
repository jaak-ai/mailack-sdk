using System.Text.Json;
using System.Text.Json.Serialization;

namespace Mailack;

public sealed class SendRequest
{
    [JsonPropertyName("from")]
    public string From { get; set; } = "";

    [JsonPropertyName("to")]
    public string To { get; set; } = "";

    [JsonPropertyName("subject")]
    public string? Subject { get; set; }

    [JsonPropertyName("text")]
    public string? Text { get; set; }

    [JsonPropertyName("html")]
    public string? Html { get; set; }

    [JsonPropertyName("headers")]
    public Dictionary<string, string>? Headers { get; set; }

    [JsonPropertyName("template_id")]
    public string? TemplateId { get; set; }

    [JsonPropertyName("variables")]
    public Dictionary<string, string>? Variables { get; set; }

    /// <summary>
    /// Omit (leave null) to use the account default (default_certified);
    /// plain messages (certified=false) cannot be sealed.
    /// </summary>
    [JsonPropertyName("certified")]
    public bool? Certified { get; set; }
}

public sealed class BatchItem
{
    [JsonPropertyName("idempotency_key")]
    public string IdempotencyKey { get; set; } = "";

    [JsonPropertyName("from")]
    public string From { get; set; } = "";

    [JsonPropertyName("to")]
    public string To { get; set; } = "";

    [JsonPropertyName("subject")]
    public string Subject { get; set; } = "";

    [JsonPropertyName("text")]
    public string? Text { get; set; }

    [JsonPropertyName("html")]
    public string? Html { get; set; }

    /// <summary>
    /// Omit (leave null) to use the account default (default_certified);
    /// plain messages (certified=false) cannot be sealed.
    /// </summary>
    [JsonPropertyName("certified")]
    public bool? Certified { get; set; }
}

public sealed class SendResult
{
    public required Dictionary<string, object?> Message { get; init; }
    public bool Replay { get; init; }

    public string? Id => Message.TryGetValue("id", out var v) ? v?.ToString() : null;
    public string? State => Message.TryGetValue("state", out var v) ? v?.ToString() : null;

    /// <summary>Whether the message was queued as certified (present in the message response).</summary>
    public bool? Certified => Message.TryGetValue("certified", out var v)
        && v is JsonElement je && je.ValueKind is JsonValueKind.True or JsonValueKind.False
        ? je.GetBoolean() : null;
}

/// <summary>Result of verifying a message against its Merkle proof (POST /v1/verify).</summary>
public sealed class VerifyResult
{
    [JsonPropertyName("valid")]
    public bool Valid { get; init; }

    [JsonPropertyName("merkle_root")]
    public string? MerkleRoot { get; init; }

    [JsonPropertyName("certificate_id")]
    public string? CertificateId { get; init; }

    [JsonPropertyName("sealed_at")]
    public string? SealedAt { get; init; }
}
