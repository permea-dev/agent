# Tasks: Medición fiel y publicable

**Feature**: `006-medicion-fiel` · **Fecha**: 2026-10-02 · **Spec**: [spec.md](./spec.md) ·
**Plan**: [plan.md](./plan.md) · **Research**: [research.md](./research.md) · **Contratos**:
[event-id.md](./contracts/event-id.md) · [cli.md](./contracts/cli.md) · [tarifas.md](./contracts/tarifas.md) ·
**Validación**: [quickstart.md](./quickstart.md)

**Base**: `cbee0cd` (spec `03d3734` + plan, con las enmiendas E-006-P1 a P6 del 2026-10-02).
**Línea base a preservar**: `go test ./...` → **9 paquetes `ok`, 297 tests pass, 0 `FAIL`** (medido
2026-10-02 sobre `0311fa1`; el código no ha cambiado desde entonces).

---

## Format: `[ID] [P?] [✋?] Description`

- **[P]**: sin dependencia entre sí, **escribibles en cualquier orden**. Igual que en 005, **NO**
  significa «ficheros distintos»: dos `[P]` que tocan el mismo `_test.go` se escriben una detrás de
  otra. El paralelismo es **de decisión**, no de edición concurrente.
- **✋**: la ejecuta **el dueño** (Basilio), no Claude: commits, decisiones, la PR, la etiqueta y los
  ensayos en Windows. Claude no hace git de escritura ni abre la PR.
- **Numeración de creación**: T001… en el orden en que se escribieron. **No se renumera**: una tarea
  añadida después recibe el siguiente número libre, aunque se ejecute antes.
- **(1)…(26)** son los rojos y **(m1)…(m19)** las mutaciones de la tabla de bloques de `plan.md`, con
  la misma numeración. Las mutaciones propias de este fichero que el plan no numeraba se llaman
  **M-B0** y **M-B1**.
- Ruta de fichero exacta en cada tarea.

### Convenio: «Bloque» aquí, «Phase N del plan» allí

Este fichero se ordena por **bloques** (B0 … B5, Cierre), los de `plan.md` §Bloques de implementación.
El plan usa además «Phase 0/1/2 del plan» (research · diseño · tasks). **Toda cita al plan se escribe
«Phase N del plan»**, cualificada.

## Path Conventions

Proyecto único Go. Rutas desde la raíz del repositorio: `cmd/permea/`, `internal/`,
`specs/006-medicion-fiel/`.

---

## Disciplinas transversales — aplican a TODA tarea de test de este fichero

**Las ocho de 005, sin cambios** (`specs/005-adhesion-a-proyecto/tasks.md`, §Disciplinas
transversales). No se repiten; se dan por incluidas:

1. Una garantía por tarea, anclada al **contrato**. Si contrato y spec discrepan, se para.
2. Rojo antes de verde, **transcribiendo la razón real** del fallo.
3. Todo test que nazca verde se valida por mutación:
   - la mutación **compila y mata sólo su hecho**, y se revierte **por edición inversa**, nunca con
     `git checkout`;
   - **`t.Errorf` para aserciones independientes**; `t.Fatalf` sólo para precondiciones.
4. Tests de proceso: se compara `ExitCode()`, nunca texto.
5. La ausencia se comprueba por **canal vacío**.
6. Aislamiento obligatorio (`HOME` / `USERPROFILE` / `XDG_CONFIG_HOME` temporales).
7. Los dos canales se capturan **por separado**.
8. En comentarios de código se cita **por nombre**, nunca por línea. Y **en este fichero tampoco** se
   cita por línea el código que 006 edita: aquí también caduca antes que la fecha.

Una más, **nueva en 006** (`plan.md` §Disciplinas):

9. **Sobre la copia congelada de logs reales, sólo recuentos.** Nada imprime un `message.id`, un
   `requestId` ni un `event_id` de un log real. Los vectores de prueba usan identificadores
   **sintéticos**. La copia vive fuera del repositorio y se borra al terminar su tramo.

### Protocolo de mutación de 006 — el censo va ANTES

Toda mutación de este fichero sigue estos pasos, en este orden:

1. **Declarar el censo en la propia tarea, antes de tocar nada.** Por nombre, qué tests deben caer.
2. **Declarar las co-caídas**: qué otros tests caerán por arrastre, y por qué.
3. Aplicar la mutación, ejecutar `go test ./... 2>&1`, y **comparar el conjunto de `FAIL` con lo
   declarado**.
4. **Si coincide exactamente**: revertir por edición inversa, comprobar el verde completo y transcribir
   en la tarea el mensaje de fallo de cada test del censo.
5. **Si cae algo no declarado, o no cae algo declarado, la mutación SE DEJA PUESTA y se PARA.** Se
   reporta al orquestador con la salida. Lo primero es que un hecho que creíamos separado no lo está.
   Lo segundo es que el test no mira lo que dice mirar. Ninguna de las dos cosas la resuelve quien
   ejecuta.
6. **Mutación inválida** (no compila, o panica): se descarta y se busca otra forma de alterar el mismo
   hecho (disciplina 3). **No cuenta como superada.**

### Puertas del bloque — se ejecutan al final de cada bloque, ANTES de su ✋ commit

```sh
gofmt -l .                                   # → vacío
go vet ./...                                 # → sin hallazgos
golangci-lint run                            # → 0 issues (2.12.2; desde B0; A SECAS y SIN TOPE desde T089, E-006-P7)
go test -count=1 ./...                       # → 9 paquetes ok (B1: salvo sus 4 rojos declarados)
git diff 0311fa1 -- internal/event           # → vacío (SC-005, D-006-3)
grep -rnE '\.(go|md|jsonl|json|sh|yaml|yml|tsv|mod|gitignore):[0-9]+|[(`]:[0-9]+' \
     --include='*.go' --include='*.sh' .     # → nada (disciplina 8)
grep -rn "event.NewID" --include=*.go . | grep -v '^./internal/event/'   # → vacío (desde B2a)
```

### Nota de ejecución

**Claude Code no ejecuta git de escritura.** Cada bloque termina en una tarea **✋ commit**: el marcado
de casillas y las transcripciones de rojos y mutaciones viajan **en el mismo commit** que el código
que documentan. El **mensaje de commit va previsto en cada ✋, sin tildes ni ñ.**

**Dónde se registran rojos y mutaciones**: aquí, en la tarea que los espera (convención de 001–005;
no hay `registro-*.md` en este repositorio).

---

## Bloque B0 · Linter a 0 (D-006-13, `research.md` R9)

> **Por qué va primero**: la constitución exige `golangci-lint` limpio **para cerrar cualquier tarea**
> (`.specify/memory/constitution.md:72-75`). Con 7 avisos heredados, nada de lo que sigue podría
> cerrarse limpio (`plan.md` D-006-P10).
>
> **Enmendado el 2026-10-02 (E-006-P7).** Los avisos reales son **15**, no 7: el tope por defecto
> `max-same-issues: 3` escondía 8 `errcheck` más en `project.go`. El bloque empieza por **T089**, que
> quita el tope en `.golangci.yml`, y desde ahí todo se mide con `golangci-lint run` a secas.

- [x] **T089** *(nueva, 2026-10-02, E-006-P7; va la PRIMERA del bloque)* En `.golangci.yml`, quitar
  el tope de avisos: `issues.max-same-issues: 0` e `issues.max-issues-per-linter: 0`. Clave exacta de
  la 2.12.2, validada con `golangci-lint config verify`. **Rojo**: `golangci-lint run` **sin
  opciones** pasa de **7** a **15**. Ningún otro cambio en la configuración.
  > **Hecho el 2026-10-02.** `issues.max-issues-per-linter: 0` e `issues.max-same-issues: 0` en
  > `.golangci.yml`, con un comentario del motivo. `golangci-lint config verify` → sin error.
  > **Rojo medido**: `golangci-lint run` a secas, **7 issues** antes (errcheck 4 · revive 2 ·
  > staticcheck 1) → **15 issues** después (errcheck 12 · revive 2 · staticcheck 1).
- [x] **T001** **Rojo**: ejecutar `golangci-lint run` (2.12.2) y transcribir aquí los avisos. Se
  esperan **7**: 4 `errcheck` en `cmd/permea/project.go` (las cuatro escrituras de error de
  `runProject` y `projectJoin`), 1 `revive` `error-return` en `internal/config/endpoint.go`
  (`JuzgarEndpoint`), 1 `revive` `unused-parameter` y 1 `staticcheck` SA1007 en
  `internal/transport/adhesion_test.go`. Si salen otros o más, **se para**.
  *(Enmendado 2026-10-02, E-006-P7: tras T089 se esperan **15**: 12 `errcheck` en
  `cmd/permea/project.go` y los 3 de arriba. Si salen otros o más que esos 15, se para.)*
  > **Medido el 2026-10-02** (2.12.2, a secas, tras T089): **15**. Son `project.go` en las líneas 93,
  > 103, 129, 139, 147, 154, 168, 184, 198, 212, 221 y 227 (`errcheck`), `endpoint.go` 83
  > (`revive` error-return), y `adhesion_test.go` 267 (`revive` unused-parameter) y 194
  > (`staticcheck` SA1007). Coinciden con los 15 previstos por E-006-P7.
- [x] **T002** [P] Avisos 1–4: en `cmd/permea/project.go`, `_, _ =` delante de las cuatro escrituras
  de `fmt.Fprintln`/`fmt.Fprintf` a `stderr` («falta el verbo», «verbo desconocido», el error de la
  entrada y el del directorio actual). Es el idioma que el repositorio ya usa en `status.go` y
  `enroll.go`.
  *(Enmendado 2026-10-02, E-006-P7: cubre las **12** escrituras `fmt.Fprint*` al `stderr`/`stdout`
  de `runProject` y `projectJoin` que señala el linter sin tope, no sólo cuatro. Mismo idioma.)*
  > **Hecho.** 12 × `_, _ = fmt.Fprint…` en `project.go`. La escritura a `os.Stderr` de
  > `runProjectOS` no se toca: `errcheck` la excluye por defecto y no estaba señalada.
- [x] **T003** [P] Aviso 5: reordenar `config.JuzgarEndpoint` a `(admisible bool, errAnalisis error)`
  en `internal/config/endpoint.go`, y sus llamantes: `ParseEnrollmentString` (`enrollment.go`),
  `Config.Validate` (`config.go`), `Client.Send` y `Client.Adherir` (`internal/transport/transport.go`),
  y los cinco sitios de `internal/config/endpoint_test.go`. **Los tests cambian sólo el orden del
  destructurado, ninguna aserción.** El comentario de la función se actualiza por nombre.
  > **Hecho.** Firma `(admisible bool, errAnalisis error)`. Se cambió el destructurado en los cuatro
  > llamantes de producción y en los cinco sitios de `endpoint_test.go`, que sólo cambian de orden
  > (`_, errAnalisis` ×3 y `admisible, errAnalisis` ×2): **ninguna aserción tocada**. El comentario de
  > la función gana la sección «El orden de los resultados: el error al final (P-006 B0)».
- [x] **T004** [P] Avisos 6–7 en `internal/transport/adhesion_test.go`:
  - el parámetro `r` del manejador de `backendAdhesion` pasa a `_`;
  - en el `url.Parse` de `TestAdherir_ConservaLaCausaDelParseo`, `//nolint:staticcheck`, con el motivo
    en la misma línea: la URL inválida es el **sujeto** del test (E-006-P4). Será la única directiva
    `nolint` del repositorio.
  > **Hecho.** `r` → `_` en `backendAdhesion`; `//nolint:staticcheck // SA1007: …` con su motivo en
  > la misma línea. `grep -rn nolint --include=*.go .` → 1.
- [x] **T005** **Verde**: `golangci-lint run` → **0** *(a secas y sin tope, E-006-P7)*. `go test ./...` → 9 paquetes ok y **297 pass**.
  `git diff` de los `_test.go` sólo contiene el destructurado de T003, el `_` y el `nolint`.
  > **Verde medido**: `golangci-lint run` → **0 issues**. `go test -count=1 ./...` → **9 paquetes
  > ok, 297 pass**, 0 fail, 0 skip. El diff de los `_test.go` contiene sólo los 5 destructurados, el
  > `_` y el `nolint`.
- [x] **T006** **Mutación M-B0**: `JuzgarEndpoint` devuelve `admisible = true` siempre.
  - **Censo** (completarlo **antes** de mutar con `grep -ln 'http://' --include=*_test.go -r .` y
    escribir aquí los nombres que falten):
    - `TestJuzgarEndpoint_HechoEsquema` y `TestJuzgarEndpoint_NoAnalizableNoAfirmaEsquema`
      (`internal/config/endpoint_test.go`);
    - `TestValidate_RejectsNonHTTPS` (`config_test.go`);
    - `TestSend_RejectsHTTP` (`transport_test.go`);
    - `TestAdherir_RechazaCanalEnClaro` (`adhesion_test.go`);
    - los tests de rechazo de `http://` de `enrollment_test.go`, `enrollment_higiene_test.go` y
      `cmd/permea/enroll_reject_test.go`.
  - **Co-caídas**: ninguna fuera de la lista.
  - Protocolo de mutación. Acredita que el reorden no desenganchó ningún testigo.
  > **Censo DECLARADO el 2026-10-02, antes de mutar** (completado con el `grep` de arriba; la
  > mutación concreta es `return true, err` y `return true, nil` en las dos salidas de
  > `JuzgarEndpoint`). Deben caer **exactamente** estos 10 tests de nivel superior:
  > - `internal/config`: `TestJuzgarEndpoint_HechoEsquema`,
  >   `TestJuzgarEndpoint_NoAnalizableNoAfirmaEsquema`, `TestValidate_RejectsNonHTTPS`,
  >   `TestParseEnrollmentString_Rejects` (el subcaso http) y
  >   `TestParseEnrollmentString_ElErrorNoReproduceFragmentosDelArgumento` (su rama 5, «endpoint en
  >   claro»);
  > - `internal/transport`: `TestSend_RejectsHTTP`, `TestAdherir_RechazaCanalEnClaro`,
  >   `TestAdherir_LosDosHechosSonDistinguibles` y `TestSC008_SinTransporteSeguroNoSeCompleta`;
  > - `cmd/permea`: `TestEnroll_Reject_Malformed_AbortsBeforePing` (el subcaso http).
  >
  > **Co-caídas declaradas: ninguna.** Los demás paquetes, verdes.
  >
  > **Ejecutada el 2026-10-02.**
  > - **Primera forma descartada** (disciplina 3, mutación inválida): `return true, nil` deja `u` sin
  >   usar → `internal/config/endpoint.go: u declared and not used`. Revertida por edición inversa,
  >   md5 idéntico al previo.
  > - **Forma válida**: `return true, err` y `return u != nil, nil`. Misma semántica, «siempre
  >   admisible»; compila y pasa `vet`.
  > - **Resultado**: cayeron **exactamente los 10 tests de nivel superior declarados**. Ningún otro
  >   paquete cayó (`event`, `ingest`, `pricing`, `project`, `state`, `testutil`: ok).
  > - **⚠️ Precisión a nivel de subtest**: en `TestParseEnrollmentString_Rejects` cayó, además del
  >   subcaso `endpoint http (no https)`, el subcaso **`endpoint vacío`**, que el censo no nombraba
  >   (decía «el subcaso http»). Es el mismo hecho: `""` se analiza sin error y su esquema vacío no
  >   es admisible. **El censo vinculante era el de nivel superior y coincidió**. Se anota en vez de
  >   ocultarlo.
  > - **Reversión** por edición inversa: md5 `c617a5816134dfc19cf7084b1c30486c`, idéntico al previo.
  >   Suite: 297 pass.
