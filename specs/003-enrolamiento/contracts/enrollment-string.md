# Contrato — Enrollment string

Formato del **enrollment string**: el envoltorio que empaqueta la **URL del backend**, el
**device token** que el backend genera y revela **una sola vez**, y el **`dev_id`** que el backend
asigna al desarrollador. Cumple FR-013.

**Fuente de verdad compartida por dos repos**: este contrato lo **define y decodifica el agente**
(003, en `internal/config/enrollment.go`) y lo **emite el backend** (P-002b). Ambos lados deben
coincidir bit a bit en el formato descrito aquí. Este contrato **NO** redefine el esquema del device
token ni el endpoint de ingesta, que siguen siendo fuente de verdad de P-002 / P-001
(`transport.md`). El enrollment string es solo transporte de entrega: el token que viaja dentro es
exactamente el que `/ingest` verifica, y el `dev_id` que viaja dentro es exactamente el que el
backend consideró autoritativo al emitirlo.

## Superficie

```
pmea2.<base64url-sin-padding( JSON )>
```

- **Prefijo de versión**: literal `pmea2.` (incluye el punto). Es la **única versión válida**. Un
  prefijo distinto o ausente es un enrollment string no reconocido → error, sin persistir nada
  (ver [Política de versión](#política-de-versión)).
- **Cuerpo**: `base64url` **sin padding** (`base64.RawURLEncoding`) de la serialización JSON del
  objeto de abajo. Alfabeto url-safe (`-` y `_`, sin `+` `/` `=`) para pegado seguro en terminales.

### Objeto JSON (payload)

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["endpoint", "token", "dev_id"],
  "properties": {
    "endpoint": { "type": "string", "format": "uri", "description": "URL del backend; DEBE ser https:// y apuntar a /ingest" },
    "token":    { "type": "string", "minLength": 1, "description": "device token en claro (secreto)" },
    "dev_id":   { "type": "string", "minLength": 1, "maxLength": 64, "pattern": "^[A-Za-z0-9._-]+$", "description": "identificador del desarrollador asignado por el backend (Owner/Admin); autoritativo" }
  }
}
```

Struct **cerrada** (`additionalProperties: false`): exactamente estos tres campos, ni uno más ni uno
menos. Un payload con campos extra o con alguno de los tres ausente es malformado.

**Sobre `dev_id`:**

- Es **asignado por el backend** (por un Owner/Admin) y es **autoritativo**: el agente lo adopta tal
  cual como su propio `dev_id`. El agente **ya NO genera** un `dev_id` en local — el valor del
  enrollment string es la única fuente.
- Constraint: **no vacío**, longitud **1–64** caracteres, charset `[A-Za-z0-9._-]` (alfanuméricos
  más punto, guion bajo y guion). El límite y el charset existen para que el valor sea seguro de usar
  en rutas, cabeceras y nombres de fichero sin escaping.

### Ejemplo (token ficticio) — round-trip verificable

Este ejemplo es **real**: el enrollment string de abajo decodifica exactamente al payload de arriba.
El token es ficticio (no es un secreto real), pero el encoding es genuino, así que sirve para
verificar la implementación a mano o en un test.

Payload (JSON compacto, sin espacios, orden de campos `endpoint`, `token`, `dev_id`):

```json
{"endpoint":"https://api.permea.example/ingest","token":"dev_tok_3f9a2b1c8e7d6f5a4b3c2d1e0f9a8b7c","dev_id":"acme-dev-01"}
```

Enrollment string resultante (`pmea2.` + `base64url` sin padding del JSON de arriba):

```
pmea2.eyJlbmRwb2ludCI6Imh0dHBzOi8vYXBpLnBlcm1lYS5leGFtcGxlL2luZ2VzdCIsInRva2VuIjoiZGV2X3Rva18zZjlhMmIxYzhlN2Q2ZjVhNGIzYzJkMWUwZjlhOGI3YyIsImRldl9pZCI6ImFjbWUtZGV2LTAxIn0
```

Decodificación (los tres campos que produce la terna `(endpoint, token, dev_id)`):

| Campo      | Valor                                             |
|------------|---------------------------------------------------|
| `endpoint` | `https://api.permea.example/ingest`               |
| `token`    | `dev_tok_3f9a2b1c8e7d6f5a4b3c2d1e0f9a8b7c`        |
| `dev_id`   | `acme-dev-01`                                      |

Round-trip a mano. El cuerpo (todo lo que va tras `pmea2.`) mide **163** caracteres; como
`163 % 4 == 3`, para herramientas que exigen padding hay que añadir **un** `=` (base64url sin padding
no lo lleva; `base64.RawURLEncoding` de Go decodifica sin este paso):

```sh
# tr traduce el alfabeto url-safe (_ -) al estándar (/ +); el "=" final es el padding
$ echo -n 'eyJlbmRwb2ludCI6Imh0dHBzOi8vYXBpLnBlcm1lYS5leGFtcGxlL2luZ2VzdCIsInRva2VuIjoiZGV2X3Rva18zZjlhMmIxYzhlN2Q2ZjVhNGIzYzJkMWUwZjlhOGI3YyIsImRldl9pZCI6ImFjbWUtZGV2LTAxIn0=' \
    | tr '_-' '/+' | base64 -d
{"endpoint":"https://api.permea.example/ingest","token":"dev_tok_3f9a2b1c8e7d6f5a4b3c2d1e0f9a8b7c","dev_id":"acme-dev-01"}
```

## Reglas de decodificación (agente)

El agente DEBE, en este orden, y ante cualquier fallo **abortar sin persistir y sin reproducir el
argumento** en el error:

1. Comprobar el prefijo exacto `pmea2.`; si no, aplicar la [Política de versión](#política-de-versión).
2. `base64.RawURLEncoding.Decode` del cuerpo; si falla, error `enrollment string malformado`.
3. `json.Unmarshal` en la struct cerrada `{endpoint, token, dev_id}`; si falla (campo ausente,
   campo extra por `additionalProperties:false`, tipo incorrecto), error `enrollment string malformado`.
4. Validar `endpoint`: URL parseable con esquema **`https`** (una `http://` u otro esquema se
   rechaza, coherente con `transport.md` y `config.Validate`).
5. Validar `token`: no vacío.
6. Validar `dev_id`: no vacío, longitud 1–64, charset `[A-Za-z0-9._-]`.

Un enrollment string que pasa todas las reglas produce la terna `(endpoint, token, dev_id)`. El par
`(endpoint, token)` alimenta el ping de verificación; `dev_id` se adopta como identidad del
desarrollador. **La decodificación no confirma el enrolamiento**: la validez del token la decide el
backend en el ping (2xx/401), no el formato.

## Política de versión

**Solo `pmea2.` es válido.** Los demás casos se rechazan de forma explícita, sin persistir nada:

- **`pmea1.` (formato viejo, `{endpoint, token}`)**: se **RECHAZA** con un error claro y accionable:
  `formato de enrollment obsoleto; solicita uno nuevo desde el panel`. No se intenta migrar ni
  adivinar un `dev_id`: el `dev_id` autoritativo solo puede venir del backend, así que un `pmea1.`
  es intrínsecamente insuficiente.
- **Prefijo desconocido o ausente** (`pmea3.`, `pmea.`, texto sin prefijo, etc.): error
  `enrollment string no reconocido`.

El prefijo versiona el formato; los binarios rechazan versiones que no entienden con un error claro
en vez de adivinar. Una futura `pmea3.` se introduciría en una spec posterior con su propia política.

## Seguridad

- El enrollment string es **encoding, no cifrado**: contiene el token en claro. Es un **secreto del
  mismo calibre que el `salt`** y hereda su disciplina (FR-007): NUNCA se loguea, se imprime ni se
  incluye en mensajes de error. Que ahora incluya `dev_id` no relaja esta disciplina: el string
  entero sigue siendo secreto porque **contiene el token**.
- El prefijo `pmea2.` es un patrón reconocible a propósito, para que los escáneres de secretos puedan
  detectarlo si se filtra por accidente.
- El agente NUNCA persiste el enrollment string envuelto; persiste solo `endpoint`, `token` y
  `dev_id` por separado en `config.json` (0600).

## Compatibilidad y evolución

- Este contrato es **fuente de verdad compartida**: lo consumen el **agente** (003, decodifica) y el
  **backend** (P-002b, emite). Cualquier cambio de formato debe coordinarse entre ambos repos y
  materializarse en un nuevo prefijo de versión.
- El backend (P-002b) DEBE emitir exactamente `pmea2.<base64url(json)>` con las mismas reglas —
  struct cerrada de tres campos y `dev_id` autoritativo— para ser compatible.
- `pmea1.` queda **retirado**: el backend ya no lo emite y el agente lo rechaza (ver
  [Política de versión](#política-de-versión)).
