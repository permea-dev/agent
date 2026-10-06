# 007 · «Coste fiel» — Especificación

**Feature Branch**: `007-coste-fiel` · **Created**: 2026-10-06 · **Status**: Ratificada el 2026-10-06 *(E-2)* · **Cabeza de partida**: `222c824` · **Enmiendas**: E-1, E-2, E-3
**Input**: las decisiones del dueño `D-1`…`D-3` · el descubrimiento del 06-10 *(`soporte/descubrimiento.md`; las referencias, desde E-1,
sobre la **copia del dueño 2026-10-06-wsl** y la **copia del dueño 2026-10-06-windows**)* · el «Hallazgo de W2» de 006 *(`specs/006-medicion-fiel/spec.md:377-389`)* · el catálogo de la plataforma
`permea-dev/permea-platform` · `backend/config/pricing.php` · **`8f147d1`** *(P-031, en producción desde el 2026-10-06)*.

> ⛔ Toda afirmación sobre el código lleva `fichero:línea`, medida sobre `222c824`. Las propuestas eran **✋**; desde **E-2** son
> decisiones **✅**, y conservan sus alternativas como rastro. Lo que añade la spec sin que nadie lo dijera va marcado **`DECIDÍ YO`**.
> §Preguntas abiertas queda **a cero** *(E-2)*.

## Contexto — la `0.3.0` tiene dos defectos de coste

1. **Tarifa toda la escritura de caché a 5 minutos.** `Cost` multiplica `cacheCreate × CacheWrite` (`internal/pricing/pricing.go:55-62`),
   y la cabecera lo declara como «Limitación 1» (`:27-28`). De **1 hora** es el **99,24 %** de la escritura de caché en la copia del
   dueño 2026-10-06-wsl, y el **95,64 %** en la -windows *(2 × la entrada, frente a 1,25 × la de 5 minutos)*. El coste observado sale **infravalorado** en esa partida. La plataforma
   lo usa tal cual y no lo recalcula *(P-012 FR-005)*.
2. **Cuenta la primera línea de un mensaje.** `registrar` emite la primera y descarta las demás (`internal/ingest/pasada.go:101-122`;
   006 FR-005). En **W2**, 143 de 5 286 mensajes crecen en la salida, y faltan **88 511** tokens *(1,6 %)*. Son **conversaciones de
   subagente** de Claude Code 2.1.283–2.1.287: en la copia del dueño 2026-10-06-windows están **los mismos 143** *(de 6 074)*, y en la
   -wsl, cuyos subagentes son todos de 2.1.288, **0**. **No se atribuye al sistema** *(N-5)*.

**El log ya trae lo que falta**: cada línea facturable de las dos copias trae `usage.cache_creation.ephemeral_5m_input_tokens` y
`…ephemeral_1h_input_tokens`, y su suma es exactamente `cache_creation_input_tokens` *(22 180 de 22 180 en -wsl; 13 867 de 13 867 en -windows)*.

## Decisiones del dueño *(no se discuten aquí)*

| # | Decisión |
|---|---|
| **D-1** · Sin contrato | El evento *(17 campos, `internal/event/event.go:17-35`)*, `SchemaVersion = 1` (`:12`), la allowlist y la derivación del `event_id` (`internal/ingest/eventid.go:49-85`; `specs/006-medicion-fiel/contracts/event-id.md`) **no cambian**. `tokens_cache_creation` sigue siendo **la suma** de las dos duraciones. |
| **D-2** · Dos partes, una publicación | **(a)** el coste tarifa la escritura de caché según su duración; **(b)** un mensaje escrito en varias líneas se cuenta entero. |
| **D-3** · Tarifas en espejo | De `permea-dev/permea-platform` · `backend/config/pricing.php` · **`8f147d1`**: **17 claves y cinco cifras** (`input`, `output`, `cache_write`, `cache_write_1h`, `cache_read`). |

## Propuestas — ratificadas por el dueño el 2026-10-06 a las 20:00 (Madrid) *(E-2)*

