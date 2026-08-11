# SDKs de Mailack

Clientes oficiales de la API de [Mailack](https://mailack.com), correo certificado con
evidencia NOM-151 residente en México.

Este repositorio es público y no requiere credenciales para instalarse. El servidor vive
en un repositorio privado aparte; aquí solo están los clientes.

| Lenguaje | Directorio | Paquete |
|---|---|---|
| **Go 1.25+** | [`go/`](go/) | `github.com/jaak-ai/mailack-sdk/go` |
| **Python 3.9+** | [`python/`](python/) | `mailack` |
| **Node.js / TypeScript** | [`nodejs/`](nodejs/) | `@mailack/sdk` |
| **JavaScript** (ESM, sin build) | [`javascript/`](javascript/) | `@mailack/sdk-js` |
| **Rust** | [`rust/`](rust/) | crate `mailack` |
| **Java 17+** | [`java/`](java/) | `com.mailack:mailack-sdk` |
| **C# / .NET 8** | [`csharp/`](csharp/) | `Mailack.Sdk` |

Cada directorio tiene su propio README con la instalación y un ejemplo completo.

## Empezar

```bash
export MAILACK_API_URL=https://api.mailack.com
export MAILACK_API_KEY=mlk_…        # se crea en el portal, no la escribas en el código
```

La API key se emite desde el [portal del cliente](https://portal.mailack.com); el tenant
se resuelve en el servidor a partir de ella.

```bash
go get github.com/jaak-ai/mailack-sdk/go
pip install "git+https://github.com/jaak-ai/mailack-sdk.git#subdirectory=python"
npm install "https://github.com/jaak-ai/mailack-sdk#main:nodejs"
```

## Qué cubren

Todos los SDKs hablan el mismo contrato:

- Envío con `Idempotency-Key`, y envío por lotes de hasta 100
- Plantillas con variables (`template_id` + `variables`)
- Dominios de envío: alta y verificación por DNS
- Webhooks del ciclo de vida del mensaje
- Tasas de entregabilidad

Autenticación por API key Bearer (`mlk_…`).

## Errores que conviene tratar

| Código | Significado |
|---|---|
| `quota_exceeded` | Se agotó la cuota mensual de la cuenta |
| `recipient_suppressed` | El destinatario está en la lista de supresión |
| `domain_not_verified` | El `From` no pertenece a un dominio verificado |
| `domain_taken` / `domain_exists` | El dominio ya está registrado, por otra cuenta o por la tuya |
| `template_not_found` / `template_render_error` | Problema con la plantilla |

## Versiones

Los SDKs se etiquetan con semver. El módulo de Go vive en un subdirectorio, así que sus
etiquetas llevan el prefijo del directorio: `go/v0.1.0` para
`github.com/jaak-ai/mailack-sdk/go@v0.1.0`. Los demás lenguajes usan etiquetas simples
(`v0.1.0`).

## Contrato de la API

La referencia canónica es el OpenAPI publicado en la documentación de Mailack. Si un SDK
y el OpenAPI no coinciden, manda el OpenAPI y es un fallo del SDK: abre una incidencia.

## Licencia

MIT. Ver [`LICENSE`](LICENSE).
