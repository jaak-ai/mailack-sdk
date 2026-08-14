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

string messageId;
try
{
    var result = await client.SendAsync(idem, new SendRequest
    {
        From = Environment.GetEnvironmentVariable("MAILACK_FROM") ?? "noreply@example.com",
        To = Environment.GetEnvironmentVariable("MAILACK_TO") ?? "you@example.com",
        Subject = "mailack C# SDK example",
        Text = "Hello from the C# SDK.",
        Certified = true, // omit to use the account default (default_certified)
    });
    Console.WriteLine($"id={result.Id} state={result.State} hash={result.Message.GetValueOrDefault("canonical_hash")} replay={result.Replay} certified={result.Certified}");
    messageId = result.Id!;
}
catch (ApiError e)
{
    Console.Error.WriteLine($"{e.Code}: {e.Message}");
    return 1;
}

// Seal the message into a Merkle batch, then fetch its evidence and verify it.
try
{
    var seal = await client.SealMessageAsync(messageId);
    Console.WriteLine($"sealed: batch={seal.GetValueOrDefault("batch_id")} merkle_root={seal.GetValueOrDefault("merkle_root")}");

    var evidence = await client.GetMessageEvidenceAsync(messageId);
    Console.WriteLine($"evidence: canonical_hash={evidence.GetValueOrDefault("canonical_hash")} leaf_index={evidence.GetValueOrDefault("leaf_index")}");

    var verify = await client.VerifyAsync(messageId);
    Console.WriteLine($"verify: valid={verify.Valid} merkle_root={verify.MerkleRoot}");
}
catch (ApiError e) when (e.Is("not_certified") || e.Is("missing_proof_data"))
{
    Console.Error.WriteLine($"seal/verify skipped: {e.Code}: {e.Message}");
}

var rates = await client.RatesAsync(7);
Console.WriteLine($"rates(7d): delivery={rates.GetValueOrDefault("delivery_rate")}% bounce={rates.GetValueOrDefault("bounce_rate")}% ingested={rates.GetValueOrDefault("ingested")}");
return 0;