- [x] **T007** Puertas del bloque *(con `golangci-lint run` a secas, sin tope: E-006-P7)*.
  > **Puertas, 2026-10-02**:
  > - `gofmt -l .` vacío;
  > - `go vet ./...` limpio;
  > - `golangci-lint run` **0 issues** (a secas, sin tope);
  > - `go test -count=1 ./...` **9/9 ok, 297 pass** (6 s);
  > - `git diff 0311fa1 -- internal/event` vacío;
  > - `GOOS=windows go build ./...` y `GOOS=darwin go build ./...` compilan;
  > - el `grep` de la disciplina 8, sin resultados.
- [ ] **T008** ✋ **Commit** (dueño): `006 B0: golangci-lint a 0 (los 7 avisos heredados de 005)`
  *(Enmendado 2026-10-02, E-006-P7: mensaje propuesto ahora
  `006 B0: golangci-lint a 0 y sin tope (15 avisos heredados de 005)`.)*

---

## Bloque B1 · Contratos y testigos — la frontera primero (Principio IV)

> **El golden NO puede nacer rojo, y se dice**: hoy el lector no decodifica `message.id` ni
> `requestId`, así que los centinelas no pueden aparecer. Nace verde y se valida por mutación (T014).
> **El rojo de B1 son los testigos de la derivación** (T013), que se escriben contra la API que ya
> existe (`ingest.FromClaudeCodeLine`). Así caen **por conducta** y no por compilación.

- [x] **T009** [P] Redescribir el origen del `event_id` en `specs/001-agente-inicial/contracts/transport.md`
  (§Deduplicación, la frase «lo genera el agente con `crypto/rand` (`event.NewID`)») y en
  `specs/001-agente-inicial/data-model.md` (la fila de `event_id`). Determinista, derivado de los
  identificadores del mensaje, remitiendo a `specs/006-medicion-fiel/contracts/event-id.md`. **La forma
  del campo no cambia.** FR-011.
  > **Hecho el 2026-10-02.** En `transport.md`, la frase de `crypto/rand` queda tachada y se añade
  > «Redescrito por P-006 FR-011»: determinista, hash sin sal de `message.id` y `requestId`, remisión
  > a `006/contracts/event-id.md`, misma forma (32 hex). En `data-model.md`, el origen de la fila
  > `event_id` pasa a «derivado del mensaje», con el anterior tachado.
- [x] **T010** [P] **E-006-P1**: en `internal/ingest/testdata/claude_code_sample.jsonl`, añadir
  `message.id` y `requestId` **sintéticos y distintos** a las dos líneas `assistant`
  (`msg_FIXTUREREF00000000000001` / `req_FIXTUREREF00000000000001` y `…02`). En
  `internal/project/testdata/README.md` §Lo que estos fixtures NO son, una nota fechada: se editó el
  2026-10-02 (006 E-006-P1) para añadir los dos campos; ninguna de las tres columnas del baseline los
  lee. **Condición**: `TestSC009_RegresionCeroDelCaminoDeIngesta` y `TestBoundary_NoDenylistLeaks`
  siguen **verdes sin tocar sus aserciones**. Si alguno cae, se para.
  > **Hecho.** Los ids `msg_FIXTUREREF0000000000000{1,2}` / `req_FIXTUREREF0000000000000{1,2}`
  > (28 caracteres) se insertaron como texto, sin reserializar: el resto de cada línea queda byte a
  > byte. La nota fechada está en `internal/project/testdata/README.md`. **Condición medida**:
  > `TestSC009_…` y `TestBoundary_NoDenylistLeaks` PASS, sin tocar sus aserciones, y el md5 de
  > `baseline-sc004.tsv` sin cambios (`bfb54db8f7a2d5036b75adc662cb4e19`).
- [x] **T011** [P] En `internal/ingest/testdata/boundary_sample.jsonl`, `message.id` y `requestId`
  **centinela** en las dos líneas `assistant` (`msg_CENTINELAID00000000001`, `req_CENTINELAID00000000001`
  y `…02`). Los cuatro valores, más los cuatro de T010, a la `denylist` de
  `internal/ingest/boundary_test.go`, con un bloque de comentario por nombre: «P-006 FR-013 · los
  identificadores del proveedor». FR-013.
  > **Hecho.** `msg_/req_CENTINELAID0000000000{1,2}` en `boundary_sample.jsonl`. La denylist gana el
  > bloque «P-006 FR-013 · los identificadores del proveedor», con 8 valores: 4 centinelas y 4 de
  > T010. `internal/ingest` verde: el golden nace verde, como se preveía.
- [x] **T012** [P] Añadir `message.id` y `requestId` (sintéticos, distintos entre sí) a las seis líneas
  `assistant` escritas en tests:
  - las cuatro de `internal/ingest/boundary_test.go` (en `TestBoundary_UnknownFutureFieldDoesNotLeak`,
    `TestBoundary_CostAvailable` ×2 y `TestBoundary_KeepsMetrics`);
  - las dos de `cmd/permea/main_test.go` (en `TestAgentVersion_ReachesEvent` y en el subtest
    `--scan con "plain" presente`).

  Sin ellas, FR-006 las dejaría sin evento (`research.md` R7), y la del subtest seguiría verde **sin
  ejercer nada**.
  > **Hecho.** Seis literales con `msg_/req_TESTLITERAL000000000000{1…6}`: 4 en `boundary_test.go`
  > (`UnknownFutureFieldDoesNotLeak`, `CostAvailable` ×2, `KeepsMetrics`) y 2 en `main_test.go`
  > (`TestAgentVersion_ReachesEvent` y el subtest de `--scan`). Suite completa verde tras T010–T012.
- [x] **T013** **Rojo** — nuevo `internal/ingest/eventid_test.go`, cuatro tests independientes, todos
  por `FromClaudeCodeLine`:
  - (1) `TestEventID_LaMismaLineaDaElMismoIDEnDosInstalaciones`: la misma línea con dos `Context` de
    sal, máquina y desarrollador distintos da el **mismo** `event_id`. **Cae** porque `event.NewID` es
    aleatorio.
  - (2) `TestEventID_VectorDelPar`: el par sintético de `contracts/event-id.md` da exactamente
    `43b8b3b6446704ae3cb8bb74683bf3e2`. **Cae** por la misma razón.
  - (3) `TestSintetica_NoSeEmite`: `model=<synthetic>` → `nil`. **Cae** porque hoy sólo se descarta
    el modelo vacío.
  - (4) `TestCasoLimite_SinIdentificadorNoSeEmite`: una línea `assistant` sin ninguno de los dos
    identificadores → `nil`. **Cae**: hoy se emite.

  Transcribir aquí el mensaje real de cada rojo. **Superficie nueva: ninguna.**
  > **🔴 Medido el 2026-10-02: los cuatro caen, cada uno por su razón.**
  > ```
  > eventid_test.go:45: P-006 FR-002: la misma línea dio dos event_id distintos en dos instalaciones:
  >       A: 616a9238b162b2e5190f8aa25dcbb5dd
  >       B: 62a88b6158eb4986599903afe15e4a34
  > eventid_test.go:58: P-006 FR-002: event_id del par = "c54d0a9a9c517ce61354a1c0477eca80", want "43b8b3b6446704ae3cb8bb74683bf3e2" (contracts/event-id.md, §Vectores de prueba)
  > eventid_test.go:71: P-006 FR-007: una línea <synthetic> NO debe producir evento; se produjo uno de modelo "<synthetic>"
  > eventid_test.go:85: P-006 FR-006: una línea sin message.id ni requestId NO debe producir evento; se produjo event_id="cbe9998c3f1f77a005bb4be8b82556bc"
  > ```
  > Los `event_id` del mensaje son los aleatorios de `event.NewID`, no identificadores del proveedor.
- [x] **T014** **Mutación M-B1** (el golden, nacido verde): en `internal/ingest/claudecode.go`,
  decodificar `message.id` **sólo para la mutación** y escribirlo en claro en `SessionRef`.
  - **Censo**: `TestBoundary_TresCaminosHaciaElExterior` (los tres caminos, por el centinela de T011)
    y `TestBoundary_NoDenylistLeaks` (por los valores de T010 en la denylist).
  - **Co-caída**: `TestSC009_RegresionCeroDelCaminoDeIngesta`, porque cambia la columna `session_ref`.
  - Ninguna otra.
  - Protocolo de mutación; revertir por edición inversa.
  > **Censo DECLARADO el 2026-10-02, antes de mutar.** La mutación concreta: añadir
  > `ID string \`json:"id"\`` a `rawRecord.Message` y `SessionRef: r.Message.ID` en
  > `FromClaudeCodeLine`. Ninguno de los tres tests usa `t.Run`, así que no hay subtests que nombrar.
  > - **Caen**: `TestBoundary_TresCaminosHaciaElExterior`, en los tres caminos (evento serializado,
  >   `queue.jsonl` y cuerpo transmitido), por `msg_CENTINELAID…`; y `TestBoundary_NoDenylistLeaks`, por
  >   `msg_FIXTUREREF…`.
  > - **Co-caída**: `TestSC009_RegresionCeroDelCaminoDeIngesta`, porque cambia la columna
  >   `session_ref`.
  > - **Los 4 rojos de T013 siguen en rojo**: ya caían antes de mutar, así que no cuentan como caída.
  > - Todo lo demás, verde.
  > **Ejecutada.** Compiló y pasó `vet`. Cayeron **exactamente** los declarados:
  > - `TestBoundary_TresCaminosHaciaElExterior`: los 3 caminos × los 2 centinelas `msg_CENTINELAID…`
  >   (`FUGA DE FRONTERA [camino 1 · evento serializado]`, `[camino 2 · queue.jsonl]` y
  >   `[camino 3 · cuerpo transmitido]`);
  > - `TestBoundary_NoDenylistLeaks`: los 2 `msg_FIXTUREREF…`;
  > - co-caída `TestSC009_…`: «conjunto de identidades: got 2 filas, want 1».
  >
  > Los 4 rojos de T013 siguieron en rojo; no cayó nada más (8 paquetes ok, sólo `ingest` en
  > FAIL). **Reversión** por edición inversa: md5 de `claudecode.go`
  > `ad24460e2e8e3a11ac2dc3a415057fb1` antes y después; `git diff` del fichero, vacío.
- [x] **T015** Puertas del bloque. **`go test` cierra con exactamente los cuatro rojos de T013** y nada
  más en rojo; se transcribe la lista de `FAIL`.
  > **Puertas, 2026-10-02**:
  > - `gofmt -l .` vacío;
  > - `go vet ./...` limpio;
  > - `golangci-lint run` **0**;
  > - `go test -count=1 ./...`: **297 pass, 4 fail**. Los FAIL son exactamente los de T013, todos en
  >   `internal/ingest`; los otros 8 paquetes ok;
  > - `git diff 0311fa1 -- internal/event` vacío;
  > - `git diff a702de8 --stat`: ningún fichero de producción (sólo 2 tests, 2 fixtures, el README del
  >   fixture, 2 docs de 001 y `tasks.md`; el nuevo `eventid_test.go` está sin seguimiento).
- [ ] **T016** ✋ **Commit** (dueño):
  `006 B1: contratos de 001 redescritos, ids en fixtures y testigos del event_id EN ROJO`

---

## Bloque B2a · La derivación (`contracts/event-id.md`, `plan.md` D-006-P1/P2)

- [x] **T090** *(nueva, 2026-10-02, añadido del orquestador; va la PRIMERA del bloque)* FR-013 dice
  «ni entero ni como fragmento reconocible», y la `denylist` de `internal/ingest/boundary_test.go`
  sólo busca los identificadores enteros. Añadir los **núcleos** de los centinelas, `CENTINELAID` y
  `FIXTUREREF`, que no pueden aparecer por azar en una salida hexadecimal. **Acreditarlo con una
  mutación declarada antes de mutar**: filtrar un identificador **truncado** a un campo del evento
  debe caer; con la denylist anterior no caería.
  > **Censo DECLARADO el 2026-10-02, antes de mutar.**
  > - **Mutación**: decodificar `message.id` en `rawRecord` y poner en `SessionRef` el identificador
  >   **truncado** `id[4:15]`, con guarda de longitud para no panicar (`msg_CENTINELAID…` →
  >   `CENTINELAID`; `msg_FIXTUREREF…` → `FIXTUREREF0`). Ninguno de los tests afectados usa `t.Run`.
  > - **Con la denylist nueva caen** `TestBoundary_TresCaminosHaciaElExterior` (los 3 caminos, por
  >   el núcleo `CENTINELAID`) y `TestBoundary_NoDenylistLeaks` (por el núcleo `FIXTUREREF`).
  >   Co-caída: `TestSC009_RegresionCeroDelCaminoDeIngesta`, porque cambia `session_ref`. Los 4 rojos
  >   de B1 siguen en rojo. Nada más.
  > - **Contraprueba con la denylist anterior** (núcleos quitados por edición inversa, con la mutación
  >   puesta): `TresCaminos` y `NoDenylistLeaks` **pasan**, y sólo quedan `TestSC009` y los 4 rojos de
  >   B1. Es la prueba de que sin núcleos no caería.
  >
  > **Ejecutada el 2026-10-02.** Núcleos `CENTINELAID` y `FIXTUREREF` en la denylist; `internal/ingest`
  > sólo con los 4 rojos de B1.
  > - **Con la mutación y la denylist nueva**, cayeron exactamente los declarados:
  >   `NoDenylistLeaks` (`boundary_test.go:120`, «el evento contiene "FIXTUREREF"»), `TresCaminos`
  >   (`:180` camino 1, `:193` camino 2 y `:216` camino 3, «contiene "CENTINELAID"») y la co-caída
  >   `TestSC009`, más los 4 rojos previos.
  > - **Contraprueba**: núcleos quitados por edición inversa, con md5 igual a la denylist de B1. Con
  >   la mutación puesta, `TresCaminos` y `NoDenylistLeaks` **pasaron**, y sólo quedaron `TestSC009` y
  >   los 4 rojos. Sin núcleos, la fuga truncada no cae.
  > - **Reversión**: núcleos repuestos (md5 de `boundary_test.go` igual al de después de T090,
  >   `29a190ca64fd6d9beaba2224b3d2ad6c`) y mutación revertida (md5 de `claudecode.go`
  >   `ad24460e2e8e3a11ac2dc3a415057fb1`, sin diff).
- [x] **T017** [P] **Rojo** (5) `TestCasoLimite_UnSoloIdentificador` en `internal/ingest/eventid_test.go`.
  Las tres formas (par, sólo `message.id`, sólo `requestId`) dan `event_id` **distintos entre sí**,
  estables y de 32 hex. Incluye **el mismo valor** usado como `message.id` solo y como `requestId`
  solo, que debe dar dos `event_id` distintos (ver m1). **Cae** por aleatoriedad. SC-008 (b).
  > **🔴 Medido**: cae sólo el subtest `estable` (`eventid_test.go:142`, «forma par: la misma línea
  > dio dos event_id distintos», y lo mismo para `solo_message_id` y `solo_request_id`).
  > `tres_formas_distintas`, `mismo_valor_en_las_dos_formas_solas` y `forma_32_hex` nacen verdes y los
  > validan m1, m-formas y m-hex.