| # | Propuesta → **decisión** | Alternativas *(rastro)* | Recomendación *(rastro)* |
|---|---|---|---|
| **P-1** ✅ **(a)** | **Una línea sin desglose**: toda su escritura de caché va a **1 hora** *(la misma hipótesis que declara la cabecera del catálogo, `pricing.php@8f147d1:19-30`)* | (a) a 1 hora · (b) a 5 minutos, como la 0.3.0 · (c) coste no disponible | **(a)**. Es coherente con la plataforma, y hoy **no hay ninguna** línea así *(0 en las dos copias del dueño)*: es un caso de defensa, no de volumen |
| **P-2** ✅ **(a)** | **Regla del mensaje**: cada partida vale **el máximo** entre sus líneas; **nunca se suman** | (a) el máximo por partida · (b) la última línea · (c) la primera, como la 0.3.0 | **(a)**. **Windows** *(E-1, W-A2)*: 143 mensajes crecen, sólo en la salida y sólo en subagentes; el máximo es la última en los 143; **0** decrecen; la primera da 1 306 y el máximo 89 817. En WSL las tres reglas coinciden *(0 diferencias, 0 decrecimientos)*. La medida no pide otra regla: «máximo» y «última» dan lo mismo en las dos copias. El máximo no depende del orden de lectura y no baja si una línea llega fuera de sitio. **Ajuste de la medida**: el desglose 5 m / 1 h se toma **de la línea que da el máximo de la escritura de caché** *(la última, si empatan)*, para que la suma siga siendo exacta *(`DECIDÍ YO`)* |
| **P-3** ✅ **(a) sin (d)**, con Q-1 (a), Q-6 (a) y la regla (ii) de FR-010 *(E-2)* | **Retención**: un mensaje se emite **cuando está cerrado**. Lo que sigue abierto **no se consume**: el offset de su fichero se queda en el comienzo del mensaje y la pasada siguiente lo relee *(`--run` seguidos y ciclos de `--daemon`; Q-3 (d), E-1)*. «Cerrado» = **(i)** empieza **otro mensaje posterior en el mismo fichero**, o **(ii)** pasan **T = 10 minutos, fijo**, sin líneas nuevas suyas **ni cambios en su fichero** *(FR-010, E-2)*, o **(iii)** pasan **24 horas** desde su última línea, cambie o no su fichero *(tope de la espera, E-3)*. `stop_reason` **no** cuenta como cierre *(Q-1)* | T: (a) 10 min · (b) 5 min · (c) 30 min · señal: (d) además `stop_reason` no nulo | **(a) sin (d)**. Primera → última línea de un mensaje: máx. **259,9 s** en WSL *(p99 81,6 s)* y **144,6 s** en Windows *(90,9 s en los 143 que crecen)*; 10 min es más del doble. **0** huecos ≥ 5 min en las dos copias. Regla (i): **0** mensajes reciben una línea tras empezar el siguiente, también en Windows *(W-A4)*. `stop_reason`: ver Q-1 |
| **P-4** ✅ **(a)** | **Versión `0.4.0`**. Lo ya enviado **no se corrige ni se reenvía**: un mensaje que la 0.3.0 envió con su primera línea conserva su `event_id`, y la plataforma descarta el repetido por (`org_id`, `event_id`). **Se declara** en el CHANGELOG | (a) 0.4.0 · (b) 0.3.1 | **(a)**. Cambia el coste que emite el agente y su conducta entre pasadas. No es un parche |
| **P-5** ✅ | **Los textos nuevos que ve el usuario**, literales, en §Textos aprobados *(el CHANGELOG, sustituido en E-2)* | aprobarlos tal cual o corregirlos | Aprobarlos **antes de escribir sus tests** *(como D-006-14)* |

## User Scenarios & Testing

### Historia 1 — La escritura de caché sale a su precio *(P1)*
1. **Dado** una línea con 12 345 tokens de escritura a 5 minutos y 33 334 a 1 hora de `claude-opus-5-5`, **entonces** el coste usa 5,00 para la
   primera cifra y **8,00** para la segunda, y `tokens_cache_creation` = 45 679.
2. **Dado** una línea **sin desglose**, **entonces** toda su escritura va a la tarifa de 1 hora *(P-1)*.

### Historia 2 — Un mensaje escrito en varias líneas se cuenta entero *(P1)*
1. **Dado** un mensaje de tres líneas cuya salida crece *(7 → 1 303 → 89 817)*, **entonces** sale **un** evento con salida **89 817**.
2. **Dado** que la línea parcial se lee en una pasada y la final en la siguiente, **entonces** sigue saliendo **un** evento, con la final.

### Historia 3 — Actualizar desde la 0.3.0 no reenvía nada *(P1)*
1. **Dado** una instalación 0.3.0 con estado y cola, **cuando** se actualiza, **entonces** no se relee el historial, la cola se envía tal cual,
   y un mensaje que ya salió con su primera línea no se vuelve a emitir con otro `event_id`.

### Edge Cases
- **El último mensaje de un fichero, sin otro detrás**: se cierra por inactividad: T desde su última línea **y** desde el último cambio
  del fichero *(FR-010, E-2)*; y, aunque el fichero siga cambiando, a las 24 h de su última línea *(FR-010 (iii), E-3)*.
- **Un subagente**: `stop_reason` nulo en las líneas parciales, y en 61 de 219 mensajes de Windows y en todos los de WSL también en la
  última *(W-A3)*; se cierra por (i) o (ii), igual que los demás.
- **Una línea llega después de que su mensaje se emitiera** *(tras T)*: dentro de la pasada, cuenta como tardía y no se emite *(FR-013)*.
  En una pasada posterior no hay memoria: sale con el **mismo** `event_id` y la plataforma lo descarta.
- **El fichero se trunca o rota** (`internal/state/state.go:85-87`): se relee desde 0, como hoy. Si tenía un mensaje abierto y sus líneas
  ya no están, ese mensaje **se pierde**. Es el modo de pérdida propio de Q-3 (d), sin casos medibles *(E-1, `DECIDÍ YO`)*.
- **El desglose no suma el total**: hoy hay 0 casos *(A1)*; regla en FR-004 *(Q-4)*.

## Requirements

### A · El coste por duración *(D-2 (a), D-3)*

