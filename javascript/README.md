# @mailack/sdk-js (JavaScript)

Cliente oficial en **JavaScript puro** (ESM) para la API de correo certificado de mailack.

- Node.js **18+** (fetch nativo)
- Navegadores con `fetch`
- **Sin TypeScript ni build**

> Si prefieres tipos y `tsc`, usa también [`nodejs`](../nodejs/) (`@mailack/sdk`).

## Instalación

```bash
npm install "https://github.com/jaak-ai/mailack-sdk#main:javascript"

# o import por path
import { Client } from './mailack-sdk/javascript/mailack.js';
```

## Uso

```js
import { Client, APIError } from '@mailack/sdk-js';

const client = new Client({
  baseUrl: process.env.MAILACK_API_URL || 'https://api.mailack.com',
  apiKey: process.env.MAILACK_API_KEY,
});

try {
  const { message, replay } = await client.send('order-42', {
    from: 'noreply@acme.mx',
    to: 'cliente@example.com',
    subject: 'Recibo',
    text: 'Gracias por su compra.',
  });
  console.log(message.id, message.state, replay);
} catch (e) {
  if (e instanceof APIError && e.is('quota_exceeded')) {
    // cuota mensual
  }
  throw e;
}
```

### Batch

```js
const result = await client.sendBatch([
  {
    idempotency_key: 'a1',
    from: 'a@acme.mx',
    to: '1@x.com',
    subject: 'Hi',
    text: '…',
  },
]);
```

### Plantilla

```js
await client.send('welcome-9', {
  from: 'hola@acme.mx',
  to: 'user@x.com',
  template_id: '<uuid>',
  variables: { nombre: 'Ada' },
});
```

### Dominios, webhooks, rates

```js
const d = await client.createDomain('mail.acme.mx');
// TXT: d.dns.challenge_host = d.dns.challenge_value
await client.verifyDomain(d.id);

const { webhook, secret } = await client.createWebhook(
  'https://api.suempresa.mx/hooks/mailack',
  ['email.queued', 'email.sent', 'email.bounced', 'email.sealed'],
);

const rates = await client.rates(14);
console.log(rates.delivery_rate, rates.bounce_rate);
```

## Ejemplo

```bash
export MAILACK_API_URL=http://localhost:8080
export MAILACK_API_KEY=mlk_…
node javascript/examples/send.js
```

## API key

Scopes habituales: `messages:send`, `evidence:read`. El tenant lo resuelve el servidor a partir de la key.
