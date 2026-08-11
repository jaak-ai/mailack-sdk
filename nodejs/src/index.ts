/**
 * Official Node.js / TypeScript SDK for the mailack certified email API.
 *
 * @example
 * ```ts
 * const client = new Client({ baseUrl: "https://api.mailack.com", apiKey: process.env.MAILACK_API_KEY! });
 * const { message, replay } = await client.send("idem-1", {
 *   from: "noreply@acme.mx",
 *   to: "cliente@example.com",
 *   subject: "Recibo",
 *   text: "Gracias.",
 * });
 * ```
 */

export class APIError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string
  ) {
    super(`mailack: HTTP ${status} ${code}: ${message}`);
    this.name = "APIError";
  }

  is(code: string): boolean {
    return this.code === code;
  }
}

export type SendRequest = {
  from: string;
  to: string;
  subject?: string;
  text?: string;
  html?: string;
  headers?: Record<string, string>;
  template_id?: string;
  variables?: Record<string, string>;
};

export type BatchItem = {
  idempotency_key: string;
  from: string;
  to: string;
  subject: string;
  text?: string;
  html?: string;
  headers?: Record<string, string>;
};

export type ClientOptions = {
  baseUrl: string;
  apiKey?: string;
  fetch?: typeof fetch;
};

export class Client {
  private baseUrl: string;
  private apiKey: string;
  private fetchImpl: typeof fetch;

  constructor(opts: ClientOptions) {
    this.baseUrl = opts.baseUrl.replace(/\/$/, "");
    this.apiKey = opts.apiKey?.trim() ?? "";
    this.fetchImpl = opts.fetch ?? globalThis.fetch.bind(globalThis);
  }

  private async request<T>(
    method: string,
    path: string,
    opts: {
      body?: unknown;
      idempotencyKey?: string;
      query?: Record<string, string>;
    } = {}
  ): Promise<{ status: number; headers: Headers; data: T }> {
    let url = this.baseUrl + path;
    if (opts.query) {
      const q = new URLSearchParams(opts.query);
      url += `?${q.toString()}`;
    }
    const headers: Record<string, string> = {
      Accept: "application/json",
      "User-Agent": "mailack-node/0.1.0",
    };
    if (opts.body !== undefined) {
      headers["Content-Type"] = "application/json";
    }
    if (opts.idempotencyKey) {
      headers["Idempotency-Key"] = opts.idempotencyKey;
    }
    if (this.apiKey) {
      headers.Authorization = `Bearer ${this.apiKey}`;
    }
    const res = await this.fetchImpl(url, {
      method,
      headers,
      body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
    });
    const text = await res.text();
    let data: unknown = null;
    if (text) {
      try {
        data = JSON.parse(text);
      } catch {
        data = text;
      }
    }
    if (!res.ok) {
      const err = (data as { error?: { code?: string; message?: string } })?.error;
      throw new APIError(
        res.status,
        err?.code ?? "http_error",
        err?.message ?? (typeof data === "string" ? data : res.statusText)
      );
    }
    return { status: res.status, headers: res.headers, data: data as T };
  }

  /** POST /v1/messages */
  async send(
    idempotencyKey: string,
    req: SendRequest
  ): Promise<{ message: Record<string, unknown>; replay: boolean }> {
    const { headers, data } = await this.request<Record<string, unknown>>(
      "POST",
      "/v1/messages",
      { body: req, idempotencyKey }
    );
    return {
      message: data,
      replay: headers.get("Idempotent-Replay") === "true",
    };
  }

  /** POST /v1/messages/batch */
  async sendBatch(messages: BatchItem[]): Promise<Record<string, unknown>> {
    const { data } = await this.request<Record<string, unknown>>("POST", "/v1/messages/batch", {
      body: { messages },
    });
    return data;
  }

  async getMessage(id: string): Promise<Record<string, unknown>> {
    const { data } = await this.request<{ message?: Record<string, unknown> }>(
      "GET",
      `/v1/messages/${id}`
    );
    return data.message ?? (data as Record<string, unknown>);
  }

  async rates(days = 14): Promise<Record<string, unknown>> {
    const { data } = await this.request<Record<string, unknown>>("GET", "/v1/rates", {
      query: { days: String(days) },
    });
    return data;
  }

  async listDomains(): Promise<unknown[]> {
    const { data } = await this.request<{ items?: unknown[] }>("GET", "/v1/domains");
    return data.items ?? [];
  }

  async createDomain(domain: string): Promise<Record<string, unknown>> {
    const { data } = await this.request<Record<string, unknown>>("POST", "/v1/domains", {
      body: { domain },
    });
    return data;
  }

  async verifyDomain(id: string): Promise<Record<string, unknown>> {
    const { data } = await this.request<Record<string, unknown>>(
      "POST",
      `/v1/domains/${id}/verify`,
      { body: {} }
    );
    return data;
  }

  async listWebhooks(): Promise<{ items: unknown[]; available_events: string[] }> {
    const { data } = await this.request<{ items?: unknown[]; available_events?: string[] }>(
      "GET",
      "/v1/webhooks"
    );
    return { items: data.items ?? [], available_events: data.available_events ?? [] };
  }

  async createWebhook(
    url: string,
    events: string[],
    description = ""
  ): Promise<{ webhook: Record<string, unknown>; secret: string }> {
    const body: Record<string, unknown> = { url, events };
    if (description) body.description = description;
    const { data } = await this.request<{ webhook?: Record<string, unknown>; secret?: string }>(
      "POST",
      "/v1/webhooks",
      { body }
    );
    return { webhook: data.webhook ?? {}, secret: data.secret ?? "" };
  }

  async disableWebhook(id: string): Promise<void> {
    await this.request("DELETE", `/v1/webhooks/${id}`);
  }

  async listTemplates(): Promise<unknown[]> {
    const { data } = await this.request<{ items?: unknown[] }>("GET", "/v1/templates");
    return data.items ?? [];
  }

  async createTemplate(
    name: string,
    subject: string,
    text = "",
    html = ""
  ): Promise<Record<string, unknown>> {
    const { data } = await this.request<Record<string, unknown>>("POST", "/v1/templates", {
      body: { name, subject, text, html },
    });
    return data;
  }
}
