/**
 * Official JavaScript SDK for the mailack certified email API.
 * Works in Node.js 18+ and browsers that provide global `fetch`.
 *
 * @example
 * ```js
 * import { Client, APIError } from '@mailack/sdk-js';
 * // or: import { Client, APIError } from './mailack.js';
 *
 * const client = new Client({
 *   baseUrl: 'https://api.mailack.com',
 *   apiKey: process.env.MAILACK_API_KEY,
 * });
 * const { message, replay } = await client.send('idem-1', {
 *   from: 'noreply@acme.mx',
 *   to: 'cliente@example.com',
 *   subject: 'Recibo',
 *   text: 'Gracias.',
 * });
 * ```
 */

export class APIError extends Error {
  /**
   * @param {number} status
   * @param {string} code
   * @param {string} message
   */
  constructor(status, code, message) {
    super(`mailack: HTTP ${status} ${code}: ${message}`);
    this.name = 'APIError';
    this.status = status;
    this.code = code;
  }

  /** @param {string} code */
  is(code) {
    return this.code === code;
  }
}

export class Client {
  /**
   * @param {{ baseUrl: string, apiKey?: string, fetch?: typeof fetch }} opts
   */
  constructor(opts) {
    if (!opts?.baseUrl) {
      throw new TypeError('baseUrl is required');
    }
    this.baseUrl = String(opts.baseUrl).replace(/\/$/, '');
    this.apiKey = (opts.apiKey ?? '').trim();
    this.fetchImpl = opts.fetch ?? globalThis.fetch?.bind(globalThis);
    if (!this.fetchImpl) {
      throw new Error('fetch is not available; pass opts.fetch or use Node 18+');
    }
  }

  /**
   * @param {string} method
   * @param {string} path
   * @param {{ body?: unknown, idempotencyKey?: string, query?: Record<string, string> }} [opts]
   */
  async #request(method, path, opts = {}) {
    let url = this.baseUrl + path;
    if (opts.query) {
      const q = new URLSearchParams(opts.query);
      url += `?${q.toString()}`;
    }
    /** @type {Record<string, string>} */
    const headers = {
      Accept: 'application/json',
      'User-Agent': 'mailack-js/0.1.0',
    };
    if (opts.body !== undefined) {
      headers['Content-Type'] = 'application/json';
    }
    if (opts.idempotencyKey) {
      headers['Idempotency-Key'] = opts.idempotencyKey;
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
    let data = null;
    if (text) {
      try {
        data = JSON.parse(text);
      } catch {
        data = text;
      }
    }

    if (!res.ok) {
      const err = data?.error;
      throw new APIError(
        res.status,
        err?.code ?? 'http_error',
        err?.message ?? (typeof data === 'string' ? data : res.statusText),
      );
    }
    return { status: res.status, headers: res.headers, data };
  }

