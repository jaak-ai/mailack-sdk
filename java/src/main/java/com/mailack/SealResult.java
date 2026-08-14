package com.mailack;

import com.google.gson.JsonObject;

/** Outcome of Client.sealMessage (POST /v1/messages/{id}/seal). */
public final class SealResult {
  public final JsonObject seal;

  public SealResult(JsonObject seal) {
    this.seal = seal;
  }

  public String messageId() {
    return get("message_id");
  }

  public String batchId() {
    return get("batch_id");
  }

  public String sealType() {
    return get("seal_type");
  }

  public String canonicalHash() {
    return get("canonical_hash");
  }

  public String merkleRoot() {
    return get("merkle_root");
  }

  public String certificateId() {
    return get("certificate_id");
  }

  public String serialNumber() {
    return get("serial_number");
  }

  public String policyOid() {
    return get("policy_oid");
  }

  public String algorithmOid() {
    return get("algorithm_oid");
  }

  public String sealedAt() {
    return get("sealed_at");
  }

  private String get(String key) {
    return seal.has(key) && !seal.get(key).isJsonNull() ? seal.get(key).getAsString() : null;
  }
}
