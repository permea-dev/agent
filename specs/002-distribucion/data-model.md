# Data Model — Entidades de distribución

Esta feature no define entidades de dominio que crucen la frontera (eso es la spec 001). El
"modelo" aquí es el conjunto de **artefactos de release** y sus relaciones: lo que se produce,
cómo se nombra y cómo se verifica. Fija los invariantes de los que dependen los contratos.

## Entidades

### Release

La unidad publicada asociada a una etiqueta.

| Campo | Descripción | Regla |
|---|---|---|
| `tag` | Etiqueta git que dispara el release | DEBE casar `^v[0-9]+\.[0-9]+\.[0-9]+$` (semver, prefijo `v`) |
| `version` | Versión sin prefijo, `X.Y.Z` | Derivada de `tag` quitando la `v`; inyectada en el binario |
| `commit` | Commit al que apunta la etiqueta | El árbol DEBE estar limpio en ese commit (reproducibilidad) |
| `artifacts` | Lista de artefactos (1 por objetivo) | Exactamente **5** (ver matriz) |
| `checksums` | Fichero de checksums de la release | Exactamente **1**, cubre los 5 artefactos |
| `notes` | Notas de la versión | Generadas del changelog; informativas |

**Transiciones de estado**: `etiqueta empujada → CI construye matriz → artefactos + checksums →
release publicada → tap y bucket actualizados`. Un fallo en cualquier paso deja la release **no
publicada** (no hay estado intermedio observable por el usuario).

### Matriz de objetivos (Target)

Producto SO × arquitectura soportado. Fija cardinalidad = 5.

| os | arch | binario | formato de archivo |
|---|---|---|---|
| darwin | amd64 | `permea` | `tar.gz` |
| darwin | arm64 | `permea` | `tar.gz` |
| linux | amd64 | `permea` | `tar.gz` |
| linux | arm64 | `permea` | `tar.gz` |
| windows | amd64 | `permea.exe` | `zip` |

Reglas: todo objetivo se compila con `CGO_ENABLED=0` (estático). No hay `windows/arm64` ni
`linux/arm` (v6/v7) en esta spec; ampliaciones futuras extienden la matriz sin cambiar el modelo.

### Artefacto (Artifact)

Un archivo comprimido por objetivo.

| Campo | Descripción | Regla |
|---|---|---|
| `name` | Nombre del archivo | `permea_{version}_{os}_{arch}.{ext}` (contrato, ver `contracts/artifacts.md`) |
| `os` / `arch` | Objetivo | Uno de la matriz |
| `ext` | Extensión | `tar.gz` (darwin/linux) · `zip` (windows) |
| `contents` | Contenido del archivo | binario + `LICENSE` + `README` |
| `sha256` | Digest del archivo | Presente en `checksums`; base de la verificación |
| `version_embedded` | Versión reportada por el binario | DEBE == `Release.version` (SC-002) |

### Fichero de checksums (Checksums)

| Campo | Descripción | Regla |
|---|---|---|
| `name` | Nombre | `permea_{version}_checksums.txt` |
| `algorithm` | Algoritmo | `sha256` |
| `lines` | Una línea por artefacto | Formato `sha256sum` (`<hex>  <name>`); cubre los 5 artefactos |

Invariante: para cada `Artifact` existe **exactamente una** línea en `Checksums` con su digest.

### Canal de instalación (InstallChannel)

Vías de instalación de un comando; cada una verifica integridad antes de instalar (FR-009).

| Canal | Objetivos | Fuente de verdad del hash | Repo |
|---|---|---|---|
| Homebrew (tap) | darwin amd64/arm64 (+linux) | `sha256` en `Formula/permea.rb` | `bfgnet/homebrew-permea` |
| Scoop (bucket) | windows amd64 | `hash` en `bucket/permea.json` | `bfgnet/scoop-permea` |
| Script `install.sh` | darwin/linux amd64/arm64 | `Checksums` de la release | este repo |

Invariante transversal: el hash que cada canal verifica DEBE proceder de (o coincidir con) el
`sha256` del `Artifact` correspondiente. GoReleaser rellena la fórmula y el manifiesto con el
mismo digest que publica en `Checksums`, evitando divergencias.

### Versión del binario (embebida)

| Campo | Descripción | Regla |
|---|---|---|
| símbolo | `main.version` en `cmd/permea` | Sobreescrito por `-ldflags -X main.version={version}` |
| salida | `permea --version` | Imprime `X.Y.Z` y sale con código 0; DEBE == `Release.version` |

## Relaciones

```text
Release (tag vX.Y.Z)
├── 1..1  Checksums (sha256)
├── 5..5  Artifact ── 1..1 sha256 ∈ Checksums
│           └── version_embedded == Release.version
└── actualiza
     ├── Homebrew Formula (tap)     -> sha256 de artefactos darwin/linux
     └── Scoop Manifest (bucket)    -> sha256 del artefacto windows
```

## Invariantes verificables (mapa a Success Criteria)

- **INV-1** (SC-001): una `Release` publicada tiene 5 `Artifact` + 1 `Checksums`.
- **INV-2** (SC-002): `Artifact.version_embedded == Release.version` para los 5.
- **INV-3** (SC-003): cada `Artifact` es estático y arranca sin dependencias.
- **INV-4** (SC-005): si el `sha256` recalculado ≠ el de `Checksums`, la instalación aborta.
- **INV-5** (SC-006): misma `tag` + mismo `commit` ⇒ artefactos equivalentes (reproducible).
- **INV-6** (SC-007): ningún canal instala servicios en segundo plano.
