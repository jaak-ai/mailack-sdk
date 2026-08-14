/**
 * Example: send one certified message.
 *
 *   export MAILACK_API_URL=http://localhost:8080
 *   export MAILACK_API_KEY=mlk_…
 *   npx tsx examples/send.ts
 */
import { Client, APIError } from "../src/index";

async function main() {
  const baseUrl = process.env.MAILACK_API_URL ?? "http://localhost:8080";
  const apiKey = process.env.MAILACK_API_KEY;
  if (!apiKey) throw new Error("MAILACK_API_KEY is required");

  const client = new Client({ baseUrl, apiKey });
  const key = `node-sdk-${new Date().toISOString().replace(/[-:.TZ]/g, "")}`;

  try {
    const { message, replay } = await client.send(key, {
      from: process.env.MAILACK_FROM ?? "noreply@example.com",
      to: process.env.MAILACK_TO ?? "you@example.com",
      subject: "mailack Node SDK example",
      text: "Hello from the Node.js SDK.",
      certified: true, // omit to use the account default (default_certified)
    });
    console.log(
      `id=${message.id} state=${message.state} hash=${message.canonical_hash} replay=${replay}`
    );

    // Seal the message into the Merkle tree, then fetch evidence and verify it.
    const seal = await client.sealMessage(message.id as string);
    console.log(`sealed: batch=${seal.batch_id} root=${seal.merkle_root} at=${seal.sealed_at}`);

    const evidence = await client.getEvidence(message.id as string);
    console.log(`evidence: leaf=${evidence.leaf_index} cert=${evidence.certificate_id}`);

    const result = await client.verify(message.id as string);
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
    `rates(7d): delivery=${rates.delivery_rate}% bounce=${rates.bounce_rate}% ingested=${rates.ingested}`
  );
}

main();
