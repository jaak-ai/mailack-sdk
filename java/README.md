# mailack Java SDK

Cliente oficial en Java 17+ para la API de correo certificado de mailack.

## Instalación

```xml
<!-- Todavía no está en Maven Central. Clona https://github.com/jaak-ai/mailack-sdk e instala en tu repositorio local: -->
<!-- mvn -f mailack-sdk/java/pom.xml install -->
<dependency>
  <groupId>com.mailack</groupId>
  <artifactId>mailack-sdk</artifactId>
  <version>0.1.0</version>
</dependency>
```

Dependencia: Gson 2.x.

## Uso

```java
import com.mailack.*;

Client client = new Client("https://api.mailack.com", System.getenv("MAILACK_API_KEY"));

try {
  SendResult r = client.send("order-42", SendRequest.text(
      "noreply@acme.mx",
      "cliente@example.com",
      "Recibo",
      "Gracias por su compra."));
  System.out.println(r.id() + " " + r.state() + " replay=" + r.replay);
} catch (APIError e) {
  if (e.isCode("quota_exceeded")) { /* … */ }
  throw e;
}
```

### Plantilla

```java
SendResult r = client.send("welcome-9", SendRequest.template(
    "hola@acme.mx", "user@x.com", templateUuid,
    Map.of("nombre", "Ada")));
```

### Batch

```java
JsonObject batch = client.sendBatch(List.of(
    Map.of("idempotency_key", "a1", "from", "a@acme.mx", "to", "1@x.com",
           "subject", "Hi", "text", "…", "certified", true)));
```

Cada ítem admite `"certified"` opcional; omítelo para aplicar el default de la cuenta
(`default_certified`). Los mensajes plain (`certified=false`) entregan igual pero no se pueden sellar.

### Sellado, evidencia y verificación

```java
SendRequest req = SendRequest.text("noreply@acme.mx", "cliente@example.com", "Recibo", "…");
req.certified = true;                       // omite para usar el default de la cuenta
SendResult r = client.send("order-42", req);

SealResult seal = client.sealMessage(r.id());           // POST /v1/messages/{id}/seal
MessageEvidence ev = client.getMessageEvidence(r.id()); // GET  /v1/messages/{id}/evidence
JsonObject bundle = client.getProofBundle(r.id());      // GET  /v1/messages/{id}/proof-bundle (JSON crudo)
VerifyResult v = client.verifyMessage(r.id());          // POST /v1/verify → v.valid()
```

Errores: sellar un plain → 422 `not_certified`; verificar o pedir el bundle antes del sellado →
422 `missing_proof_data`; verificar un id inexistente → 404 `not_found`.

### Dominios / webhooks / rates

```java
JsonObject d = client.createDomain("mail.acme.mx");
client.verifyDomain(d.get("id").getAsString());

JsonObject hook = client.createWebhook(
    "https://api.suempresa.mx/hooks/mailack",
    List.of("email.queued", "email.sent", "email.bounced", "email.sealed"),
    "ERP");
// hook.get("secret") — solo una vez

JsonObject rates = client.rates(14);
```

## Ejemplo

```bash
export MAILACK_API_URL=http://localhost:8080
export MAILACK_API_KEY=mlk_…
mvn -f java/pom.xml -q exec:java \
  -Dexec.mainClass=com.mailack.examples.SendExample \
  -Dexec.classpathScope=compile
```

(O compile y ejecute `examples/SendExample.java` con el classpath de Gson.)