  /**
   * POST /v1/messages
   * @param {string} idempotencyKey
   * @param {{
   *   from: string,
   *   to: string,
   *   subject?: string,
   *   text?: string,
   *   html?: string,
   *   headers?: Record<string, string>,
   *   template_id?: string,
   *   variables?: Record<string, string>,
   *   certified?: boolean,
   * }} req
   *
   * `certified`: omit to use the account default (default_certified);
   * plain messages (certified=false) cannot be sealed.
   */
  async send(idempotencyKey, req) {
    const { headers, data } = await this.#request('POST', '/v1/messages', {
      body: req,
      idempotencyKey,
    });
    return {
      message: data,
      replay: headers.get('Idempotent-Replay') === 'true',
    };
  }

  /**
   * POST /v1/messages/batch (max 100)
   * Each item accepts an optional `certified` flag: omit to use the account
   * default (default_certified); plain messages (certified=false) cannot be
   * sealed. Undefined fields are dropped by JSON serialization, so leaving
   * `certified` out keeps the account default.
   * @param {Array<Record<string, unknown>>} messages
   */
  async sendBatch(messages) {
    const { data } = await this.#request('POST', '/v1/messages/batch', {
      body: { messages },
    });
    return data;
  }

  /** @param {string} id */
  async getMessage(id) {
    const { data } = await this.#request('GET', `/v1/messages/${id}`);
    return data?.message ?? data;
  }

  /**
   * POST /v1/messages/{id}/seal
   * Seals a certified message into the Merkle tree. Sealing a plain message
   * (certified=false) fails with 422 `not_certified`.
   * @param {string} id
   * @returns {Promise<{
   *   message_id: string,
   *   batch_id: string,
   *   seal_type: string,
   *   canonical_hash: string,
   *   merkle_root: string,
   *   certificate_id: string,
   *   serial_number: string,
   *   policy_oid: string,
   *   algorithm_oid: string,
   *   sealed_at: string,
   * }>}
   */
  async sealMessage(id) {
    const { data } = await this.#request('POST', `/v1/messages/${id}/seal`, {
      body: {},
    });
    return data;
  }

  /**
   * GET /v1/messages/{id}/evidence
   * @param {string} id
   * @returns {Promise<{
   *   message_id: string,
   *   canonical_hash: string,
   *   mime_sha256: string,
   *   message_id_header: string,
   *   date_header: string,
   *   batch_id: string,
   *   merkle_root: string,
   *   sealed_at: string,
   *   certificate_id: string,
   *   leaf_index: number,
   * }>}
   */
  async getEvidence(id) {
    const { data } = await this.#request('GET', `/v1/messages/${id}/evidence`);
    return data;
  }

  /**
   * GET /v1/messages/{id}/proof-bundle
   * Returns the raw proof bundle JSON document ({version, message_id,
   * canonical_hash, leaf_index, proof_path, merkle_root, seal: {...}}).
   * Fails with 422 `missing_proof_data` if the message is not sealed yet.
   * @param {string} id
   * @returns {Promise<Record<string, unknown>>}
   */
  async getProofBundle(id) {
    const { data } = await this.#request(
      'GET',
      `/v1/messages/${id}/proof-bundle`,
    );
    return data;
  }

  /**
   * POST /v1/verify
   * Verifies a message's Merkle proof. Fails with 404 `not_found` when the
   * message does not exist, or 422 `missing_proof_data` when it is not sealed.
   * @param {string} messageId
   * @returns {Promise<{
   *   valid: boolean,
   *   merkle_root: string,
   *   certificate_id: string,
   *   sealed_at: string,
   * }>}
   */
  async verify(messageId) {
    const { data } = await this.#request('POST', '/v1/verify', {
      body: { message_id: messageId },
    });
    return data;
  }

  /** @param {number} [days] */
  async rates(days = 14) {
    const { data } = await this.#request('GET', '/v1/rates', {
      query: { days: String(days) },
    });
    return data;
  }

  async listDomains() {
    const { data } = await this.#request('GET', '/v1/domains');
    return data?.items ?? [];
  }

  /** @param {string} domain */
  async createDomain(domain) {
    const { data } = await this.#request('POST', '/v1/domains', {
      body: { domain },
    });
    return data;
  }

  /** @param {string} id */
  async verifyDomain(id) {
    const { data } = await this.#request('POST', `/v1/domains/${id}/verify`, {
      body: {},
    });
    return data;
  }

  async listWebhooks() {
    const { data } = await this.#request('GET', '/v1/webhooks');
    return {
      items: data?.items ?? [],
      available_events: data?.available_events ?? [],
    };
  }

  /**
   * @param {string} url
   * @param {string[]} events
   * @param {string} [description]
   */
  async createWebhook(url, events, description = '') {
    const body = { url, events };
    if (description) body.description = description;
    const { data } = await this.#request('POST', '/v1/webhooks', { body });
    return { webhook: data?.webhook ?? {}, secret: data?.secret ?? '' };
  }

  /** @param {string} id */
  async disableWebhook(id) {
    await this.#request('DELETE', `/v1/webhooks/${id}`);
  }

  async listTemplates() {
    const { data } = await this.#request('GET', '/v1/templates');
    return data?.items ?? [];
  }

  /**
   * @param {string} name
   * @param {string} subject
   * @param {string} [text]
   * @param {string} [html]
   */
  async createTemplate(name, subject, text = '', html = '') {
    const { data } = await this.#request('POST', '/v1/templates', {
      body: { name, subject, text, html },
    });
    return data;
  }
}
