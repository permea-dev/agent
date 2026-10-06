# Implementation Plan: 007 · «Coste fiel»

**Branch**: `007-coste-fiel` | **Date**: 2026-10-06 | **Spec**: [spec.md](./spec.md) *(E-1, E-2, E-3)* | **Base**: `222c824`
**Input**: la spec ratificada · [`soporte/descubrimiento.md`](./soporte/descubrimiento.md) · [`contracts/tarifas.md`](./contracts/tarifas.md) · [`quickstart.md`](./quickstart.md)

## Summary

La `0.4.0` tarifa la escritura de caché por su duración *(D-2 (a))* y cuenta cada mensaje entero, con el máximo por partida *(D-2 (b))*.
Un mensaje se emite cuando está cerrado, y lo abierto se relee del log *(Q-3 (d))*. **El evento no cambia** *(D-1)*. El riesgo está en
tres sitios:
1. **La emisión deja de ser por línea.** Hoy `FromClaudeCodeLine` devuelve el evento de la primera línea. Los tests de frontera la
   llaman **sin pasada** y no se pueden tocar *(FR-018)*. Por eso, con la pasada a `nil`, la conducta es la de hoy, y la acumulación
   vive en la `Pasada` *(D-007-P2)*.
2. **El offset deja de avanzar hasta el final.** Es lo único que cambia en `internal/state`, y `ScanFile` se queda como envoltorio *(D-007-P3)*.
3. **Los tests de proceso no pueden inyectar el reloj.** Un `--run` en un subproceso retiene el último mensaje de cada fichero, porque
   el `mtime` es de ahora. Medido: un único test fuera del censo depende de eso, y sigue verde gracias a la regla (i) *(Riesgos, R-3)*.

## Technical Context

- **Go** 1.22.2 · `golangci-lint` 2.12.2 · `goreleaser` v2.16.0. **Ninguna dependencia ni paquete nuevo** *(9 paquetes)*.
- **Línea base**, medida el 2026-10-06 sobre `222c824`: `go test -count=1 ./...` → **9 paquetes ok, 432 pass, 0 fail**;
  `golangci-lint run` → **0 issues**, sin tope.
- **Storage**: **ninguno nuevo**. `state.json` conserva sus cuatro campos *(FR-011)*, y la cola no cambia.
- **Restricciones**:
  - `internal/event`, `eventid.go`, `eventid_test.go`, `boundary_test.go` y `contracts/event-id.md` de 006, **sin diff** *(SC-009, FR-018)*;
  - lint a 0;
  - el espejo de tarifas no lee el otro repositorio en tiempo de test.

## Constitution Check *(antes de Phase 0 y tras Phase 1 del plan: sin cambios)*

| Principio | Veredicto | Razón |
|---|:--:|---|
| I · Frontera inviolable | ✅ | `rawRecord` gana **dos números** de consumo (FR-001). El desglose no cruza la frontera (FR-005). El mensaje abierto vive en memoria y nunca en disco (FR-011). El golden y la denylist, intactos |
| II · Local-first | ✅ | El coste sigue en local, con la tabla empaquetada (`contracts/tarifas.md`) |
| III · Binario único | ✅ | Ni dependencias ni paquetes nuevos |
| IV · Test-first en la frontera | ✅ | La frontera **no se toca**. SC-009 siembra centinelas en el camino nuevo *(B4)* |
| V · Specs | ✅ | Spec ratificada (E-2), contrato antes del código |
| Puertas | ✅ | vet, lint a 0 y suite verde al cerrar cada bloque |

## Project Structure

