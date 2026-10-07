# 007 · Descubrimiento — medidas y lectura del código

**Fecha**: 2026-10-06 · **Rama**: `007-coste-fiel` sobre `222c824` · **Sólo lectura**: ni código ni tests.

> **Enmienda E-1.** Las secciones A1–A5 se midieron sobre una copia temporal **ya borrada**: sus cifras no se pueden repetir.
> Las referencias de la spec salen ahora de las **copias del dueño**, que se conservan: §«Copia del dueño 2026-10-06-wsl» y
> §«Windows», al final. La sección C contrasta la opción (d) de Q-3 con el código.

> ⛔ **Cómo se midió.** Copia **congelada** de `~/.claude/projects` *(los logs de Claude Code de WSL)* en un directorio de
> `mktemp -d /tmp/permea-007-XXXXXX`, borrado al acabar tras comprobar el prefijo. Un script de Python leyó la copia y escribió
> **sólo** recuentos, sumas, fechas y nombres de campo. No imprimió ninguna línea, ninguna clave de `message.content`, ningún
> identificador *(de mensaje, sesión, petición o herramienta)* ni ninguna ruta. Las líneas que cuentan son las mismas que lee el
> agente: `type = "assistant"` con `message.model` no vacío (`internal/ingest/claudecode.go:85`). Se descartan la sintética y las
> que no traen identificador. Un **mensaje** es el par (`message.id`, `requestId`): la misma clave que usa `derivarEventID`
> (`internal/ingest/eventid.go:49-59`).

## La copia

| Qué | Cifra |
|---|---:|
| Ficheros `.jsonl` | **31** *(3 de subagente)* |
| Líneas facturables | **22 139** |
| …sintéticas (`<synthetic>`) | 1 |
| …sin identificador | 0 |
| …corruptas | 0 |
| **Líneas útiles** | **22 138** *(149 de subagente)* |
| **Mensajes distintos** | **10 101** *(76 de subagente)* |
| …escritos en varias líneas | **7 335** |
| …repartidos entre varios ficheros | **0** |

*(006 midió en su día 23 778 líneas y 11 105 mensajes en otra copia de la misma máquina. Las cifras bajan porque Claude Code
borra los logs viejos: ver A4.)*

## A1 · El desglose de la escritura de caché

| Qué | Líneas |
|---|---:|
| Traen `usage.cache_creation` | **22 138 de 22 138** |
| …con `ephemeral_5m_input_tokens` | 22 138 |
| …con `ephemeral_1h_input_tokens` | 22 138 |
| **5m + 1h ≠ `cache_creation_input_tokens`** | **0** |
| **Sin desglose** | **0** |

| Reparto de tokens | 5 minutos | 1 hora | Total | % de 1 hora |
|---|---:|---:|---:|---:|
| Contando **por línea** *(con repetidas)* | 509 252 | 69 174 515 | 69 683 767 | 99,27 % |
| Contando **por mensaje**, regla «máximo» | **247 506** | **32 462 896** | **32 710 402** | **99,24 %** |
| …de ello, en ficheros **principales** | **0** | 32 462 896 | | 100 % |
| …de ello, en ficheros **de subagente** | **247 506** | **0** | | 0 % |

- **La escritura de 5 minutos sólo aparece en los subagentes**, y en ellos es toda de 5 minutos. Lo mismo que midió P-031 en la
  plataforma el 04-10 *(«la de 5 minutos sólo aparece en subagentes»)*.
- Dentro de un mensaje, el desglose **nunca decrece** de una línea a la siguiente: 0 mensajes.

## A2 · Mensajes escritos en varias líneas: primera línea frente a «máximo por partida»

| Partida | Mensajes donde primera ≠ máximo | Mensajes donde alguna línea decrece | Mensajes donde el máximo ≠ la última |
|---|---:|---:|---:|
| entrada | 0 | 0 | 0 |
| salida | 0 | 0 | 0 |
| escritura de caché | 0 | 0 | 0 |
| lectura de caché | 0 | 0 | 0 |

