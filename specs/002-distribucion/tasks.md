---
description: "Task list for feature implementation"
---

# Tasks: Distribución — release multiplataforma del binario `permea`

**Input**: Design documents from `/specs/002-distribucion/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅ (release-workflow.md, artifacts.md, install-contract.md), quickstart.md ✅

**Tests**: Se incluye test SOLO donde aporta y el usuario lo pidió: el test del flag `--version` (único cambio de código, test-first por cultura del repo) y una prueba de que `install.sh` **aborta ante checksum manipulado** (SC-005). El resto de la validación es de release (checks/dry-run), no test unitario.

**Organization**: Tareas agrupadas por historia de usuario. Las tres historias son P1; US1 (publicar por etiqueta) es el MVP y produce los artefactos que US2/US3 instalan.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Puede ejecutarse en paralelo (fichero distinto, sin dependencias sobre tareas incompletas)
- **[Story]**: US1, US2, US3 (mapea a las historias de spec.md)
- Toda descripción incluye ruta de fichero exacta

## Path Conventions

El binario ya existe (spec 001). Esta feature añade **ficheros de release en la raíz** del repo
(`.goreleaser.yaml`, `.github/workflows/release.yml`, `install.sh`, `scripts/test-install.sh`,
`LICENSE`) y un flag menor en `cmd/permea/main.go`. La fórmula de Homebrew y el manifiesto de
Scoop los **genera GoReleaser** en repos externos propios (tap/bucket); no viven en este repo.

## Estado de partida (no reinventar)

- ✅ `cmd/permea/main.go` — `var version = "0.0.1-dev"` (sobreescribible por `-ldflags`), flags `--scan/--run/--daemon`. **Falta** `--version`.
- ✅ `cmd/permea/main_test.go` — existe (spec 001); se le añade el test del flag.
- ⛔ No existen: `.goreleaser.yaml`, `.github/workflows/`, `install.sh`, `LICENSE`.
- ℹ️ Contrato de nombres de artefacto (`contracts/artifacts.md`) es la fuente de verdad compartida por script/fórmula/manifiesto.

---

## Prerrequisitos operativos (NO son tareas de código)

**Deben completarse una vez, fuera del repo, antes de cortar una release real (US1 en CI).**
El dry-run local (`--snapshot`) NO los necesita; sí los necesita la publicación real y tap/bucket.

- [ ] P001 **PRERREQUISITO (operativo, manual)** Crear el repo del tap de Homebrew `github.com/bfgnet/homebrew-permea` (vacío, público). No es tarea de código; GoReleaser commiteará ahí la fórmula.
- [ ] P002 **PRERREQUISITO (operativo, manual)** Crear el repo del bucket de Scoop `github.com/bfgnet/scoop-permea` (vacío, público). No es tarea de código; GoReleaser commiteará ahí el manifiesto.
- [ ] P003 **PRERREQUISITO (operativo, manual)** Crear un PAT con scope `repo` y añadirlo como secreto `TAP_GITHUB_TOKEN` en `bfgnet/agente_permea` (Settings → Secrets → Actions). Necesario para que GoReleaser escriba en el tap y el bucket (el `GITHUB_TOKEN` no puede escribir en otros repos).

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Herramientas locales e higiene del repo necesarias para el empaquetado.

- [X] T001 [P] Verificar el tooling local de release: `goreleaser --version` (v2) y `shellcheck --version` presentes; documentar en `quickstart.md` la versión usada si difiere de la anotada
- [X] T002 [P] Añadir `LICENSE` en la raíz del repo (proyecto de código abierto, Principio II); los `archives` de GoReleaser lo incluyen y fallarían si falta

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: El **único** cambio de código del binario: exponer la versión de forma verificable
(`permea --version`). Lo consumen la verificación de US1 (SC-002) y los tres canales de instalación.

**⚠️ CRITICAL**: Bloquea la verificación de versión de todas las historias. Test-first.

- [X] T003 Añadir un test del flag en `cmd/permea/main_test.go` (`TestVersionFlag`): ejecutar la lógica de `--version` y aseverar que imprime **solo** `version` en stdout y termina sin error — **debe FALLAR** antes de T004
- [X] T004 Implementar el flag `--version` en `cmd/permea/main.go`: `flag.Bool("version", …)`; cuando se pasa, imprimir `version` en **stdout** (una línea, sin el prefijo "Permea") y `return`/exit 0, antes de resolver config o escanear — deja T003 en verde

**Checkpoint**: `go build ./cmd/permea && ./permea --version` imprime la versión; suite en verde.

---

## Phase 3: User Story 1 — Publicar una versión con una etiqueta (Priority: P1) 🎯 MVP

**Goal**: Empujar una etiqueta `vX.Y.Z` construye la matriz de 5 binarios estáticos, inyecta la
versión desde la etiqueta y publica una release con los artefactos y su fichero de checksums SHA256.

**Independent Test**: `goreleaser release --snapshot --clean` produce en `dist/` los **5** archivos
con el nombre del contrato + `*_checksums.txt`; el binario extraído reporta la versión con
`permea --version` (SC-001, SC-002). En CI real, empujar la etiqueta publica la release equivalente.

### Implementation for User Story 1

- [X] T005 [US1] Crear `.goreleaser.yaml` (config completa, per `contracts/artifacts.md` + `contracts/install-contract.md`):
  - `builds`: `main: ./cmd/permea`, `env: [CGO_ENABLED=0]`, `flags: [-trimpath]`, `ldflags: -s -w -X main.version={{.Version}}`, `goos: [darwin,linux,windows]`, `goarch: [amd64,arm64]`, `ignore: windows/arm64`, `mod_timestamp: {{.CommitTimestamp}}`
  - `archives`: `name_template: permea_{{.Version}}_{{.Os}}_{{.Arch}}`, `format_overrides` zip para windows, incluir `LICENSE` y `README.md`
  - `checksum`: `name_template: permea_{{.Version}}_checksums.txt`, `algorithm: sha256`
  - `brews`: tap `bfgnet/homebrew-permea`, token `{{ .Env.TAP_GITHUB_TOKEN }}`, `test: system "#{bin}/permea", "--version"` (sirve a US2)
  - `scoops`: bucket `bfgnet/scoop-permea`, token `{{ .Env.TAP_GITHUB_TOKEN }}`, `bin: permea.exe` (sirve a US3)
- [X] T006 [US1] Validar la config: `goreleaser check` sin errores (iterar sobre `.goreleaser.yaml` hasta que pase)
- [X] T007 [US1] Dry-run local sin publicar: `goreleaser release --snapshot --clean`; verificar que `dist/` contiene los **5** archivos con el nombre exacto del contrato (darwin amd64/arm64, linux amd64/arm64 en `.tar.gz`; windows amd64 en `.zip`) y un `permea_*_checksums.txt`
- [X] T008 [US1] Verificar la versión inyectada (SC-002): extraer el binario de `dist/permea_*_linux_amd64.tar.gz` y comprobar que `permea --version` imprime la versión del snapshot; confirmar que es estático (`file` → "statically linked")
- [X] T009 [US1] Crear `.github/workflows/release.yml` (per `contracts/release-workflow.md`): `on: push: tags: ['v*.*.*']`; `permissions: contents: write`; job en `ubuntu-latest`; `actions/checkout` con `fetch-depth: 0`; `actions/setup-go` (Go 1.22); `goreleaser/goreleaser-action` con `args: release --clean` y `env` `GITHUB_TOKEN` + `TAP_GITHUB_TOKEN`

**Checkpoint**: US1 funcional — el dry-run produce los 5 artefactos + checksums con versión embebida; el workflow queda listo para disparar en la próxima etiqueta.

---

## Phase 4: User Story 2 — Instalar en macOS/Linux con un comando (Priority: P1)

**Goal**: Un desarrollador en macOS/Linux instala `permea` con un solo comando (tap de Homebrew o
`install.sh`) y obtiene un binario cuya integridad se verifica por SHA256 antes de instalar.

**Independent Test**: Contra los artefactos del snapshot servidos en local, `install.sh` instala un
binario funcional; con un checksum **manipulado**, aborta y **no** deja binario (SC-005); la fórmula
del tap declara `permea --version` como test (SC-004).

### Implementation for User Story 2

- [ ] T010 [US2] Crear `install.sh` (POSIX sh, per `contracts/install-contract.md`): detectar SO/arch con `uname` (`x86_64→amd64`, `aarch64|arm64→arm64`); mapear al nombre de artefacto del contrato; soportar `PERMEA_VERSION` (defecto: última) y `PREFIX` (defecto `/usr/local/bin`, fallback `~/.local/bin`); descargar `.tar.gz` + `checksums.txt` de la release; **verificar el SHA256 ANTES de extraer**; instalar con permisos de ejecución; abortar con código ≠ 0 y sin instalar ante SO/arch no soportado, fallo de descarga o checksum no coincidente
- [ ] T011 [US2] `shellcheck install.sh` sin hallazgos (sin bashismos; POSIX)
- [ ] T012 [US2] Crear `scripts/test-install.sh` que valide `install.sh` contra un directorio de release local simulado: (a) caso feliz → instala y `permea --version` responde; (b) **caso manipulado** → checksum alterado ⇒ `install.sh` termina con código ≠ 0 y `PREFIX` queda sin binario (SC-005). El script sale ≠ 0 si algún aserto falla
- [ ] T013 [US2] Verificar el canal Homebrew (cask, solo macOS): el bloque `homebrew_casks` de `.goreleaser.yaml` pasa `goreleaser check` y define `binary "permea"` sin `service` (GoReleaser v2.16 deprecó `brews:`); confirmar en el snapshot que se genera `dist/homebrew/Casks/permea.rb`; documentar en `quickstart.md` que la publicación real del tap se valida con `brew install bfgnet/permea/permea` tras la primera release, y que **Linux se instala por `install.sh`, no por Homebrew**

**Checkpoint**: US2 funcional — `install.sh` verificado (incl. abort ante manipulación) y el canal Homebrew validado en config.

---

## Phase 5: User Story 3 — Instalar en Windows con un comando (Priority: P1)

**Goal**: Un desarrollador en Windows instala `permea` con un solo comando vía Scoop desde el bucket
del proyecto, con verificación de hash antes de instalar.

**Independent Test**: El manifiesto que GoReleaser genera para windows/amd64 declara `url`, `hash`
(sha256) y `bin: permea.exe`; `scoop install permea` desde el bucket instala un binario que reporta
la versión (SC-004). La verificación de hash es intrínseca a Scoop (FR-009).

### Implementation for User Story 3

- [ ] T014 [US3] Verificar el canal Scoop: el bloque `scoops` de `.goreleaser.yaml` pasa `goreleaser check` (bucket `bfgnet/scoop-permea`, `bin: permea.exe`); confirmar en el snapshot que existe el artefacto `permea_*_windows_amd64.zip` y su línea en checksums; documentar en `quickstart.md` los comandos `scoop bucket add permea …` + `scoop install permea` para validar tras la primera release

**Checkpoint**: Las tres historias quedan cubiertas — publicación por etiqueta (US1) e instalación de un comando en los tres SO (US2 macOS/Linux, US3 Windows), todas con verificación de integridad.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentación de instalación, validación end-to-end y puertas de calidad.

- [ ] T015 [P] Actualizar `README.md`: sección de **instalación** con los tres canales (Homebrew tap, Scoop bucket, `install.sh` en una línea `curl … | sh`), nota de que la versión se inyecta desde la etiqueta, y enlace a la página de releases
- [ ] T016 [P] Ejecutar la validación de `quickstart.md` (V1–V7) y dejar constancia de resultados (config válida, snapshot con 5 artefactos + checksums, binarios estáticos, versión == etiqueta, checksums OK, abort ante manipulación, instalación por canal)
- [ ] T017 [P] Puertas de calidad de la constitución por el cambio de código: `go vet ./...` sin hallazgos, `golangci-lint run` limpio, `go test ./...` en verde (incl. `TestVersionFlag`); confirmar que el golden test de frontera sigue verde (esta feature no toca la frontera)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Prerrequisitos operativos (P001–P003)**: manuales, fuera del repo. No bloquean el dry-run local; **sí** la publicación real en CI (US1 en GitHub) y la escritura en tap/bucket (US2/US3 publicados).
- **Setup (Phase 1)**: sin dependencias — empieza de inmediato.
- **Foundational (Phase 2)**: depende de Setup. Aporta `--version`, necesario para la verificación de versión de US1 (T008) y de los canales.
- **US1 (Phase 3)**: depende de Foundational. Crea `.goreleaser.yaml` (incluye los bloques `brews`/`scoops` que sirven a US2/US3, por el orden pedido) y el workflow.
- **US2 (Phase 4)**: depende de US1 (consume el contrato de nombres y el bloque `brews` de `.goreleaser.yaml`). `install.sh` es independiente del CI.
- **US3 (Phase 5)**: depende de US1 (bloque `scoops` y artefacto windows del snapshot).
- **Polish (Phase 6)**: depende de las historias deseadas completas.

### User Story Dependencies

- **US1 (P1, MVP)**: arranca tras Foundational. Base de todo: sin artefactos publicados no hay nada que instalar.
- **US2 (P1)**: arranca tras US1. Testeable de forma independiente contra el snapshot local (no requiere CI ni una release real).
- **US3 (P1)**: arranca tras US1. La parte de config se valida en local; la instalación real requiere el bucket publicado.

### Within Each Story

- Foundational: test del flag (T003) **antes** de implementarlo (T004).
- US1: `.goreleaser.yaml` (T005) → `check` (T006) → `snapshot` (T007) → verificación de versión (T008); el workflow (T009) es independiente del snapshot pero cierra la historia.
- US2: `install.sh` (T010) antes de `shellcheck` (T011) y de la prueba de manipulación (T012).

### Parallel Opportunities

- Setup: T001 y T002 en paralelo.
- Prerrequisitos P001–P003 pueden hacerse en cualquier momento en paralelo al desarrollo local.
- Polish: T015, T016, T017 en paralelo (ficheros/acciones distintas).
- US2/US3: una vez `.goreleaser.yaml` existe (US1), la verificación de canales (T013, T014) y `install.sh` (T010) pueden avanzar en paralelo entre sí.

---

## Parallel Example: Setup + arranque

```bash
# Setup en paralelo:
Task: "Verificar tooling local (goreleaser, shellcheck)"      # T001
Task: "Añadir LICENSE en la raíz"                              # T002