```text
internal/pricing/pricing.go · pricing_test.go     # B1 (tabla 17×5, cabecera) · B2 (Cost por duración, hipótesis P-1)
internal/ingest/claudecode.go                      # B2 (desglose en rawRecord) · B3 (acumular con pasada; sin pasada, como hoy)
internal/ingest/pasada.go · pasada_test.go         # B2 (consumo + desglose, SinDesglose) · B3 (máximo, Crecieron) · B4 (cierre, resumen)
internal/ingest/desglose_test.go                   # NUEVO, B2
internal/ingest/maximo_test.go                     # NUEVO, B3
internal/ingest/cierre_test.go                     # NUEVO, B4: reglas (i) y (ii), tardías y releídas
internal/state/state.go                            # B4: método nuevo con el comienzo de cada línea y el offset retenido
internal/state/retener_test.go                     # NUEVO, B4 (state_test.go NO se toca)
cmd/permea/main.go · main_test.go                  # B2 (línea evento:) · B3 (emitir al final del fichero) · B4 (reloj, offset, --run, tick)
cmd/permea/retencion_test.go                       # NUEVO, B4: SC-006 a SC-009 y FR-021
README.md · CHANGELOG.md                           # B5
specs/006-medicion-fiel/contracts/tarifas.md       # B0: una línea «sustituido por 007»
```

## Decisiones de plan

| ID | Decisión | Por qué |
|---|---|---|
| **D-007-P1** | `Cost(model, in, out, cw5m, cw1h, cr)`. **Quien llama** decide el desglose: sin desglose o incoherente → `cw5m = 0`, `cw1h = total` | `pricing` sólo tarifa. P-1 y Q-4 son reglas de lectura del log |
| **D-007-P2** | Con `Pasada` a `nil`, `FromClaudeCodeLine` emite por línea, como hoy. Con pasada, **acumula** y devuelve `nil`; los cerrados salen por un método de la `Pasada`. El desglose viaja con el cerrado, **no** en el evento | Es la única forma de dejar `boundary_test.go` sin tocar (`eventosDelFixture` va sin pasada) y respetar D-1 |
| **D-007-P3** | En `internal/state`, un método nuevo da al callback el **comienzo** de cada línea y fija el offset en un valor que decide quien llama. `ScanFile` lo envuelve con la conducta de hoy | Q-3 (d). Los cuatro tests de `state_test.go` quedan fuera del censo |
| **D-007-P4** | El cierre por T usa el `info.ModTime()` del `os.Stat` de **la pasada**, a resolución completa. No usa `FileState.ModTime`, que se guarda en segundos y no lo lee nadie | E-2. Ver la tabla siguiente |
| **D-007-P5** | El evento toma el `timestamp`, el modelo y las referencias de la **primera** línea, y los tokens del máximo. `occurred_at` coincide con el que emitía la 0.3.0 *(`DECIDÍ YO`)* | La spec no lo fija; así lo ya enviado y lo nuevo se ordenan igual |
| **D-007-P6** | El reloj es un campo `func() time.Time` del `agent`, `time.Now` por defecto. Los tests lo fijan, y fijan el `mtime` con `os.Chtimes` | FR-010. Hoy no hay ningún `time.Now` en producción |

**FR-010 (ii) contra el código *(E-2)*.**
- **Es viable.** `ScanFile` ya hace `os.Stat` (`state.go:80`) y guarda `ModTime` sin que nadie lo lea (`:118`), así que basta con
  exponer el del stat.
- **Efecto 1**: el `mtime` cambia con **cualquier** línea, incluidas las de usuario o herramienta. El último mensaje espera a que **toda la
  conversación** lleve T quieta, y el CHANGELOG aprobado lo dice así («tras 10 minutos sin cambios»).
- **Efecto 2**: un fichero que se siga modificando sin líneas nuevas retendría su último mensaje **indefinidamente**. No está medido *(R-2)*.
  **E-3 lo cierra** con la regla **(iii)**: a las ≥ 24 h del `timestamp` de su última línea, el mensaje se cierra cambie o no el fichero.
- **Efecto 3, Windows**: `os.Stat` lee el `LastWriteTime` de NTFS, y si se actualiza tarde **sin comprobar**. Si se queda atrás, la regla
  vuelve a ser la del `timestamp`, nunca más agresiva *(se comprueba en W1)*.