| Suma por mensaje | entrada | salida | escritura de caché | lectura de caché |
|---|---:|---:|---:|---:|
| primera línea | 26 352 | 11 667 839 | 32 710 402 | 4 493 141 746 |
| máximo | 26 352 | 11 667 839 | 32 710 402 | 4 493 141 746 |
| última línea | 26 352 | 11 667 839 | 32 710 402 | 4 493 141 746 |

⚠️ **En WSL no hay ningún mensaje que crezca**, igual que en la copia de 006 *(0 de 23 778 líneas)*. Así que **esta copia no
distingue** «primera», «máximo» y «última»: las tres reglas dan lo mismo. Lo que sí hay medido está en otra máquina. En **W2**
*(Windows, 006 FR-005, «Hallazgo de W2»)*:
- 143 de 5 286 mensajes crecen, y **sólo en la salida**;
- **la última línea es el máximo en los 143**;
- la primera da 1 306 tokens de salida y la última 89 817.

## A3 · El cierre de un mensaje

| Qué | Cifra |
|---|---:|
| Campo | **`message.stop_reason`** |
| Líneas con `stop_reason` **no nulo** | **21 989**, **todas** de ficheros principales |
| Líneas con `stop_reason` **nulo** | **149**, **todas** de ficheros de subagente |
| Mensajes de varias líneas con `stop_reason` no nulo **ya en la primera** | **7 271 de 7 335** *(los 64 restantes, de subagente)* |
| Mensajes cuya **última** línea trae `stop_reason` nulo («sin cerrarse») | **76**: **todos de subagente**, y los 76 mensajes de subagente lo son |
| …de ellos, el último mensaje de su fichero | 3 |
| Mensajes con una línea **después** de que empiece otro mensaje del mismo fichero | **0** |
| Mensajes con sus líneas **entrelazadas** con las de otro mensaje | **0** |

| Tiempo | n | máx. | p99 | p50 |
|---|---:|---:|---:|---:|
| Primera → última línea, **mensajes de varias líneas** | 7 335 | **259,9 s** | 81,0 s | 2,9 s |
| Primera → última línea, todos los mensajes | 10 101 | 259,9 s | 68,1 s | 1,8 s |
| Hueco entre dos líneas seguidas del mismo mensaje | 12 037 | 259,5 s | 63,4 s | 1,4 s |

⚠️ **En esta copia, `stop_reason` no sirve para saber si un mensaje está cerrado.** En los ficheros principales viene informado
desde la **primera** línea, también en los mensajes que tienen varias. En los de subagente viene **nulo en todas**. Se cumple, en
cambio, que **un mensaje no vuelve a escribirse después de que empiece el siguiente** en su fichero *(0 casos)*. Lo que no está medido
es si en Windows las líneas parciales traen `stop_reason` nulo.

## A4 · Cuánto conserva Claude Code

| Qué | Valor |
|---|---|
| Ficheros | 31 |
| Fecha de modificación más antigua | **2026-09-08 20:38 UTC** |
| Fecha de modificación más nueva | 2026-10-06 17:24 UTC |
| Línea más antigua *(por `timestamp`)* | 2026-09-08 09:50 UTC |
| `cleanupPeriodDays` | **no está declarado** en `~/.claude/settings.json`, en `~/.claude.json` ni en los ajustes de los dos repositorios |

El historial abarca unos **28 días**. Encaja con una retención por defecto de unos 30 días, pero **ese valor por defecto no se ha
comprobado aquí**.

## A5 · Campos de `usage` que hoy no se usan *(sólo la lista; fuera de alcance)*

Los de `usage`: `cache_creation` *(que pasa a usarse)*, `fallback_credit`, `inference_geo`, `iterations`, `output_tokens_details`,
`server_tool_use`, **`service_tier`** y **`speed`**. Hoy se usan `input_tokens`, `output_tokens`, `cache_creation_input_tokens` y
`cache_read_input_tokens`. **`speed`** existe: podría servir algún día para la «Limitación 2» *(el modo rápido)*. No se ha mirado
qué valores toma.

---

## B · Lectura del código: qué toca cada parte

