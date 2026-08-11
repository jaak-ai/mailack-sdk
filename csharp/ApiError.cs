namespace Mailack;

/// <summary>Decoded mailack error envelope.</summary>
public sealed class ApiError : Exception
{
    public int Status { get; }
    public string Code { get; }

    public ApiError(int status, string code, string message)
        : base($"mailack: HTTP {status} {code}: {message}")
    {
        Status = status;
        Code = code;
    }

    public bool Is(string code) => Code == code;
}