- **Efecto 4**: en los tests de proceso el `mtime` es el de ahora, y el último mensaje se retiene *(R-3)*.
- **Ninguno lo hace inviable**, así que **no hay PREMISA CAÍDA**.

## Disciplinas

**Las nueve de 006** *(`specs/006-medicion-fiel/plan.md` §Disciplinas)*, sin cambios. **Una nueva:**

10. **Las copias del dueño se leen en su sitio.** Nada se copia, se escribe ni se borra en ellas, y se toma su huella antes y después. De
    ellas salen sólo recuentos y sumas. Los ficheros temporales con `event_id` viven en un `mktemp -d` y se borran al terminar el tramo.

## Bloques

**Fase 0** de cada bloque: puertas en verde sobre el commit anterior, el censo del bloque declarado en `tasks.md` y el esqueleto que
haga compilar los rojos, para que fallen como test y no como compilación. **Corte**: ✋ commit del bloque, con las transcripciones.
Tamaños: **S** < 50 líneas de producción · **M** 50–150 · **L** > 150.

| Bloque | FR · SC | Producto | Censo | Rojos | Mutaciones *(el test que debe caer, y sólo ése)* | Tamaño |
|---|---|---|---|---|---|---|
| **B0 · Documentos** | — | sólo `specs/` | — | — | — | docs |
| **B1 · Tarifas** — F0: `Rate.CacheWrite1h` vacío | FR-006, FR-007 *(commit, verificación, aprobación)*, FR-008 · SC-010 | `pricing.go` | `pricing_test.go`: `esperadaDelCatalogo` a 17×5; `CifrasClaveAClave` compara la quinta | **(1)** recuento 16 ≠ 17 · **(2)** `CifrasClaveAClave`: 16 subtests por `cache_write_1h` = 0 y `claude-fable-5-1` ausente | **(m1)** `CacheWrite1h` de `claude-fable-5-1` a 0 → subtest `claude-fable-5-1` · **(m2)** clave sobrante → `NingunaClaveSobra` + recuento *(declaradas)* · **(m3)** cruzar 5 m y 1 h en `claude-opus-5-5` → su subtest y `TestCost_Opus55AMano` *(co-caída declarada en T007)* | **S** |
| **B2 · Desglose y coste** — F0: `Cost` con la firma nueva y la conducta vieja; suite verde | FR-001 a FR-005, FR-007 *(hipótesis P-1)*, FR-016 *(línea)* · SC-004 | `pricing.go`, `claudecode.go`, `pasada.go`, `main.go` (`dryRun`) | `pricing_test.go`: `TestCost`, `TestCost_UnknownModel`, `TestCost_Opus55AMano` → los vectores de SC-004 | **(3)** SC-004 con desglose: 1,1775756 · **(4)** `FromClaudeCodeLine` con desglose: coste y `tokens_cache_creation` = 45 679 · **(5)** sin desglose: 1,2146106 · **(6)** desglose que no suma → como (5), total del log · **(7)** `SinDesglose` = 2 · **(8)** `--scan` imprime `cw5m=12345 cw1h=33334` | **(m4)** cruzar las dos tarifas en `Cost` → (3), (4), (5), (6) y `TestCost` *(declarado en T015)* · **(m5)** sin desglose a 5 min → (5) y (6) · **(m6)** aceptar el desglose incoherente → (6) y (7) · **(m7)** `tokens_cache_creation` = 5 m + 1 h → (6) · **(m8)** `cw5m` con el total → (8) | **M** |
| **B3 · Máximo en la pasada** — F0: el acumulador vacío y nil-seguro | FR-009, FR-013 *(pasada)*, FR-016 *(emitir al final del fichero)* · SC-002 *(sintético)*, SC-003 | `pasada.go`, `claudecode.go`, `main.go` (`generate` y `dryRun` emiten al final de cada fichero) | `pasada_test.go`: `leerEnUnaPasada` cierra al final; `TestCasoLimite_ConsumoDistinto` pasa al máximo | **(9)** 7 → 1 303 → 89 817 → **89 817** · **(10)** 7 → 89 817 → 1 303 → 89 817 · **(11)** desglose de la línea del máximo de la escritura · **(12)** `Crecieron` = 1 · **(13)** `--scan` con (9) → `out=89817` | **(m9)** máximo → primera → (9), (10), (13) · **(m10)** máximo → última → (10) · **(m11)** sumar → (9), (10), (14) · **(m12)** desglose de la última → (11). **(14)** una línea duplicada no cambia nada *(nace verde; la valida m11)* | **M** |
| **B4 · Retención** — F0: el método nuevo de `state` con el offset de hoy, y el reloj en el `agent` | FR-010 a FR-015, FR-017, FR-019, FR-021 · SC-006 a SC-009 | `state.go`, `pasada.go`, `claudecode.go`, `main.go` (`generate`, `runOnce`, `tick`) | `main_test.go`: `GenerateEncolaUnoPorMensaje` y `Actualizar…` fijan el reloj a T + 1 s; `pasada_test.go`: `ElResumen…` mira las dos líneas | **(15)** SC-006 en proceso: el primer `--run` no encola nada de ese mensaje; el offset queda en su comienzo; el segundo *(final + otro mensaje)* da 1 evento con la final · **(16)** SC-007 b y c retenidos · **(17)** SC-008, mismo fichero y otro · **(18)** resumen en dos líneas, la primera byte a byte · **(19)** aviso de `--run` · **(20)** FR-021: el predicado de `tick` calla si todo es releído · **(21)** línea tardía contada · **(23)** *(E-3)* SC-007 tope: con el `mtime` de ahora, `timestamp` a 24 h − 1 s → retenido *(rojo)*; a 24 h → emitido *(nace verde)* | **(m13)** avanzar el offset al final → (15) · **(m14)** quitar la condición del `mtime` → (16 b) **y** (23, 24 h − 1 s), que depende de ella *(declarada)* · **(m15)** quitar la del `timestamp` → (16 c) · **(m16)** regla (i) entre ficheros → (17, otro fichero) · **(m17)** emitir lo abierto al acabar `--run` → (15) y (19) · **(m18)** contar releídas como nuevas → (20) · **(m19)** guardar el abierto en `pendientes.json` con sus ids → **(22)** SC-009 *(nace verde)* · **(m20)** quitar la regla (i) → (17, mismo fichero) **y**, declarada, `TestProjectJoin_LaPeticionNuncaSeEncola/CASO_POSITIVO` *(R-3)* · **(m21)** SC-007 a: no emitir nunca por T → (16 a) *(nace verde)* · **(m22)** *(E-3)* quitar el tope de 24 h → (23, 24 h), y sólo ése | **L** |
| **B5 · README y CHANGELOG** | FR-020 · SC-011 *(CHANGELOG)* | `README.md`, `CHANGELOG.md` | — | `grep` de la «Limitación conocida» de W2 y de la «Limitación 1» > 0 · sin `## 0.4.0` | — *(documentación; `cmp` del cuerpo con la spec)* | **S** |

