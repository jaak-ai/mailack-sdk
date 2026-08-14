package com.mailack;

import com.google.gson.JsonObject;

/** Outcome of Client.getMessageEvidence (GET /v1/messages/{id}/evidence). */
public final class MessageEvidence {
  public final JsonObject evidence;

  public MessageEvidence(JsonObject evidence) {
    this.evidence = evidence;
  }

  public String messageId() {
    return get("message_id");
  }

  public String canonicalHash() {
    return get("canonical_hash");
  }

  public String mimeSha256() {
    return get("mime_sha256");
  }

  public String messageIdHeader() {
    return get("message_id_header");
  }

  public String dateHeader() {
    return get("date_header");
  }

  public String batchId() {
    return get("batch_id");
  }

  public String merkleRoot() {
    return get("merkle_root");
  }

  public String sealedAt() {
    return get("sealed_at");
  }

  public String certificateId() {
    return get("certificate_id");
  }

  public Long leafIndex() {
    return evidence.has("leaf_index") && !evidence.get("leaf_index").isJsonNull()
        ? evidence.get("leaf_index").getAsLong()
        : null;
  }

  private String get(String key) {
    return evidence.has(key) && !evidence.get(key).isJsonNull()
        ? evidence.get(key).getAsString()
        : null;
  }
}
