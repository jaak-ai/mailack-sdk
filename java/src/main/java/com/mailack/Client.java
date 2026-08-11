package com.mailack;

import com.google.gson.Gson;
import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URI;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;

/**
 * Official Java client for the mailack certified email API.
 *
 * <pre>{@code
 * Client client = new Client("https://api.mailack.com", System.getenv("MAILACK_API_KEY"));
 * SendResult r = client.send("idem-1", SendRequest.text(
 *     "noreply@acme.mx", "cliente@example.com", "Recibo", "Gracias."));
 * }</pre>
 */
public final class Client {
  private final String baseUrl;
  private final String apiKey;
  private final int timeoutMs;
  private final Gson gson = new Gson();

  public Client(String baseUrl, String apiKey) {
    this(baseUrl, apiKey, 30_000);
  }

  public Client(String baseUrl, String apiKey, int timeoutMs) {
    this.baseUrl = baseUrl.endsWith("/") ? baseUrl.substring(0, baseUrl.length() - 1) : baseUrl;
    this.apiKey = apiKey == null ? "" : apiKey.trim();
    this.timeoutMs = timeoutMs;
  }

  /** POST /v1/messages */
  public SendResult send(String idempotencyKey, SendRequest req) throws APIError, IOException {
    HttpResult r = request("POST", "/v1/messages", req, idempotencyKey, null);
    boolean replay = "true".equalsIgnoreCase(r.headers.get("Idempotent-Replay"));
    return new SendResult(r.json.getAsJsonObject(), replay);
  }

  /** POST /v1/messages/batch */
  public JsonObject sendBatch(List<Map<String, Object>> messages) throws APIError, IOException {
    JsonObject body = new JsonObject();
    body.add("messages", gson.toJsonTree(messages));
    return request("POST", "/v1/messages/batch", body, null, null).json.getAsJsonObject();
  }

  public JsonObject getMessage(String id) throws APIError, IOException {
    JsonObject root = request("GET", "/v1/messages/" + id, null, null, null).json.getAsJsonObject();
    if (root.has("message")) {
      return root.getAsJsonObject("message");
    }
    return root;
  }

  public JsonObject rates(int days) throws APIError, IOException {
    return request("GET", "/v1/rates", null, null, Map.of("days", String.valueOf(days)))
        .json.getAsJsonObject();
  }

  public JsonArray listDomains() throws APIError, IOException {
    JsonObject root = request("GET", "/v1/domains", null, null, null).json.getAsJsonObject();
    return root.has("items") ? root.getAsJsonArray("items") : new JsonArray();
  }

  public JsonObject createDomain(String domain) throws APIError, IOException {
    return request("POST", "/v1/domains", Map.of("domain", domain), null, null).json.getAsJsonObject();
  }

  public JsonObject verifyDomain(String id) throws APIError, IOException {
    return request("POST", "/v1/domains/" + id + "/verify", Map.of(), null, null).json.getAsJsonObject();
  }

  public JsonObject listWebhooks() throws APIError, IOException {
    return request("GET", "/v1/webhooks", null, null, null).json.getAsJsonObject();
  }

  public JsonObject createWebhook(String url, List<String> events, String description)
      throws APIError, IOException {
    JsonObject body = new JsonObject();
    body.addProperty("url", url);
    body.add("events", gson.toJsonTree(events));
    if (description != null && !description.isEmpty()) {
      body.addProperty("description", description);
    }
    return request("POST", "/v1/webhooks", body, null, null).json.getAsJsonObject();
  }

  public void disableWebhook(String id) throws APIError, IOException {
    request("DELETE", "/v1/webhooks/" + id, null, null, null);
  }

  public JsonArray listTemplates() throws APIError, IOException {
    JsonObject root = request("GET", "/v1/templates", null, null, null).json.getAsJsonObject();
    return root.has("items") ? root.getAsJsonArray("items") : new JsonArray();
  }

  public JsonObject createTemplate(String name, String subject, String text, String html)
      throws APIError, IOException {
    JsonObject body = new JsonObject();
    body.addProperty("name", name);
    body.addProperty("subject", subject);
    body.addProperty("text", text == null ? "" : text);
    body.addProperty("html", html == null ? "" : html);
    return request("POST", "/v1/templates", body, null, null).json.getAsJsonObject();
  }

  private static final class HttpResult {
    final JsonElement json;
    final Map<String, String> headers;

    HttpResult(JsonElement json, Map<String, String> headers) {
      this.json = json;
      this.headers = headers;
    }
  }

  private HttpResult request(
      String method,
      String path,
      Object body,
      String idempotencyKey,
      Map<String, String> query)
      throws APIError, IOException {
    String url = baseUrl + path;
    if (query != null && !query.isEmpty()) {
      StringBuilder sb = new StringBuilder(url).append('?');
      boolean first = true;
      for (Map.Entry<String, String> e : query.entrySet()) {
        if (!first) sb.append('&');
        first = false;
        sb.append(URLEncoder.encode(e.getKey(), StandardCharsets.UTF_8))
            .append('=')
            .append(URLEncoder.encode(e.getValue(), StandardCharsets.UTF_8));
      }
      url = sb.toString();
    }

    HttpURLConnection conn = (HttpURLConnection) URI.create(url).toURL().openConnection();
    conn.setRequestMethod(method);
    conn.setConnectTimeout(timeoutMs);
    conn.setReadTimeout(timeoutMs);
    conn.setRequestProperty("Accept", "application/json");
    conn.setRequestProperty("User-Agent", "mailack-java/0.1.0");
    if (apiKey != null && !apiKey.isEmpty()) {
      conn.setRequestProperty("Authorization", "Bearer " + apiKey);
    }
    if (idempotencyKey != null && !idempotencyKey.isEmpty()) {
      conn.setRequestProperty("Idempotency-Key", idempotencyKey);
    }
    if (body != null && !"GET".equals(method) && !"DELETE".equals(method)) {
      byte[] raw = gson.toJson(body).getBytes(StandardCharsets.UTF_8);
      conn.setDoOutput(true);
      conn.setRequestProperty("Content-Type", "application/json");
      try (OutputStream os = conn.getOutputStream()) {
        os.write(raw);
      }
    }

    int status = conn.getResponseCode();
    InputStream stream = status >= 400 ? conn.getErrorStream() : conn.getInputStream();
    String raw = stream == null ? "" : new String(stream.readAllBytes(), StandardCharsets.UTF_8);
    Map<String, String> hdrs = new java.util.HashMap<>();
    for (Map.Entry<String, List<String>> e : conn.getHeaderFields().entrySet()) {
      if (e.getKey() != null && e.getValue() != null && !e.getValue().isEmpty()) {
        hdrs.put(e.getKey(), e.getValue().get(0));
      }
    }

    if (status < 200 || status >= 300) {
      String code = "http_error";
      String message = raw;
      try {
        JsonObject env = JsonParser.parseString(raw).getAsJsonObject();
        if (env.has("error")) {
          JsonObject err = env.getAsJsonObject("error");
          if (err.has("code")) code = err.get("code").getAsString();
          if (err.has("message")) message = err.get("message").getAsString();
        }
      } catch (Exception ignored) {
        // keep defaults
      }
      throw new APIError(status, code, message);
    }
    JsonElement json = raw.isEmpty() ? new JsonObject() : JsonParser.parseString(raw);
    return new HttpResult(json, hdrs);
  }
}