Estimación: B1 ~40 + 50 de test · B2 ~70 + 160 · B3 ~110 + 220 · B4 ~200 + 420 · B5 ~60 líneas de texto. **~420 de producción y ~850 de test.**

**El orden propuesto se mantiene, con dos ajustes que pide el código:**
- La cabecera de tarifas se parte entre **B1** *(commit, verificación, aprobación)* y **B2** *(la hipótesis P-1 sustituye a la «Limitación 1»
  cuando `Cost` ya tarifa por duración)*. Si no, la cabecera mentiría en el commit de B1.
- **B3 emite al final de cada fichero** en `generate()` y en `dryRun()`. Es un corte coherente: el máximo dentro de la pasada sin
  retención. Es seguro porque hay **0** mensajes repartidos entre ficheros en las dos copias. `GenerateEncolaUnoPorMensaje` sigue verde
  en B3 y entra en el censo en B4.

## El contador independiente *(SC-001, SC-002, SC-003 y SC-005)*

- **Dónde vive**: en `quickstart.md` §Contador. Es Python con `-I` y no comparte código con el agente, como el de 006
  *(`specs/006-medicion-fiel/quickstart.md` §El contador independiente)*.
- **Qué corrige del de 006**: aquel aplicaba `setdefault`, «la primera manda», y **no era independiente** en ese punto *(Hallazgo de W2)*.
  Éste aplica el **máximo por partida**. Saca además:
  - la suma con «primera», para ver la diferencia;
  - los mensajes que crecen;
  - el reparto 5 m / 1 h con el desglose de la línea del máximo;
  - las líneas sin desglose o que no suman.
