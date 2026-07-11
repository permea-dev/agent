# Research — Enrolamiento de dispositivo (003)

Decisiones de diseño previas a la implementación. Cada punto resuelve una incógnita del
`plan.md` sin introducir detalle en el `spec.md`. No queda ningún **NEEDS CLARIFICATION**.

## R1 — Formato del enrollment string

**Decisión**: `pmea1.<base64url-sin-padding(JSON)>`, donde el JSON es un objeto compacto
cerrado `{"endpoint":"https://…","token":"<device_token>"}`. Prefijo de versión `pmea1.`
literal; el cuerpo es `base64.RawURLEncoding` de la serialización JSON.

**Rationale**:
- **Stdlib pura** (`encoding/base64` + `encoding/json`): cero dependencias (Principio III).
- **Reproducible en P-002** (backend PHP/Laravel): `base64url(json_encode(...))` es trivial en
  ambos lados; 003 define el formato, P-002 lo emitirá igual.
- **Versionado explícito** (`pmea1.`): permite evolucionar el envoltorio sin romper binarios viejos
  (un prefijo desconocido se rechaza con un error claro).
- **Prefijo reconocible**: `pmea1.` da a los escáneres de secretos un patrón que detectar; el
  string es *encoding, no cifrado* y contiene el token en claro, así que debe tratarse como secreto.
- **Copy-paste seguro**: `base64url` sin padding evita `+`, `/`, `=` que se rompen al pegar en
  terminales/URLs.

**Alternativas consideradas**:
- *Token pelado + URL por separado* (dos argumentos): peor UX y reabre la ambigüedad de la URL que
  el spec ya cerró; el usuario tendría que copiar dos valores. Rechazado.
- *JWT/base64 firmado*: introduce cripto y esquema de claves innecesarios; el token **ya** es el
  secreto autenticado por el backend. El envoltorio no necesita integridad propia. Rechazado.
- *base64 estándar (no url)*: `+//=` se corrompen al pegar. Rechazado.

## R2 — Verificación por ping de lote vacío (sin endpoint nuevo)

**Decisión**: verificar reutilizando `transport.Client.Send(events)` con
`events := []event.Event{}` (slice **no-nil** vacío). Se añade un método fino
`func (c *Client) Verify() error { return c.Send([]event.Event{}) }` que documenta la intención.

**Rationale**:
- `json.Marshal([]event.Event{})` produce **`[]`** (lote vacío, cero eventos, cero metadato);
  `json.Marshal(nil-slice)` produciría `null` — por eso el ping DEBE usar un slice **no-nil**.
  Verificado empíricamente: `[]struct{}{}` → `"[]"`, `var n []struct{}` → `"null"`.
- `Send` ya cumple `contracts/transport.md`: POST HTTPS con `Authorization: Bearer`, rechaza
  esquema no-https (`ErrScheme`), y clasifica `2xx`=ok, `401/403`=auth. Reutilizarlo garantiza que
  **no se inventa** ningún endpoint de verificación (FR-002/FR-011). Nota: en el flujo de `enroll`,
  una URL `http://` ya se rechaza en la **decodificación** (R1, pre-ping), por lo que `ErrScheme` es
  una **guardia secundaria** que no llega a ejercitarse aquí.
- El backend deduplica por `event_id`; un lote vacío no aporta ninguno, así que **no crea ni altera
  estado** más allá de confirmar la autenticación (SC-009).

**Alternativas consideradas**:
- *Endpoint `/verify` dedicado*: viola FR-002 (reutilizar el contrato existente). Rechazado.
- *`sendWithRetry` (con backoff)*: en enrolamiento un 5xx/red significa "no pude verificar → no
  persisto" (US2 escenario 4); no queremos bloquear al usuario con reintentos largos. Se usa `Send`
  (un solo intento). Rechazado el retry para este camino.

## R3 — Semántica de respuestas en enrolamiento

**Decisión**: sobre el resultado de `Verify()`:
- `nil` (2xx) → **confirmar** y persistir.
- `IsAuth(err)` (401/403) → **rechazar**: token inválido/revocado; **no** persistir; error al usuario
  sin el token.
- cualquier otro error (5xx, red, `ErrScheme`, timeout) → **no verificado**: **no** persistir;
  informar "no se pudo verificar"; sin filtrar el token.

