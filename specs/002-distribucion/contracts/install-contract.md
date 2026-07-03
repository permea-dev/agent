# Contrato — Canales de instalación (script, Homebrew, Scoop)

Los tres canales instalan **un binario ya publicado y verificado**; ninguno recompila. Todos
verifican integridad por SHA256 antes de dejar el binario en el sistema (FR-009). Ninguno
instala servicios en segundo plano (FR-011 / SC-007).

## 1. Script de instalación `install.sh` (macOS/Linux)

### Interfaz

```bash
# Última versión, prefijo por defecto:
curl -fsSL https://raw.githubusercontent.com/bfgnet/agente_permea/main/install.sh | sh

# Con variables:
PERMEA_VERSION=v1.4.0 PREFIX="$HOME/.local/bin" sh install.sh
```

| Variable | Defecto | Significado |
|---|---|---|
| `PERMEA_VERSION` | última release | Etiqueta a instalar (`vX.Y.Z`) |
| `PREFIX` | `/usr/local/bin` (o `~/.local/bin` si no hay permisos) | Destino del binario |

### Flujo (postcondiciones y errores)

1. **Detección**: `uname -s` → `darwin|linux`; `uname -m` → `amd64` (de `x86_64`) / `arm64`
   (de `aarch64|arm64`). Par no soportado ⇒ **aborta** con mensaje y código ≠ 0 (no instala).
2. **Descarga**: el `tar.gz` según el contrato de nombres (`contracts/artifacts.md`) y el
   fichero `*_checksums.txt` de la misma release. Fallo de descarga ⇒ aborta.
3. **Verificación**: recalcula el SHA256 del archivo (`sha256sum`/`shasum -a 256`) y lo compara
   con la línea correspondiente de checksums. **No coincide ⇒ aborta y no instala nada**
   (SC-005). Este paso ocurre **antes** de extraer.
4. **Instalación**: extrae el binario y lo coloca en `PREFIX` con permisos de ejecución.
5. **Postcondición**: `permea --version` imprime la versión instalada. No se crea ningún
   servicio, demonio ni entrada de autoarranque.

### Requisitos del script

- POSIX **sh** (no bashismos); `shellcheck`-limpio.
- Solo herramientas presentes en macOS/Linux base: `curl` (o `wget`), `tar`, `sha256sum` o
  `shasum`.
- Idempotente: reinstalar la misma versión deja el mismo binario.

## 2. Fórmula de Homebrew (tap `bfgnet/homebrew-permea`)

Generada y publicada por GoReleaser (`brews:`). Contrato de la fórmula `permea`:

```ruby
class Permea < Formula
  desc "Medidor local de coste de IA (frontera de datos)"
  homepage "https://github.com/bfgnet/agente_permea"
  version "1.4.0"
  # bloque por objetivo con url al artefacto y su sha256 (rellenado por GoReleaser)
  on_macos { on_arm { url "...darwin_arm64.tar.gz"; sha256 "..." } ... }
  def install; bin.install "permea"; end
  test { system "#{bin}/permea", "--version" }
end
```

- **Instalación**: `brew tap bfgnet/permea && brew install permea` (un comando efectivo por
  tap+install; `brew install bfgnet/permea/permea` en uno).
- **Integridad**: Homebrew verifica el `sha256` declarado antes de instalar (FR-009).
- **Test de la fórmula**: `brew test permea` ejecuta `permea --version` (verifica SC-002/SC-003).
- **Sin servicios**: la fórmula NO declara bloque `service` (FR-011).

## 3. Manifiesto de Scoop (bucket `bfgnet/scoop-permea`)

Generado y publicado por GoReleaser (`scoops:`). Contrato de `permea.json`:

```json
{
  "version": "1.4.0",
  "architecture": {
    "64bit": {
      "url": "https://github.com/bfgnet/agente_permea/releases/download/v1.4.0/permea_1.4.0_windows_amd64.zip",
      "hash": "<sha256>",
      "bin": "permea.exe"
    }
  },
  "homepage": "https://github.com/bfgnet/agente_permea",
  "license": "..."
}
```

- **Instalación**: `scoop bucket add permea https://github.com/bfgnet/scoop-permea && scoop install permea`.
- **Integridad**: Scoop verifica el `hash` antes de instalar (FR-009).
- **Sin servicios**: el manifiesto NO usa `post_install` para registrar servicios ni autoarranque
  (FR-011).

## Verificación transversal

| Criterio | Cómo se comprueba |
|---|---|
| SC-004 (un comando) | Instalar por cada canal en su SO y ejecutar `permea --version`. |
| SC-005 (tamper) | Alterar el artefacto o el checksum y confirmar que `install.sh` aborta. |
| SC-007 (sin servicio) | Tras instalar, comprobar que no hay demonio/tarea/servicio nuevo. |
