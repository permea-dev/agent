---
description: "Task list for feature implementation"
---

# Tasks: Enrolamiento de dispositivo — emparejar el agente `permea` con su backend

**Input**: Design documents from `/specs/003-enrolamiento/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅ (enrollment-string.md, cli.md), quickstart.md ✅

**Tests**: Se incluyen tests donde aportan y la constitución los exige (**Principio IV: test-first en la frontera**). Prioridad de test-first: (a) la decodificación del enrollment string (función pura), y (b) que el **ping de verificación es un lote vacío `[]`** sin metadato (garantía adyacente a la frontera). El resto son tests de comportamiento de los comandos (persistencia 0600, redacción del secreto, rechazo sin persistir).

**Organization**: Tareas agrupadas por historia de usuario. US1 (enrolar con token válido) es el MVP; US2 (rechazo seguro) y US3 (status) se apoyan en las piezas foundational compartidas.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Puede ejecutarse en paralelo (fichero distinto, sin dependencias sobre tareas incompletas)
- **[Story]**: US1, US2, US3 (mapea a las historias de spec.md)
- Toda descripción incluye ruta de fichero exacta

## Path Conventions

Proyecto único en Go con layout `cmd/` + `internal/` (constitución). 003 **reutiliza** paquetes
existentes; **no** crea paquetes nuevos ni toca la frontera (`internal/event`, `internal/ingest`).

## Estado de partida (no reinventar)

- ✅ `internal/config/config.go` — `Config` ya tiene `Endpoint` y `DeviceToken`; `Save` escribe **0600** atómico (temp+chmod+rename); `Load` devuelve `Default()` si no hay fichero; `Validate()` exige `https`. `DataDir()` resuelve la ruta por SO. **No** se añade ningún campo.
- ✅ `internal/transport/transport.go` — `Client.Send(events)` cumple `contracts/transport.md` (POST HTTPS + `Bearer`, rechaza no-https con `ErrScheme`, `2xx`→nil, `401/403`→auth); `New(endpoint, token)`, `IsAuth(err)`, `Retryable(err)` existen. `json.Marshal([]event.Event{})` → **`[]`** (nil-slice → `null`, por eso el ping usa slice **no-nil**).
- ✅ `cmd/permea/main.go` — despacho por flags (`--scan/--run/--daemon/--version`). **Falta** el despacho de subcomandos `enroll`/`status`.
- ⛔ No existen: `internal/config/enrollment.go`, `cmd/permea/enroll.go`, `cmd/permea/status.go`, ni `transport.Verify()`.
- ℹ️ `contracts/enrollment-string.md` es la fuente de verdad del formato `pmea1.<base64url(json)>` (lo define 003); `contracts/transport.md` (de P-001) es la fuente de verdad del `/ingest` y del token — 003 lo **consume**, no lo redefine.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirmar la superficie reutilizada y que no se introducen dependencias nuevas.

- [X] T001 [P] Verificar el estado de partida (no reinventar) leyendo `internal/config/config.go` y `internal/transport/transport.go`: confirmar `Config.Endpoint`/`Config.DeviceToken`, `Save` 0600, `Validate` https, `Client.Send`/`New`/`IsAuth`; confirmar que `go.mod` NO necesita dependencias nuevas (solo stdlib: `encoding/base64`, `encoding/json`, `bufio`, `os`).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Las dos piezas compartidas por US1 y US2 —decodificar el enrollment string y el ping de verificación de lote vacío—, ambas funciones aisladas y testeables. **Test-first** (Principio IV).

**⚠️ CRITICAL**: Ninguna historia puede completarse sin estas piezas. Las dos parejas test→impl están en ficheros distintos (`config` vs `transport`) y pueden avanzar en paralelo entre sí.

- [X] T002 [P] Añadir tests test-first en `internal/config/enrollment_test.go` para la decodificación del enrollment string (per `contracts/enrollment-string.md`): (a) `pmea1.<base64url(json)>` válido → `(endpoint, token)`; (b) prefijo desconocido → error; (c) base64 inválido → error; (d) JSON malformado → error; (e) `endpoint` con esquema `http://`/no-https → error; (f) `token` vacío → error; (g) **ningún error reproduce el argumento de entrada** (FR-007/FR-013). **Debe FALLAR** antes de T003.
- [X] T003 [P] Implementar `internal/config/enrollment.go`: `ParseEnrollmentString(s string) (endpoint, token string, err error)` (strip prefijo `pmea1.` → `base64.RawURLEncoding.Decode` → `json.Unmarshal` en struct cerrada `{endpoint, token}` → validar `https` y token no vacío; errores **sin** eco del argumento) y `IsEnrolled(cfg config.Config) bool` (`Endpoint!="" && DeviceToken!="" && Validate()==nil`) — deja T002 en verde (FR-001, FR-012, FR-013).
- [X] T004 [P] Añadir un test test-first en `internal/transport/transport_test.go` (`TestVerifyEmptyBatch`) usando `httptest.NewTLSServer`: aseverar que `Verify()` envía un cuerpo **exactamente `[]`** (cero eventos, cero metadato), con `Authorization: Bearer`, y que `2xx`→`nil`, `401`→`IsAuth`, `5xx`/red→error no-auth. **Debe FALLAR** antes de T005 (FR-002, FR-009, FR-010, SC-006, SC-009).
- [X] T005 [P] Implementar `func (c *Client) Verify() error` en `internal/transport/transport.go`: `return c.Send([]event.Event{})` (slice **no-nil** → `[]`), un solo intento (sin `sendWithRetry`); documentar "ping de lote vacío; reutiliza el contrato de transporte, sin endpoint nuevo" — deja T004 en verde (FR-002, FR-011).

