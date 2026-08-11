using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;

namespace Mailack;

/// <summary>
/// Official C# client for the mailack certified email API.
/// </summary>
/// <example>
/// <code>
/// var client = new Client("https://api.mailack.com", Environment.GetEnvironmentVariable("MAILACK_API_KEY")!);
/// var result = await client.SendAsync("idem-1", new SendRequest {
///     From = "noreply@acme.mx", To = "cliente@example.com",
///     Subject = "Recibo", Text = "Gracias." });
/// </code>
/// </example>
public sealed class Client : IDisposable
{
    private static readonly JsonSerializerOptions JsonOpts = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
        DefaultIgnoreCondition = System.Text.Json.Serialization.JsonIgnoreCondition.WhenWritingNull,
        PropertyNameCaseInsensitive = true,
    };

    private readonly HttpClient _http;
    private readonly bool _ownsHttp;

    public Client(string baseUrl, string apiKey, HttpClient? http = null)
    {
        _ownsHttp = http is null;
        _http = http ?? new HttpClient { Timeout = TimeSpan.FromSeconds(30) };
        _http.BaseAddress = new Uri(baseUrl.TrimEnd('/') + "/");
        if (!string.IsNullOrWhiteSpace(apiKey))
        {
            _http.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", apiKey.Trim());
        }
        _http.DefaultRequestHeaders.Accept.Add(new MediaTypeWithQualityHeaderValue("application/json"));
        _http.DefaultRequestHeaders.UserAgent.ParseAdd("mailack-csharp/0.1.0");
    }

    public async Task<SendResult> SendAsync(string idempotencyKey, SendRequest req, CancellationToken ct = default)
    {
        using var msg = new HttpRequestMessage(HttpMethod.Post, "v1/messages");
        msg.Headers.TryAddWithoutValidation("Idempotency-Key", idempotencyKey);
        msg.Content = JsonContent(req);
        using var resp = await _http.SendAsync(msg, ct).ConfigureAwait(false);
        var dict = await ReadObjectAsync(resp, ct).ConfigureAwait(false);
        var replay = resp.Headers.TryGetValues("Idempotent-Replay", out var vals)
            && vals.Any(v => v.Equals("true", StringComparison.OrdinalIgnoreCase));
        return new SendResult { Message = dict, Replay = replay };
    }

    public async Task<Dictionary<string, object?>> SendBatchAsync(IEnumerable<BatchItem> items, CancellationToken ct = default)
    {
        using var resp = await _http.PostAsync("v1/messages/batch",
            JsonContent(new { messages = items }), ct).ConfigureAwait(false);
        return await ReadObjectAsync(resp, ct).ConfigureAwait(false);
    }

    public async Task<Dictionary<string, object?>> GetMessageAsync(string id, CancellationToken ct = default)
    {
        using var resp = await _http.GetAsync($"v1/messages/{id}", ct).ConfigureAwait(false);
        var root = await ReadObjectAsync(resp, ct).ConfigureAwait(false);
        if (root.TryGetValue("message", out var m) && m is JsonElement je && je.ValueKind == JsonValueKind.Object)
        {
            return JsonSerializer.Deserialize<Dictionary<string, object?>>(je.GetRawText(), JsonOpts) ?? root;
        }
        return root;
    }

    public async Task<Dictionary<string, object?>> RatesAsync(int days = 14, CancellationToken ct = default)
    {
        using var resp = await _http.GetAsync($"v1/rates?days={days}", ct).ConfigureAwait(false);
        return await ReadObjectAsync(resp, ct).ConfigureAwait(false);
    }

    public async Task<List<Dictionary<string, object?>>> ListDomainsAsync(CancellationToken ct = default)
    {
        using var resp = await _http.GetAsync("v1/domains", ct).ConfigureAwait(false);
        var root = await ReadObjectAsync(resp, ct).ConfigureAwait(false);
        return ReadItems(root);
    }

    public async Task<Dictionary<string, object?>> CreateDomainAsync(string domain, CancellationToken ct = default)
    {
        using var resp = await _http.PostAsync("v1/domains", JsonContent(new { domain }), ct).ConfigureAwait(false);
        return await ReadObjectAsync(resp, ct).ConfigureAwait(false);
    }

    public async Task<Dictionary<string, object?>> VerifyDomainAsync(string id, CancellationToken ct = default)
    {
        using var resp = await _http.PostAsync($"v1/domains/{id}/verify", JsonContent(new { }), ct).ConfigureAwait(false);
        return await ReadObjectAsync(resp, ct).ConfigureAwait(false);
    }

    public async Task<Dictionary<string, object?>> ListWebhooksAsync(CancellationToken ct = default)
    {
        using var resp = await _http.GetAsync("v1/webhooks", ct).ConfigureAwait(false);
        return await ReadObjectAsync(resp, ct).ConfigureAwait(false);
    }

    public async Task<Dictionary<string, object?>> CreateWebhookAsync(
        string url, IEnumerable<string> events, string? description = null, CancellationToken ct = default)
    {
        var body = new Dictionary<string, object?> { ["url"] = url, ["events"] = events };
        if (!string.IsNullOrEmpty(description)) body["description"] = description;
        using var resp = await _http.PostAsync("v1/webhooks", JsonContent(body), ct).ConfigureAwait(false);
        return await ReadObjectAsync(resp, ct).ConfigureAwait(false);
    }

    public async Task DisableWebhookAsync(string id, CancellationToken ct = default)
    {
        using var resp = await _http.DeleteAsync($"v1/webhooks/{id}", ct).ConfigureAwait(false);
        await EnsureSuccessAsync(resp, ct).ConfigureAwait(false);
    }

    public async Task<List<Dictionary<string, object?>>> ListTemplatesAsync(CancellationToken ct = default)
    {
        using var resp = await _http.GetAsync("v1/templates", ct).ConfigureAwait(false);
        var root = await ReadObjectAsync(resp, ct).ConfigureAwait(false);
        return ReadItems(root);
    }

    public async Task<Dictionary<string, object?>> CreateTemplateAsync(
        string name, string subject, string? text = null, string? html = null, CancellationToken ct = default)
    {
        using var resp = await _http.PostAsync("v1/templates", JsonContent(new
        {
            name,
            subject,
            text = text ?? "",
            html = html ?? "",
        }), ct).ConfigureAwait(false);
        return await ReadObjectAsync(resp, ct).ConfigureAwait(false);
    }

    private static StringContent JsonContent(object body) =>
        new(JsonSerializer.Serialize(body, JsonOpts), Encoding.UTF8, "application/json");

    private static async Task EnsureSuccessAsync(HttpResponseMessage resp, CancellationToken ct)
    {
        if (resp.IsSuccessStatusCode) return;
        var raw = await resp.Content.ReadAsStringAsync(ct).ConfigureAwait(false);
        var code = "http_error";
        var message = raw;
        try
        {
            using var doc = JsonDocument.Parse(raw);
            if (doc.RootElement.TryGetProperty("error", out var err))
            {
                if (err.TryGetProperty("code", out var c)) code = c.GetString() ?? code;
                if (err.TryGetProperty("message", out var m)) message = m.GetString() ?? message;
            }
        }
        catch { /* keep defaults */ }
        throw new ApiError((int)resp.StatusCode, code, message);
    }

    private static async Task<Dictionary<string, object?>> ReadObjectAsync(HttpResponseMessage resp, CancellationToken ct)
    {
        await EnsureSuccessAsync(resp, ct).ConfigureAwait(false);
        var raw = await resp.Content.ReadAsStringAsync(ct).ConfigureAwait(false);
        if (string.IsNullOrWhiteSpace(raw)) return new Dictionary<string, object?>();
        return JsonSerializer.Deserialize<Dictionary<string, object?>>(raw, JsonOpts)
               ?? new Dictionary<string, object?>();
    }

    private static List<Dictionary<string, object?>> ReadItems(Dictionary<string, object?> root)
    {
        if (!root.TryGetValue("items", out var items) || items is not JsonElement je || je.ValueKind != JsonValueKind.Array)
            return new List<Dictionary<string, object?>>();
        var list = new List<Dictionary<string, object?>>();
        foreach (var el in je.EnumerateArray())
        {
            var d = JsonSerializer.Deserialize<Dictionary<string, object?>>(el.GetRawText(), JsonOpts);
            if (d != null) list.Add(d);
        }
        return list;
    }

    public void Dispose()
    {
        if (_ownsHttp) _http.Dispose();
    }
}
