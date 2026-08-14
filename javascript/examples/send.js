/**
 * Example: send one certified message (plain JavaScript).
 *
 *   export MAILACK_API_URL=http://localhost:8080
 *   export MAILACK_API_KEY=mlk_…
 *   node examples/send.js
 */
import { Client, APIError } from '../mailack.js';

const baseUrl = process.env.MAILACK_API_URL || 'http://localhost:8080';
const apiKey = process.env.MAILACK_API_KEY;
if (!apiKey) {
  console.error('MAILACK_API_KEY is required');
  process.exit(1);
}

const client = new Client({ baseUrl, apiKey });
const key = `js-sdk-${new Date().toISOString().replace(/[-:.TZ]/g, '')}`;

try {
  const { message, replay } = await client.send(key, {
    from: process.env.MAILACK_FROM || 'noreply@example.com',
    to: process.env.MAILACK_TO || 'you@example.com',
    subject: 'mailack JavaScript SDK example',
    text: 'Hello from the JavaScript SDK.',
    // Omit `certified` to use the account default; plain messages
    // (certified=false) cannot be sealed.
    certified: true,
  });
  console.log(
    `id=${message.id} state=${message.state} hash=${message.canonical_hash} certified=${message.certified} replay=${replay}`,
  );

  const seal = await client.sealMessage(message.id);
  console.log(`sealed: batch=${seal.batch_id} root=${seal.merkle_root}`);

  const evidence = await client.getEvidence(message.id);
  console.log(`evidence: leaf=${evidence.leaf_index} mime=${evidence.mime_sha256}`);

  const result = await client.verify(message.id);
  console.log(`verify: valid=${result.valid}`);
} catch (e) {
  if (e instanceof APIError) {
    console.error(`${e.code}: ${e.message}`);
    process.exit(1);
  }
  throw e;
}

const rates = await client.rates(7);
console.log(
  `rates(7d): delivery=${rates.delivery_rate}% bounce=${rates.bounce_rate}% ingested=${rates.ingested}`,
);
