# Contrato — Enrollment string

Formato del **enrollment string**: el envoltorio que empaqueta la **URL del backend** y el
**device token** que el backend (P-002) genera y revela **una sola vez**. Cumple FR-013.

**Fuente de verdad**: este contrato lo **define 003** (es el lado que lo **decodifica**, en
`internal/config/enrollment.go`). **P-002 (backend) lo consumirá para emitirlo** con el mismo
formato. Este contrato **NO** redefine el esquema del device token ni el endpoint de ingesta, que
siguen siendo fuente de verdad de P-002 / P-001 (`transport.md`). El enrollment string es solo
transporte de entrega: el token que viaja dentro es exactamente el que `/ingest` verifica.

## Superficie

```
pmea1.<base64url-sin-padding( JSON )>
```

- **Prefijo de versión**: literal `pmea1.` (incluye el punto). Un prefijo distinto o ausente es un
  enrollment string no reconocido → error, sin persistir nada.
- **Cuerpo**: `base64url` **sin padding** (`base64.RawURLEncoding`) de la serialización JSON del
  objeto de abajo. Alfabeto url-safe (`-` y `_`, sin `+` `/` `=`) para pegado seguro en terminales.

### Objeto JSON (payload)

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["endpoint", "token"],
  "properties": {
    "endpoint": { "type": "string", "format": "uri", "description": "URL del backend; DEBE ser https://" },
    "token":    { "type": "string", "minLength": 1, "description": "device token en claro (secreto)" }
  }
}
```

### Ejemplo (token ficticio)

Payload:

```json
{"endpoint":"https://api.permea.example/ingest","token":"dev_tok_3f9a…REDACTED"}
```

Enrollment string resultante (ilustrativo, no real):

```
pmea1.eyJlbmRwb2ludCI6Imh0dHBzOi8vYXBpLnBlcm1lYS5leGFtcGxlL2luZ2VzdCIsInRva2VuIjoiZGV2X3Rva18zZjlhIn0
```

## Reglas de decodificación (agente)

El agente DEBE, en este orden, y ante cualquier fallo **abortar sin persistir y sin reproducir el
argumento** en el error:

1. Comprobar el prefijo exacto `pmea1.`; si no, error `enrollment string no reconocido`.
2. `base64.RawURLEncoding.Decode` del cuerpo; si falla, error `enrollment string malformado`.
3. `json.Unmarshal` en la struct cerrada `{endpoint, token}`; si falla, error `enrollment string malformado`.
4. Validar `endpoint`: URL parseable con esquema **`https`** (una `http://` u otro esquema se
   rechaza, coherente con `transport.md` y `config.Validate`).
5. Validar `token`: no vacío.

Un enrollment string que pasa las 5 reglas produce el par `(endpoint, token)` que alimenta el ping
de verificación. **La decodificación no confirma el enrolamiento**: la validez del token la decide el
backend en el ping (2xx/401), no el formato.

## Seguridad

- El enrollment string es **encoding, no cifrado**: contiene el token en claro. Es un **secreto del
  mismo calibre que el `salt`** y hereda su disciplina (FR-007): NUNCA se loguea, se imprime ni se
  incluye en mensajes de error.
- El prefijo `pmea1.` es un patrón reconocible a propósito, para que los escáneres de secretos puedan
  detectarlo si se filtra por accidente.
- El agente NUNCA persiste el enrollment string envuelto; persiste solo `endpoint` y `token` por
  separado en `config.json` (0600).

## Compatibilidad y evolución

- El prefijo versiona el formato. Un `pmea2.` futuro (p. ej. con más campos) se introduciría en una
  spec posterior; los binarios que solo entienden `pmea1.` rechazan versiones desconocidas con un
  error claro en vez de adivinar.
- P-002 DEBE emitir exactamente `pmea1.<base64url(json)>` con las mismas reglas para ser compatible.