- **FR-001** *(lo que se lee)*: `rawRecord` (`internal/ingest/claudecode.go:31-47`) **DEBE** decodificar también
  `message.usage.cache_creation.ephemeral_5m_input_tokens` y `…ephemeral_1h_input_tokens`. Son números de consumo; la guarda de frontera
  (`:17-24`) **NO** se relaja para nada más.
- **FR-002** *(cinco cifras)*: `Rate` (`internal/pricing/pricing.go:6-11`) **DEBE** llevar `CacheWrite1h`. `Cost` **DEBE** tarifar la
  escritura de 5 minutos a `CacheWrite` y la de 1 hora a `CacheWrite1h`. Las otras tres partidas, como hoy.
- **FR-003** *(sin desglose — P-1 ✅)*: una línea sin las dos claves del desglose **DEBE** tarifar toda su escritura a `CacheWrite1h`, y
  **DEBE** contarse en el resumen de la pasada.
- **FR-004** *(desglose incoherente — Q-4 ✅ (a))*: si 5 m + 1 h ≠ `cache_creation_input_tokens`, la línea **DEBE** tratarse como **sin desglose**
  *(FR-003)*, y `tokens_cache_creation` es el total del log.
- **FR-005** *(el evento no cambia — D-1)*: `tokens_cache_creation` = `cache_creation_input_tokens`. El desglose **NUNCA** cruza la frontera.
- **FR-006** *(espejo — D-3)*: `Table` (`pricing.go:33-50`) **DEBE** tener las **17** claves de `pricing.php@8f147d1` con sus **cinco**
  cifras, incluida `claude-fable-5-1` *(10,00 / 50,00 / 12,50 / 20,00 / 0,25)*. Lo comprobado el 06-10 con `php -r` sobre ese commit:
  17 filas, `cache_write_1h` = 2 × la entrada en todas.
- **FR-007** *(cabecera)*: la cabecera de `Table` (`pricing.go:13-32`) **DEBE** decir:
  - el commit replicado: `8f147d1`;
  - la verificación (2026-10-04) y la aprobación de la plataforma;
  - la **hipótesis de P-1** en lugar de la «Limitación 1»;
  - la «Limitación 2» *(modo rápido)*, que sigue.
- **FR-008** *(vigilancia)*: la tabla esperada del test (`internal/pricing/pricing_test.go:45-62`) **DEBE** pasar a 17 × 5, escrita **aparte**
  y con su procedencia. Las tres aserciones (`:65-110`) comparan las cinco cifras. El contrato de tarifas de 006 queda **sustituido** por
  `specs/007-coste-fiel/contracts/tarifas.md`, con los mismos cinco sitios por cambio de fila *(006 `contracts/tarifas.md` §«Cómo se absorbe»)*.

### B · El mensaje entero *(D-2 (b))*

- **FR-009** *(regla — P-2 ✅)*: el evento de un mensaje **DEBE** llevar, en cada partida, el **máximo** entre sus líneas leídas. El desglose
  de la escritura de caché, de la línea que da ese máximo *(la última, si empatan)*. Las líneas **NUNCA** se suman.
- **FR-010** *(retención — P-3 ✅)*: un mensaje **NUNCA** se emite antes de estar cerrado: **(i)** ha empezado otro mensaje posterior en el
  mismo fichero, o **(ii)** han pasado **≥ T** *(10 min, fijo)* **desde el `timestamp` de su última línea Y desde la última modificación de
  su fichero** *(el `os.Stat` de la pasada, `internal/state/state.go:80`; no el `ModTime` guardado, que va en segundos y nadie lee)*, las dos
  contra el reloj del agente, que **DEBE** poder inyectarse en los tests *(E-2: la medida no pudo contrastar el `timestamp` con la hora
  real de escritura)*, o **(iii)** han pasado **≥ 24 horas** desde el `timestamp` de su última línea, **cambie o no su fichero**, contra el
  mismo reloj *(E-3: tope de la espera)*. Efectos declarados: el último mensaje espera a que **toda la conversación** lleve T sin escribirse,
  y nunca más de 24 h *(la (iii) cierra el caso de un fichero que siga cambiando sin líneas nuevas, `plan.md` R-2)*; un `mtime` que se
  quede atrás *(Windows, sin comprobar)* sólo devuelve la regla (ii) a la del `timestamp`, nunca cierra antes que ella.
- **FR-011** *(memoria entre pasadas — Q-3 ✅ (d))*: lo abierto **NO se guarda**. El offset de un fichero **NO DEBE** pasar del
  comienzo de la primera línea de su mensaje abierto, y la pasada siguiente lo relee del log. `state.json` conserva sus cuatro campos
  (`internal/state/state.go:15-20`) y **no gana ninguno**. `ScanFile` (`:79-122`) se conserva como envoltorio con su conducta de hoy; el
  comienzo de cada línea y el offset retenido van en un método nuevo *(`DECIDÍ YO`: así `state_test.go` queda fuera del censo)*.
- **FR-012** *(orden de la durabilidad)*: en `generate()` (`cmd/permea/main.go:236-292`), lo cerrado se **encola antes** de guardar el
  estado. Una caída entre medias puede **re-encolar** *(la plataforma descarta el repetido)*, pero **nunca perder** un mensaje
  *(el mismo criterio y los mismos dos pasos de hoy, `:231-232`)*.