- [x] **T018** [P] **Rojo** (6) `TestEventID_VectoresDeUnaSolaFormaYAmbiguedad`, en el mismo fichero:
  - los vectores de «sólo `message.id`» y «sólo `requestId`» de `contracts/event-id.md`;
  - los dos de la ambigüedad sin prefijo de longitud.

  Para que el vector de ambigüedad sea alcanzable por la API pública, se usa como `message.id`/`requestId`
  el par de componentes del contrato. **Cae** por aleatoriedad.
  > **🔴 Medido**: caen `solo_message_id`, `solo_request_id`, `ambiguedad_a_bc` y `ambiguedad_ab_c`
  > (`eventid_test.go:186`, «event_id = <aleatorio>, want <vector>»). `ambiguedad_distintos` nace
  > verde y lo valida m2.
- [x] **T019** Nuevo `internal/ingest/eventid.go`: la derivación del contrato, sin exportar:
  - dominio `permea/event_id/v1` + `claude_code` + tipo;
  - componentes con prefijo de longitud `uint32` big-endian;
  - SHA-256 truncado a 16 bytes, hex en minúsculas.

  **Superficie nueva**: la función, mirada por T013 (1)(2), T017 y T018.
  > **Hecho.** `derivarEventID` / `hashEventID` en `internal/ingest/eventid.go`, sin exportar. La
  > guarda de longitud usa `math.MaxInt32` para que compile también con `int` de 32 bits
  > (`GOARCH=386 go build ./internal/ingest` compila). `gosec` no marca nada.
- [x] **T020** En `internal/ingest/claudecode.go`:
  - `rawRecord` decodifica `message.id` y `requestId`. El comentario de la guarda de frontera se
    actualiza **por nombre**: son metadatos técnicos, nunca salen en claro, y sólo derivan el
    `event_id` (spec §Verificación de la frontera);
  - `<synthetic>` → `nil`;
  - sin ningún identificador → `nil`;
  - el `event_id` sale de T019, y **`event.NewID` deja de llamarse**.

  **Verde**: (1)–(6). El golden y `TestSC009` siguen verdes.
  > **Verde**: (1)–(6) en verde, y **los vectores del contrato salen a la primera**, sin tocar ni la
  > implementación ni el vector. Suite: **312 pass** (297 + los 4 de B1 + los 2 nuevos, con 9
  > subtests), 0 fail. Golden y `TestSC009` verdes. `event.NewID` sin llamantes de producción.
- [x] **T021** **Mutaciones m1–m4** (protocolo; cada una con su censo, todo lo demás verde):
  - **m1**: quitar el tipo del dominio. Censo: (2), (5) y (6). (5) cae por el caso del mismo valor en
    las dos formas solas.
  - **m2**: quitar el prefijo de longitud. Censo: (2) y (6). (5) **no** cae: las formas siguen
    separadas por el tipo.
  - **m3**: volver a emitir `<synthetic>`. Censo: (3).
  - **m4**: meter `ctx.Salt` en el hash. Censo: (1), (2) y (6).
  > **Censos DECLARADOS el 2026-10-02, antes de mutar**, nombrando subtests. Un test que contiene
  > un subtest caído cae también como padre, y no se repite en cada línea. Todo lo no nombrado,
  > verde.
  > - **m1** · `hashEventID` ignora el tipo. Caen `TestEventID_VectorDelPar`,
  >   `TestCasoLimite_UnSoloIdentificador/mismo_valor_en_las_dos_formas_solas` y
  >   `TestEventID_VectoresDeUnaSolaFormaYAmbiguedad/{solo_message_id, solo_request_id,
  >   ambiguedad_a_bc, ambiguedad_ab_c}`.
  > - **m2** · sin prefijo de longitud. Caen `TestEventID_VectorDelPar` y
  >   `TestEventID_VectoresDeUnaSolaFormaYAmbiguedad/{solo_message_id, solo_request_id,
  >   ambiguedad_a_bc, ambiguedad_ab_c, ambiguedad_distintos}`. (5) entero verde.
  > - **m3** · `<synthetic>` vuelve a emitirse. Cae `TestSintetica_NoSeEmite`.
  > - **m4** · el `event_id` se rehace con `ctx.Salt` (`hashEventID(tipoPar, id, ctx.Salt)`). Caen
  >   `TestEventID_LaMismaLineaDaElMismoIDEnDosInstalaciones`, `TestEventID_VectorDelPar` y
  >   `TestEventID_VectoresDeUnaSolaFormaYAmbiguedad/{solo_message_id, solo_request_id,
  >   ambiguedad_a_bc, ambiguedad_ab_c}`. (5) verde, porque usa una sola sal.
  >
  > **Dos mutaciones más, añadidas por la disciplina 3**: (5)/`tres_formas_distintas` y
  > (5)/`forma_32_hex` nacieron verdes, y ninguna de m1–m4 las tumba.
  > - **m-hex** · truncar a 15 bytes (30 hex). Caen `TestEventID_VectorDelPar`,
  >   `TestCasoLimite_UnSoloIdentificador/forma_32_hex` y
  >   `TestEventID_VectoresDeUnaSolaFormaYAmbiguedad/{solo_message_id, solo_request_id,
  >   ambiguedad_a_bc, ambiguedad_ab_c}`.
  > - **m-formas** · el par se deriva sólo de `message.id` (`hashEventID(tipoSoloMessageID, messageID)`).
  >   Caen `TestEventID_VectorDelPar`, `TestCasoLimite_UnSoloIdentificador/tres_formas_distintas` y
  >   `TestEventID_VectoresDeUnaSolaFormaYAmbiguedad/{ambiguedad_a_bc, ambiguedad_ab_c}`.
  > **Ejecutadas las 6.** Cada una cayó **exactamente** según su censo, contando los padres, y se
  > revirtió por edición inversa con md5 idéntico (`eventid.go` `4ba99d404f547a0d557172ed1f78378b`,
  > `claudecode.go` `0613145e83d6a21cb58b9ce357325a44`):
  > - **m1**: `VectorDelPar`, (5)/`mismo_valor_en_las_dos_formas_solas`, y (6) con sus 4 vectores;
  > - **m2**: `VectorDelPar`, y (6) con sus 4 vectores más `ambiguedad_distintos` («falta el prefijo de
  >   longitud»);
  > - **m3**: `TestSintetica_NoSeEmite`;
  > - **m4**: (1), `VectorDelPar`, y (6) con sus 4 vectores;
  > - **m-hex**: `VectorDelPar`, (5)/`forma_32_hex` y (6) con sus 4 vectores;
  > - **m-formas**: `VectorDelPar`, (5)/`tres_formas_distintas` y (6)/`ambiguedad_a_bc` y
  >   `ambiguedad_ab_c`.
- [x] **T022** Puertas del bloque (incluido el `grep` de `event.NewID`).
  > **Puertas, 2026-10-02**:
  > - `gofmt -l .` vacío;
  > - `go vet ./...` limpio;
  > - `golangci-lint run` **0**;
  > - `go test -count=1 ./...` **9/9 ok, 312 pass, 0 fail**;
  > - `git diff 0311fa1 -- internal/event` vacío;
  > - Windows y darwin compilan;
  > - `grep event.NewID` fuera de `internal/event`: vacío;
  > - disciplina 8, sin resultados.
- [ ] **T023** ✋ **Commit** (dueño):
  `006 B2a: event_id determinista por hash con dominio y sin sal; synthetic y lineas sin ids no se emiten`

---

## Bloque B2b · La pasada (`plan.md` D-006-P3/P4, `research.md` R3/R4)

- [x] **T024** Nuevo `internal/ingest/pasada.go`, **andamiaje**:
  - un tipo «pasada», como puntero en `ingest.Context` y **nil válido** (patrón del `Resolutor`);
  - métodos que **no deduplican, no cuentan** y devuelven un resumen vacío.

  Existe para que T025–T029 caigan **por conducta** y no por compilación. **Superficie nueva**,
  mirada por T025–T029.
  > **Hecho el 2026-10-02.** `Pasada`, `Recuentos`, `NuevaPasada`, y `Recuentos()`/`Resumen()`
  > devolviendo ceros y vacío. El campo `Pasada *Pasada` en `ingest.Context`, nil válido, con
  > comentario por nombre. Compila; sin cablear.
- [x] **T025** [P] **Rojo** (7) `TestPasada_UnMensajeDeTresLineasEsUnEvento` (`internal/ingest/pasada_test.go`):
  tres líneas con el mismo par, en una pasada → **un** evento, con los tokens de **una** línea. **Cae**:
  el andamiaje no deduplica.
  > **🔴 Medido**: caen los 3 subtests. `un_evento` (`pasada_test.go:62`, «produjeron 3 eventos; se
  > esperaba 1»), `tokens_de_una_linea` (`:67`, «suman 420 tokens») y `recuentos` (`:73`, todos a
  > 0).
- [x] **T026** [P] **Rojo** (8) `TestCasoLimite_ConsumoDistinto`: segunda línea del mismo par con
  `usage` distinto → **no** se emite, el evento conserva el de la primera, y el contador de «consumo
  distinto» = 1. **Cae** por el contador. SC-008 (a).
  > **🔴 Medido**: caen los 3. `un_solo_evento` (`:87`, «2 eventos»), `conserva_el_consumo_de_la_primera`
  > (`:92`, «suman 1140 tokens») y `cuenta_la_discrepancia` (`:97`, «ConsumoDistinto = 0»).
- [x] **T027** [P] **Rojo** (9) `TestCasoLimite_SinIdentificador`: una línea sin identificadores →
  contador «sin identificador» = 1. **Cae** por el contador. SC-008 (c).
  > **🔴 Medido**: caen `cuenta_sin_identificador` (`:112`, «SinIdentificador = 0») y
  > `sinteticas_aparte` (`:117`, «Sinteticas = 0»). El segundo es una ampliación: el contador de
  > sintéticas no tenía testigo.
- [x] **T028** [P] **Rojo** (10) `TestPasada_ElResumenNoLlevaIdentificadores`, con tres aserciones
  independientes:
  - el resumen **no está vacío** y contiene el número de facturables;
  - no contiene ninguno de los identificadores de entrada;
  - no contiene ningún `event_id` emitido.

  **Cae** por la primera: el andamiaje devuelve vacío. Sin ella el test nacería verde y vacuo.
  > **🔴 Medido**: cae `no_vacio_con_facturables` (`:143`, «resumen = ""»).
  > `sin_identificadores_de_entrada` y `sin_event_id` nacen verdes y los validan m-ids y m7.
- [x] **T029** **Rojo** (11) `TestPasada_GenerateEncolaUnoPorMensaje` en `cmd/permea/main_test.go`. En
  sandbox, con `logs_root` a un temporal que contiene un log con un mensaje de tres líneas, una pasada
  de `generate()` deja **un** evento en `queue.jsonl`. **Cae**: deja tres.
  > **🔴 Medido**: `main_test.go:443`, «un mensaje de tres líneas dejó 3 eventos en la cola; se
  > esperaba 1». En T031, la llamada a `a.generate()` del test pasó de `_, err :=` a `_, _, err :=`,
  > por la firma nueva; **ninguna aserción cambió**.
- [x] **T030** Implementar la pasada en `internal/ingest/pasada.go`:
  - conjunto con **clave de 16 bytes**, no la cadena hex (`research.md` R3.4), y el `usage` de la
    primera línea;
  - los seis contadores de `data-model.md`;
  - el resumen.

  Cablearla en `FromClaudeCodeLine`: una repetición devuelve `(nil, nil)`; con la pasada a nil, se
  emite.
  > **Hecho.** Conjunto `map[[16]byte]consumo` (clave: los 16 bytes del `event_id`), seis
  > contadores, y `Resumen()` sólo con recuentos. Cableado en `FromClaudeCodeLine`, en este orden:
  > facturable → sintética → sin identificador → `registrar` (la primera emite; las repetidas no, se
  > cuentan, y nunca se suman).
- [x] **T031** En `cmd/permea/main.go`:
  - `generate()` instancia una pasada **por llamada** (en `--daemon`, una por ciclo);
  - `runOnce` escribe el resumen por **stderr**;
  - `tick` lo escribe sólo si la pasada leyó alguna línea facturable.

  **Verde**: (7)–(11).
  > **Verde**: `generate()` devuelve `(int, *ingest.Pasada, error)` e instancia una pasada por
  > llamada. `runOnce` escribe el resumen por stderr; `tick`, sólo si `Facturables > 0`. (7)–(11)
  > verdes. Suite: **328 pass**, 0 fail.
- [x] **T032** **Medida** (no es test): memoria de una pasada sobre **10 698 mensajes sintéticos**,
  tamaño de la referencia M2. `runtime.MemStats` antes y después, en un programa o test temporal que
  **no se commitea**. Transcribir aquí la cifra, que contrasta la estimación de `research.md` R3.4.
  > **Medido el 2026-10-02**, con un test temporal (`internal/ingest/zz_medida_memoria_test.go`,
  > **borrado, no se commitea**):
  > - 10 698 mensajes sintéticos en 22 925 líneas (2 por mensaje, y 3 en uno de cada 7);
  > - `runtime.GC()` y `HeapAlloc` antes de crear la pasada y después de leerlo todo, con la pasada
  >   viva;
  > - 3 ejecuciones: **891 592 / 895 008 / 900 456 bytes ≈ 0,90 MB, unos 84 B por mensaje**.
  >
  > Queda por debajo de la estimación de 1–2 MB de `research.md` R3.4.
- [x] **T033** **Mutaciones m5–m8**:
  - **m5**: el conjunto nunca recuerda. Censo: (7) y (11). Co-caída: (8), porque sin conjunto no hay
    primera línea con la que comparar.
  - **m6**: no contar la discrepancia. Censo: (8).
  - **m7**: escribir el último `event_id` en el resumen. Censo: (10).
  - **m8**: `generate()` sin instanciar la pasada. Censo: (11). **(7) queda verde**, y eso demuestra
    que (11) mira el camino real.
  > **Censos DECLARADOS el 2026-10-02, antes de mutar**, nombrando subtests. Un test con un subtest
  > caído cae también como padre. Todo lo no nombrado, verde.
  > - **m5** · `registrar` nunca recuerda (`… ; visto && false {`). Caen
  >   `TestPasada_UnMensajeDeTresLineasEsUnEvento/{un_evento, tokens_de_una_linea, recuentos}`, la
  >   co-caída `TestCasoLimite_ConsumoDistinto/{un_solo_evento, conserva_el_consumo_de_la_primera,
  >   cuenta_la_discrepancia}` y `TestPasada_GenerateEncolaUnoPorMensaje`.
  > - **m6** · no se cuenta la discrepancia. Cae `TestCasoLimite_ConsumoDistinto/cuenta_la_discrepancia`.
  > - **m7** · la pasada guarda el último `event_id` y el resumen lo añade. Cae
  >   `TestPasada_ElResumenNoLlevaIdentificadores/sin_event_id`.
  > - **m8** · `generate()` no asigna `ictx.Pasada`. Cae `TestPasada_GenerateEncolaUnoPorMensaje`. **(7)
  >   verde.**
  >
  > **Una más, añadida por la disciplina 3**: (10)/`sin_identificadores_de_entrada` nació verde, y
  > ninguna de m5–m8 la tumba.
  > - **m-ids** · `FromClaudeCodeLine` pasa el último `message.id` a la pasada y el resumen lo añade.
  >   Cae `TestPasada_ElResumenNoLlevaIdentificadores/sin_identificadores_de_entrada`.
  > **Ejecutadas las 5.** Cada una cayó **exactamente** según su censo, contando los padres, y se
  > revirtió por edición inversa con md5 idéntico (`pasada.go` `c0c7d117f685b1024dcb1e53efb9b8d7`,
  > `claudecode.go` `669ba50d21a0ccaa120d4f1afc733362`, `main.go` `253c09721fd129b9ce59f87ffd4fd0ce`):
  > - **m5**: (7) ×3, (8) ×3 y (11);
  > - **m6**: (8)/`cuenta_la_discrepancia`;
  > - **m7**: (10)/`sin_event_id`;
  > - **m8**: (11), con (7) verde;
  > - **m-ids**: (10)/`sin_identificadores_de_entrada`.
  >
  > `vet` limpio con cada mutación.
