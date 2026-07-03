# Quickstart — Validación de la distribución de `permea`

Guía ejecutable para validar que el pipeline de release cumple el [spec.md](./spec.md). Enlaza
a los [contratos](./contracts/) y al [data-model](./data-model.md); no duplica su detalle.

## Prerrequisitos

- Go 1.22+ (`go version`).
- **GoReleaser** v2 instalado en local (`goreleaser --version`) para los dry-run.
- `shellcheck` para validar el script de instalación.
- Repos externos creados: `bfgnet/homebrew-permea` (tap) y `bfgnet/scoop-permea` (bucket).
- Secreto de CI `TAP_GITHUB_TOKEN` (PAT con scope `repo`) configurado en el repo.

## Escenarios de validación (Given/When/Then → comando)

### V1 — Config de release válida (FR-001, FR-002)

```bash
goreleaser check
```
**Esperado**: `.goreleaser.yaml` válido; la matriz declara los 5 objetivos (darwin amd64/arm64,
linux amd64/arm64, windows amd64) con `CGO_ENABLED=0`. Ver [contracts/artifacts.md](./contracts/artifacts.md).

### V2 — Build local sin publicar (FR-002, FR-003, SC-001)

```bash
goreleaser release --snapshot --clean
ls dist/
```
**Esperado**: en `dist/` aparecen los **5** archivos con el nombre del contrato
(`permea_<ver>_<os>_<arch>.tar.gz|zip`) y un `permea_<ver>_checksums.txt`. Ningún push a GitHub.

### V3 — Binarios estáticos (FR-003, SC-003)

```bash
# Extraer el binario linux del snapshot y comprobar que es estático:
tar -xzf dist/permea_*_linux_amd64.tar.gz -C /tmp
file /tmp/permea            # -> "statically linked"
go version -m /tmp/permea | head -1
```
**Esperado**: ELF estático, sin dependencias dinámicas; equivalente para darwin (Mach-O) y
windows (PE32+). Coherente con lo ya validado en la spec 001.

### V4 — Versión inyectada desde la etiqueta (FR-004, SC-002)

```bash
# El snapshot embebe una versión derivada; en una release real es la etiqueta:
/tmp/permea --version
```
**Esperado**: imprime `X.Y.Z` en stdout y sale con código 0. En una release por etiqueta
`vX.Y.Z`, la salida es exactamente `X.Y.Z`. Ver [contracts/artifacts.md](./contracts/artifacts.md).

### V5 — Checksums correctos (FR-005, FR-009)

```bash
cd dist
sha256sum -c permea_*_checksums.txt --ignore-missing   # macOS: shasum -a 256 -c
```
**Esperado**: `OK` para cada artefacto presente; el fichero de checksums cubre los 5 artefactos.

### V6 — El instalador verifica integridad y aborta ante manipulación (FR-008, SC-005)

```bash
shellcheck install.sh
# Prueba de tamper: servir dist/ localmente, corromper un artefacto y verificar el abort:
#   (con un checksum alterado, install.sh DEBE terminar con código != 0 y no instalar)
```
**Esperado**: `shellcheck` limpio; con un checksum que no coincide, `install.sh` **aborta y no
coloca ningún binario**. Ver [contracts/install-contract.md](./contracts/install-contract.md).

### V7 — Instalación de un comando por canal (SC-004, SC-007)

```bash
# macOS (Homebrew cask — SOLO macOS):
brew install bfgnet/permea/permea && permea --version
# Linux y macOS (script — canal PRINCIPAL en Linux):
curl -fsSL https://raw.githubusercontent.com/bfgnet/agente_permea/main/install.sh | sh && permea --version
# Windows (Scoop):
scoop bucket add permea https://github.com/bfgnet/scoop-permea; scoop install permea; permea --version
```
**Esperado**: cada canal instala con un único comando efectivo; `permea --version` coincide con
la versión publicada; **no** se crea ningún servicio en segundo plano (SC-007, spec 003 aparte).

**Canales por SO** (GoReleaser v2.16 deprecó `brews:` → Homebrew usa **casks**):

- **macOS**: cask de Homebrew (`homebrew_casks:`, `binary "permea"`, sin `service`) o `install.sh`.
- **Linux**: **`install.sh`** es el canal principal; el cask de Homebrew es solo macOS.
- **Windows**: Scoop (`scoops:`, `bin: permea.exe`, verificación de `hash` intrínseca).

### V6-bis — Verificación local de `install.sh` (antes de una release real)

```bash
shellcheck install.sh
sh scripts/test-install.sh   # caso feliz + caso tamper (checksum manipulado -> abort sin binario)
```
**Esperado**: `shellcheck` limpio; el harness imprime `PASS` para ambos casos y sale con 0. El
caso tamper prueba SC-005: un checksum alterado aborta la instalación sin dejar binario.

## Cortar una release real (procedimiento)

```bash
# 1. Estar en la rama principal con el árbol limpio.
git tag v1.4.0
git push origin v1.4.0
# 2. GitHub Actions dispara GoReleaser (ver contracts/release-workflow.md).
# 3. Verificar la release publicada:
gh release view v1.4.0            # 5 artefactos + checksums
```
**Esperado**: release `v1.4.0` con los 5 artefactos y su fichero de checksums; tap y bucket
actualizados a `1.4.0`.

## Puertas de calidad (constitución)

```bash
go vet ./...          # el cambio --version no introduce hallazgos
golangci-lint run     # limpio
go test ./...         # frontera + suite completa siguen en verde (sin cambios de frontera)
```

## Nota de frontera (Principio I)

Esta feature **no toca la frontera**: no modifica `internal/event` ni `internal/ingest`. El
único cambio de binario es el flag `--version`, que no serializa ni transmite ningún dato del
log. El golden test de frontera sigue siendo la línea de defensa y debe permanecer en verde.
