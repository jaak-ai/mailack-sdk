package com.mailack;

import java.util.Map;

/** JSON body for POST /v1/messages. */
public final class SendRequest {
  public String from;
  public String to;
  public String subject;
  public String text;
  public String html;
  public Map<String, String> headers;
  public String template_id;
  public Map<String, String> variables;

  public static SendRequest text(String from, String to, String subject, String text) {
    SendRequest r = new SendRequest();
    r.from = from;
    r.to = to;
    r.subject = subject;
    r.text = text;
    return r;
  }

  public static SendRequest template(String from, String to, String templateId, Map<String, String> vars) {
    SendRequest r = new SendRequest();
    r.from = from;
    r.to = to;
    r.template_id = templateId;
    r.variables = vars;
    return r;
  }
}
