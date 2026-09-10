import { test } from 'node:test';
import assert from 'node:assert/strict';
import { Client, APIError } from './mailack.js';

for (const [method, path, field, header] of [
  ['getMessageRaw', '/v1/messages/m/raw', 'canonicalHash', 'x-mailack-canonical-hash'],
  ['getEventRaw', '/v1/messages/m/events/e/raw', 'rawSha256', 'x-mailack-raw-sha256'],
]) {
  test(method, async () => {
    const bytes = new Uint8Array([0, 255, 13, 10, 128]);
    const client = new Client({ baseUrl: 'https://example.test', apiKey: 'test', fetch: async (url, opts) => {
      assert.equal(url, 'https://example.test' + path);
      assert.equal(opts.method, 'GET');
      assert.equal(opts.headers.Authorization, 'Bearer test');
      return new Response(bytes, { headers: { [header]: 'digest' } });
    }});
    const result = await client[method]('m', 'e');
    assert.deepEqual(new Uint8Array(result.data), bytes);
    assert.equal(result[field], 'digest');
  });
  test(method + ' errors', async () => {
    const client = new Client({ baseUrl: 'https://example.test', fetch: async () =>
      new Response(JSON.stringify({error: {code: 'not_found', message: 'missing'}}), {status: 404}) });
    await assert.rejects(client[method]('m', 'e'), e => e instanceof APIError && e.code === 'not_found');
  });
}

for (const [method, verb, suffix] of [
  ['sealMessage', 'POST', 'seal'],
  ['getEvidence', 'GET', 'evidence'],
  ['getProofBundle', 'GET', 'proof-bundle'],
]) {
  test(method + ' regression', async () => {
    const expected = { canonical_hash: 'digest', proof_path: [] };
    const client = new Client({ baseUrl: 'https://example.test', fetch: async (url, opts) => {
      assert.equal(url, `https://example.test/v1/messages/m/${suffix}`);
      assert.equal(opts.method, verb);
      return Response.json(expected);
    }});
    assert.deepEqual(await client[method]('m'), expected);
  });
}