# Prerrequisitos operativos (manual, en paralelo al código):
Task: "Crear repo tap homebrew-permea"                        # P001
Task: "Crear repo bucket scoop-permea"                        # P002
Task: "Crear secreto TAP_GITHUB_TOKEN"                        # P003
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Fase 1: Setup (tooling + LICENSE).
2. Fase 2: Foundational (`--version` test-first). **Bloquea** la verificación de versión.
3. Fase 3: US1 — `.goreleaser.yaml` + `check` + `snapshot` + workflow.
4. **PARAR y VALIDAR**: dry-run produce 5 artefactos + checksums; `permea --version` == versión (SC-001, SC-002).
5. Cortar la primera release real empujando una etiqueta (requiere P001–P003).

### Incremental Delivery

1. Setup + Foundational → base lista.
2. + US1 → publicación por etiqueta (MVP: hay artefactos verificables).
3. + US2 → instalación macOS/Linux (Homebrew + `install.sh` con verificación e "abort" ante manipulación).
4. + US3 → instalación Windows (Scoop).
5. Cada historia añade un canal sin romper los anteriores.

---

## Notes

- [P] = ficheros distintos, sin dependencias.
- El **nombre de los artefactos** (`contracts/artifacts.md`) es el contrato compartido por
  `install.sh`, la fórmula y el manifiesto: no cambiarlo sin actualizar los tres.
- Esta feature **no toca la frontera** (`internal/event`, `internal/ingest`): el golden test debe
  seguir verde. El único cambio de código es `--version`.
- Cero dependencias externas en el binario; GoReleaser y GitHub Actions son tooling de release.
- Commit tras cada tarea o grupo lógico; parar en cualquier checkpoint para validar la historia.