**Checkpoint**: `go test ./internal/config/ ./internal/transport/` en verde; decode y ping de verificación disponibles.

---

## Phase 3: User Story 1 — Enrolar el agente con un enrollment string válido (Priority: P1) 🎯 MVP

**Goal**: `permea enroll` recibe el enrollment string por **argv o stdin**, lo decodifica, verifica el token con el ping de lote vacío y, si el backend responde 2xx, persiste `{endpoint, token}` en `config.json` a **0600**. El secreto nunca se filtra.

**Independent Test**: Contra un `/ingest` HTTPS que devuelve 200 (`httptest.NewTLSServer`), `permea enroll <string>` y `… | permea enroll -` confirman el enrolamiento; `config.json` queda a 0600 con `endpoint`+`device_token`; ni la salida ni el argv contienen el token (SC-001, SC-002, SC-003, SC-011).

### Implementation for User Story 1

- [X] T006 [US1] Implementar el flujo en `cmd/permea/enroll.go` (`runEnroll(args []string) error`): resolver la entrada por **dos vías** —argv si hay argumento posicional; si no, **stdin** solo cuando es un **pipe** (no una TTY interactiva): comprobar con `os.Stdin.Stat()` que el modo **NO** tiene `os.ModeCharDevice` (una TTY sí lo tiene) para distinguir pipe de terminal —solo stdlib, sin dependencias nuevas—; leer con `bufio`, recortar espacios/salto final, **sin eco**. Si no hay argumento y stdin es una **TTY interactiva** (o no hay input) → **error de uso, exit ≠ 0**, sin colgarse (implementa el edge case ya fijado en spec/cli, evita el bloqueo de lectura en CI). El convenio `-` fuerza stdin explícito. Con la entrada resuelta: `config.ParseEnrollmentString` → `transport.New(endpoint, token).Verify()`; en `nil` (2xx) cargar config (`config.Load`), fijar `Endpoint`+`DeviceToken` y `config.Save` (0600, ruta por `config.DataDir`); en error **NO** persistir y devolver un error **sin** el token. Éxito imprime confirmación + **URL**, nunca el token (FR-001, FR-002, FR-004, FR-005, FR-006, FR-007, FR-011, FR-014, FR-015).
- [X] T007 [US1] Añadir el despacho del subcomando en `cmd/permea/main.go`: si `os.Args[1] == "enroll"` → `runEnroll(os.Args[2:])` y salir con código según error; **conservar** intactos los flags `--scan/--run/--daemon/--version`.
- [X] T008 [US1] Test de camino feliz en `cmd/permea/enroll_test.go` con `httptest.NewTLSServer` (200) y `HOME`/config dir temporal: `runEnroll` con el string por **argv** persiste `endpoint`+`device_token`; `config.json` a **0600**; la salida incluye la **URL** y **no** el token (SC-001, SC-002, SC-003).
- [X] T009 [US1] Test de la vía **stdin** y SC-011 en `cmd/permea/enroll_test.go`: alimentar el string por stdin (`-` y sin argumento) produce el mismo estado que por argv; aseverar que el token **no** aparece en `os.Args` durante la llamada (SC-011) y que **no** se hace eco a la salida (FR-015). Añadir la aserción del **camino de error** (edge case A1): `runEnroll` sin argumento y con stdin **no-pipe** (TTY simulada / sin input) → **error de uso, exit ≠ 0**, **no** se cuelga y **no** persiste `config.json` (FR-001).

