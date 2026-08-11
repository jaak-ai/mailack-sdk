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
}

public sealed class SendResult
{
    public required Dictionary<string, object?> Message { get; init; }
    public bool Replay { get; init; }

    public string? Id => Message.TryGetValue("id", out var v) ? v?.ToString() : null;
    public string? State => Message.TryGetValue("state", out var v) ? v?.ToString() : null;
}
