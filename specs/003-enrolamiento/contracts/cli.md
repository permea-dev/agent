# Contrato — Comandos `permea enroll` y `permea status`

Contrato observable de los dos comandos de usuario de 003. Define entradas, salidas, códigos de
salida y las garantías de redacción del secreto. No prescribe implementación.

## `permea enroll [<enrollment-string>]`

Empareja el agente con su backend: decodifica el enrollment string, verifica el token y, si es
válido, lo persiste.

### Entrada

El enrollment string (`pmea1.…`, ver [enrollment-string.md](./enrollment-string.md)) se acepta por
**dos vías equivalentes** que producen el **mismo flujo**:

| Vía | Invocación | Notas |
|---|---|---|
| Argumento posicional | `permea enroll <enrollment-string>` | Cómodo, pero deja el secreto en la lista de procesos (argv) y en el historial del shell. |
| stdin (**recomendada**) | `permea enroll` (sin argumento) o `permea enroll -` | Se lee de stdin (p. ej. `… \| permea enroll -`); no aparece en argv ni en el historial. NUNCA se hace eco. |

El **mensaje de ayuda** del comando DEBE mencionar la vía **stdin** como la **recomendada**, para no
dejar el secreto en el historial del shell.

| Situación de entrada | Efecto |
|---|---|
| Con argumento posicional | Usa ese valor (una vía). |
| Sin argumento, con stdin disponible (pipe) | Lee el enrollment string de stdin y sigue el flujo. |
| Sin argumento y sin stdin (terminal interactiva sin pipe) | **Error de uso, exit ≠ 0** (seguro en CI/no-interactivo); NUNCA un prompt que se cuelgue esperando. |

### Comportamiento y salidas

| Situación | Efecto | stdout/stderr | Exit |
|---|---|---|---|
| Enrollment string malformado / prefijo desconocido / endpoint no-https | Aborta; no persiste | Error **sin** reproducir el argumento ni el token | ≠ 0 |
| Verificación **2xx** (token válido) | Persiste `{endpoint, token}` en `config.json` (0600) | Confirmación + **URL** del backend; **nunca** el token | 0 |
| Verificación **401/403** (token inválido/revocado) | **No** persiste | Error "token rechazado por el backend"; **sin** el token | ≠ 0 |
| Verificación **5xx / red / timeout** | **No** persiste | Error "no se pudo verificar (backend no disponible)"; **sin** el token | ≠ 0 |

### Garantías (DEBE / NUNCA)

- DEBE verificar **antes** de persistir (ping de lote vacío contra `/ingest`, contrato existente).
- Tras un fallo (rechazo o no verificable), el estado DEBE ser **indistinguible** de no haber
  enrolado: no queda token en disco.
- El re-enrolamiento con un string válido reemplaza el token y NUNCA deja el token viejo en disco
  (sin `.bak` ni temporales con el secreto).
- El token y el enrollment string NUNCA aparecen en stdout, stderr ni logs (salvo el argumento que
  el usuario pega).
- Leído por **stdin**, el enrollment string NUNCA se hace eco a la terminal ni queda en salida
  alguna, y NUNCA aparece en la lista de procesos (argv).

## `permea status`

Informa el estado de enrolamiento. Operación **local**, sin red.

### Salidas

| Estado | Condición | stdout | Exit |
|---|---|---|---|
| Enrolado | `endpoint` y `device_token` presentes, endpoint https | `enrolado` + **URL** del backend | 0 |
| No enrolado | falta token y/o endpoint | `no enrolado` | 0 |

### Garantías (DEBE / NUNCA)

- DEBE informar si está enrolado y, si lo está, contra qué backend (**URL**).
- NUNCA DEBE mostrar el `device_token` en claro ni de forma que permita reconstruirlo (a lo sumo un
  indicador de presencia tipo `token: configurado`).
- No contacta al backend: el estado se deriva de la config local persistida.

## Notas

- Estos verbos se añaden al binario existente **sin** alterar los flags de P-001/P-002
  (`--scan/--run/--daemon/--version`).
- Ningún comando de 003 amplía la allowlist de la frontera ni añade contenido a payload alguno
  (Principio I): el único tráfico que generan es el ping de lote vacío de `enroll`.
- Una URL `http://` (no-https) se rechaza ya en la **decodificación** del enrollment string (antes de
  cualquier ping); el `ErrScheme` del transporte queda como **guardia secundaria** y no llega a
  ejercitarse en el flujo de `enroll`.