- [x] **T034** Puertas del bloque.
  > **Puertas, 2026-10-02**:
  > - `gofmt` vacío;
  > - `vet` limpio;
  > - lint **0**;
  > - `go test -count=1 ./...` **9/9 ok, 328 pass, 0 fail**;
  > - `internal/event` sin diff;
  > - Windows y darwin compilan;
  > - `event.NewID` sin llamantes;
  > - disciplina 8, sin resultados.
- [ ] **T035** ✋ **Commit** (dueño):
  `006 B2b: un mensaje un evento dentro de la pasada, con resumen por stderr`

---

## Bloque B2c · Dry-run y actualización (`plan.md` D-006-P5, FR-009, FR-010)

- [x] **T036** [P] **Rojo** (12) `TestScan_UnEventoPorMensaje` en `cmd/permea/main_test.go`. Proceso:
  `--scan` sobre un fichero con un mensaje de tres líneas → exit 0 y exactamente **una** línea
  `evento:` en stdout. **Cae**: salen tres.
  > **🔴 Medido el 2026-10-02**: `main_test.go:484`, «`--scan` imprimió 3 líneas `evento:` para un
  > mensaje de tres líneas; se esperaba 1». En sandbox, con identificadores sintéticos.
- [x] **T037** [P] **Rojo** (13) `TestScan_LineaConCuatroPartidasYEventID`: la línea `evento:` lleva
  `in=`, `out=`, `cw=`, `cr=` y `event_id=` con 32 hex. **Estos nombres de campo los usan las medidas
  V3 y V4 del quickstart.** **Cae**: hoy no hay `cw`, `cr` ni `event_id`.
  > **🔴 Medido**: caen los dos subtests. `cuatro_partidas` (`main_test.go:507`, «la línea no lleva
  > "cw=7"» y «"cr=3"») y `event_id_32_hex` (`:514`, «la línea no lleva `event_id=`»).
- [x] **T038** (14) `TestActualizar_NoReenviaNiReescribeLaCola` en `cmd/permea/main_test.go`. Sandbox
  con `state.json` a mitad de un log y un evento antiguo con `event_id` aleatorio en `queue.jsonl`. Una
  pasada de `generate()`:
  - encola sólo lo posterior al offset;
  - deja la línea antigua **byte a byte** igual.

  **Nace verde** (no hay nada que cambiar en el estado), así que lo valida T040.
  > **Nace verde**, con sus dos subtests (`la_cola_previa_byte_a_byte` y
  > `solo_lo_posterior_al_offset`). Estado previo con `state.New` + `Save` (offset al final de la
  > primera línea), y cola sembrada con un evento de `event_id` `00112233…eeff` y
  > `agent_version` `0.2.1`. Lo acreditan m9 y m10.
- [x] **T039** En `cmd/permea/main.go`, `dryRun()`:
  - pasada propia;
  - formato de T037;
  - resumen por stderr, que sustituye a «N eventos generados» o lo amplía.

  **Verde**: (12) y (13).
  > **Verde**: `dryRun()` con pasada propia. Línea `evento: … in= out= cw= cr= cost= cost_avail=
  > project_ref= event_id=`. Por stderr, la línea «N eventos generados» se conserva y se añade el
  > resumen de la pasada. (12) y (13) verdes. Suite **335 pass**, 0 fail.
- [x] **T040** **Mutaciones m9–m10**:
  - **m9**: en `generate()`, empezar con un estado nuevo en vez de cargar `state.json`. Censo: (14).
    **Antes de mutar**, buscar con `grep` en `cmd/permea/*_test.go` otros tests que ejecuten dos
    pasadas, y declararlos como co-caídas si los hay.
  - **m10**: vaciar la cola al empezar `generate()`. Censo: (14). Co-caídas: ninguna, porque (11) parte
    de una cola vacía.
  > **Censos DECLARADOS el 2026-10-02, antes de mutar**, nombrando subtests. El `grep` previo
  > encontró dos tests más que ejecutan `--run` como proceso: los de `TestRetirada_*` en
  > `main_test.go` y los casos positivos de `TestProjectJoin_*` en `project_test.go`. Ninguno depende
  > de lo que alteran m9 o m10: los primeros paran antes de generar, y los segundos parten de estado
  > y cola vacíos. **Co-caídas: ninguna.** Todo lo no nombrado, verde.
  > - **m9** · `generate()` empieza con `state.New()` en vez de `state.Load`. Cae
  >   `TestActualizar_NoReenviaNiReescribeLaCola/solo_lo_posterior_al_offset`, porque relee la primera
  >   línea y salen 2 eventos nuevos. `la_cola_previa_byte_a_byte`, verde.
  > - **m10** · `generate()` borra `queue.jsonl` al empezar. Cae
  >   `TestActualizar_NoReenviaNiReescribeLaCola/la_cola_previa_byte_a_byte`. `solo_lo_posterior_al_offset`
  >   verde, porque cuenta sólo los eventos nuevos.
  > **Ejecutadas.** Cada una cayó **exactamente** según su censo y se revirtió por edición inversa
  > (md5 de `main.go` `0949b8361da7aa7a8f3a0a570a8802c1`, idéntico):
  > - **m9**: (14)/`solo_lo_posterior_al_offset` (`main_test.go:584`, «la pasada encoló 2 eventos
  >   nuevos»);
  > - **m10**: (14)/`la_cola_previa_byte_a_byte` (`main_test.go:569`, «la línea que ya estaba en la
  >   cola cambió o desapareció»).
  >
  > `vet` limpio con las dos.
- [x] **T041** Puertas del bloque.
  > **Puertas, 2026-10-02**:
  > - `gofmt` vacío;
  > - `vet` limpio;
  > - lint **0**;
  > - `go test -count=1 ./...` **9/9 ok, 335 pass, 0 fail**;
  > - `internal/event` sin diff;
  > - Windows y darwin compilan;
  > - `make run` (el `--scan` del fixture) sale con 0: 2 líneas `evento:` con `cw=`/`cr=`/`event_id=`
  >   y el resumen «pasada: 2 líneas facturables · 2 eventos · …»;
  > - `event.NewID` sin llamantes;
  > - disciplina 8, sin resultados.
- [ ] **T042** ✋ **Commit** (dueño): `006 B2c: dry-run con cuatro partidas y event_id; actualizar no reenvia`

---

## Q-006-1 · `claude-sonnet-5` — BLOQUEANTE antes de B3