- **Cómo se lanza**: `python3 -I - <copia>`, desde un directorio que no es la copia, leyendo la copia **en su sitio**. Sólo imprime
  recuentos y sumas.
- **Las medidas** *(quickstart §M)*:
  - el binario de la rama, con `--scan`, **por fichero**, dentro de `env -i` con `HOME` y `XDG_CONFIG_HOME` temporales y sin enrolar;
  - la salida, a un `mktemp -d`;
  - por `awk`: el número de `evento:`, las sumas de `in`, `out`, `cw`, `cw5m`, `cw1h` y `cr`, y los `event_id` repetidos *(0)*;
  - SC-005 recalcula el coste de cada evento con la tabla del contrato. El `--scan` imprime 4 decimales, así que la comparación tolera
    5 × 10⁻⁵ por evento.

## Actualizar desde la 0.3.0 con (d) *(FR-019, P-4)*

- `state.json` **no cambia de forma**, así que no hay migración. La 0.3.0 dejó el offset tras la última línea completa de cada fichero.
- La primera pasada de la 0.4.0 lee **desde ahí**, sin mensajes abiertos *(FR-019)*. Lo que la 0.3.0 ya consumió no se relee.
- Si llegan líneas de un mensaje que la 0.3.0 ya envió con la primera, esta pasada las ve como un mensaje nuevo. Lo emite al cerrarse, con
  el **mismo** `event_id`, y la plataforma lo descarta. Es lo que declara P-4: «se queda como llegó».
- La cola se envía tal cual.
- **Volver a la 0.3.0 también es seguro**: leería desde el offset retenido y reenviaría la primera línea del mensaje abierto, con el
  mismo `event_id`.
- Lo acredita `TestActualizar_NoReenviaNiReescribeLaCola`, con el reloj fijado *(B4)*.

## Cierre — en tramos, uno por mensaje *(si un tramo falla, se para y se rehace desde C1)*

| Tramo | Qué | Quién |
|---|---|---|
| **C1** · Puertas | `gofmt -l`, `go vet`, `golangci-lint run` → 0, `go test -count=1 ./...` → 432 + nuevos *(SC-012)*. `git diff 222c824` de la frontera, vacío *(SC-009, FR-018)*. La cabecera cita `8f147d1`. Compilan Windows y darwin. `PENDIENTE` = 1 | Claude |
| **C2** · Medidas | Sobre las dos copias del dueño *(quickstart §M)*: SC-001, SC-003 y SC-005 en -wsl, y SC-002 en -windows. La huella de las copias, antes = después | Claude |
| **C3** · Snapshot | `goreleaser release --snapshot --clean`, la huella del zip de Windows, y `strings` del binario con los textos aprobados *(SC-011)* | Claude |
| **C4** · **W1** | En Windows, el binario del snapshot: `--version` *(ni `0.0.1-dev` ni `0.3.0`)*. **Sólo después**: `status`, `--scan` de un log contra el contador, y `--run` dos veces separadas por > 10 min sin usar Claude Code. La primera avisa «en espera»; la segunda los envía. Así se comprueba FR-010 (ii) en NTFS | ✋ dueño |
| **C5** · PR | El cuerpo del PR y `## 0.4.0 — <fecha prevista de la etiqueta>` *(`PENDIENTE` → 0)* | Claude redacta |
| **C6** · Fusión | Con merge commit, como #1–#3 | ✋ dueño |
| **C7** · Etiqueta | `git tag -a v0.4.0` sobre `main` | ✋ dueño |
| **C8** · Canales | `gh release view v0.4.0`; Scoop y el cask en `0.4.0`; `strings` del binario publicado *(SC-011)* | Claude |
| **C9** · **W2** | `scoop update` → `--version` `0.4.0` → `status` → `--run`. En la plataforma, los eventos `0.4.0` de la ventana = los mensajes distintos del contador **menos** «en espera» de la segunda línea del resumen. Tras T, otro `--run`, y deben cuadrar con todos | ✋ dueño |

