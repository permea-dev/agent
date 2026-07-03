# Research — Distribución multiplataforma de `permea`

Fase 0 del plan. Resuelve las decisiones técnicas del pipeline de release. No hay
`NEEDS CLARIFICATION` pendientes: el input del usuario fija herramienta (GoReleaser), CI
(GitHub Actions), matriz, integridad (SHA256) y canales (Homebrew/Scoop/script).

## R1 — Herramienta de release: GoReleaser v2

- **Decisión**: usar **GoReleaser v2** invocado desde GitHub Actions.
- **Rationale**: la matriz de build, el empaquetado por objetivo, el fichero de checksums y la
  publicación de fórmula Homebrew y manifiesto Scoop son declarativos en un solo
  `.goreleaser.yaml`. Reduce el código de release y concentra en un contrato la parte sensible
  (integridad). Reproducible con `--snapshot` en local.
- **Alternativas descartadas**: matriz artesanal de `go build` + `shasum` + `gh release create`
  en el propio workflow (más YAML imperativo, más superficie de error en checksums y en la
  publicación cruzada a tap/bucket); `ko` (orientado a contenedores, no aplica a un CLI nativo).

## R2 — Disparo: etiqueta semver

- **Decisión**: el workflow se dispara con `on: push: tags: ['v*.*.*']`. GoReleaser exige que la
  etiqueta apunte al `HEAD` que se compila y que el árbol esté limpio.
- **Rationale**: una etiqueta es el identificador canónico e inmutable de la versión; el release
  queda atado a un commit concreto (reproducibilidad, FR-010).
- **Alternativas descartadas**: `workflow_dispatch` manual (no automático, FR-001 pide sin pasos
  manuales); disparo por *GitHub Release created* (invierte el flujo: obliga a crear la release a
  mano antes de construir).
- **Etiqueta inválida**: un patrón que no case `v*.*.*` no dispara el workflow (FR-001 / US1-3).

## R3 — Inyección de versión desde la etiqueta

- **Decisión**: `ldflags` `-s -w -X main.version={{.Version}}` en `.goreleaser.yaml`. GoReleaser
  expone `{{.Version}}` = etiqueta sin la `v` inicial. En `cmd/permea` se añade un flag
  `--version` que imprime `main.version` y sale con código 0.
- **Rationale**: SC-002 exige que la versión reportada por el binario == etiqueta. El binario ya
  tiene `var version` sobreescribible por `-ldflags` (spec 001); solo falta exponerla de forma
  verificable. `-s -w` reduce tamaño sin afectar a la corrección.
- **Alternativas descartadas**: `git describe` embebido (la etiqueta ya es la fuente de verdad,
  y `--snapshot` sin etiqueta usaría un pseudo-valor que confundiría la verificación); leer la
  versión de un fichero (otra fuente que mantener sincronizada).

## R4 — Binario estático sin CGO

- **Decisión**: `env: [CGO_ENABLED=0]`, `flags: [-trimpath]`, y la matriz `goos/goarch` exacta.
  `mod_timestamp: {{.CommitTimestamp}}` para builds deterministas.
- **Rationale**: FR-003 y Principio III. Puro Go cross-compila a los cinco objetivos desde un
  runner Linux sin toolchains nativas. `-trimpath` + `mod_timestamp` acercan la reproducibilidad
  (FR-010, SC-006).
- **Alternativas descartadas**: builds nativos por SO en runners `macos`/`windows` (más lentos y
  caros, innecesarios para Go puro); habilitar CGO (rompe el binario estático y la portabilidad).

## R5 — Integridad: checksums SHA256

- **Decisión**: bloque `checksum` de GoReleaser con `algorithm: sha256`, un único fichero
  `permea_{{.Version}}_checksums.txt` que lista el SHA256 de cada artefacto (formato `sha256sum`).
- **Rationale**: FR-005/FR-009. Un fichero de checksums estándar es verificable por el script de
  instalación (`sha256sum -c`/`shasum -a 256 -c`), por Homebrew (`sha256` en la fórmula, que
  GoReleaser rellena) y por Scoop (`hash` en el manifiesto).
- **Alternativas descartadas**: un `.sha256` por artefacto (más ficheros, misma garantía); firma
  GPG/cosign (mayor garantía pero no la pide el input; se anota como mejora futura, no bloquea).
- **Mejora futura anotada**: firmar `checksums.txt` con cosign/GPG para cadena de confianza
  completa; fuera de alcance de esta spec.

## R6 — Homebrew: tap propio

- **Decisión**: bloque `brews` de GoReleaser publicando en el repo externo propio
  **`github.com/bfgnet/homebrew-permea`** (convención `homebrew-<nombre>`), fórmula `permea`.
  Cubre darwin amd64/arm64 (y linux via Homebrew) descargando el `tar.gz` correcto y validando
  su `sha256`. La publicación cruzada usa un **PAT** en el secreto `TAP_GITHUB_TOKEN`.
- **Rationale**: FR-006/SC-004. Un tap propio evita el proceso de homebrew-core y da control de
  versión inmediato. GoReleaser genera y commitea la fórmula con el checksum ya calculado.