| Sitio | Hoy | Parte (a), el coste por duración | Parte (b), el mensaje entero |
|---|---|---|---|
| `internal/ingest/claudecode.go:31-47` (`rawRecord`) | Decodifica 4 partidas de `usage`; la guarda de frontera, en `:17-24` | **Cambia**: añade `usage.cache_creation.ephemeral_5m_input_tokens` y `…_1h_…` (números, sin contenido) | — |
| `claudecode.go:99-105` | La línea se registra y, si es la primera, se tarifa y se emite | `Cost` recibe las dos escrituras | **Cambia**: la emisión deja de ser inmediata |
| `internal/ingest/pasada.go:30-33` (`consumo`) | 4 enteros | +2 enteros del desglose | — |
| `pasada.go:101-122` (`registrar`) | La **primera** manda; las demás se cuentan; `ConsumoDistinto` | — | **Cambia**: máximo por partida y retención |
| `pasada.go:17-22` | «Entre pasadas no hay memoria» *(006 FR-033)* | — | **Cambia**: la memoria de lo retenido persiste |
| `pasada.go:66-74` (`Resumen`) | Una línea de recuentos | +líneas sin desglose | +mensajes que crecieron, +retenidos |
| `internal/state/state.go:79-122` (`ScanFile`) | Avanza el offset hasta la última línea completa; una línea parcial no se consume (`:103`) | — | **No cambia**. Lo retenido ya queda por debajo del offset, así que hace falta guardarlo en otro sitio |
| `internal/state/state.go:15-25` (`FileState`, `Store`) | `path`, `size`, `mod_time`, `offset` | — | Sin cambio si lo retenido vive en un fichero propio *(Q-3)* |
| `internal/pricing/pricing.go:6-11` (`Rate`) | 4 cifras | **Cambia**: 5 cifras | — |
| `pricing.go:33-50` (`Table`) | 16 claves, espejo de `e50d0a5` | **Cambia**: 17 claves, espejo de `8f147d1` | — |
| `pricing.go:55-62` (`Cost`) | `cacheCreate × CacheWrite` | **Cambia**: `5m × CacheWrite + 1h × CacheWrite1h` | — |
| `pricing_test.go:45-62` (`esperadaDelCatalogo`) y `:65-110` | Espejo de 16 × 4, en tres aserciones | **Cambia**: 17 × 5 | — |
| `internal/event/event.go:17-35` | 17 campos; `SchemaVersion = 1` (`:12`) | **No cambia** *(D-1)* | **No cambia** |
| `internal/ingest/eventid.go:49-85` | La derivación v1 | **No cambia** *(D-1)* | **No cambia** |
| `cmd/permea/main.go:236-292` (`generate`) | Una pasada nueva por llamada; encola al leer; guarda el estado después | — | **Cambia**: carga lo retenido, emite lo que se ha cerrado y guarda lo que sigue abierto |
| `main.go:308-330` (`runOnce`) · `:335-363` (`runDaemon`) · `:367-396` (`tick`) | `generate` + `sync`; el demonio cada `sync_interval` *(60 s por defecto, `internal/config/config.go:47`)* | — | **Cambia**: qué hace `--run` con lo que sigue abierto *(Q-2)* |
| `main.go:406-442` (`dryRun`) | Una pasada por fichero, sin estado ni cola | Imprime el desglose | Al final del fichero emite lo que quede |
| `specs/006-medicion-fiel/contracts/tarifas.md` | Espejo de `e50d0a5` y «Limitación 1» | Lo **sustituye** un contrato de 007 | — |
| `specs/006-medicion-fiel/contracts/event-id.md` | Los vectores normativos | **No cambia** | **No cambia** |

**Lo que hoy garantiza la frontera, y seguirá garantizando**:
- **Un struct cerrado de 17 campos** (`event.go:14-35`) **y un decodificador que sólo lee lo declarado** (`claudecode.go:13-24`): lo
  que no se declara no entra en el proceso.
- **Los identificadores del proveedor sólo entran en el hash** (`claudecode.go:26-30`; `eventid.go:22-27`).
- **Tres tests**: el golden y la denylist (`boundary_test.go:89`), el de un campo futuro desconocido que no se filtra (`:273`) y el de
  los tres caminos hacia el exterior (`:166`).