- **FR-013** *(una vez)*: el agente **NUNCA** encola dos veces el mismo `event_id` **en la misma pasada**.
  Una línea de un mensaje ya emitido se cuenta como **tardía** y no se emite. Entre pasadas, tras emitirlo, no guarda memoria de lo
  emitido: rige 006 FR-033, y la plataforma descarta *(`DECIDÍ YO`: guardar todo lo emitido sería un estado que crece sin límite)*.
- **FR-014** *(`--run` con mensajes abiertos — Q-2 ✅ (a))*: al terminar `--run`, lo que siga abierto **DEBE** quedarse sin consumir para la
  pasada siguiente *(recomendación de Q-2)*; el resumen dice cuántos.
- **FR-015** *(`--daemon`)*: cada ciclo relee desde el offset retenido, emite lo cerrado y deja el resto sin consumir. Un mensaje abierto
  se cierra por T aunque su fichero no crezca: su cola se relee en cada ciclo *(`offset < size`)*.
- **FR-021** *(relectura — E-1; ✅ E-2, decisión de método)*: una línea que empieza por debajo del `Size` guardado de su fichero *(`state.go:17`)* es
  **releída**. Cuenta en la segunda línea del resumen *(FR-017)*, y el demonio **no** escribe el resumen de un ciclo que sólo ha releído
  *(hoy lo escribe si hay facturables, `main.go:376`)*: lo escribe si hay facturables **no releídas** o eventos emitidos. Medido: lo releído por pasada va de p50 11–14 KB a un máximo de 1,16 MB, durante T
  como mucho *(soporte, §C)*.
- **FR-016** *(`--scan`)*: el dry-run (`main.go:406-442`) **DEBE** aplicar FR-001…FR-009, y **al final del fichero emite todo** lo que tenga:
  la copia está completa *(`DECIDÍ YO`)*. **DEBE** imprimir por evento el desglose, además de las cuatro partidas de hoy. Es el instrumento
  de SC-001 a SC-003.
- **FR-017** *(resumen — P-5 ✅)*: el resumen de la pasada (`pasada.go:66-74`) **DEBE** pasar a **dos líneas**: la de hoy, **sin cambiar
  un byte**, y una segunda con los mensajes que crecieron entre líneas, los que esperan a cerrarse, las líneas releídas, las tardías y
  las que no traen desglose. Sólo recuentos: **nunca** identificadores ni rutas.

### C · La frontera y la actualización

- **FR-018** *(frontera — D-1)*: el golden y la denylist (`internal/ingest/boundary_test.go:89`), el campo futuro (`:273`) y los tres caminos
  (`:166`) **NO** se tocan y siguen en verde. Los vectores de `eventid_test.go` tampoco.
- **FR-019** *(actualización — P-4 ✅)*: actualizar desde la 0.3.0 **NUNCA** relee lo que ya está por debajo del offset *(006 FR-009)*, y la
  cola se envía tal cual. La primera pasada de la 0.4.0 empieza **sin** retenidos.
- **FR-020** *(publicación)*: `0.4.0`, con el CHANGELOG aprobado *(P-5 ✅)*, encabezado `## 0.4.0 — PENDIENTE` hasta el día de la etiqueta. El README **retira** la «Limitación conocida» de W2 y la de la escritura de caché
  a 5 minutos, y declara P-1. Lo que deba ir en el paquete va **antes** de la etiqueta *(cosecha 299 de la plataforma)*.

### Key Entities
- **Mensaje abierto**: el máximo por partida, el desglose 5 m / 1 h, el `event_id` y el `timestamp` de su última línea. Vive **sólo en
  memoria, durante la pasada**; entre pasadas lo que persiste es el offset retenido *(FR-011, E-1)*.
- **Tarifa**: cinco cifras por millón de tokens.

## Success Criteria

- **SC-001** *(recuento independiente, cuatro partidas, WSL)*: sobre una copia congelada de los logs de WSL, el `--scan` de la 0.4.0 suma
  **lo mismo en cada partida** que un contador **independiente** que aplica «máximo por partida» y **no comparte código** con el agente
  *(cosecha 297 de la plataforma)*. Referencia, **copia del dueño 2026-10-06-wsl**: **10 121** mensajes · entrada **26 392** · salida
  **11 713 955** · escritura **32 765 802** · lectura **4 499 158 833**. ⚠️ En WSL no crece ningún mensaje: esta SC **no distingue** las reglas.
  La que lo hace es SC-002.
- **SC-002** *(lo que se recupera — falsable)*: con un fixture sintético de **un mensaje en tres líneas** con salida 7 → 1 303 → 89 817, el evento
  lleva **89 817**. Si se sustituye «máximo» por «primera», cae. Si se sustituye por «última» en un fixture con la línea mayor **en medio**,
  también cae. Y sobre la **copia del dueño 2026-10-06-windows** *(Q-5, E-1)*, el `--scan` y el contador coinciden en las cuatro partidas:
  **6 074** mensajes · entrada **12 206** · salida **6 723 801** · escritura **22 307 249** *(971 559 a 5 min y 21 335 690 a 1 h)* · lectura
  **2 423 738 410**. Los **143** mensajes que crecen suman **89 817** de salida y no 1 306; con «primera», la salida total cae a 6 635 290.