- **Alternativas descartadas**: enviar a homebrew-core (revisión externa lenta, no apto para un
  release automático); Cask (es para apps `.app`/GUI, no un CLI).

## R7 — Scoop: bucket propio

- **Decisión**: bloque `scoops` de GoReleaser publicando en el repo externo propio
  **`github.com/bfgnet/scoop-permea`** (bucket), manifiesto `permea.json` para windows/amd64 con
  el `hash` SHA256 del `zip`. Mismo `TAP_GITHUB_TOKEN` (o uno análogo `SCOOP_GITHUB_TOKEN`).
- **Rationale**: FR-007/SC-004. Scoop es el gestor idiomático en Windows para CLIs de usuario sin
  privilegios de administrador; el manifiesto verifica el hash antes de instalar (FR-009).
- **Alternativas descartadas**: instalador MSI/winget (winget exige revisión de manifiestos en un
  repo central; MSI implica firma de código y más complejidad — desproporcionado para un CLI).

## R8 — Script de instalación (macOS/Linux)

- **Decisión**: `install.sh` en **POSIX sh**, publicado en el repo (y referenciado en el README).
  Flujo: detecta SO/arch con `uname -s`/`uname -m` → mapea a los nombres de artefacto del
  contrato (R9) → descarga el `tar.gz` y `checksums.txt` de la release (por defecto la última, o
  `PERMEA_VERSION=vX.Y.Z`) → **verifica el SHA256 antes de extraer** → instala en `PREFIX`
  (`/usr/local/bin` por defecto, o `~/.local/bin` sin permisos). Aborta con mensaje y código ≠ 0
  ante SO/arch no soportado, descarga fallida o **checksum no coincidente** (FR-008/SC-005).
- **Rationale**: cubre distribuciones sin Homebrew y el patrón `curl … | sh`. POSIX sh + `curl`
  y `sha256sum`/`shasum` están disponibles en macOS y Linux base. Se valida con `shellcheck`.
- **Alternativas descartadas**: script en Bash con extensiones (menos portable en `sh` de macOS);
  binario instalador dedicado (introduce otro artefacto que a su vez habría que verificar).
- **Windows**: cubierto por Scoop; el input pide *un* script de instalación (POSIX). Un
  `install.ps1` queda como posible extensión, no requerido por esta spec.

## R9 — Nomenclatura y formato de artefactos (contrato compartido)

- **Decisión**: nombre `permea_{{.Version}}_{{.Os}}_{{.Arch}}` con formato **`tar.gz`** para
  darwin/linux y **`zip`** para windows. `{{.Os}}` ∈ {darwin, linux, windows}; `{{.Arch}}` ∈
  {amd64, arm64}. Cada archivo contiene el binario (`permea`/`permea.exe`), `LICENSE` y `README`.
- **Rationale**: este nombre es el **contrato** del que dependen el script de instalación, la
  fórmula y el manifiesto; fijarlo evita desajustes silenciosos. `tar.gz`/`zip` son los formatos
  que Homebrew y Scoop esperan respectivamente.
- **Alternativas descartadas**: binarios "desnudos" sin archivar (Homebrew/Scoop prefieren
  archivos; complica el empaquetado de LICENSE); nombres con mayúsculas de SO (los templates de
  GoReleaser y `uname` en minúscula son más simples de mapear).

## R10 — Permisos y secretos de CI

- **Decisión**: el job usa `permissions: contents: write` para publicar la release con el
  `GITHUB_TOKEN` del repo. La publicación cruzada al tap y al bucket usa un **PAT** de alcance
  `repo` en el secreto `TAP_GITHUB_TOKEN` (el `GITHUB_TOKEN` no puede escribir en otros repos).
  `fetch-depth: 0` en checkout para que GoReleaser vea etiquetas e historial.
- **Rationale**: mínimo privilegio para la release propia; el PAT es imprescindible para commitear
  en `homebrew-permea` y `scoop-permea`. Se documenta como prerrequisito operativo.
- **Alternativas descartadas**: publicar tap/bucket en el mismo repo (mezcla artefactos de gestor
  con el código; rompe la convención `homebrew-*`/bucket que los gestores esperan); GitHub App
  (más setup del necesario para dos repos propios).

## Resumen de decisiones

| # | Tema | Decisión |
|---|---|---|
| R1 | Herramienta | GoReleaser v2 sobre GitHub Actions |
| R2 | Disparo | `push` de etiqueta `v*.*.*` |
| R3 | Versión | `ldflags -X main.version={{.Version}}` + flag `--version` |
| R4 | Build | `CGO_ENABLED=0`, `-trimpath`, matriz de 5 objetivos |
| R5 | Integridad | `checksums.txt` SHA256 |
| R6 | Homebrew | tap propio `homebrew-permea` |
| R7 | Scoop | bucket propio `scoop-permea` |
| R8 | Script | `install.sh` POSIX con verificación SHA256 |
| R9 | Artefactos | `permea_{ver}_{os}_{arch}` tar.gz/zip |
| R10 | CI | `contents: write` + PAT `TAP_GITHUB_TOKEN` |
