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

## 2. Cask de Homebrew (tap `bfgnet/homebrew-permea`, solo macOS)

Generado y publicado por GoReleaser (`homebrew_casks:`). GoReleaser v2.16 **deprecó `brews:`**
(fórmula) para binarios de CLI en favor de **casks**; este contrato refleja esa migración. El
cask es **solo macOS** (amd64/arm64): Homebrew Cask no opera en Linux. Contrato del cask `permea`:

```ruby
cask "permea" do
  version "1.4.0"
  # bloque por arquitectura con url al artefacto y su sha256 (rellenado por GoReleaser)
  on_arm do
    url "https://github.com/bfgnet/agente_permea/releases/download/v1.4.0/permea_1.4.0_darwin_arm64.tar.gz"
    sha256 "..."
  end
  on_intel do
    url "https://github.com/bfgnet/agente_permea/releases/download/v1.4.0/permea_1.4.0_darwin_amd64.tar.gz"
    sha256 "..."
  end
  name "permea"
  desc "Medidor local de coste de IA (frontera de datos)"
  homepage "https://github.com/bfgnet/agente_permea"
  binary "permea"
end
```

- **Plataforma**: **solo macOS** (darwin amd64/arm64). **Linux NO se cubre por Homebrew**: su
  canal es `install.sh` (ver Sección 1), que es el **canal principal de instalación en Linux**.
- **Instalación**: `brew install bfgnet/permea/permea` (un comando; `brew tap bfgnet/permea &&
  brew install permea` de forma equivalente).
- **Integridad**: Homebrew verifica el `sha256` declarado en el cask antes de instalar (FR-009).
- **Sin servicios**: el cask NO declara `service` ni un `postflight` que registre autoarranque
  (FR-011). GoReleaser añade el `postflight` estándar para quitar la cuarentena del binario.

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