Los dos campos nuevos de `rawRecord` son **números de consumo**, no contenido. La lista de retenidos guarda **campos del evento ya
derivado** y el `event_id`, nunca el identificador del proveedor *(spec FR-011)*.

---

# Enmienda E-1 · Las copias del dueño

> ⛔ **Cómo se midió.** Las dos copias congeladas del dueño, **leídas en su sitio**: *copia del dueño 2026-10-06-wsl* (31 `.jsonl`)
> y *copia del dueño 2026-10-06-windows* (29). No se copiaron ni se modificaron. Antes y después se tomó una huella SHA-256 del
> conjunto de cada copia y del número de ficheros, y **coinciden**. Cuatro scripts de Python, ejecutados con `python3 -I` desde un
> directorio aparte, escribieron **sólo** recuentos, sumas, duraciones, fechas, versiones de Claude Code, nombres de campo y los
> valores de `stop_reason`. Ninguna línea, ningún identificador, ninguna ruta. Las mismas reglas de arriba: `type = "assistant"`
> con modelo, sin `<synthetic>` ni líneas sin identificador; un **mensaje** es (`message.id`, `requestId`). Un fichero es **de
> subagente** si está bajo un directorio `subagents/` *(3 en WSL, 9 en Windows; coincide con los que se llaman `agent-*.jsonl`)*.
> Una línea **crece** si alguna partida supera a todas las anteriores del mensaje.

## Copia del dueño 2026-10-06-wsl · A1–A3 repetidos

| Qué | Cifra | | Qué | Cifra |
|---|---:|---|---|---:|
| Ficheros | 31 *(3 de subagente)* | | Mensajes | **10 121** *(76 de subagente)* |
| Líneas facturables | 22 181 | | …de varias líneas | 7 350 *(64 de subagente)* |
| …sintéticas · sin identificador · corruptas | 1 · 0 · 0 | | …repartidos entre ficheros | 0 |
| **Líneas útiles** | **22 180** *(149 de subagente)* | | Líneas de más de 1 MiB | 0 |

**A1.** Desglose en **22 180 de 22 180** líneas; **0** no suman; **0** sin desglose. Por línea: 509 252 a 5 min *(todo de subagente)* y
69 296 082 a 1 h *(todo principal)*. **Por mensaje** *(máximo, con el desglose de la línea que da el máximo de la escritura)*:
**247 506** a 5 min y **32 518 296** a 1 h → **99,24 %** a 1 hora. El desglose no cambia dentro de ningún mensaje *(0)*.

**A2.** Primera = máximo = última en las cuatro partidas: **0** diferencias, **0** decrecimientos.

| Suma por mensaje *(las tres reglas)* | entrada | salida | escritura de caché | lectura de caché |
|---|---:|---:|---:|---:|
| primera = máximo = última | **26 392** | **11 713 955** | **32 765 802** | **4 499 158 833** |

**A3.** `stop_reason` informado en las 22 031 líneas principales y nulo en las 149 de subagente. Los 7 286 mensajes principales de
varias líneas lo traen **desde la primera**; los 64 de subagente, **nunca**. 76 mensajes terminan sin cerrarse, todos de subagente.
Regla (i): **0** mensajes reciben una línea después de que empiece otro. **0** entrelazados. Primera → última, mensajes de varias
líneas: máx. **259,9 s** · p99 81,6 s · p50 2,9 s. Huecos ≥ 5 min: **0**.

**A4.** Líneas del 2026-09-08 09:38 UTC al 2026-10-06 17:36 UTC *(unos 28 días)*.

*(Frente a la copia temporal de arriba, 20 mensajes más: la copia del dueño se hizo unos minutos después.)*

## Windows · copia del dueño 2026-10-06-windows

| Qué | Cifra | | Qué | Cifra |
|---|---:|---|---|---:|
| Ficheros | 29 *(9 de subagente)* | | Mensajes | **6 074** *(219 de subagente)* |
| Líneas facturables | 13 869 | | …de varias líneas | 4 428 *(187 de subagente)* |
| …sintéticas · sin identificador · corruptas | 2 · 0 · 0 | | …repartidos entre ficheros | **0** |
| **Líneas útiles** | **13 867** *(599 de subagente)* | | Líneas de más de 1 MiB | 0 |

