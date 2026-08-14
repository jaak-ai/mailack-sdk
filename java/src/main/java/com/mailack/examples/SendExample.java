package com.mailack.examples;

import com.google.gson.JsonObject;
import com.mailack.APIError;
import com.mailack.Client;
import com.mailack.MessageEvidence;
import com.mailack.SealResult;
import com.mailack.SendRequest;
import com.mailack.SendResult;
import com.mailack.VerifyResult;

import java.time.Instant;

/**
 * Example: send one certified message, then seal it and verify its proof.
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
    String messageId;
    try {
      SendRequest req = SendRequest.text(
          env("MAILACK_FROM", "noreply@example.com"),
          env("MAILACK_TO", "you@example.com"),
          "mailack Java SDK example",
          "Hello from the Java SDK.");
      // Omit `certified` to use the account default (default_certified);
      // plain messages (certified=false) cannot be sealed.
      req.certified = true;
      SendResult r = client.send(idem, req);
      messageId = r.id();
      System.out.printf("id=%s state=%s certified=%s hash=%s replay=%s%n",
          messageId, r.state(), r.certified(),
          r.message.has("canonical_hash") ? r.message.get("canonical_hash").getAsString() : "",
          r.replay);
    } catch (APIError e) {
      System.err.println(e.getCode() + ": " + e.getMessage());
      System.exit(1);
      return;
    }

    try {
      // seal → evidence → verify
      SealResult seal = client.sealMessage(messageId);
      System.out.printf("sealed: batch=%s merkle_root=%s at=%s%n",
          seal.batchId(), seal.merkleRoot(), seal.sealedAt());

      MessageEvidence ev = client.getMessageEvidence(messageId);
      System.out.printf("evidence: canonical_hash=%s leaf_index=%s%n",
          ev.canonicalHash(), ev.leafIndex());

      VerifyResult v = client.verifyMessage(messageId);
      System.out.printf("verify: valid=%s merkle_root=%s%n", v.valid(), v.merkleRoot());
    } catch (APIError e) {
      // 422 not_certified: the message was sent plain; 422 missing_proof_data: not sealed yet
      System.err.println(e.getCode() + ": " + e.getMessage());
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
