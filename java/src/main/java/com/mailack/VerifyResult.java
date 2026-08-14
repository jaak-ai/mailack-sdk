package com.mailack;

import com.google.gson.JsonObject;

/** Outcome of Client.verifyMessage (POST /v1/verify). */
public final class VerifyResult {
  public final JsonObject result;

  public VerifyResult(JsonObject result) {
    this.result = result;
  }

  /** Whether the Merkle proof for the message verifies against its seal. */
  public boolean valid() {
    return result.has("valid") && result.get("valid").getAsBoolean();
  }

  public String merkleRoot() {
    return get("merkle_root");
  }

  public String certificateId() {
    return get("certificate_id");
  }

  public String sealedAt() {
    return get("sealed_at");
  }

  private String get(String key) {
    return result.has(key) && !result.get(key).isJsonNull() ? result.get(key).getAsString() : null;
  }
}