- **SC-003** *(un evento por mensaje)*: en la copia de SC-001, el número de eventos del `--scan` es **igual** al de mensajes distintos *(10 121)*.
  Ningún `event_id` se repite. Si se duplica una línea a propósito, el recuento no cambia.
- **SC-004** *(coste a mano, cinco cifras, vector irregular, comparación absoluta ≤ 1e-9)*: `claude-opus-5-5` con 123 457 de entrada,
  7 891 de salida, **12 345 a 5 min**, **33 334 a 1 h** y 987 653 de lectura:
  - 0,493828 + 0,157820 + 0,061725 + **0,266672** + 0,1975306 = **1,1775756 USD**.
  - **Sin desglose** *(45 679 a 1 h, P-1)*: 0,493828 + 0,157820 + **0,365432** + 0,1975306 = **1,2146106 USD**. Es la misma cifra que la
    plataforma da al reparar ese vector, redondeada a 6 decimales: `'1.214611'` *(P-031 SC-001)*.
  - Si se intercambian `CacheWrite` y `CacheWrite1h`, cae: daría 0,061725 → 0,098760 y 0,266672 → 0,166670.
- **SC-005** *(reparto sobre la copia)*: en la copia de SC-001, el coste de cada evento del `--scan` es igual al calculado a mano con
  `5m × CacheWrite + 1h × CacheWrite1h`. Referencia, copia del dueño 2026-10-06-wsl: **247 506** tokens a 5 min *(todos de subagente)* y
  **32 518 296** a 1 h.
- **SC-006** *(entre pasadas)*: un test con dos `generate()` y **dos procesos distintos** *(el mensaje abierto se relee tras un reinicio)*:
  - la línea parcial cae en la primera pasada y la final en la segunda → **un** evento con la final;
  - con la cola inspeccionada entre las dos, la primera pasada **no** ha encolado nada de ese mensaje;
  - entre las dos, el offset del fichero en `state.json` es **el comienzo de la primera línea** de ese mensaje *(FR-011)*.
- **SC-007** *(cierre por T — E-2, E-3)*: con el reloj inyectado y el `mtime` del fichero fijado *(`os.Chtimes`)*, un mensaje sin líneas nuevas:
  - `timestamp` y `mtime` a **T** del reloj → **emitido**;
  - `timestamp` a T y `mtime` a **T − 1 s** → **retenido**; `mtime` a T y `timestamp` a **T − 1 s** → **retenido**;
  - *(E-3, tope)* con el `mtime` **de ahora**: `timestamp` a **24 h** → **emitido**; a **24 h − 1 s** → **retenido**.
  Quitar la condición del `timestamp` hace caer su caso, y sólo ése. Quitar la del `mtime` hace caer su caso **y** el de 24 h − 1 s, que
  depende de ella. Quitar el tope hace caer el caso de 24 h, y sólo ése.
- **SC-008** *(cierre por mensaje posterior)*: un mensaje A seguido de la primera línea de B, en el mismo fichero, emite A en esa pasada.
  Si B está en **otro** fichero, A sigue retenido.
- **SC-009** *(frontera intacta)*: los tests de FR-018 **sin tocar** y en verde. `git diff 222c824 -- internal/event/ internal/ingest/eventid.go
  internal/ingest/eventid_test.go specs/006-medicion-fiel/contracts/event-id.md` **vacío**. Tras una pasada con un mensaje abierto sembrado
  con centinelas de `message.id` y `requestId`, **ningún fichero** del directorio de datos los contiene, y `state.json` sigue con sus cuatro
  campos *(E-1: ya no hay fichero de retenidos)*.
- **SC-010** *(espejo)*: el test del espejo da **17** claves y **cinco** cifras exactas. La mutación que quita `CacheWrite1h` de
  `claude-fable-5-1` cae por su subtest.
- **SC-011** *(textos)*: los textos nuevos, **byte a byte** iguales a los aprobados *(P-5)*, comparados con `cmp`, también sobre el binario publicado.
  El CHANGELOG se compara contra el texto **con citas** de §Textos aprobados *(E-3)*.
- **SC-012** *(suites)*: `go test -count=1 ./...` en verde, **432 + los nuevos**; `golangci-lint run` **0**, sin tope.

## Censo de tests — **autorización para la fase de tareas**

| Fichero | Por qué |
|---|---|
| `internal/pricing/pricing_test.go` | `esperadaDelCatalogo` a 17 × 5 (`:45-62`); aserciones del espejo (`:65-110`); `TestCost`, `TestCost_UnknownModel` y `TestCost_Opus55AMano` (`:13-34`, `:119-130`), por la firma de `Cost` |
| `internal/ingest/pasada_test.go` | `TestCasoLimite_ConsumoDistinto` (`:80`) asevera «la primera manda»; `TestPasada_UnMensajeDeTresLineasEsUnEvento` (`:54`) y `TestPasada_ElResumenNoLlevaIdentificadores` (`:124`), por la retención y el resumen |
| `cmd/permea/main_test.go` | `TestPasada_GenerateEncolaUnoPorMensaje` (`:426`) espera el evento en la misma pasada; `TestScan_UnEventoPorMensaje` (`:471`); `TestScan_LineaConCuatroPartidasYEventID` (`:490`), si cambia la línea `evento:`; `TestActualizar_NoReenviaNiReescribeLaCola` (`:526`) |
| **Nuevos** | los de retención y desglose, en ficheros nuevos de `internal/ingest/` y `cmd/permea/`; los del método nuevo de offset retenido *(FR-011)*, en un fichero **nuevo** de `internal/state/`; fixtures nuevos en `testdata/` |