Líneas del **2026-09-06** 16:08 UTC al 2026-10-06 17:39 UTC *(unos 30 días)*. Ficheros sin salto de línea final: 0.

### W-A1 · El desglose

| Qué | Líneas |
|---|---:|
| Traen `cache_creation` con las dos claves | **13 867 de 13 867** |
| **5m + 1h ≠ `cache_creation_input_tokens`** | **0** |
| **Sin desglose** | **0** |

| Reparto de tokens | 5 minutos | 1 hora | Total | % de 1 hora |
|---|---:|---:|---:|---:|
| Por **línea** *(con repetidas)* | 2 915 631 | 51 619 599 | 54 535 230 | 94,65 % |
| Por **mensaje**, regla «máximo» | **971 559** | **21 335 690** | **22 307 249** | **95,64 %** |
| …en ficheros **principales** | **0** | 21 335 690 | | 100 % |
| …en ficheros **de subagente** | **971 559** | **0** | | 0 % |

El mismo patrón que en WSL: 5 minutos sólo en subagentes, 1 hora sólo en principales. El desglose no cambia dentro de ningún
mensaje *(0)*. La proporción baja a 95,64 % porque Windows tiene más trabajo de subagente.

### W-A2 · Primera, máximo y última

| Partida | primera ≠ máximo | alguna línea decrece | máximo ≠ última |
|---|---:|---:|---:|
| entrada | 0 | 0 | 0 |
| **salida** | **143** | 0 | **0** |
| escritura de caché | 0 | 0 | 0 |
| lectura de caché | 0 | 0 | 0 |

| Suma por mensaje | entrada | salida | escritura de caché | lectura de caché |
|---|---:|---:|---:|---:|
| primera *(la 0.3.0)* | 12 206 | 6 635 290 | 22 307 249 | 2 423 738 410 |
| **máximo** | **12 206** | **6 723 801** | **22 307 249** | **2 423 738 410** |
| última | 12 206 | 6 723 801 | 22 307 249 | 2 423 738 410 |
| *los 143 que crecen*: primera → máximo | 286 → 286 | **1 306 → 89 817** | 752 812 → 752 812 | 9 008 614 → 9 008 614 |

- **143** mensajes crecen, **sólo en la salida**, y **los 143 en ficheros de subagente** *(0 de 5 855 principales)*. Están en los **9**
  ficheros de subagente, entre el **2026-09-27 y el 2026-10-02**. Son **los mismos 143** del «Hallazgo de W2», con las mismas cifras
  *(1 306 → 89 817; faltan 88 511, el 1,32 % de la salida de la copia)*.
- **Ninguno decrece. El máximo es la última en los 143.** Tienen 457 líneas; en 128 la subida está en la última, y en los 15 restantes
  la sigue una línea idéntica.
- ⚠️ **No se puede separar «Windows» de «versión».** Los subagentes de Windows son todos de Claude Code **2.1.283 a 2.1.287**, y todas
  esas versiones tienen mensajes que crecen. Los de WSL son todos de **2.1.288** *(posteriores al 10-02)*, y ninguno crece. Windows
  no tiene ninguna línea de subagente después del 10-02.

### W-A3 · `stop_reason`

| Qué | Cifra |
|---|---:|
| Líneas principales con `stop_reason` informado | **13 268 de 13 268** |
| Líneas de subagente: informado · nulo | **173** · 426 |
| Mensajes principales de varias líneas con `stop_reason` **desde la primera** | **4 241 de 4 241** |
| Mensajes de subagente de varias líneas: lo traen **más tarde** · **nunca** | **143** *(los que crecen)* · 44 |
| Mensajes de subagente de una línea: informado · nulo | 15 · 17 |
| **En los 143 que crecen**: líneas no finales con `stop_reason` **nulo** | **299** *(todas antes de la subida)* |
| …líneas no finales **informadas** | **15** *(todas después de la subida, idénticas a la final)* |
| …líneas finales **informadas** | **143 de 143** *(`tool_use`)* |
| …mensajes donde `stop_reason` llega **antes** que el máximo | **0** |
| …mensajes donde llega **en la misma línea** que el máximo | **143** |
| Mensajes que terminan sin cerrarse *(última línea con `stop_reason` nulo)* | **61**, todos de subagente *(44 de varias líneas y 17 de una)* |

