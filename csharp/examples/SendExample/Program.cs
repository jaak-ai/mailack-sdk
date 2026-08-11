using Mailack;

var baseUrl = Environment.GetEnvironmentVariable("MAILACK_API_URL") ?? "http://localhost:8080";
var apiKey = Environment.GetEnvironmentVariable("MAILACK_API_KEY");
if (string.IsNullOrWhiteSpace(apiKey))
{
    Console.Error.WriteLine("MAILACK_API_KEY is required");
    return 1;
}

using var client = new Client(baseUrl, apiKey);
var idem = "csharp-sdk-" + DateTime.UtcNow.ToString("yyyyMMddTHHmmss");

try
{
    var result = await client.SendAsync(idem, new SendRequest
    {
        From = Environment.GetEnvironmentVariable("MAILACK_FROM") ?? "noreply@example.com",
        To = Environment.GetEnvironmentVariable("MAILACK_TO") ?? "you@example.com",
        Subject = "mailack C# SDK example",
        Text = "Hello from the C# SDK.",
    });
    Console.WriteLine($"id={result.Id} state={result.State} hash={result.Message.GetValueOrDefault("canonical_hash")} replay={result.Replay}");
}
catch (ApiError e)
{
    Console.Error.WriteLine($"{e.Code}: {e.Message}");
    return 1;
}

var rates = await client.RatesAsync(7);
Console.WriteLine($"rates(7d): delivery={rates.GetValueOrDefault("delivery_rate")}% bounce={rates.GetValueOrDefault("bounce_rate")}% ingested={rates.GetValueOrDefault("ingested")}");
return 0;
