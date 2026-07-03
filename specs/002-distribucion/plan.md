# Implementation Plan: Distribución — release multiplataforma del binario `permea`

**Branch**: `002-distribucion` | **Date**: 2026-07-03 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-distribucion/spec.md`

## Summary

Automatizar la publicación del binario `permea` (compilado por la spec 001) como binario
estático único para macOS (amd64/arm64), Linux (amd64/arm64) y Windows (amd64). El release
se dispara al empujar una etiqueta semver (`vX.Y.Z`) y lo orquesta **GoReleaser** sobre
**GitHub Actions**: compila la matriz con `CGO_ENABLED=0`, inyecta la versión desde la
etiqueta por `ldflags`, empaqueta cada objetivo, genera un fichero de checksums **SHA256**,
publica la **fórmula de Homebrew** en un tap propio y el **manifiesto de Scoop** en un bucket
propio, y adjunta un **script de instalación** para macOS/Linux que verifica el checksum
antes de instalar. Esta feature no toca la frontera de datos ni la funcionalidad; solo
empaqueta y distribuye. **No** incluye servicio en segundo plano (spec 003). El único cambio
de código en el binario es exponer un modo de reporte de versión (`--version`) para que la
verificación de release e instalación sea comprobable.

## Technical Context

**Language/Version**: Go 1.22 (toolchain go1.22.2). El binario ya existe (`cmd/permea`); esta
feature añade tooling de release, no lógica de producto.

**Primary Dependencies**: **Tooling de release, no dependencias del binario**:
- **GoReleaser** v2 (herramienta de build/release; no se enlaza en el binario).
- **GitHub Actions** (CI) con `actions/checkout`, `actions/setup-go`, `goreleaser/goreleaser-action`.
- El binario publicado sigue siendo **solo stdlib**, estático, `CGO_ENABLED=0` (Principio III intacto).

**Storage**: N/A. Los artefactos viven en **GitHub Releases**; la fórmula y el manifiesto en
repos externos propios (tap de Homebrew, bucket de Scoop).

**Testing**:
- `goreleaser check` (validación de config) y `goreleaser release --snapshot --clean` (dry-run
  local sin publicar) — produce los cinco artefactos + `checksums.txt` en `dist/`.
- Verificación de versión: `permea --version` == etiqueta, sobre el binario del snapshot.
- Verificación de estático: `file`/`go version -m` y arranque en SO limpio.
- `shellcheck` del script de instalación + prueba de que aborta ante checksum manipulado.

**Target Platform**: matriz de release — darwin/amd64, darwin/arm64, linux/amd64, linux/arm64,
windows/amd64. La CI corre en `ubuntu-latest` (cross-compile puro Go; no requiere runners por SO).

**Project Type**: Release engineering / distribución de un CLI de proyecto único.

**Performance Goals**: N/A (proceso de release; minutos por publicación es aceptable).

**Constraints**: binario estático `CGO_ENABLED=0`; release reproducible desde la etiqueta;
integridad por SHA256 en todos los canales; **sin servicio en segundo plano** (spec 003); sin
dependencias de runtime; el nombre de los artefactos es un **contrato** compartido entre
GoReleaser, el script de instalación, la fórmula y el manifiesto.

**Scale/Scope**: 5 objetivos de build, 1 fichero de checksums, 1 workflow de CI, 1 config de
GoReleaser, 1 fórmula de Homebrew, 1 manifiesto de Scoop, 1 script de instalación, 1 flag
`--version` en `cmd/permea`.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Puerta | Estado |
|---|---|---|
| **I. Frontera de datos inviolable** | La distribución no construye ni serializa el evento; no toca `internal/event` ni `internal/ingest`. Ningún contenido cruza nada nuevo. | ✅ PASS — feature de empaquetado; la frontera queda intacta y sigue auditándose donde ya lo hace. |
| **II. Privacidad auditable, local-first** | Código abierto; la confianza se verifica leyendo el código. El coste sigue en local; el binario funciona sin red. | ✅ PASS — checksums SHA256 + build reproducible desde la etiqueta **refuerzan** la auditabilidad. La distribución no introduce llamadas de red en el binario. |
| **III. Binario único y auditable** | Binario estático único, instalable con un comando en macOS/Linux/Windows; se favorece stdlib; toda dependencia externa se justifica; rutas por SO. | ✅ PASS — **objetivo central**. `CGO_ENABLED=0`, sin runtime externo, un comando por SO (Homebrew/Scoop/script). GoReleaser y Actions son *tooling de release*, no dependencias del binario: se justifican como el medio de cumplir el Principio III y no se enlazan en el artefacto. |
| **IV. Test-first en la frontera** | El golden test de frontera es intocable y debe seguir verde. Nada se cierra con la frontera en rojo. | ✅ PASS — no hay cambios de frontera. El único cambio de código (`--version`) no toca el evento; el golden test sigue verde. Las "pruebas" de esta feature son validaciones de release (snapshot, checksum, arranque). |
| **V. Desarrollo dirigido por especificaciones** | Ciclo specify→plan→tasks→implement; artefactos en `specs/NNN-slug/`; consistencia con la constitución. | ✅ PASS — este plan deriva de `spec.md` (corregida a distribución) y no redefine la frontera. |

**Resultado del gate**: sin violaciones. La sección *Complexity Tracking* queda vacía.

**Justificación de dependencias externas (Principio III)**: GoReleaser y GitHub Actions son
herramientas de **entorno de release**, no dependencias del binario distribuido. El artefacto
final no las contiene ni las requiere en el sistema del usuario. Se eligen sobre un pipeline
artesanal (matriz de `go build` + `shasum` a mano en Actions) porque reducen el código de
release, hacen la matriz declarativa y publican Homebrew/Scoop de forma reproducible — menos
superficie de error en la parte que reparte confianza (checksums e integridad).

## Project Structure

### Documentation (this feature)

```text
specs/002-distribucion/
├── plan.md              # Este fichero (/speckit.plan)
├── research.md          # Fase 0 (/speckit.plan)
├── data-model.md        # Fase 1 (/speckit.plan) — entidades de distribución
├── quickstart.md        # Fase 1 (/speckit.plan)
├── contracts/           # Fase 1 (/speckit.plan)
│   ├── release-workflow.md   # Disparo por etiqueta, permisos, secretos, salidas
│   ├── artifacts.md          # Nombres, formatos, checksums, inyección de versión
│   └── install-contract.md   # Script de instalación + fórmula Homebrew + manifiesto Scoop
└── tasks.md             # Fase 2 (/speckit.tasks — NO lo crea /speckit.plan)
```

### Source Code (repository root)

Esta feature añade **ficheros de release en la raíz** y un flag menor en el binario; no crea
paquetes `internal/` nuevos ni toca los existentes de la frontera.

```text
.github/
└── workflows/
    └── release.yml         # CI: on push tag v*.*.* -> GoReleaser (NUEVO)