⚠️ **Esto contradice dos cosas de la spec**: que en los subagentes `stop_reason` viene «siempre nulo» *(en Windows llega en 173
líneas)* y que «en Windows cerraría en la línea parcial» *(0 de 143: las parciales lo traen nulo y llega con la subida)*. Pero no
sirve solo: en los principales viene desde la primera línea y 61 mensajes de subagente no lo traen nunca.

### W-A4 · El cierre anticipado *(P-3)*

**Regla (i)**, «empieza otro mensaje en el mismo fichero»:

| Qué | Cifra |
|---|---:|
| Mensajes que reciben una línea **después** de que empiece otro en su fichero | **0** |
| …de ellos, que crecen en esa línea | **0** |
| Ficheros con mensajes entrelazados | **0** |
| Máximo de mensajes abiertos a la vez en un fichero | **1** *(en los 28 ficheros con mensajes)* |

**Regla (ii)**, «hueco ≥ T sin líneas»:

| Mensajes con un hueco entre dos líneas seguidas ≥ T | T = 5 min | T = 10 min | T = 30 min |
|---|---:|---:|---:|
| todos | **0** | **0** | **0** |
| los que crecen | **0** | **0** | **0** |

| Tiempo *(por `timestamp`)* | n | máx. | p99 | p50 |
|---|---:|---:|---:|---:|
| Primera → última, todos los mensajes | 6 074 | **144,6 s** | 52,9 s | 1,3 s |
| Primera → última, mensajes de varias líneas | 4 428 | 144,6 s | 65,1 s | 2,1 s |
| **Primera → última, los que crecen** | **143** | **90,9 s** | 60,1 s | 1,9 s |
| Hueco entre líneas seguidas, todos | 7 793 | 144,2 s | 45,4 s | 1,1 s |
| Hueco entre líneas seguidas, los que crecen | 314 | 90,9 s | 3,7 s | 0,8 s |

En los 143, la última línea tiene un `timestamp` **distinto** del de la primera: el reloj de T, medido sobre la última línea, avanza
con el mensaje. *(La diferencia entre el `mtime` de cada fichero y su último `timestamp` va de 0,07 s a 10 h: hay líneas sin
`timestamp` al final de algunos ficheros, así que no mide el retraso de escritura.)*

### W-A5 · Sin cerrarse y repartidos

61 mensajes terminan sin `stop_reason` *(todos de subagente)*; **0** mensajes repartidos entre ficheros, también en WSL. Un `--scan`
por fichero ve cada mensaje entero.

## C · La opción (d) de Q-3: no avanzar el punto de lectura

**Qué es.** El offset de un fichero no pasa del comienzo de la primera línea de su mensaje abierto. La pasada siguiente lo relee del log.
No hay fichero de retenidos.

**Lo que dice el código.**
- `ScanFile` (`internal/state/state.go:79-122`) llama a `fn(line)` y avanza `consumed` tras **cada** línea completa (`:109`). El callback
  no sabe dónde empieza la línea ni puede retener el offset. Al final guarda `Offset: consumed` (`:115-120`). **Hay que cambiarlo**: el
  callback recibe el comienzo de cada línea, y quien llama fija el offset al terminar *(p. ej. un método nuevo junto a `ScanFile`, con
  `ScanFile` como envoltorio que lo deja todo como hoy)*.
- `generate()` (`cmd/permea/main.go:236-292`) encola en el callback (`:277`) y guarda el estado al final (`:288`). Con (d): encola lo que
  se cierra al leerlo *(regla (i))* o al final del fichero *(regla (ii))*, y fija el offset del fichero en el comienzo del mensaje que
  sigue abierto. **El orden de durabilidad es el de hoy**: encolar y después guardar `state.json` (`:231-232`).