**Medido y FUERA del censo** *(no deben cambiar)*:
- `internal/ingest/boundary_test.go`, entero *(FR-018)*;
- `internal/ingest/eventid_test.go`;
- `internal/event/event_test.go`;
- `internal/ingest/baseline_regresion_test.go` *(no mira coste)*;
- `cmd/permea/project_test.go` *(E-2)*: `TestProjectJoin_LaPeticionNuncaSeEncola/CASO_POSITIVO` lanza `--run` en un subproceso, sin reloj
  inyectable, y espera que la cola crezca. Sigue verde porque su fixture trae **dos** mensajes y la regla (i) cierra el primero. Es la
  co-caída declarada de la mutación que quita la regla (i) *(`plan.md` R-3)*;
- `TestScan_LineaConCuatroPartidasYEventID` *(E-2)*: sigue verde sin tocarlo, porque busca `" cw=7 "` y `cw=` no cambia;
- `internal/state/state_test.go`: con Q-3 (d), `ScanFile` se conserva como envoltorio y sus cuatro tests (`:18`, `:65`, `:94`, `:115`)
  siguen igual y en verde *(FR-011)*. Si el plan no puede conservarlo, entra en el censo y se dice.

Los fixtures de hoy (`internal/ingest/testdata/*.jsonl`) **no** traen `cache_creation` *(0 casos)*, así que sus líneas pasan a ser «sin desglose»
*(P-1)*. ⚠️ Los dos tests de `boundary_test.go` que miran coste sólo aseveran `> 0` y `cost_available` (`:302-355`): no cambian de veredicto.

## Fuera de alcance

| # | Qué | Por qué |
|---|---|---|
| **N-1** | Cambiar el evento, `SchemaVersion` o la allowlist | D-1 |
| **N-2** | Corregir o reenviar lo que envió la 0.3.0 | P-4 |
| **N-3** | El modo rápido | sigue como «Limitación 2». `usage.speed` existe *(A5)*, pero no se ha medido |
| **N-4** | `service_tier`, `inference_geo`, `iterations`, `output_tokens_details`, `server_tool_use`, `fallback_credit` | no se han pedido *(A5)* |
| **N-5** | Explicar **por qué** la salida se escribe creciendo | la regla de P-2 la cubre sin saberlo; queda como pregunta de método. ⚠️ *(E-1, E-2)* **No se atribuye a Windows** en ningún texto de 007: los 143 son de subagentes de Claude Code 2.1.283–2.1.287, y todos los subagentes de WSL son de 2.1.288. Sistema y versión no se pueden separar con estas copias |
| **N-6** | Cambios en la plataforma | D-1. ⚠️ La cabecera de `pricing.php@8f147d1:32-35` dice que el agente «0.3.0» calcula la caché a 5 minutos; con la 0.4.0 eso deja de ser cierto para los eventos nuevos. La enmienda va en un encargo de la plataforma |

## Textos aprobados *(P-5 ✅ — literales; E-2)*

**Resumen de la pasada** *(stderr; E-1: dos líneas. La primera es la de hoy, `pasada.go:71-73`, **sin cambiar un byte**; la segunda es nueva)*:
```
pasada: %d líneas facturables · %d eventos · %d repetidas del mismo mensaje · %d sintéticas · %d sin identificador (no contables) · %d con consumo distinto de la primera
pasada: %d mensajes que crecieron entre líneas · %d en espera de cerrarse · %d líneas releídas de un mensaje en espera · %d líneas tardías · %d líneas sin desglose de caché (a 1 hora)
```
**Línea de `--scan`** *(stdout; añade `cw5m=` y `cw1h=` detrás de `cw=`, que no cambia)*:
```
evento: tool=%s model=%s in=%d out=%d cw=%d cw5m=%d cw1h=%d cr=%d cost=$%.4f cost_avail=%t project_ref=%s event_id=%s
```
**`--run`, cuando quedan mensajes abiertos** *(stderr, una línea más)*:
```
%d mensajes siguen abiertos: se enviarán en la próxima pasada
```
**CHANGELOG `0.4.0`** *(E-2: sustituye al propuesto. La aprobación cubre el cuerpo; el encabezado lleva la fecha del día de la
etiqueta y, hasta entonces, `PENDIENTE`, como en 006. **E-3**: cada punto lleva al final la cita de su especificación, con el formato de
la 0.3.0; las frases aprobadas no cambian. Éste es el texto contra el que compara SC-011)*:
```
## 0.4.0 — PENDIENTE

### Cambia
- El coste de la escritura de caché usa su duración: la de 5 minutos a su tarifa y la de 1 hora a la suya (el doble de la entrada).
  Antes toda iba a la de 5 minutos, y el coste salía por debajo.
  (`specs/007-coste-fiel/spec.md`, FR-001, FR-002, FR-007)
- Un mensaje que Claude Code escribe en varias líneas se cuenta entero: cada partida vale lo más alto que alcanza. Antes contaba
  la primera línea, y en las conversaciones de subagentes podía faltar parte de la salida.
  (`specs/007-coste-fiel/spec.md`, FR-009)
- El último mensaje de cada conversación se envía cuando empieza el siguiente o tras 10 minutos sin cambios. Si la pasada
  termina antes, sale en la siguiente.
  (`specs/007-coste-fiel/spec.md`, FR-010, FR-011, FR-014)
- Tarifas de 17 modelos, con la escritura de caché a 1 hora; nueva: claude-fable-5-1.
  (`specs/007-coste-fiel/spec.md`, FR-006, FR-008; `specs/007-coste-fiel/contracts/tarifas.md`)

### Lo ya enviado
- No se corrige ni se reenvía. Un mensaje que la 0.3.0 envió incompleto se queda como llegó.
  (`specs/007-coste-fiel/spec.md`, FR-019)

### Limitaciones conocidas
- El «modo rápido» no se distingue: un mensaje en modo rápido queda por debajo de su coste.
  (`specs/007-coste-fiel/spec.md`, FR-007, N-3)
- Una línea sin el desglose de la caché se tarifa entera a 1 hora.
  (`specs/007-coste-fiel/spec.md`, FR-003, FR-004)
```

