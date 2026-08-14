# @mailack/sdk (Node.js / TypeScript)

Cliente oficial para la API de correo certificado de mailack. Usa `fetch` nativo (Node 18+).

> ¿Solo JavaScript sin TypeScript? Usa el paquete ESM listo en [`../javascript`](../javascript/) (`@mailack/sdk-js`).

## Instalación

```bash
npm install "https://github.com/jaak-ai/mailack-sdk#main:nodejs"
```

O clonando el repositorio:

```bash
git clone https://github.com/jaak-ai/mailack-sdk.git
cd mailack-sdk/nodejs && npm install && npm run build
```

## Uso

```ts
import { Client, APIError } from "@mailack/sdk";

const client = new Client({
  baseUrl: process.env.MAILACK_API_URL ?? "https://api.mailack.com",
  apiKey: process.env.MAILACK_API_KEY!,
});

try {
  const { message, replay } = await client.send("order-42", {
    from: "noreply@acme.mx",
    to: "cliente@example.com",
    subject: "Recibo",
    text: "Gracias por su compra.",
  });
  console.log(message.id, message.state, replay);
} catch (e) {
  if (e instanceof APIError && e.is("quota_exceeded")) {
    // …
  }
  throw e;
}
```

### Batch

```ts
const result = await client.sendBatch([
  { idempotency_key: "a1", from: "a@acme.mx", to: "1@x.com", subject: "Hi", text: "…" },
]);
```

### Plantilla

```ts
await client.send("welcome-9", {
  from: "hola@acme.mx",
  to: "user@x.com",
  template_id: "<uuid>",
  variables: { nombre: "Ada" },
});
```

### Sello y verificación

```ts
// `certified` es opcional en send/sendBatch; si se omite aplica el default de la cuenta.
const { message } = await client.send("order-43", {
  from: "noreply@acme.mx",
  to: "cliente@example.com",
  subject: "Contrato",
  text: "…",
  certified: true,
});

const seal = await client.sealMessage(message.id as string); // POST /v1/messages/{id}/seal
const evidence = await client.getEvidence(message.id as string); // GET /v1/messages/{id}/evidence
const bundle = await client.getProofBundle(message.id as string); // GET /v1/messages/{id}/proof-bundle (JSON crudo)
const { valid } = await client.verify(message.id as string); // POST /v1/verify
```

### Dominios, webhooks, rates

```ts
const d = await client.createDomain("mail.acme.mx");
await client.verifyDomain(d.id as string);

const { webhook, secret } = await client.createWebhook(
  "https://api.suempresa.mx/hooks/mailack",
  ["email.queued", "email.sent", "email.bounced", "email.sealed"]
);

const rates = await client.rates(14);
```

## Ejemplo

```bash
export MAILACK_API_URL=http://localhost:8080
export MAILACK_API_KEY=mlk_…
npx tsx examples/send.ts
```
