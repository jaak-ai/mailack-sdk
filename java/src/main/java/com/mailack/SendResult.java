package com.mailack;

import com.google.gson.JsonObject;

/** Outcome of Client.send. */
public final class SendResult {
  public final JsonObject message;
  public final boolean replay;

  public SendResult(JsonObject message, boolean replay) {
    this.message = message;
    this.replay = replay;
  }

  public String id() {
    return message.has("id") ? message.get("id").getAsString() : null;
  }

  public String state() {
    return message.has("state") ? message.get("state").getAsString() : null;
  }
}