## Preguntas abiertas — **ninguna** *(E-2)*

Las seis, resueltas; la tabla queda como rastro. Q-1, Q-2 y Q-6: el dueño, 2026-10-06 20:00. Q-3 y Q-4: decisión de método del
orquestador, 2026-10-06. Q-5: resuelta por la medida *(E-1)*.

| # | Pregunta | Opciones | Recomendación *(rastro)* |
|---|---|---|---|
| **Q-1** ✅ **(a)** | ¿Cuenta `stop_reason` **no nulo** como cierre? | (a) **no** *(P-3 ajustada)* · (b) sí, después de medir una copia de Windows y comprobar que las líneas parciales lo traen nulo · (c) sí, ya | **(a)**, con la medida de Windows *(E-1, W-A3)*. La condición de (b) **se cumple** en los subagentes de Windows: las 299 líneas parciales lo traen nulo y llega **en la misma línea** que el máximo en los 143. Pero: en los principales viene **desde la primera línea** *(4 241 de 4 241 en Windows, 7 286 de 7 286 en WSL)*, así que allí equivale a «la primera»; 61 mensajes de subagente de Windows y los 76 de WSL **no lo traen nunca**; y su conducta cambia con la versión de Claude Code. Lo único que ganaría es cerrar antes el último mensaje de un subagente; (i) y (ii) ya lo cubren |
| **Q-2** ✅ **(a)** | ¿Qué hace un `--run` suelto con los mensajes aún abiertos al terminar? | (a) **quedan retenidos** para la siguiente pasada *(FR-014)* · (b) se emiten al terminar, como la 0.3.0 · (c) `--run` espera hasta T antes de salir | **(a)**. (b) reintroduce el defecto cuando la línea final llega después. (c) bloquea hasta 10 min. Coste de (a): quien ejecute `--run` una sola vez y nunca más deja sus últimos mensajes sin enviar. Se dice en el resumen *(P-5)* |
| **Q-3** ✅ **(d)** | ¿Dónde viven los retenidos? | (a) un fichero propio, `pendientes.json`, junto a `state.json`, con la misma escritura atómica · (b) dentro de `state.json` · (c) en la cola, con una marca de «retenido» · **(d)** *(E-1)* en ningún sitio: el offset no pasa del comienzo del mensaje abierto y la pasada siguiente lo relee del log | **(d)** *(E-1; antes, (a))*. No añade estado en disco, y (a) sí: un fichero cuya pérdida o corrupción, con `state.json` intacto, pierde mensajes sin aviso. Conserva los dos pasos de durabilidad de hoy *(FR-012)*. A lo sumo hay **un** mensaje abierto por fichero *(regla (i); medido: 1 en todos)*. Cuesta: releer la cola *(p50 11–14 KB, máx. 1,16 MB, hasta T)*, contar las releídas *(FR-021)* y un modo de pérdida propio: el fichero truncado o borrado con un mensaje abierto. Comparación completa en `soporte/descubrimiento.md` §C |
| **Q-4** ✅ **(a)** | ¿Qué se hace con una línea cuyo desglose **no suma** el total? *(0 casos hoy)* | (a) **sin desglose**, todo a 1 hora *(FR-004)* · (b) usar el desglose tal cual · (c) repartir el total en proporción | **(a)**. El total es lo que cruza la frontera *(D-1)*, y P-1 ya dice qué hacer cuando el desglose no sirve |
| **Q-5** ✅ | ¿Hay una copia congelada de los logs de **W2** *(Windows)* para SC-002? | (a) el dueño la aporta y se mide como en A *(sólo recuentos)* · (b) sólo el fixture sintético | **Resuelta** *(E-1)*: la copia del dueño 2026-10-06-windows contiene los 143 de W2, con las mismas cifras. SC-002 la usa, además del fixture |
| **Q-6** ✅ **(a)** | ¿T = 10 minutos? | (a) 10 min · (b) 5 min · (c) 30 min · (d) configurable | **(a)**, fijo: más del doble del máximo medido *(259,9 s en WSL; 144,6 s en Windows, 90,9 s en los que crecen)*. **0** huecos ≥ 5 min en las dos copias, pero 5 min sólo deja un 15 % de margen sobre WSL. Configurable añade superficie sin una medida que lo pida |

