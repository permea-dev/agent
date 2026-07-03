# Contrato — Artefactos, nomenclatura, checksums y versión

El **nombre de los artefactos es el contrato central** del que dependen el script de
instalación, la fórmula de Homebrew y el manifiesto de Scoop. Cambiarlo rompe los tres canales,
así que se fija aquí. Cumple FR-002/FR-003/FR-004/FR-005.

## Nomenclatura

```
permea_{version}_{os}_{arch}.{ext}
```

- `version` = etiqueta sin `v` (`1.4.0`).
- `os` ∈ `darwin | linux | windows` (minúsculas, como GoReleaser y `uname` normalizado).
- `arch` ∈ `amd64 | arm64`.
- `ext` = `tar.gz` para darwin/linux, `zip` para windows.

Ejemplos (matriz completa para `v1.4.0`):

```
permea_1.4.0_darwin_amd64.tar.gz
permea_1.4.0_darwin_arm64.tar.gz
permea_1.4.0_linux_amd64.tar.gz
permea_1.4.0_linux_arm64.tar.gz
permea_1.4.0_windows_amd64.zip
permea_1.4.0_checksums.txt
```

## Contenido de cada archivo

- El binario: `permea` (unix) o `permea.exe` (windows).
- `LICENSE` y `README.md` (documentación mínima junto al binario).

## Build (invariantes)

- `CGO_ENABLED=0` — binario **estático**, sin dependencias de runtime (FR-003).
- `flags: -trimpath`; `ldflags: -s -w -X main.version={version}` (FR-004).
- `mod_timestamp` = timestamp del commit → builds deterministas (FR-010).
- Matriz exacta: `goos: [darwin, linux, windows]`, `goarch: [amd64, arm64]`, con
  `windows/arm64` excluido (`ignore`).

## Checksums

- Fichero único `permea_{version}_checksums.txt`, algoritmo **sha256**.
- Una línea por artefacto en formato `sha256sum`:

  ```
  <64-hex>  permea_1.4.0_linux_amd64.tar.gz
  <64-hex>  permea_1.4.0_darwin_arm64.tar.gz
  ...
  ```

- Verificación estándar por el consumidor:

  ```bash
  # Linux
  sha256sum -c permea_1.4.0_checksums.txt --ignore-missing
  # macOS
  shasum -a 256 -c permea_1.4.0_checksums.txt --ignore-missing
  ```

## Contrato de versión embebida

- `main.version` en `cmd/permea` se sobreescribe por `ldflags` con `{version}`.
- El binario DEBE exponer la versión de forma verificable:

  ```
  permea --version   ->  imprime "X.Y.Z" en stdout y sale con código 0
  ```

- Postcondición (SC-002): la salida de `permea --version` de cada artefacto == `version` de la
  etiqueta que lo produjo.

## Reglas de consumo (para los canales)

- El script de instalación construye el nombre del artefacto a partir de `uname` mapeado a este
  esquema; si el par SO/arch no está en la matriz, aborta sin instalar.
- Homebrew y Scoop referencian estos nombres exactos y el `sha256`/`hash` publicado en checksums.
- Ningún canal debe reconstruir el binario: siempre consume el artefacto publicado y verificado.