**Rationale**: reutiliza los clasificadores ya existentes `transport.IsAuth`/`Retryable`. Distingue
"rechazado" (auth) de "no verificable" (transitorio) para dar un mensaje correcto, pero **ninguno de
los dos persiste** (FR-004): el estado tras el fallo es indistinguible de no haber enrolado.

## R4 — Persistencia segura y re-enrolamiento sin residuos

**Decisión**: reutilizar `config.Load` + `config.Save`. Enrolar = cargar config actual, fijar
`Endpoint` y `DeviceToken` con los valores decodificados, `Save` (que ya hace temp+chmod `0600`+rename
atómico). El re-enrolamiento sobrescribe `config.json` por `os.Rename`, que reemplaza el contenido
antiguo; el temporal `.tmp-*` se crea con `0600` y se elimina (defer) tras el rename.

**Rationale**:
- `atomicWrite` ya cumple FR-005 (0600) y la durabilidad; no hay que reimplementar nada.
- FR-014 (sin residuos): tras el `rename`, el token viejo del `config.json` anterior desaparece
  (reemplazo atómico); el único temporal es `0600` y de vida efímera. No se generan ficheros `.bak`.
- **Riesgo detectado**: `os.CreateTemp` crea el temporal con `0600` en la mayoría de SO, pero
  `atomicWrite` hace `Chmod(0600)` explícito **después** de crear, cerrando la ventana. Se mantiene.

**Alternativas consideradas**:
- *Fichero de token separado del config.json*: duplicaría la lógica de rutas/permisos; `config.json`
  ya alberga `device_token`. Rechazado (más superficie, menos auditable).

## R5 — Higiene del secreto (token y enrollment string)

**Decisión**: el token y el enrollment string se tratan como el `salt` (R6 de P-001): **nunca** a
logs, stdout ni mensajes de error. Reglas concretas:
- Errores de decodificación **no** incluyen el argumento (`enrollment string malformado`, sin eco).
- `status` imprime solo `endpoint` y un booleano de estado; **nunca** el token (ni truncado de forma
  reconstruible; se muestra a lo sumo un indicador tipo `configurado`/`ausente`).
- `enroll` en éxito imprime confirmación + URL, nunca el token.
- Leído por **stdin**, el enrollment string no se hace eco a la terminal ni queda en salida alguna
  (FR-015).
- Revisar rutas existentes que vuelquen `Config` con `%+v` (p. ej. `main.go` rama `default:` imprime
  `config.Default()`, cuyo token está vacío — sin riesgo, pero se evita imprimir configs cargadas).

**Exposición en argv y su mitigación (en alcance)**: pasar el enrollment string como **argumento de
comando** lo expone en la lista de procesos (argv) y en el historial del shell. Por eso `enroll`
acepta también el string por **stdin** (`permea enroll` sin argumento, o `permea enroll -`) como vía
recomendada, que no deja el secreto en argv ni en el historial (FR-001/FR-015, SC-011). El *override
por variable de entorno* para CI sigue **fuera de alcance** (posible spec futura).

## R6 — Modelo de subcomandos en el CLI

**Decisión**: despacho por `os.Args[1]`: si es `enroll` o `status`, se enruta a su handler; en otro
caso se mantiene el parseo de flags actual (`--scan/--run/--daemon/--version`). Solo stdlib (`os.Args`
+ `flag` para el resto).

**Rationale**: introduce los verbos que pide el spec sin romper la CLI de P-001/P-002 ni añadir un
framework de subcomandos (Principio III: se favorece stdlib). `enroll` toma exactamente un argumento
posicional (el enrollment string); su ausencia es error de uso.

**Alternativas consideradas**:
- *Migrar todo a un router de subcomandos (cobra, etc.)*: dependencia externa injustificada. Rechazado.
- *Flag `--enroll=<string>`*: mete el secreto en un flag con `=`, menos natural que un verbo; y el
  spec fija la forma `permea enroll <enrollment-string>`. Rechazado.

## R7 — Estado de enrolamiento (para `status`)

**Decisión**: "enrolado" ⇔ `config.Endpoint != "" && config.DeviceToken != ""` y `config.Validate()`
(endpoint https) no falla. `status` deriva de ahí; no persiste ni verifica contra el backend (lectura
local, sin red).

**Rationale**: el estado es una propiedad local derivada de la config ya persistida; `status` no debe
depender de red (Principio II, local-first). La confirmación contra el backend es cosa de `enroll`.