## Assumptions
1. **El orden de líneas de un fichero es el orden de escritura**: 0 entrelazados en 31 ficheros de WSL y en 29 de Windows; 0 mensajes con
   una línea después de que empiece otro, y 0 repartidos entre ficheros, en las dos copias *(E-1, W-A4)*.
2. **El desglose viene en todas las líneas de la versión actual de Claude Code** *(22 180 de 22 180 en WSL; 13 867 de 13 867 en Windows)*.
   P-1 cubre las que no lo traigan.
3. **El historial cubre unos 28–30 días** *(A4: desde el 09-08 en WSL y el 09-06 en Windows)*. `cleanupPeriodDays` no está declarado. La primera pasada de un usuario nuevo envía ese historial,
   como en la 0.3.0.

## Dependencias
- **Plataforma `8f147d1`**: el catálogo que se replica *(D-3)*. Ver N-6 para la enmienda de su cabecera.
- **006**: FR-005 y FR-033 se **sustituyen** por FR-009 a FR-013 y FR-021. FR-009 *(no reenviar)* se conserva. El contrato `tarifas.md` se sustituye *(FR-008)*.

## Registro de enmiendas

| # | Fecha | Qué cambia | Por qué |
|---|---|---|---|
| **E-1** | 2026-10-06 | **Referencias**: SC-001, SC-003 y SC-005 pasan a la copia del dueño 2026-10-06-wsl *(10 121 mensajes)*; SC-002, a la -windows *(6 074, con los 143 de W2)*. Q-5, resuelta. **Q-3**: nueva opción (d) y recomendación (d) en lugar de (a); se reescriben P-3, FR-011, FR-012, FR-013, FR-014, FR-015, SC-006, SC-009, el caso del truncado, la entidad y el censo de `internal/state`; nuevo FR-021 *(relectura)*. **P-2, Q-1, Q-6**: misma recomendación, con la evidencia de Windows. **Q-1**: cae la razón «en Windows cerraría en la línea parcial». **Assumption 1**: medida en Windows. **P-5**: el resumen pasa a dos líneas *(la de hoy intacta)* y el CHANGELOG dice que el último mensaje espera 10 minutos. **N-5**: sistema y versión no se separan | La copia temporal del descubrimiento se borró: sus cifras no se podían repetir. Las copias del dueño sí. La medida de Windows *(`soporte/descubrimiento.md` §Windows y §C)* |
| **E-2** | 2026-10-06 | **Ratificadas por el dueño** *(20:00, Madrid)*: P-1 (a), P-2 (a), P-3 (a) sin (d) con Q-1 (a) y Q-6 (a), P-4 (a), Q-2 (a) y P-5, con el **CHANGELOG sustituido** por el aprobado *(encabezado `PENDIENTE` hasta la etiqueta)*. **Decisiones de método del orquestador**: Q-3 (d), Q-4 (a), FR-021 *(el demonio calla si sólo relee)* y **FR-010 (ii) más conservadora**: ≥ T desde el `timestamp` **y** desde el `mtime` del fichero; SC-007 pasa a tres casos. El crecimiento deja de atribuirse a Windows *(Contexto, N-5)*. §Preguntas abiertas, a cero. ✋ → ✅ en P, FR y Q, con las alternativas como rastro. Censo: `project_test.go` y `TestScan_LineaConCuatroPartidasYEventID`, medidos y fuera | Ratificación del dueño y decisiones del orquestador del 06-10. Contraste de la regla (ii) con el código: `plan.md` §Decisiones |
| **E-3** | 2026-10-06 | **Enmienda del orquestador.** **(1) Tope de la espera**: FR-010 gana la regla **(iii)**, que cierra un mensaje a las ≥ 24 h del `timestamp` de su última línea, cambie o no su fichero. SC-007 gana su caso *(con el `mtime` de ahora: a 24 h, emitido; a 24 h − 1 s, retenido)*. Se tocan P-3 y el caso límite del último mensaje; el CHANGELOG no cambia por esto. **(2) Citas en el CHANGELOG**: cada punto del cuerpo aprobado lleva al final la cita de su especificación, con el formato de la 0.3.0, sin cambiar las frases. SC-011 compara contra el texto con citas. **(3)** `tasks.md` T003 pasa a «007 B0: enmienda E-3 y contrato de tarifas de 006 sustituido» | Cierra el riesgo R-2 del plan *(un fichero que cambia sin líneas nuevas retendría su último mensaje sin límite)*. La cabecera de `CHANGELOG.md` exige citar la especificación en cada punto |
