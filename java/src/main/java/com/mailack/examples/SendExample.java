package com.mailack.examples;

import com.google.gson.JsonObject;
import com.mailack.APIError;
import com.mailack.Client;
import com.mailack.SendRequest;
import com.mailack.SendResult;

import java.time.Instant;

/**
 * Example: send one certified message.
 *
 * <pre>
 *   export MAILACK_API_URL=http://localhost:8080
 *   export MAILACK_API_KEY=mlk_…
 *   mvn -f java/pom.xml -q package
 *   java -cp java/target/classes:… com.mailack.examples.SendExample
 * </pre>
 */
public final class SendExample {
  public static void main(String[] args) throws Exception {
    String base = env("MAILACK_API_URL", "http://localhost:8080");
    String key = System.getenv("MAILACK_API_KEY");
    if (key == null || key.isBlank()) {
      throw new IllegalStateException("MAILACK_API_KEY is required");
    }
    Client client = new Client(base, key);
    String idem = "java-sdk-" + Instant.now().toString().replaceAll("[:.]", "");
    try {
      SendResult r = client.send(idem, SendRequest.text(
          env("MAILACK_FROM", "noreply@example.com"),
          env("MAILACK_TO", "you@example.com"),
          "mailack Java SDK example",
          "Hello from the Java SDK."));
      System.out.printf("id=%s state=%s hash=%s replay=%s%n",
          r.id(), r.state(),
          r.message.has("canonical_hash") ? r.message.get("canonical_hash").getAsString() : "",
          r.replay);
    } catch (APIError e) {
      System.err.println(e.getCode() + ": " + e.getMessage());
      System.exit(1);
    }
    JsonObject rates = client.rates(7);
    System.out.printf("rates(7d): delivery=%s%% bounce=%s%% ingested=%s%n",
        rates.get("delivery_rate"), rates.get("bounce_rate"), rates.get("ingested"));
  }

  private static String env(String k, String def) {
    String v = System.getenv(k);
    return v == null || v.isBlank() ? def : v;
  }
}