## Trazabilidad

| FR | Bloque | | SC | Bloque |
|---|---|---|---|---|
| 001–005 | B2 | | 001, 003 | B3 (tests) · C2 (copia -wsl) |
| 006, 008 | B1 | | 002 | B3 (sintético) · C2 (copia -windows) |
| 007 | B1 + B2 | | 004 | B2 |
| 009 | B3 | | 005 | B2 (instrumento) · C2 |
| 010–012, 014, 015, 017, 021 | B4 | | 006–009 | B4 *(009 también C1)* |
| 013 | B3 (pasada) · B4 (tardías) | | 010 | B1 |
| 016 | B2 (línea) · B3 (emitir al final) | | 011 | B4 (resumen) · B5 (CHANGELOG) · C3/C8 (binario) |
| 018 | todos (puerta) · C1 | | 012 | cada bloque · C1 |
| 019 | B4 | | | |
| 020 | B5 · C5–C8 | | | |

**21 / 21 requisitos y 12 / 12 criterios con bloque.**

## Riesgos medidos

| # | Riesgo | Medida o mitigación |
|---|---|---|
| R-1 | El `mtime` de Windows se queda atrás | Dirección segura: la regla vuelve a la del `timestamp` *(D-007-P4)*. Se comprueba en W1 |
| R-2 | Un fichero que se modifica sin líneas nuevas retiene su último mensaje sin límite | **Cerrado por E-3**: la regla (iii) lo cierra a las 24 h, con su rojo (23) y su mutación (m22). W1 y W2 siguen leyendo «en espera» |
| R-3 | Los tests de proceso no inyectan el reloj | Medido: `TestProjectJoin_LaPeticionNuncaSeEncola/CASO_POSITIVO` sigue verde porque su fixture tiene **2** mensajes y la regla (i) cierra el primero. Lo declara (m20). Los demás `--run` de test sólo miran el código de salida |
| R-4 | Releer la cola | p50 11–14 KB y máx. 1,16 MB por pasada, durante T como mucho *(soporte §C)* |
| R-5 | Truncado o borrado con un mensaje abierto *(modo de pérdida de (d))* | Sin casos medibles. Declarado en la spec |
| R-6 | El crecimiento sólo se ha visto en 2.1.283–2.1.287 *(N-5)* | SC-002 se apoya en la copia -windows y en el fixture, no en logs nuevos |
| R-7 | `--scan` usa un `Scanner` de 1 MiB | 0 líneas mayores en las dos copias. Fuera de alcance |

**Continuación, tras publicar**: la cabecera de `pricing.php@8f147d1:32-35` dice que el agente «0.3.0» tarifa la caché a 5 minutos.
Necesita su enmienda en un **encargo aparte de la plataforma** *(N-6)*.

## Complexity Tracking

*Vacío: sin violaciones.*

## Enmiendas

- **E-3** *(2026-10-06, orquestador; registro en `spec.md`)*: regla (iii) de FR-010, el tope de 24 h, que cierra R-2. En B4 entran el rojo (23) y la mutación (m22), y (m14) declara una co-caída más. Las citas del CHANGELOG cambian B5 *(el `cmp` de T036 compara contra el texto con citas)*.