.goreleaser.yaml            # Config de GoReleaser: matriz, ldflags, archives,
                            #   checksums, brews (tap), scoops (bucket)          (NUEVO)

install.sh                  # Instalador macOS/Linux: detecta SO/arch, descarga,
                            #   verifica SHA256 y aborta ante discrepancia        (NUEVO)

cmd/
└── permea/
    └── main.go             # + flag --version (imprime la versión y sale)       (MODIFICADO)

README.md                   # + sección de instalación (Homebrew/Scoop/script)   (MODIFICADO)

# Repos externos propios (fuera de este repo; GoReleaser publica en ellos):
#   github.com/bfgnet/homebrew-permea   -> Formula/permea.rb   (tap de Homebrew)
#   github.com/bfgnet/scoop-permea      -> bucket/permea.json  (bucket de Scoop)
```

**Structure Decision**: Proyecto único en Go ya existente. La distribución se concentra en la
raíz del repo (workflow de CI + config de GoReleaser + script de instalación) y en dos repos
externos propios para el tap y el bucket, que GoReleaser actualiza en cada release. El único
cambio dentro de `cmd/permea` es el flag `--version`, aislado de la frontera. El **nombre de
los artefactos** definido en `.goreleaser.yaml` es el contrato del que dependen el script de
instalación, la fórmula y el manifiesto; se fija en `contracts/artifacts.md`.

## Complexity Tracking

> Sin violaciones de la Constitution Check. No hay complejidad que justificar.