- [x] **T043** ✋ **Decisión del dueño**: `claude-sonnet-5`, ¿2 / 10 / 2,50 / 0,20 (página oficial) o
  3 / 15 / 3,75 / 0,30 (catálogo de entonces *(Enmendado 2026-10-02, Q-006-1 resuelta: era `865bba0`; ahora `e50d0a5`)*)?
  (spec §Preguntas abiertas.)
  - **Si el catálogo de la plataforma cambia**, el dueño comunica el **commit nuevo** de
    `backend/config/pricing.php`.
  - **B3 no empieza** sin esta respuesta.
  > **Decidido el 2026-10-02 (dueño)**: `claude-sonnet-5` = **2.00 / 10.00 / 2.50 / 0.20**. La nota oficial
  > dice que 2 / 10 ya es el precio estándar y que la subida a 3 / 15 no se producirá. La plataforma lo
  > corrigió en **`permea-platform@e50d0a5`** (PR #78), que es el catálogo de referencia nuevo
  > (`git -C ~/dev/permea-platform rev-parse --short origin/main` = `e50d0a5`).
- [x] **T044** *(condicional: sólo si T043 cambia el catálogo)* Enmienda fechada de la spec:
  - M4 rehecho sobre el commit nuevo;
  - las notas de FR-014, FR-019 (limitación 3) y SC-009;
  - Q-006-1 marcada como resuelta, con fecha.

  Mensaje para el ✋ commit, que el dueño puede unir a T051:
  `006 spec: Q-006-1 resuelta, M4 sobre el catalogo <commit>`
  > **Hecho el 2026-10-02.** Enmiendas fechadas («Enmendado 2026-10-02, Q-006-1 resuelta: …»):
  > - en la spec: M4 sobre `e50d0a5`, con la fila de Sonnet 5 a 2.00 / 10.00 / 2.50 / 0.20, las citas de
  >   línea rehechas y dos limitaciones; FR-014 replica `e50d0a5`; FR-019 pierde la tercera
  >   limitación; SC-009 cita `e50d0a5`; SC-011 cuenta dos limitaciones; Q-006-1 queda RESUELTA;
  > - en T045, T047, T048 y T072 de este fichero;
  > - en el barrido de `plan.md`, `research.md`, `quickstart.md` y `contracts/tarifas.md`.
  >
  > Mensaje propuesto: `006 spec: Q-006-1 resuelta, M4 sobre el catalogo e50d0a5`

---

## Bloque B3 · Tarifas (`contracts/tarifas.md`, `plan.md` D-006-P6)

- [x] **T045** [P] En `internal/pricing/pricing_test.go`, una **tabla esperada escrita aparte**, con
  comentario de procedencia (repositorio · fichero · commit de T043 — **`e50d0a5`**, Q-006-1 resuelta el
  2026-10-02), y tres tests independientes:
  - (15) `TestEspejo_RecuentoDeClaves`: 16 claves. **Cae**: hay 3.
  - (16) `TestEspejo_CifrasClaveAClave`: cada clave esperada existe con sus **cuatro** cifras exactas.
    **Cae**: faltan 13 claves y `claude-opus-4-6` difiere.
  - (17) `TestEspejo_NingunaClaveSobra`. **Nace verde** (las 3 actuales están entre las 16) → m12.
  > **Paso 0 (2026-10-02), antes de escribir nada.** `grep` de `cost_usd|CostUSD|cost_avail|CostAvailable|pricing.`
  > en `.go`, `.json`, `.jsonl` y `.golden`. **Sólo (18) cambia de resultado.** Los demás tests que
  > tocan coste (`TestBoundary_CostAvailable`, `TestBoundary_KeepsMetrics`) exigen `true` y > 0
  > para `claude-opus-4-6`, o `false` y 0 para un modelo desconocido, y eso no cambia; el literal de
  > `event_test.go` no sale de la tabla. No hay goldens ni fixtures con cifras de coste.
  >
  > **🔴 Medido**: la tabla esperada se generó por script desde
  > `git show e50d0a5:backend/config/pricing.php`, no desde `pricing.go`.
  > - (15) `pricing_test.go:67`: «la tabla tiene 3 claves; el catálogo replicado tiene 16».
  > - (16): 14 subtests en rojo, 13 por «falta la clave …» (`:84`) y `claude-opus-4-6` por «entrada = 15,
  >   want 5 …» (`:88`–`:97`). `claude-haiku-4-5` y `claude-sonnet-4-6` nacen verdes → m-h45 y m-s46.
  > - (17) nace verde → m12.
- [x] **T046** [P] En el mismo fichero:
  - (18) `TestCost` pasa a `claude-opus-4-6` a 5 / 25 / 6,25 / 0,50. **Cae**: la tabla dice 15/75.
  - (19) `TestCost_Opus55AMano`: un evento de `claude-opus-5-5` con las **cuatro** partidas, contra el
    cálculo a mano a 4 / 20 / 5 / 0,20. **Cae**: no hay fila.
  > **🔴 Medido**: (18) `pricing_test.go:21`, «coste fuera de ±1%: got 110.25 want 36.75». (19)
  > `:117`, «claude-opus-5-5 debe tener fila (ok=true)». El vector de (19) es el del orquestador:
  > 300 000 / 40 000 / 100 000 / 1 500 000 → literal `2.80`, con la suma en un comentario y la
  > comparación ±1 % de `TestCost`. `Cost` lo admite tal cual: no redondea.
  > *(Enmendado 2026-10-02, remate de B3: el vector anterior —300 000 / 40 000 / 100 000 / 1 500 000 →
  > `2.80`, ±1 %— no veía una tarifa desviada en un céntimo (m11 no lo tumbó) ni un coste redondeado a
  > céntimos. **(19) pasa a tokens irregulares 123 457 / 7 891 / 45 679 / 987 653 → literal
  > `1.0775736`** (0,493828 + 0,157820 + 0,228395 + 0,1975306, calculado por el orquestador), con
  > **diferencia absoluta ≤ 1e-9**. Es el único test que distingue las cuatro partidas. (18) no se toca:
  > su ±1 % viene de SC-001 de 001. **Nace verde** con la tabla de T047: `Cost` devuelve exactamente
  > `1.0775736`, con diferencia 0.)*
- [x] **T047** `internal/pricing/pricing.go`: las 16 claves del catálogo de T043, y la **cabecera** de
  `contracts/tarifas.md` (fuente, verificación, aprobación, catálogo replicado, casamiento exacto, las
  ~~tres~~ **dos** limitaciones). **Verde**: (15), (16), (18) y (19). *(Enmendado 2026-10-02, Q-006-1
  resuelta: catálogo `e50d0a5`; la limitación de Sonnet 5 desaparece.)*
  > **Verde**: las 16 claves de `e50d0a5` (filas generadas por el mismo script) y la cabecera con los
  > siete elementos. (15), (16), (18) y (19) verdes; `TestCost_UnknownModel` sigue verde (FR-017).
  > Suite: **355 pass** (335 + 4 tests y 16 subtests), 0 fail.
- [x] **T048** **Revisión SC-011**: una casilla por elemento de la cabecera. Transcribir las siete
  casillas aquí. *(Enmendado 2026-10-02, Q-006-1 resuelta: recalculado a partir de SC-011 enmendado
  —fuente, fecha de verificación, aprobación, versión replicada del catálogo y las dos limitaciones de
  FR-019— son ~~**seis**~~ casillas, no siete.)* *(Enmendado 2026-10-02, coherencia con FR-018: SC-011
  añade el casamiento, así que son **siete** casillas: fuente · verificación · aprobación · catálogo
  replicado · casamiento · limitación 1 (caché de 5 minutos) · limitación 2 (modo rápido). El «seis»
  de la nota anterior queda corregido.)*
  > **Revisión, 2026-10-02**: las siete casillas, contra la cabecera de `Table` en `pricing.go`
  > (líneas del 2026-10-02):
  > - ☑ **fuente**, l. 20: `Fuente: https://platform.claude.com/docs/en/about-claude/pricing`;
  > - ☑ **verificación**, l. 21: `2026-08-07 (catálogo); filas claude-opus-5-5 y claude-sonnet-5, 2026-10-02`;
  > - ☑ **aprobación**, l. 22: `Basilio, 2026-08-07 (catálogo); filas … 2026-10-02`;
  > - ☑ **catálogo replicado**, l. 23: `permea-dev/permea-platform · backend/config/pricing.php · e50d0a5`;
  > - ☑ **casamiento**, l. 24–26: «exacto … no se normalizan sufijos de fecha, prefijos ni mayúsculas
  >   (P-006 FR-018)»;
  > - ☑ **limitación 1**, l. 27–28: escritura de caché a 5 minutos, la de 1 hora quedaría infravalorada;
  > - ☑ **limitación 2**, l. 29: «modo rápido» no distinguido, quedaría infravalorado.
- [x] **T049** **Mutaciones m11–m13**:
  - **m11**: `claude-opus-5-5` `CacheRead` 0.20 → 0.21. Censo: (16). Co-caída: (19), porque usa las
    cuatro partidas.
  - **m12**: añadir una clave `claude-inventado`. Censo: (15) y (17).
  - **m13**: quitar `claude-haiku-3-5`, que no usa ningún otro test. Censo: (15) y (16).
  > **Censos DECLARADOS el 2026-10-02, antes de mutar**, nombrando subtests. Un test con un subtest
  > caído cae también como padre. Todo lo no nombrado, verde.
  > - **m11** · `claude-opus-5-5` `CacheRead` 0.20 → 0.21. Cae `TestEspejo_CifrasClaveAClave/claude-opus-5-5`.
  >   **⚠️ Discrepancia con el enunciado: (19) NO caerá.** Con 0,21, (19) da 1,20 + 0,80 + 0,50 + 0,315
  >   = 2,815 USD, un 0,54 % sobre 2,80, **dentro** de la tolerancia de ±1 % que (19) copia de
  >   `TestCost`, tal como mandaba el encargo. Se declara lo que se prevé y no la co-caída del
  >   enunciado, que la tolerancia hace imposible.
  > - **m12** · añadir `claude-inventado`. Caen `TestEspejo_RecuentoDeClaves` y
  >   `TestEspejo_NingunaClaveSobra`.
  > - **m13** · quitar `claude-haiku-3-5`. Caen `TestEspejo_RecuentoDeClaves` y
  >   `TestEspejo_CifrasClaveAClave/claude-haiku-3-5`.
  >
  > **Dos más, por la disciplina 3**: (16)/`claude-sonnet-4-6` y (16)/`claude-haiku-4-5` nacieron
  > verdes (ya estaban en la tabla con sus cifras) y ninguna de m11–m13 los tumba.
  > - **m-s46** · `claude-sonnet-4-6` `Input` 3.00 → 3.01. Cae `TestEspejo_CifrasClaveAClave/claude-sonnet-4-6`.
  > - **m-h45** · `claude-haiku-4-5` `Input` 1.00 → 1.01. Cae `TestEspejo_CifrasClaveAClave/claude-haiku-4-5`.
  >
  > **Ejecutadas las 5.** Cada una cayó **exactamente** según el censo declarado y se revirtió por
  > edición inversa (md5 de `pricing.go` `c8239bfe4338458dd1df149bb51135b7`, idéntico):
  > - **m11**: (16)/`claude-opus-5-5` («lectura de caché = 0.21, want 0.2»). **(19) quedó verde, como se
  >   había previsto.**
  > - **m12**: (15) («17 claves») y (17) («"claude-inventado" está en la tabla y no en el catálogo»).
  > - **m13**: (15) («15 claves») y (16)/`claude-haiku-3-5` («falta la clave»). Se expresó como sustituir
  >   las líneas de haiku-4-5 y haiku-3-5 por la de haiku-4-5, para que la reversión fuera unívoca.
  > - **m-s46**: (16)/`claude-sonnet-4-6` («entrada = 3.01, want 3»).
  > - **m-h45**: (16)/`claude-haiku-4-5` («entrada = 1.01, want 1»).
  >
  > **Remate de B3 (2026-10-02) · censos DECLARADOS antes de mutar**, con el (19) nuevo. Previsión del
  > orquestador contrastada con análisis propio: `Cost` sólo lo llama `FromClaudeCodeLine`. Los tests
  > que miran una cifra de coste son (18), (19), `TestBoundary_CostAvailable` (coste > 0) y
  > `TestBoundary_KeepsMetrics` (coste > 0). Valores calculados con la misma aritmética `float64`.
  > **El censo propio coincide con el del orquestador en las tres.**
  > - **m11 (repetida)** · `claude-opus-5-5` `CacheRead` 0.20 → 0.21. Caen
  >   `TestEspejo_CifrasClaveAClave/claude-opus-5-5` **y `TestCost_Opus55AMano`** (daría 1,08745013, a
  >   9,9e-3 > 1e-9). Censo restaurado: la co-caída (19) del enunciado vuelve a ser posible.
  > - **m-cruce** · en `Cost`, cruzar las tarifas de caché (`cacheCreate` con `r.CacheRead`, `cacheRead` con
  >   `r.CacheWrite`). Cae sólo `TestCost_Opus55AMano` (daría 5,5990488). (18) queda verde porque sus
  >   cuatro partidas llevan los mismos tokens, y los `TestBoundary_*` siguen con coste > 0.
  > - **m-centimos** · en `Cost`, redondear a céntimos (`math.Round(x*100)/100`). Cae sólo
  >   `TestCost_Opus55AMano` (daría 1,08). (18) da 36,75 exacto, y los `TestBoundary_*` dan 0,03 > 0:
  >   verdes.
  >
  > **Ejecutadas (remate de B3).** Las tres cayeron **exactamente** según el censo declarado, y
  > `pricing.go` volvió a `c8239bfe4338458dd1df149bb51135b7` tras cada una; `vet` limpio con todas:
  > - **m11**: (16)/`claude-opus-5-5` («lectura de caché = 0.21, want 0.2») y (19) («coste = 1.0874501300,
  >   want 1.0775736000 (diferencia absoluta 0.00988 > 1e-9)»);
  > - **m-cruce**: sólo (19) («coste = 5.5990488000 …»). (18) verde;
  > - **m-centimos**: sólo (19) («coste = 1.0800000000 … (diferencia absoluta 0.00243 > 1e-9)»). (18) y los
  >   `TestBoundary_*` verdes.
  >
  > *(Enmendado 2026-10-02: renombradas; m14 y m15 son de B4. Las dos mutaciones de `Cost` de este remate
  > se llamaban «m14» y «m15»; ahora son **m-cruce** (tarifas de caché cruzadas) y **m-centimos** (redondeo
  > a céntimos).)*
- [x] **T050** Puertas del bloque.
  > **Puertas, 2026-10-02**:
  > - `gofmt` vacío;
  > - `vet` limpio;
  > - lint **0**;
  > - `go test -count=1 ./...` **9/9 ok, 355 pass, 0 fail**;
  > - `internal/event` sin diff;
  > - Windows y darwin compilan;
  > - `make run` sale con 0. El evento de `claude-opus-4-6` del fixture cuesta ahora $0.0304, antes
  >   $0.0911;
  > - `event.NewID` sin llamantes;
  > - disciplina 8, sin resultados.
- [ ] **T051** ✋ **Commit** (dueño): `006 B3: tarifas espejo del catalogo de la plataforma (16 claves)`

---

## Bloque B4 · CLI (`contracts/cli.md`, `plan.md` D-006-P8, E-006-P3)

> Todos los rojos son **de proceso**, contra el binario de prueba que ya compila `TestMain`, así que
> caen por conducta sin necesidad de andamiaje. **Los canales se capturan por separado** (disciplina
> 7).
>
> Para los casos enrolados contra un servidor de prueba, el hijo confía en el certificado por
> `SSL_CERT_FILE`, con el montaje que ya usa `entornoDeAdhesion` en `cmd/permea/project_test.go`, y se
> cuentan las peticiones con `bancoDeAdhesion.recibidas`. **Sin confianza TLS el servidor nunca vería
> la petición y el 0 sería vacuo.**

- [x] **T052** [P] **Rojo** (20) `TestAyuda_LasCuatroFormasSonIdenticas` (`cmd/permea/ayuda_test.go`):
  sin argumentos, `help`, `-h` y `--help` → stdout no vacío e **idéntico byte a byte**, stderr vacío,
  exit 0. **Cae**: `-h` sale por stderr con el uso de Go, y `help` lleva banner por stderr.
  > **Paso 0b (2026-10-02), antes de escribir nada.** `grep` en `*_test.go` de ayuda/uso por stderr,
  > `printUsage`, banner, exit de la invocación sin argumentos / de un argumento desconocido / de una
  > opción desconocida, y `runStatus`. El único test existente que cambia es la llamada a `runStatus`
  > de `status_test.go` (2 sitios), que ya preveía T061. Ningún otro fija la conducta vieja.
  >
  > **🔴 Medido** (5 formas: el contrato añade `-help` a las cuatro de la spec). `-h`/`--help`/`-help`
  > (`ayuda_test.go:90`): «stderr debe estar VACÍO; trae "Usage of …permea-test:\n  -daemon…"», y
  > `:93` «stdout vacío». `help` y `sin_argumentos`: stderr con «Permea 0.0.1-dev\nuso: permea
  > <subcomando | flag>…». `identicas` (`:100`): «la ayuda de `help` está vacía».
- [x] **T053** [P] **Rojo** (21) `TestAyuda_ContenidoMinimo`, sólo sobre la forma `help`:
  - `enroll`, `status`, `project join`, `--scan`, `--run`, `--daemon` y `--version`;
  - la recomendación de stdin, en los dos sitios.

  **Cae**: hoy sale por stderr y stdout está vacío.
  > **🔴 Medido**: los 9 subtests (`enroll`, `status`, `project_join`, las 4 opciones y las 2 vías
  > stdin) caen por «la ayuda (stdout de `permea help`) no contiene …» (`:131`): hoy sale por stderr.
- [x] **T054** [P] **Rojo** (22) `TestAyudaSubcomando_NoHaceNada`: las 8 invocaciones de
  `contracts/cli.md` §Las ayudas de subcomando, con exit 0 y la ayuda por stdout. En dos montajes, y
  cada aserción con `t.Errorf`:
  - **(a) sandbox vacío**: el árbol queda idéntico antes y después;
  - **(b) enrolado**, desde un árbol con raíz: el banco recibe **0** peticiones y el árbol no cambia.

  **Cae**: `status -h` crea el directorio de datos y `project join -h` emite.
  > **🔴 Medido, 16/16**, tras una corrección previa al verde. **Primera versión ciega para `status`**:
  > `testutil.Sandbox` crea el directorio de datos por adelantado y `status` imprime su estado por
  > stdout con exit 0, así que `vacio/status_*` y `enrolado/status_*` pasaban. Corregido antes del
  > verde: en (a) se borra el directorio vacío que deja el sandbox, y se exige que stdout empiece por
  > `uso: permea <subcomando>`. Razones medidas:
  > - `vacio/status_-h`: «stdout debe ser la ayuda … trae "no enrolado"» y «APARECIÓ .config/permea»;
  > - `vacio/enroll_-h`: «ExitCode() = 1» y «enrollment string inválido: prefijo no reconocido»;
  > - `vacio/project_-h`: «verbo desconocido "-h"»;
  > - `vacio/project_join_-h`: «no está enrolada» y «APARECIÓ .config/permea»;
  > - `enrolado/project_join_-h`: «trae "unido al Proyecto \"RecetApp\""», «APARECIÓ
  >   .config/permea/salt» y **«emitió 1 peticiones al banco»**. Era el defecto real: pedir ayuda emitía.
  >
  > Confianza TLS por `SSL_CERT_FILE` (montaje de `entornoDeAdhesion`); m16 demuestra que el
  > contador ve la petición.
- [x] **T055** [P] **Rojo** (23) `TestDesconocido_SubcomandoNombradoYSalida1`: `permea enrol` → exit
  1, stderr contiene `enrol`, stdout vacío. **Cae**: hoy exit 0.
  > **🔴 Medido**: `:204` «ExitCode() = 0, se esperaba 1» y `:207` «stderr debe nombrar el subcomando
  > desconocido ("enrol"); trae "Permea 0.0.1-dev\nuso: …"».
- [x] **T056** [P] **Rojo** (24) `TestDesconocido_NoReproduceSecretos`: `pmea2.<centinela>`,
  `pmeaj1.<centinela>` y `pmea1.<centinela>` → exit 1, y el centinela no aparece en ninguno de los dos
  canales. **Cae** por el exit.
  > **🔴 Medido**: los 3 prefijos, `:223` «ExitCode() = 0, se esperaba 1». El centinela no se reproducía
  > ya (la ayuda no hace eco); lo acredita m17.
- [x] **T057** [P] (25) `TestTokenCentinela_NuncaSale`: un `config.json` de prueba con un token
  centinela y todas las invocaciones de `contracts/cli.md` (las cuatro generales, las ocho de
  subcomando, `status`, subcomando inexistente y opción desconocida). El centinela no aparece en
  ningún canal. **Nace verde** (hoy no se imprime) → m18.
  > **Nace verde**: 17 subtests (5 ayudas generales, `status`, `enrol`, `--bogus`, `--scan` y las 8 de
  > subcomando). Lo acreditan m18, m-token-ayuda, m-token-errores y m-token-status.
- [x] **T058** [P] **Rojo** (26) `TestOpcionDesconocida_StderrYSalida2` (E-006-P3): `--bogus`, y
  `--scan` sin valor. Exit 2, stdout **vacío**, y stderr nombra la opción y contiene `permea help`.
  **Cae** porque Go no remite a `permea help`. Transcribir también si hoy añade el uso por defecto.
  > **🔴 Medido**: `desconocida` y `sin_valor` caen sólo por `:281` «stderr debe remitir a `permea help`;
  > trae "flag provided but not defined: -bogus\nUsage of …"» (y «flag needs an argument: -scan …»). Hoy
  > Go **sí añade su uso por defecto, por stderr**. stdout ya estaba vacío y el exit ya era 2.
- [x] **T059** Nuevo `cmd/permea/ayuda.go`: **la fuente única**, una tabla de subcomandos (nombre,
  sinopsis, líneas) y de opciones. Compone la ayuda general y la de cada subcomando. El literal de
  `printUsage` en `main.go` se sustituye por ella; su comentario de cabecera se traslada **por nombre**.
  > **Hecho.** `cmd/permea/ayuda.go`: tablas `filasDeSubcomandos` y `filasDeOpciones`; la ayuda general
  > y la de cada subcomando se componen de ellas, con la misma sangría (fragmento literal). El literal
  > `printUsage` se retira de `main.go`; su razón se traslada por nombre a la cabecera de `ayuda.go`.
- [x] **T060** `cmd/permea/main.go`:
  - **escalera**: sin argumentos, `help`, `-h`, `--help` y `-help` → ayuda por stdout, exit 0, **antes**
    del parseo y del banner;
  - **subcomando inexistente**: primer argumento que no empieza por `-` y no es conocido → error por
    stderr que lo nombra, salvo los prefijos `pmea2.`, `pmeaj1.` y `pmea1.`, con exit 1;
  - **parseo** con un `FlagSet` de `ContinueOnError` y salida descartada: `flag.ErrHelp` → ayuda por
    stdout, exit 0; otro error → mensaje propio por stderr con la opción y `permea help`, exit 2
    (`research.md` R8, enmienda E-006-P3).
  > **Hecho.**
  > - Escalera: sin argumentos y `help`/`-h`/`--help`/`-help` → ayuda por stdout, exit 0, antes del
  >   parseo y del banner.
  > - Subcomando inexistente → exit 1 por stderr; con forma de secreto, sin reproducirlo.
  > - `FlagSet` con `ContinueOnError` y salida descartada: `ErrHelp` → ayuda; otro error → mensaje
  >   propio («opción desconocida: -bogus» / «la opción -scan necesita un valor») más «Ayuda: permea
  >   help», exit 2. **Ese mensaje tampoco reproduce un secreto** (FR-024).
  > - Rama `default` (opciones válidas sin modo): la ayuda por stderr, exit 0, como siempre.
- [x] **T061** Ayudas de subcomando, en la **primera** línea de cada camino:
  - `runEnroll` (`enroll.go`), antes de inspeccionar stdin;
  - `runStatus` (`status.go`), que pasa a recibir los argumentos (se ajusta su llamada en
    `status_test.go`), antes de `config.DataDir()`;
  - `runProject` para `project -h` y `projectJoin` para `project join -h` (`project.go`).

  **Verde**: (20)–(26). `TestProject_ErroresDeUsoDeLaGramatica` sigue verde sin cambios.
  > **Verde**: (20)–(26) verdes. `TestProject_ErroresDeUsoDeLaGramatica` verde sin tocar.
  > `status_test.go`: `runStatus(nil, &out)` en los 2 sitios. Suite **415 pass** (355 + 7 tests y 53
  > subtests). Comprobado a mano en un HOME temporal (sin enrolar, `logs_root` vacío) que no cambian
  > `--version`/`-version`, `--scan`/`-scan`, `--run`/`-run` («sync omitido») ni `--daemon`/`-daemon`
  > (arranca su ciclo y lo corta `timeout`).
- [x] **T062** **Mutaciones m14–m19**:
  - **m14**: la invocación sin argumentos escribe por stderr. Censo: (20). (21) no cae, porque mira
    sólo `help`.
  - **m15**: en `runStatus`, la comprobación de `-h` detrás de `DataDir()`. Censo: (22) (a).
  - **m16**: quitar el `-h` de `projectJoin`. Censo: (22) (b), por peticiones.
  - **m17**: reproducir siempre lo tecleado en el error de subcomando. Censo: (24).
  - **m18**: la ayuda de `status` carga la configuración e imprime `DeviceToken`. Censo: (25).
    Co-caída: (22) (a), porque cargar la configuración crea el directorio en el sandbox vacío.
  - **m19**: ante una opción desconocida, escribir también la ayuda por stdout. Censo: (26).
  > **Censos DECLARADOS el 2026-10-02, antes de mutar**, nombrando subtests (padres implícitos). Análisis
  > propio sobre el código verde. Todo lo no nombrado, verde.
  > - **m14** · sin argumentos → ayuda por **stderr**. Caen `TestAyuda_LasCuatroFormasSonIdenticas/sin_argumentos`
  >   y `…/identicas`. (21) no cae porque mira `help`, y (25)/`sin_argumentos` tampoco: no hay token.
  > - **m15** · en `runStatus`, `DataDir()` antes de atender `-h`. Caen
  >   `TestAyudaSubcomando_NoHaceNada/vacio/status_-h` y `…/vacio/status_--help`, por «APARECIÓ
  >   .config/permea». `enrolado/status_*` no cae: el directorio ya existe.
  > - **m16** · `projectJoin` sin su bloque de ayuda. Caen (22)/`vacio/project_join_-h` y
  >   `…/vacio/project_join_--help` (no enrolado → exit 1, sin cabecera), y (22)/`enrolado/project_join_-h`
  >   y `…/enrolado/project_join_--help` (**emite al banco**, crea el `salt` y responde «unido»). (25) no
  >   cae: su endpoint es inalcanzable y no imprime el token.
  > - **m17** · el error de subcomando reproduce siempre lo tecleado. Caen
  >   `TestDesconocido_NoReproduceSecretos/pmea2.`, `…/pmeaj1.` y `…/pmea1.`.
  > - **m18** · la ayuda de `status`, tras escribirse, carga la configuración e imprime `DeviceToken`.
  >   Caen `TestTokenCentinela_NuncaSale/status_-h` y `…/status_--help`, y la co-caída
  >   (22)/`vacio/status_-h` y `…/vacio/status_--help` (cargar la configuración crea el directorio).
  >   `enrolado/status_*` no cae: la cabecera va primero y el árbol no cambia.
  > - **m19** · ante una opción desconocida o sin valor, escribir también la ayuda por stdout. Caen
  >   `TestOpcionDesconocida_StderrYSalida2/desconocida` y `…/sin_valor`.
  >
  > **Tres más, por la disciplina 3**: los 17 subtests de (25) nacieron verdes, y m18 sólo tumba dos.
  > Cada una inyecta el volcado de `config.json` en un camino de salida distinto:
  > - **m-token-ayuda** · `escribirAyudaGeneral` y `escribirAyudaDe` vuelcan `config.json` al final.
  >   Caen (25)/`sin_argumentos`, `help`, `-h`, `--help`, `-help`, `enroll_-h`, `enroll_--help`,
  >   `status_-h`, `status_--help`, `project_-h`, `project_--help`, `project_join_-h` y
  >   `project_join_--help` (13). En (20), (21) y (22) no hay `config.json` en el directorio de datos, o
  >   el volcado va tras la cabecera: verdes.
  > - **m-token-errores** · los errores de subcomando desconocido y de opción inválida vuelcan
  >   `config.json` por stderr. Caen (25)/`enrol`, `--bogus` y `--scan`. (23) y (26) corren sin
  >   `config.json`: verdes.
  > - **m-token-status** · `status` enrolado imprime `DeviceToken`. Caen (25)/`status` y la co-caída
  >   `TestStatus_Enrolled_ShowsURLNotToken` (`status_test.go`).
  >
  > **Ejecutadas las 9.** Cada una cayó **exactamente** según su censo y se revirtió por edición
  > inversa con md5 idéntico (`main.go` `e699fa48…`, `status.go` `2d77da79…`, `project.go` `326ce28e…`,
  > `ayuda.go` `b6386600…`); `vet` limpio con todas:
  > - **m14**: (20)/`sin_argumentos` y `identicas`;
  > - **m15**: (22)/`vacio/status_-h` y `--help` («APARECIÓ .config/permea»);
  > - **m16**: (22)/`vacio/project_join_*` (exit 1) y `enrolado/project_join_*` («unido», «APARECIÓ
  >   …/salt», **«emitió 1 peticiones al banco»**: el 0 no es vacuo);
  > - **m17**: (24) ×3;
  > - **m18**: (25)/`status_-h` y `--help`, más (22)/`vacio/status_*`;
  > - **m19**: (26) ×2 («stdout debe estar VACÍO»);
  > - **m-token-ayuda**: (25) ×13;
  > - **m-token-errores**: (25)/`enrol`, `--bogus` y `--scan`;
  > - **m-token-status**: (25)/`status` y `TestStatus_Enrolled_ShowsURLNotToken`.
- [x] **T091** *(nueva, 2026-10-02, remate de B4, D-006-14)* **Rojos (27)–(29)** en
  `cmd/permea/ayuda_test.go`, sobre el texto de ayuda aprobado por el dueño:
  - (27) `TestAyuda_AnchoMaximo80`, un subtest por ayuda;
  - (28) `TestAyuda_PrimerosPasosYAviso`, con los subtests `bloque`, `orden` y `aviso`;
  - (29) `TestAyuda_SinJergaInterna`, un subtest por ayuda.
  > **Paso 0, antes de tocar nada**: ninguna aserción de (20)–(26) depende del texto viejo.
  > - (21) busca `enroll`, `status`, `project join`, las 4 opciones, `| permea enroll -` y
  >   `| permea project join -`, y el texto aprobado los trae todos.
  > - (22) exige `uso: permea <subcomando>`, que se conserva.
  > - (20) y (23)–(26) no miran el texto.
  >
  > Ningún test anterior a B4 se toca.
  >
  > **🔴 Medido el 2026-10-02**:
  > - (27): los 5 subtests caen (`ayuda_test.go:321`); el máximo es 105/97/105/100/100 caracteres
  >   para general/enroll/status/project/project join.
  > - (28): `bloque` (`:344`, «no contiene «Primeros pasos:»»), `orden` (`:350`, posiciones −1) y
  >   `aviso` (`:355`).
  > - (29): `general` cae («contiene jerga interna "P-001"», «"P-002"», «"sync_interval"»); `enroll`,
  >   `status`, `project` y `project_join` **nacen verdes** → m-jerga.
  >
  > **Censo DECLARADO antes de mutar**: **m-jerga** · `escribirFila` añade « (P-001)» a la línea de la
  > sinopsis. Caen `TestAyuda_SinJergaInterna/general`, `…/enroll`, `…/status`, `…/project` y
  > `…/project_join`. Nada más: las sinopsis más « (P-001)» siguen por debajo de 80, la cabecera `uso:`
  > y «Primeros pasos» no cambian, y las cinco formas siguen idénticas.
  >
  > **Ejecutada**: cayeron **exactamente** los 5 subtests declarados (`ayuda_test.go:370`, «la ayuda
  > contiene jerga interna "P-001"»). Revertida por edición inversa: md5 de `ayuda.go`
  > `8a484a9b13cdda45704cc0db10763fc8`, idéntico.
- [x] **T092** *(nueva, 2026-10-02, remate de B4, D-006-14)* **Verde**: `cmd/permea/ayuda.go` con el
  texto aprobado y la forma nueva (sinopsis sola en su línea, con sangría 2; descripción con 6;
  ejemplos con 10). La general y las de subcomando salen de la misma tabla, con la misma
  `escribirFila`. Comparar `permea help` byte a byte con el texto aprobado.
  > **Verde medido**: `cmp` de `permea help` (HOME temporal) con el texto aprobado, **idéntico byte a
  > byte** (41 líneas, termina en salto de línea, máximo 75 caracteres), con stderr vacío y rc 0.
  > (20)–(29) verdes; suite **431 pass** (415 + 3 tests y 13 subtests).
  >
  > **⚠️ Ajuste de (28)/`aviso`**: en el texto aprobado «todo el historial» cruza un salto de línea
  > («…La primera vez envía todo» / «el historial que conserve…»), y la búsqueda literal no podía
  > casar nunca. La aserción compara con los espacios normalizados (`strings.Fields`). El texto
  > aprobado no se tocó.
- [x] **T063** Puertas del bloque.
  > **Puertas, 2026-10-02**:
  > - `gofmt` vacío;
  > - `vet` limpio;
  > - lint **0**;
  > - `go test -count=1 ./...` **9/9 ok, 415 pass, 0 fail**;
  > - `internal/event` sin diff;
  > - Windows y darwin compilan;
  > - `make run` rc 0;
  > - `event.NewID` sin llamantes;
  > - disciplina 8, sin resultados.
  >
  > **Repetidas tras el remate de B4 (D-006-14), 2026-10-02**:
  > - `gofmt` vacío;
  > - `vet` limpio;
  > - lint **0**;
  > - `go test -count=1 ./...` **9/9 ok, 431 pass, 0 fail**;
  > - `internal/event` sin diff;
  > - Windows y darwin compilan;
  > - `make run` rc 0;
  > - `event.NewID` sin llamantes;
  > - disciplina 8, sin resultados;
  > - `permea help` idéntica byte a byte al texto aprobado, y las 4 de subcomando, fragmento literal de
  >   la general.
- [ ] **T064** ✋ **Commit** (dueño):
  `006 B4: ayuda unica por stdout, ayudas de subcomando sin efectos y errores de uso`

---

## Bloque B5 · README, CHANGELOG y comentarios (FR-025 a FR-029)

- [x] **T065** **Rojo** (V18): transcribir `grep -c bfgnet README.md .goreleaser.yaml .github/workflows/release.yml`
  (> 0) y la ausencia de `CHANGELOG.md`.
  > **Hecho el 2026-10-02.** `README.md:2`, `.goreleaser.yaml:1`, `.github/workflows/release.yml:1`
  > (> 0 en los tres); `CHANGELOG.md` no existía. Rojo de V18 transcrito.
- [x] **T066** `README.md`:
  - **instalación**: `brew install --cask permea-dev/permea/permea`,
    `scoop bucket add permea https://github.com/permea-dev/scoop-permea` e `install.sh` desde
    `permea-dev/agent`. Comprobar cada uno con `gh api` **antes** de escribirlo;
  - **primeros pasos para quien instala**: `echo "$ENROLL" | permea enroll -` → `permea status` →
    `permea --run` o `--daemon`. **Antes** del tercer paso, el aviso de que la primera pasada envía
    todo el historial que conserve Claude Code (D-006-10);
  - los pasos de desarrollo (`make test`…) aparte;
  - **fuera** la mención del «modo de ref» en §Configuración y rutas por SO;
  - el límite de casamiento exacto de tarifas (FR-018);
  - el formato nuevo de `--scan` y la ayuda.
  > **Hecho el 2026-10-02.** Cada dirección se comprobó **antes** de escribirla, con `curl` anónimo
  > (sin credenciales, por orden del encargo, en lugar de `gh api`): los dos repositorios, el cask, el
  > manifiesto de Scoop, `install.sh` y la página de releases, **200** los seis. Primeros pasos coherentes
  > con `permea help`: enroll en tres formas (pegar el comando de la aplicación; `echo "$ENROLL" | permea
  > enroll -`; PowerShell con `Get-Clipboard | permea enroll -`, admisible porque `readEnrollmentInput`
  > recorta con `strings.TrimSpace` y por tanto quita `\r\n` —el test `TestEnroll_Stdin_And_SC011`
  > cubre `\n`, no `\r\n`—), `permea status`, y el aviso de D-006-10 **antes** de `permea --run` /
  > `--daemon`. Fuera «modo de ref»; casamiento exacto (FR-018), las dos limitaciones (caché de 1 h,
  > modo rápido) y el coste en USD; `--scan` y la ayuda; §Desarrollo aparte. `bfgnet`: 0.
- [x] **T067** Nuevo `CHANGELOG.md` con la entrada `0.3.0`:
  - lo nuevo de 003, 004, 005 y 006;
  - la ruptura aceptada de `project_ref` (004);
  - que **las cifras bajan porque antes se contaba de más** (×2,13 en los datos medidos);
  - la ayuda sin argumentos pasa de stderr a stdout (D-006-7);
  - un subcomando inexistente ahora sale con 1, y una opción desconocida ya no imprime el uso.
  > **Hecho el 2026-10-02.** `## 0.3.0 — PENDIENTE`, con §Nuevo (003, 004, 005 y 006, cada punto con
  > su cita de spec), §Las cifras bajan (×2,13) y §Cambios que rompen (`project_ref`, modo en claro
  > retirado, canal y forma de la ayuda: sin argumentos por stdout, subcomando inexistente exit 1,
  > opción desconocida exit 2 sin uso). Añadida en el tramo C5 la comprobación
  > `grep -c PENDIENTE CHANGELOG.md` → 0 antes de fusionar.
- [x] **T068** [P] Comentarios con `bfgnet/…` en `.github/workflows/release.yml` y `.goreleaser.yaml`,
  corregidos **sin tocar ninguna línea de configuración** (FR-029).
  > **Hecho el 2026-10-02.** Una línea de comentario en cada fichero (`bfgnet/` → `permea-dev/`). Diff
  > sin comentarios: `git diff -U0 … | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | grep -vE '^[+-]\s*#'`
  > → **0** líneas. `goreleaser check` (v2.16.0): 1 fichero validado.
- [x] **T069** **Verde** (V18):
  - los dos repositorios responden y `install.sh` da 200;
  - `bfgnet` 0 en los tres ficheros;
  - el diff de configuración (sin comentarios) vacío;
  - las cuatro partes del CHANGELOG (SC-017) y el aviso del README antes de `-run` (SC-016).
  > **Hecho el 2026-10-02.** Seis direcciones, **200**; `bfgnet` 0/0/0; diff de configuración sin
  > comentarios, 0 líneas; CHANGELOG con lo nuevo de 003/004/005/006, la ruptura de `project_ref`,
  > «las cifras bajan… ×2,13» y el cambio de canal de la ayuda (SC-017); en §Primeros pasos del README
  > el aviso precede a `permea --run`, y el orden es enroll → status → `--run` (SC-016).
- [x] **T093** *(Añadida el 2026-10-02, remate de B5.)* Restos del README, dos frases del CHANGELOG y un
  test de CRLF:
  1. `README.md`:
     - el ejemplo `PERMEA_VERSION` pasa a `v0.3.0`;
     - nuevo «Actualizar», una línea por canal: `brew upgrade --cask permea`, `scoop update` +
       `scoop update permea`, y relanzar `install.sh`;
     - §Comandos en el orden de `permea help` (`enroll`, `status`, `project join`), en la lista y en
       las subsecciones;
     - fuera `(US1 + US2)` y la línea de renombrar el módulo;
     - barrido de códigos internos.
  2. `CHANGELOG.md`: fuera «el producto no estaba en producción»; en «Las cifras bajan», lo enviado
     no se corrige ni se reenvía (FR-009).
  3. Subtest `TestEnroll_Stdin_And_SC011/crlf_de_PowerShell,_igual_que_lf`: el *enrollment string* de
     prueba seguido de `\r\n` da el mismo resultado que con `\n`. Nace verde.

  **Mutación declarada antes de mutar — m-crlf**: en `readEnrollmentInput` (`cmd/permea/enroll.go`),
  las dos apariciones de `strings.TrimSpace(x)` pasan a `strings.TrimSuffix(x, "\n")`.

  **Censo**: cae **sólo** `TestEnroll_Stdin_And_SC011/crlf_de_PowerShell,_igual_que_lf`, y su test
  padre `TestEnroll_Stdin_And_SC011` por arrastre. Cae por su **primera aserción**: `readEnrollmentInput`
  devuelve la cadena con `\r` final.

  Análisis propio:
  - **Las aserciones de `enroll()` del caso NO caerían solas.** `base64.RawURLEncoding` ignora
    `\r` y `\n` (comprobado aparte), así que `es+"\r"` se decodifica igual y persiste el mismo
    token. Por eso el caso compara también la salida de `readEnrollmentInput`.
  - **Ningún otro test depende del recorte**:
    - el caso (a) usa un único `\n` final, que `TrimSuffix` también quita;
    - el (b) no lleva blancos;
    - los de argumento (`enroll_test.go`, `enroll_reject_test.go`, `main_test.go`) no llevan blancos;
    - el único subproceso con stdin es `project join`, que tiene su propio lector en `project.go`;
    - las ayudas `enroll -h` no llegan a leer.
  > **Hecho el 2026-10-02.**
  > - **README**: hechos los seis puntos. `install.sh` sin `PERMEA_VERSION` resuelve la última release
  >   y copia el binario encima, así que relanzarlo actualiza; la nota pide repetir el mismo `PREFIX` si
  >   se indicó uno. `go.mod` ya declara `github.com/permea-dev/agent`. El barrido sólo encontró
  >   `(US1 + US2)`, ya fuera.
  > - **CHANGELOG**: hechas las dos frases. FR-009 respalda «no se reenvía», y §Fuera de alcance, «no se
  >   corrige». Se añade que la cola de la 0.2.1 se envía tal cual, que es literal de FR-009.
  > - **m-crlf**: 430 pasan y 2 caen, los dos declarados: el subtest, por su primera aserción
  >   (`len 121, quiero 120`), y su padre por arrastre. Reversión por edición inversa: md5 de `enroll.go`
  >   `5af2d0c2f87ed4fb1c95297fda28e4b5` antes y después, y sin diff. Tests: **432** (431 + 1 subtest).
- [x] **T070** Puertas del bloque.
  > **Hecho el 2026-10-02.** `gofmt -l .` vacío; `go vet` sin hallazgos; `golangci-lint run` 0 issues;
  > `go test -count=1 ./...` **431** pasan, 0 fallan, 9/9 paquetes ok (sin cambios); `internal/event`
  > sin diff; disciplina 8 y `event.NewID` fuera de `internal/event`, vacíos; Windows y darwin compilan;
  > `make run` rc 0 (2 eventos, dry-run); `goreleaser check` rc 0. Ningún `.go` modificado en B5.
  > **Repetidas tras T093 (2026-10-02)**, todas en verde:
  > - `gofmt -l .` vacío, `go vet` sin hallazgos, `golangci-lint run` 0 issues;
  > - `go test -count=1 ./...`: **432** pasan y 0 fallan, 9/9 paquetes;
  > - `internal/event` sin diff; disciplina 8 y `event.NewID`, vacíos;
  > - Windows y darwin compilan; `make run` rc 0; `goreleaser check` rc 0;
  > - T069 repetida: `bfgnet` 0/0/0 y diff de configuración sin comentarios, 0 líneas.
  >
  > Único `.go` tocado: `cmd/permea/enroll_test.go`. Ningún cambio en producción.
- [ ] **T071** ✋ **Commit** (dueño): `006 B5: README instalable, CHANGELOG 0.3.0 y comentarios de distribucion`

---

## Cierre — en tramos, UNO POR MENSAJE

Cada tramo es un encargo propio y no empieza hasta que el anterior está cerrado.
**Si un tramo falla, se para**: se corrige en el bloque que corresponda, con sus puertas y su commit,
y se rehace desde el tramo C1.

### Tramo C1 · Puertas locales

- [x] **T072** Sobre la rama, con todos los bloques commiteados, transcribir aquí la salida de:
  - quickstart V1;
  - `git diff 0311fa1 -- internal/event` vacío;
  - el `grep` de `event.NewID`;
  - el `grep` de la disciplina 8;
  - la cabecera de tarifas cita el commit vigente del catálogo (T043): **`e50d0a5`** *(enmendado
    2026-10-02, Q-006-1 resuelta)*;
  - `grep -rn nolint --include=*.go .` → exactamente una.
  > **Hecho el 2026-10-02 (C1)**, sobre `b504268` (`006 B5: …`), con los doce commits de 006 desde
  > `0311fa1` y el árbol limpio. Go 1.22.2 y golangci-lint 2.12.2.
  > - **V1**:
  >   - `gofmt -l .` vacío;
  >   - `go vet ./...` sin hallazgos;
  >   - `go test -count=1 ./...`: **432 pass, 0 fail, 0 skip**, 9/9 paquetes ok;
  >   - `golangci-lint run`: «0 issues.», sin tope (`max-issues-per-linter: 0`, `max-same-issues: 0`);
  >   - `grep -rn nolint --include=*.go .`: **una**, `internal/transport/adhesion_test.go`, la de
  >     SA1007 (E-006-P4).
  > - `git diff 0311fa1 -- internal/event`: vacío.
  > - `event.NewID` fuera de `internal/event`: vacío.
  > - Disciplina 8: vacío.
  > - Cabecera de `internal/pricing/pricing.go`: «Catálogo replicado: permea-dev/permea-platform ·
  >   backend/config/pricing.php · e50d0a5».
  > - **Además, por encargo de C1**:
  >   - Windows y darwin compilan;
  >   - `make run` rc 0 (2 eventos, dry-run);
  >   - `goreleaser check` rc 0;
  >   - `go mod verify`: «all modules verified»;
  >   - `go mod tidy -diff` **no existe en Go 1.22** («flag provided but not defined: -diff»). En su
  >     lugar, `go mod tidy` sobre una copia (`git archive HEAD`) en el directorio temporal deja
  >     `go.mod` idéntico; no hay `go.sum`, porque sólo se usa la biblioteca estándar;
  >   - `bfgnet` 0 / 0 / 0;
  >   - `grep -c PENDIENTE CHANGELOG.md` → **1**: se resuelve en C5 y no se toca aquí;
  >   - `git diff --stat 0311fa1..HEAD`: **42 files changed, 5545 insertions(+), 141 deletions(-)**.
  > - **Arreglos de documentos de C1**:
  >   - README §Comandos: «Los tres exigen HTTPS» pasa a «`enroll` y `project join` exigen HTTPS…;
  >     `status` no contacta con nadie»;
  >   - §Recuento y tabla de cobertura, con T089–T093 y las 13 mutaciones añadidas.

### Tramo C2 · Copia congelada y medidas V

- [x] **T073** Copia congelada y contador independiente (quickstart §La copia congelada y §El contador
  independiente). Transcribir: fecha, nº de ficheros, facturables, sintéticas, sin identificador,
  mensajes distintos y tokens una vez por mensaje. **Sólo recuentos** (disciplina 9).
  > **Hecho el 2026-10-02 (C2)**. Raíz temporal única `/tmp/permea-006-XXXXXX`, con la copia, los
  > sandboxes, el binario y los resultados dentro.
  > - **Copia congelada** de `~/.claude/projects` tomada el **2026-10-02T23:13:50+02:00** con `cp -a`:
  >   27 ficheros `.jsonl`, 156 MB. Se puso en sólo lectura.
  > - **Binario**: `go build` desde `b504268`. El árbol sólo difería en `README.md` y `tasks.md`, sin
  >   ningún `.go`.
  > - **Contador independiente**: el `python3` del quickstart, ampliado a las cuatro partidas por
  >   separado y sin importar nada del agente.
  >   - **ficheros 27 · facturables 23 778 · sintéticas 3 · sin identificador 0**;
  >   - **mensajes distintos 11 105 · tokens una vez por mensaje 5 040 966 143**: `in` 28 350,
  >     `out` 13 246 462, `cw` 35 114 677, `cr` 4 992 576 654;
  >   - por modelo: `claude-opus-5` 7 318, `claude-opus-5-5` 3 757, `claude-sonnet-5` 30.
  >
  >   Referencia de M2: 10 698 mensajes y 4 835 167 368 tokens. La copia es posterior, y no es el
  >   valor esperado.
- [x] **T074** Medidas sobre la copia, transcritas como recuento frente a valor esperado:
  - V2 (SC-001);
  - V3 (SC-002);
  - V4 a y b (SC-003);
  - V5 (SC-004);
  - V7 (SC-006);
  - V10 (SC-021);
  - V12 (SC-010).
  > **Hecho el 2026-10-02 (C2)**. Cada `--scan` se lanzó con `env -i` dentro de los sandboxes `sb1` y
  > `sb2`, uno por fichero de la copia: 27 + 27 ejecuciones, todas con rc 0. El árbol de cada sandbox
  > quedó igual antes y después: `--scan` no escribe nada. **Las ocho medidas coinciden.**
  >
  > | Medida | Criterio | Esperado | Obtenido | ¿Coincide? |
  > |---|---|---|---|:--:|
  > | V2 | SC-001 | líneas `evento:` = 11 105 | 11 105 en `sb1` y 11 105 en `sb2` | sí |
  > | V2 falsable | SC-001 | duplicar una línea `assistant` no cambia el recuento | 445 → 445; «repetidas» 550 → 551 | sí |
  > | V3 | SC-002 | suma de las cuatro partidas = 5 040 966 143 | 5 040 966 143, con las cuatro partidas iguales una a una | sí |
  > | V4 a | SC-003 | `go test ./internal/ingest -run EventID` en verde | ok, 8 PASS y 0 FAIL | sí |
  > | V4 b | SC-003 | dos sandboxes, mismo conjunto, todo de 32 hex, sin colisiones | 11 105 y 11 105, `cmp` idéntico; 0 que no sean de 32 hex; 11 105 únicos | sí |
  > | V5 | SC-004 | golden en verde, y 0 apariciones de identificadores de la copia en la salida | 5 tests `Boundary` en verde; 44 428 identificadores buscados (enteros y sin prefijo), **0** apariciones | sí |
  > | V7 | SC-006 | `model=<synthetic>` = 0 | 0 (el contador ve 3 sintéticas) | sí |
  > | V10 | SC-021 | `event_id` repetidos = 0 | 0 | sí |
  > | V12 | SC-010 | `cost_avail=false` en los modelos en uso = 0 | 0 | sí |
  >
  > **Resumen por stderr, sumadas las 27 pasadas**: 23 778 facturables = 11 105 eventos + 12 670
  > repetidas + 3 sintéticas; 0 sin identificador; 0 con consumo distinto de la primera.
  >
  > **V12, por modelo con fila**: los tres al **100 %** de `cost_avail=true`: `claude-opus-5`
  > 7 318 / 7 318, `claude-opus-5-5` 3 757 / 3 757 y `claude-sonnet-5` 30 / 30. No apareció ningún modelo
  > sin fila. Las tarifas se contrastaron con `e50d0a5`.
  >
  > Evento de referencia de cada modelo: el primero con las cuatro partidas distintas de cero.
  > Fórmula: `(in·I + out·O + cw·CW + cr·CR) / 1e6`.
  > - `claude-opus-5-5` (4 / 20 / 5 / 0,20), con `in` 2, `out` 213, `cw` 15 507 y `cr` 25 147:
  >   0,000008 + 0,004260 + 0,077535 + 0,005029 = 0,086832 → **$0,0868**. El agente da $0,0868.
  > - `claude-opus-5` (5 / 25 / 6,25 / 0,50), con `in` 2, `out` 183, `cw` 14 071 y `cr` 24 533:
  >   0,000010 + 0,004575 + 0,087944 + 0,012266 = 0,104795 → **$0,1048**. El agente da $0,1048.
  > - `claude-sonnet-5` (2 / 10 / 2,50 / 0,20), con `in` 2, `out` 91, `cw` 114 971 y `cr` 25 215:
  >   0,000004 + 0,000910 + 0,287428 + 0,005043 = 0,293384 → **$0,2934**. El agente da $0,2934.
- [x] **T075** **V4-c** (E-006-P2): `--run` **sólo** en los dos sandboxes `env -i`, **sin enrolar** y
  sobre la copia. Primero `permea status` → «no enrolado» en cada uno; si no, **se para**.
  - **Esperado**: «sync omitido» en los dos, sales distintas, el **mismo** conjunto de `event_id` y
    tantos como mensajes.
  - Medir también la memoria máxima de cada `--run` con `/usr/bin/time -v` y contrastarla con T032.
  > **Hecho el 2026-10-02 (C2)**. Sandboxes `sb1` y `sb2`, lanzados con `env -i` y sólo `HOME`,
  > `XDG_CONFIG_HOME` (dentro del sandbox), `PATH=/usr/bin:/bin`, `HTTPS_PROXY=http://127.0.0.1:9` y
  > `HTTP_PROXY=http://127.0.0.1:9`. El `config.json` lleva sólo `logs_root` apuntando a la copia: sin
  > endpoint y sin token.
  > 1. `permea status` → «no enrolado» exacto en los dos, con rc 0 y stderr vacío.
  > 2. `permea --run` con `/usr/bin/time -v`: rc 0 en los dos. stderr: el banner, «N eventos
  >    encolados», el resumen de pasada y «sync omitido: sin endpoint configurado». Ningún hex de 32 en
  >    stderr.
  > 3. Resultados:
  >    - **«sync omitido» en los dos**;
  >    - **las dos sales son distintas**: 2 huellas sha256 distintas;
  >    - **mismo conjunto de `event_id`**: 11 105 y 11 105, con la misma huella y `cmp` idéntico;
  >    - **tantos como mensajes distintos** (11 105), todos únicos;
  >    - el conjunto de la cola es además idéntico al de `--scan` (V4 b);
  >    - resumen de pasada: 23 778 facturables, 11 105 eventos, 12 670 repetidas, 3 sintéticas, 0 sin
  >      identificador y 0 con consumo distinto.
  >
  > **Memoria máxima (RSS)**: **13 692 kB** en `sb1` y **13 224 kB** en `sb2`, con 1,05 s y 0,96 s de
  > reloj.
  >
  > Contraste con T032. T032 mide *heap* de la pasada (≈ 0,90 MB para 10 698 mensajes, unos 84 B por
  > mensaje); aquí se mide RSS del proceso. La RSS base del mismo binario con `--version` es de unos
  > **6,5 MB** (6 524 / 6 460 / 6 524 kB). El incremento, unos 7 MB por leer 156 MB de logs y escribir
  > 11 105 eventos, es del mismo orden de magnitud bajo y no contradice T032.
- [x] **T076** Borrar la copia, los dos sandboxes y los temporales de `/tmp` de las medidas. Anotarlo.
  > **Hecho el 2026-10-02 (C2)**.
  > - Se comprobó que la raíz empezaba por `/tmp/permea-006-`. Se devolvió el permiso de escritura a la
  >   copia y se borró la raíz entera con `rm -rf`, rc 0: copia, `sb1`, `sb2`, binario y resultados.
  >   Nada más se borró.
  > - No se creó ningún temporal fuera de la raíz: los `/tmp/scan-006.txt`, `/tmp/ids-*`, `/tmp/cola-*`,
  >   `/tmp/run-*` y `/tmp/sales.txt` del quickstart no existen.
  > - `ls -d /tmp/permea-006-*` → nada.
  > - **Testigo de la instalación real** (`~/.config/permea`, `stat` de nombre, tamaño y fecha de
  >   modificación, nunca el contenido): 5 ficheros antes y después, **idéntico**. No había ningún
  >   proceso `permea` en marcha.

### Tramo C3 · Snapshot

- [x] **T077** `goreleaser check`, luego `goreleaser release --snapshot --clean`. Transcribir:
  - la versión estampada (**no** `0.3.0`, `research.md` R10);
  - los 5 archivos y `sha256sum -c` del fichero de checksums.

  Localizar `permea_*_windows_amd64.zip` para W1. **No se publica nada.**
  > **Hecho el 2026-10-02 (C3)**, sobre `b504268`, con GoReleaser v2.16.0. **No se publicó nada**: sólo
  > `--snapshot` («skipping announce, publish, and validate»), sin etiqueta, release, push, tap ni
  > bucket.
  > - `goreleaser check` → rc 0.
  > - `goreleaser release --snapshot --clean` → rc 0. El hook `go mod tidy` deja `go.mod` igual (mismo
  >   md5).
  > - **Versión estampada: `0.2.1-SNAPSHOT-b504268`**, **no** `0.3.0` (`dist/metadata.json`: `tag`
  >   v0.2.1, `previous_tag` v0.2.0). R10 la suponía «probablemente `0.2.2-SNAPSHOT-<sha>`, sin
  >   comprobar»: es `0.2.1-…`. SC-022 sólo exige que no sea `0.0.1-dev`.
  > - **Los 5 archivos**:
  >   - `permea_0.2.1-SNAPSHOT-b504268_darwin_amd64.tar.gz`, 2 455 736 B;
  >   - `…_darwin_arm64.tar.gz`, 2 311 744 B;
  >   - `…_linux_amd64.tar.gz`, 2 408 501 B;
  >   - `…_linux_arm64.tar.gz`, 2 226 405 B;
  >   - `…_windows_amd64.zip`, 2 477 442 B.
  >
  >   `sha256sum -c permea_0.2.1-SNAPSHOT-b504268_checksums.txt`, dentro de `dist/`: **5 × OK**, rc 0.
  > - `git status --porcelain` después: sólo ` M README.md` y ` M specs/006-medicion-fiel/tasks.md`.
  >   `dist/` no aparece porque está ignorado (`.gitignore:20`, `/dist/`).
  > - **Binario empaquetado de Linux**, extraído en `/tmp/permea-006-snap-XXXXXX` y lanzado con
  >   `env -i` y un HOME temporal:
  >   - `permea --version` → `0.2.1-SNAPSHOT-b504268`, rc 0, stderr vacío;
  >   - `permea help` → rc 0 y stderr vacío; stdout de 1 670 B, **idéntico byte a byte** (`cmp`) a la
  >     ayuda general aprobada;
  >   - el HOME temporal quedó vacío.
  >
  >   El temporal se borró tras comprobar el prefijo.
  > - **Para W1**:
  >   - fichero `dist/permea_0.2.1-SNAPSHOT-b504268_windows_amd64.zip`, 2 477 442 B, sha256
  >     `d07efb21675a1ecaa7dcd8102b493f608d4ce0c4b35f68b5b25549eb0951ad15`, igual que en el fichero de
  >     checksums;
  >   - trae `LICENSE` (11 352 B), `README.md` (11 716 B) y `permea.exe` (5 673 472 B);
  >   - `file permea.exe`: «PE32+ executable (console) x86-64, for MS Windows».
  >
  >   `dist/` se conserva.
  > - **El README empaquetado es el del árbol de trabajo**, con el arreglo de C1 todavía sin commit, y
  >   no el de `HEAD`. Al binario no le afecta. La release real se construye desde la etiqueta.
- [ ] **T078** ✋ **Commit** (dueño) de las transcripciones C1–C3:
  `006 cierre: puertas, medidas sobre la copia congelada y snapshot`

### Tramo C4 · W1 — ensayo en Windows ANTES de la etiqueta

- [ ] **T079** ✋ **El dueño prepara el secreto de enrolamiento** de la instalación de ensayo, desde la
  plataforma, en `enroll.txt` en la máquina Windows (E-006-P6). Claude no lo ve ni lo transcribe.
- [ ] **T080** ✋ **El dueño ejecuta W1** (quickstart §W1) con el zip de T077:
  1. `--version`;
  2. `enroll` por stdin;
  3. `status`;
  4. `--scan` sobre un log real de Windows.

  Anota fecha, commit del snapshot y el resultado de cada paso. **Si alguno falla, no hay etiqueta**:
  se corrige y se repite desde C1.

### Tramo C5 · Cuerpo del PR

> *(Añadido el 2026-10-02 en B5.)* **Antes de fusionar**: la entrada `## 0.3.0 — PENDIENTE` de
> `CHANGELOG.md` lleva la fecha de la etiqueta, y `grep -c PENDIENTE CHANGELOG.md` → **0**. Si no da 0,
> no se fusiona.

- [ ] **T081** Transcribir aquí el resultado de W1. Para el paso 4, el recuento de mensajes de ese log
  contado desde WSL. Redactar el cuerpo del PR en
  `~/dev/permea-platform/tmp/agente-006-pr.md`:
  - qué entra (B0–B5);
  - las cifras de C2;
  - el resultado de W1;
  - las rupturas declaradas en el CHANGELOG;
  - las enmiendas D-006-7…13 y E-006-P1…P6;
  - la línea de atribución del repositorio.
- [ ] **T082** ✋ **El dueño**: commit de la transcripción de W1
  (`006 cierre: ensayo en Windows sobre el snapshot (W1)`), push de la rama y PR
  `006-medicion-fiel` → `main` con ese cuerpo.

### Tramo C6 · Fusión

- [ ] **T083** ✋ **El dueño fusiona la PR con merge commit**, como #1 y #2.

### Tramo C7 · Etiqueta

- [ ] **T084** ✋ **El dueño**, en `main` actualizado y limpio:
  `git tag -a v0.3.0 -m "v0.3.0: medicion fiel y publicable"` y `git push origin v0.3.0`.

### Tramo C8 · Verificación de los tres canales

- [ ] **T085** Sólo lectura (quickstart §P1):
  - `gh run list` (el workflow de release en verde);
  - `gh release view v0.3.0` (5 archivos y checksums);
  - el `permea.json` del bucket y el cask en `0.3.0`;
  - `install.sh` sirve la última release.

  Transcribir. Si el tap o el bucket no se actualizaron, se mira primero `TAP_GITHUB_TOKEN`: renovado
  el 2026-10-01 según el orquestador, sin comprobar por Claude.

### Tramo C9 · W2 — ensayo final

- [ ] **T086** ✋ **El dueño ejecuta W2** (quickstart §W2):
  1. `scoop update`;
  2. `--version` = `0.3.0`;
  3. `enroll` por stdin;
  4. `status`;
  5. `-run`.

  Después, el recuento de eventos en la plataforma para esa instalación y ventana.
- [ ] **T087** Transcribir W2: eventos en la plataforma frente a mensajes distintos en los logs, para
  esa ventana (SC-020). Cerrar el checklist del quickstart.
- [ ] **T088** ✋ **Commit de cierre** (dueño; rama o `main`, a su criterio):
  `006 cierre: ensayo final en Windows (W2) y checklist`

---

## Dependencies & Execution Order

```text
B0 ──► B1 ──► B2a ──► B2b ──► B2c ──► [T043 ✋ Q-006-1] ──► B3 ──► B4 ──► B5
                                                                          │
   C1 ──► C2 ──► C3 ──► C4 (W1 ✋) ──► C5 ──► C6 ✋ ──► C7 ✋ ──► C8 ──► C9 (W2 ✋)
```

- **B0 primero**: puerta del linter (D-006-P10).
- **B1 antes que B2a**: Principio IV; los fixtures llevan identificadores antes de que el código los
  exija.
- **T043 bloquea B3** y nada más: B4 y B5 no dependen de tarifas. Si la respuesta tarda, **el orden se
  mantiene** de todos modos; sólo B3 espera, y B4 y B5 pueden adelantarse por orden expreso del
  orquestador.
- **B4 después de B2c**: los dos tocan `cmd/permea/main.go`.
- **B5 el último**: documenta la CLI y el CHANGELOG definitivos.
- **Cierre**: un tramo por mensaje. **W1 bloquea la etiqueta.**

---

## Tabla de cobertura

### Requisitos

| FR | Tareas | | FR | Tareas |
|---|---|---|---|---|
| FR-001 | T025, T029–T031 | | FR-018 | T047, T066 |
| FR-002 | T013 (1)(2), T019, T020 | | FR-019 | T047, T048 |
| FR-003 | T011, T014, T020 | | FR-020 | T045, T049 |
| FR-004 | T013 (2), T017, T019 | | FR-021 | T052, T053, T059, T060, T091, T092 |
| FR-005 | T026, T030 | | FR-022 | T055, T056, T060 |
| FR-006 | T013 (4), T017, T027, T030 | | FR-023 | T054, T061, T091, T092 |
| FR-007 | T013 (3), T020 | | FR-024 | T057 |
| FR-008 | T013 (1), T020, T022 | | FR-025 | T066, T069, T093 |
| FR-009 | T038, T040 | | FR-026 | T066, T069, T093 |
| FR-010 | T036, T037, T039 | | FR-027 | T066 |
| FR-011 | T009 | | FR-028 | T067, T069, T093 |
| FR-012 | puertas de cada bloque, T072 | | FR-029 | T068, T069 |
| FR-013 | T011, T014, T090 | | FR-030 | T077, T082–T085 |
| FR-014 | T043–T047 | | FR-031 | T089, T001–T007, puertas, T072 |
| FR-015 | T047, T048 | | FR-032 | T086, T087 |
| FR-016 | T046, T047 | | FR-033 | T025, T029–T031, T033 |
| FR-017 | T047 (sin cambio de semántica; `TestCost_UnknownModel` sigue) | | FR-034 | T079–T081 |

### Criterios

| SC | Tareas | | SC | Tareas |
|---|---|---|---|---|
| SC-001 | T036, T074 (V2) | | SC-012 | T052 |
| SC-002 | T037, T074 (V3) | | SC-013 | T054 |
| SC-003 | T013 (1), T017, T074 (V4 a/b), T075 (V4-c) | | SC-014 | T055, T056 |
| SC-004 | T011, T014, T074 (V5) | | SC-015 | T057 |
| SC-005 | puertas, T072 | | SC-016 | T066, T069 |
| SC-006 | T013 (3), T074 (V7) | | SC-017 | T067, T069 |
| SC-007 | T038, T040 | | SC-018 | T005, puertas, T072 |
| SC-008 | T017, T026, T027 | | SC-019 | T085 |
| SC-009 | T045, T049 | | SC-020 | T086, T087 |
| SC-010 | T046, T074 (V12) | | SC-021 | T029, T033, T074 (V10) |
| SC-011 | T048 | | SC-022 | T080, T081 |

**34 / 34 requisitos y 22 / 22 criterios con tarea.**

> *(Enmendado el 2026-10-02, C1.)* Se añaden las tareas nuevas a los requisitos que cubren:
> - T090 (núcleos de la denylist) a FR-013;
> - T091 y T092 (texto aprobado, D-006-14) a FR-021 y FR-023;
> - T093 a FR-025 («Actualizar»), FR-026 (el test de CRLF respalda la variante de PowerShell) y
>   FR-028 (CHANGELOG).
>
> Ningún requisito ni criterio cambia de estado.

---

## Recuento

| Grupo | Tareas | De ellas ✋ |
|---|:--:|:--:|
| B0 | T089, T001–T008 (9) | 1 |
| B1 | T009–T016 (8) | 1 |
| B2a | T090, T017–T023 (8) | 1 |
| B2b | T024–T035 (12) | 1 |
| B2c | T036–T042 (7) | 1 |
| Q-006-1 | T043–T044 (2) | 1 |
| B3 | T045–T051 (7) | 1 |
| B4 | T052–T064, T091, T092 (15) | 1 |
| B5 | T065–T071, T093 (8) | 1 |
| Cierre C1–C9 | T072–T088 (17) | 8 (T078, T079, T080, T082, T083, T084, T086, T088) |
| **Total** | **93** | **17** |

**Mutaciones**: M-B0, M-B1, y m1–m19 del plan: **21**, cada una con censo y co-caídas declarados.
**Rojos**: (1)–(26), de los cuales nacen verdes y se validan por mutación (14), (17) y (25), y el
golden de T014.

> *(Enmendado el 2026-10-02, C1.)* El recuento recoge T089–T093. T089 ya figuraba en B0; se suman
> T090 a B2a, T091 y T092 a B4, y T093 a B5. **Total: 93 tareas** (antes 89), y siguen **17 ✋**.
>
> **Mutaciones: 34** (antes 21): las 21 del plan y 13 añadidas en los bloques, todas con censo
> declarado antes de mutar y reversión verificada por md5:
> - B2a:
>   - la de T090, sin nombre: el identificador truncado en `SessionRef`;
>   - **m-hex** y **m-formas**.
> - B2b: **m-ids**.
> - B3: **m-s46** y **m-h45**, y, en el remate, **m-cruce** y **m-centimos** (antes «m14» y «m15»).
> - B4: **m-token-ayuda**, **m-token-errores** y **m-token-status**, y, en el remate, **m-jerga**.
> - B5: **m-crlf**.
>
> **Rojos**: (1)–(29). (27)–(29) son de T091.
>
> Validados por mutación, porque nacen verdes:
> - enteros: (14), (17) y (25), y el golden de T014;
> - en parte, por subtests:
>   - (5), con m1, m-formas y m-hex;
>   - (10), con m-ids y m7;
>   - (16), con m-h45 y m-s46;
>   - (29), con m-jerga;
>   - el subtest CRLF de T093, con m-crlf.

---

## Lo que este plan de tareas NO hace

- **No toca `internal/event`** (D-006-3). `event.NewID` se queda sin llamantes de producción
  (`research.md` R1.5).
- **No toca `internal/project/testdata/`**, salvo la nota fechada de su README (T010).
- **No pone tests ni linter en la tubería de publicación** (D-006-6).
- **No repara lo que la plataforma ya recibió** (spec §Fuera de alcance).
- **No ejecuta `--run`, `--daemon` ni `enroll` fuera de T075**, que es el único caso autorizado
  (E-006-P2): sandbox `env -i`, sin enrolar, sobre la copia.