**Checkpoint**: US1 funcional — `go build ./cmd/permea` y `permea enroll` (argv y stdin) enrola contra un backend 2xx, persiste a 0600 y no filtra el secreto.

---

## Phase 4: User Story 2 — Rechazo seguro de un token inválido o revocado (Priority: P2)

**Goal**: Un enrollment string cuyo token el backend rechaza (401/403), o que no puede verificarse (5xx/red), no deja token persistido ni filtra el secreto; el estado posterior es indistinguible de no haber enrolado.

**Independent Test**: Con un backend que responde 401 al lote vacío, `permea enroll` sale ≠ 0, **no** crea/modifica `config.json` y el mensaje no contiene el token; con 5xx/red da un mensaje distinto ("no se pudo verificar") con las mismas garantías (SC-004).

### Implementation for User Story 2

- [X] T010 [US2] Endurecer las ramas de rechazo en `cmd/permea/enroll.go`: distinguir `transport.IsAuth(err)` → mensaje "token rechazado por el backend" de cualquier otro error (5xx/red/`ErrScheme`/timeout) → "no se pudo verificar (backend no disponible)"; **ambas** ramas NO persisten y NO incluyen el token en el mensaje; garantizar que ante error no se ha tocado `config.json` (FR-003, FR-004, FR-007).
- [X] T011 [US2] Tests de rechazo en `cmd/permea/enroll_test.go`: (a) 401 → exit ≠ 0, `config.json` inalterado/ausente (estado indistinguible, SC-004), sin token en la salida; (b) 5xx/red → no persiste, mensaje distinto, sin token; (c) enrollment string malformado / `http://` → aborta antes del ping, sin persistir y sin eco del argumento (FR-004, FR-007, SC-004).
- [X] T012 [US2] Test de **no-residuo tras re-enrolar** en `cmd/permea/enroll_test.go` (FR-014, SC-010): con un config dir temporal y un backend 2xx (`httptest.NewTLSServer`), ejecutar `runEnroll` dos veces con enrollment strings válidos **distintos** (token A y luego token B); aseverar que (a) `config.json` contiene **solo** el token B; y (b) el directorio de config **no** contiene ningún otro fichero que incluya el token A (ni `.bak`, ni temporal `.tmp-*` huérfano) — listar el directorio y verificar que ningún fichero contiene el token viejo.

**Checkpoint**: US1 + US2 — el enrolamiento confirma con token válido y rechaza de forma segura los inválidos/no verificables, sin residuo ni fuga.

---

## Phase 5: User Story 3 — Consultar el estado de enrolamiento (Priority: P3)

**Goal**: `permea status` informa, sin red, si el agente está enrolado y contra qué backend (URL), sin mostrar nunca el token.

**Independent Test**: Con un `config.json` enrolado, `permea status` dice "enrolado" y muestra la URL, sin el token; sin enrolar, dice "no enrolado" y sale 0 (SC-005).

### Implementation for User Story 3

- [X] T013 [US3] Implementar `cmd/permea/status.go` (`runStatus() error`): `config.Load` desde `config.DataDir`; si `config.IsEnrolled(cfg)` imprimir **enrolado** + la **URL** (`cfg.Endpoint`) y a lo sumo un indicador de token tipo `token: configurado`; si no, imprimir **no enrolado**; NUNCA imprimir `cfg.DeviceToken`. Operación local, sin contactar al backend (FR-008, SC-005).
- [X] T014 [US3] Añadir el despacho del subcomando `status` en `cmd/permea/main.go`: si `os.Args[1] == "status"` → `runStatus()`; conservar los flags existentes.
- [X] T015 [US3] Test en `cmd/permea/status_test.go` con config dir temporal: (a) enrolado → salida contiene la URL y **no** el token; (b) sin enrolar → "no enrolado", exit 0, sin secreto (FR-008, SC-005).

**Checkpoint**: Las tres historias cubiertas — enrolar (argv/stdin), rechazar de forma segura y consultar el estado.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Ayuda/usage, validación end-to-end, auditoría de secreto y puertas de calidad.

