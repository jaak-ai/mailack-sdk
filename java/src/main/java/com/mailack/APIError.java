package com.mailack;

/** Decoded mailack error envelope. */
public final class APIError extends Exception {
  private final int status;
  private final String code;

  public APIError(int status, String code, String message) {
    super(String.format("mailack: HTTP %d %s: %s", status, code, message));
    this.status = status;
    this.code = code == null ? "http_error" : code;
  }

  public int getStatus() {
    return status;
  }

  public String getCode() {
    return code;
  }

  public boolean isCode(String expected) {
    return code.equals(expected);
  }
}