- El demonio (`main.go:352-362`) llama a `tick()` → `generate()` cada `sync_interval`, y `generate()` recorre **todos** los ficheros
  (`:267`). Un fichero con un mensaje abierto tiene `offset < size`, así que se relee su cola en cada ciclo, crezca o no: así se cierra por T.
- `FileState` (`state.go:15-20`) ya guarda el `Size` de la pasada anterior. Una línea que empieza por debajo de ese `Size` ya se leyó:
  con eso se distinguen las **releídas** sin añadir ningún campo.
- **Truncado** (`state.go:85-87`): con el offset retenido, la detección `size < offset` sigue igual. Lo que no puede hacer (d) es
  «emitir lo retenido tal como esté»: si el fichero se trunca o se borra con un mensaje abierto, sus líneas ya no existen.

**Lo que dice la medida.**
- **Mensajes abiertos por fichero.** Con la regla (i), a lo sumo **uno**: el último que empezó. Medido: máximo **1** en todos los
  ficheros de las dos copias, 0 entrelazados. Si algún día hubiera más de uno, la regla (i) cierra el anterior al empezar el
  siguiente, y sus líneas posteriores cuentan como tardías. Pasa lo mismo con (a).
- **Lo que se relee por pasada** *(desde la primera línea de un mensaje hasta que empieza el siguiente o acaba el fichero)*: p50 11–14 KB,
  p99 89–96 KB, máximo **1,16 MB** *(Windows)*. El último mensaje de cada fichero, hasta el final: máximo 51 KB. Un mensaje se relee
  como mucho hasta T *(10 ciclos del demonio a 60 s)*. Los ficheros llegan a 17,8 MB.

**(a) frente a (d)**

| | (a) `pendientes.json` | (d) offset retenido |
|---|---|---|
| **Cómo se pierde un mensaje** | `pendientes.json` borrado o corrupto con `state.json` intacto: sus líneas quedan por debajo del offset y no se releen nunca, **sin aviso**. Guardarlo **después** de `state.json` y caer entre medias: lo mismo. Si está corrupto, hay que elegir entre parar la pasada o perder | El fichero se **trunca o se borra** mientras su último mensaje está abierto *(≤ T tras su última línea, o hasta la pasada siguiente)*. Lo medido *(A4)* encaja con que Claude Code borra ficheros viejos *(~30 días)* y no los trunca; no está comprobado. `state.json` perdido: se relee todo, como hoy, y la plataforma descarta |
| **Estado nuevo en disco** | Un fichero, con su formato, su escritura atómica, su orden respecto de `state.json` y su prueba de privacidad | **Ninguno**. `state.json` sigue con sus cuatro campos |
| **Complejidad** | Cargar, fusionar *(máximo entre lo guardado y lo releído tras una caída)*, guardar, limpiar entradas de ficheros que desaparecen, versionar el formato | Comunicar el comienzo de cada línea y fijar el offset; contar aparte las releídas; aceptar la relectura de la cola |
| **FR-011** | Se cumple con el fichero | Se reescribe: lo abierto no se guarda, se relee |
| **FR-012** | Tres pasos: encolar → guardar pendientes → guardar estado *(ese orden es obligatorio)* | Dos pasos, los de hoy: encolar → guardar estado |
| **FR-013** | Dentro de la pasada y de lo retenido | Dentro de la pasada; lo abierto se reconstruye en cada pasada, así que el efecto es el mismo |
| **SC-009** | Añade el fichero de retenidos con centinelas | Esa cláusula sobra: `state.json` no cambia de forma |
| **Tests** | `state_test.go` fuera | `state_test.go` **fuera** si `ScanFile` se conserva como envoltorio; los del método nuevo, en un fichero nuevo |

**Veredicto de la lectura**: la hipótesis del orquestador **se sostiene**. (d) no añade un fichero cuya pérdida pierde eventos y conserva
el orden de durabilidad de hoy. Lo que cuesta: un modo de pérdida propio, teórico *(truncado o borrado con un mensaje abierto)*, la
relectura de la cola y un recuento nuevo en el resumen.