- [ ] T016 [P] Ayuda y usage: el mensaje de ayuda de `enroll` DEBE mencionar la vía **stdin** como la **recomendada** (per `contracts/cli.md`); actualizar el usage de `cmd/permea/main.go` para listar `enroll` y `status` junto a los flags existentes.
- [ ] T017 [P] Ejecutar la validación de `quickstart.md` (pasos 2–8, incl. permisos 0600, redacción, 2xx/401, **re-enrolamiento sin residuos** SC-010, vía stdin SC-011) y dejar constancia de resultados.
- [ ] T018 [P] Auditoría de higiene del secreto (FR-007): revisar que ninguna ruta de código vuelca `Config`/token con `%+v` ni lo registra (incl. la rama `default:` y los mensajes de error de `main.go`); confirmar que `enroll`/`status` no exponen el token en ninguna salida.
- [ ] T019 [P] Puertas de calidad de la constitución: `go vet ./...` sin hallazgos, `golangci-lint run` limpio, `go test ./...` en verde; **confirmar que el golden test de frontera sigue verde** — 003 no toca la frontera (FR-009).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sin dependencias — empieza de inmediato.
- **Foundational (Phase 2)**: depende de Setup. Aporta `ParseEnrollmentString`/`IsEnrolled` y `Verify()`, que **bloquean** a US1 y US2.
- **US1 (Phase 3)**: depende de Foundational. Crea `enroll.go` + el despacho de `enroll`.
- **US2 (Phase 4)**: depende de US1 (endurece las ramas de error del mismo `enroll.go`). Testeable de forma independiente (camino de rechazo).
- **US3 (Phase 5)**: depende de Foundational (usa `IsEnrolled`); independiente de US1/US2 salvo el fichero compartido `main.go` (despacho).
- **Polish (Phase 6)**: depende de las historias deseadas completas.

### User Story Dependencies

- **US1 (P1, MVP)**: arranca tras Foundational. Sin enrolar válido no hay nada que consultar ni que rechazar en contexto real.
- **US2 (P2)**: arranca tras US1 (comparte `enroll.go`).
- **US3 (P3)**: arranca tras Foundational; puede solaparse con US1/US2 (solo comparte el despacho en `main.go`).

### Within Each Story / Phase

- Foundational: T002 antes de T003; T004 antes de T005 (test-first). Las dos parejas en paralelo entre sí.
- US1: T006 (enroll.go) → T007 (despacho); T008/T009 (tests) tras tener `runEnroll`.
- US2: T010 (endurecer) antes de T011 (tests de rechazo) y T012 (test de no-residuo tras re-enrolar).
- US3: T013 (status.go) → T014 (despacho) → T015 (test).

### Parallel Opportunities

- Setup: T001 solo.
- Foundational: **(T002→T003)** y **(T004→T005)** en paralelo (ficheros `config` vs `transport`).
- US3 puede desarrollarse en paralelo a US1/US2 si se coordina el editar `cmd/permea/main.go` (despacho) para evitar conflicto en ese fichero.
- Polish: T016–T019 en paralelo (ficheros/acciones distintas), salvo que T016 y T018 tocan `main.go` — coordinar ese fichero.

---

## Parallel Example: Foundational

```bash
# Las dos parejas test-first en paralelo (paquetes distintos):
Task: "T002 tests de decodificación en internal/config/enrollment_test.go"   # → T003 impl
Task: "T004 test del ping de lote vacío en internal/transport/transport_test.go"  # → T005 impl
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Fase 1: Setup (confirmar superficie reutilizada, sin deps nuevas).
2. Fase 2: Foundational (decode + `Verify()`, test-first). **Bloquea** las historias.
3. Fase 3: US1 — `enroll.go` (argv/stdin) + despacho + tests.
4. **PARAR y VALIDAR**: `permea enroll` contra un backend 2xx persiste a 0600 sin filtrar el secreto (SC-001/002/003/011).

### Incremental Delivery

1. Setup + Foundational → base lista.
2. + US1 → enrolar con token válido (MVP: el agente ya puede autenticarse).
3. + US2 → rechazo seguro de tokens inválidos/no verificables.
4. + US3 → `permea status`.
5. Cada historia añade valor sin romper las anteriores.

---

## Notes

- [P] = ficheros distintos, sin dependencias.
- 003 **no toca la frontera** (`internal/event`, `internal/ingest`): el golden test debe seguir verde; el ping de verificación es un lote vacío `[]` (cero eventos, cero metadato).
- **Cero dependencias externas nuevas**: todo con stdlib (`encoding/base64`, `encoding/json`, `bufio`, `os`, `net/url`).
- El **device token** y el **enrollment string** son secretos del mismo calibre que el `salt`: nunca a logs, stdout (salvo el argv que el usuario pega) ni mensajes de error.
- `contracts/enrollment-string.md` (formato) lo define 003; `contracts/transport.md` (token + `/ingest`) es de P-001/P-002 y no se redefine.
- Commit tras cada tarea o grupo lógico; parar en cualquier checkpoint para validar la historia.
