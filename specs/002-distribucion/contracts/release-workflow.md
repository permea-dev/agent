# Contrato — Workflow de release (GitHub Actions + GoReleaser)

Contrato del proceso que convierte una **etiqueta** en una **release publicada**. Cumple
FR-001 (disparo automático por etiqueta), FR-002 (matriz) y FR-010 (reproducible).

## Disparo

```yaml
on:
  push:
    tags: ['v*.*.*']       # solo etiquetas semver con prefijo v
```

- Una etiqueta que **no** case `v*.*.*` NO dispara el workflow (no se publica nada).
- La etiqueta DEBE apuntar a un commit con el árbol limpio; GoReleaser aborta si hay cambios sin
  commitear o si la etiqueta no está en `HEAD`.

## Entradas

| Entrada | Origen | Uso |
|---|---|---|
| `github.ref_name` | La etiqueta empujada (`vX.Y.Z`) | GoReleaser deriva `{{.Version}}` = `X.Y.Z` |
| `GITHUB_TOKEN` | Token efímero del repo | Publicar la GitHub Release (assets + notas) |
| `TAP_GITHUB_TOKEN` | Secreto (PAT, scope `repo`) | Commitear fórmula y manifiesto en repos externos |

## Permisos

```yaml
permissions:
  contents: write          # crear la release y subir assets en ESTE repo
```

La escritura en `bfgnet/homebrew-permea` y `bfgnet/scoop-permea` NO la cubre `GITHUB_TOKEN`
(es de otro repo): requiere el PAT `TAP_GITHUB_TOKEN`.

## Pasos (secuencia esperada)

1. `actions/checkout` con `fetch-depth: 0` (historial y etiquetas completos).
2. `actions/setup-go` con Go 1.22 (misma toolchain que el proyecto).
3. `goreleaser/goreleaser-action` con `args: release --clean`.
   - `env.GITHUB_TOKEN` = `secrets.GITHUB_TOKEN`
   - `env.TAP_GITHUB_TOKEN` = `secrets.TAP_GITHUB_TOKEN`

## Salidas (postcondiciones)

Tras un run correcto sobre `vX.Y.Z`:

- Una **GitHub Release** `vX.Y.Z` con **5 artefactos** (matriz) + **1** fichero de checksums.
- La **fórmula** `Formula/permea.rb` actualizada en el tap a `X.Y.Z` con los `sha256` correctos.
- El **manifiesto** `bucket/permea.json` actualizado en el bucket a `X.Y.Z` con el `hash` correcto.
- Cada binario reporta `X.Y.Z` con `permea --version`.

## Fallos y semántica

| Situación | Comportamiento esperado |
|---|---|
| Etiqueta no semver | El workflow no se dispara; nada se publica. |
| Árbol sucio / etiqueta no en HEAD | GoReleaser falla; release **no** publicada. |
| Falta `TAP_GITHUB_TOKEN` | La release del repo puede publicarse, pero tap/bucket fallan; se trata como fallo del release (config incompleta). |
| Re-run de una etiqueta ya publicada | No se crean releases duplicadas contradictorias; `--clean` regenera `dist/` local, y la publicación es idempotente por versión. |

## Verificación (dry-run, sin publicar)

```bash
goreleaser check                         # valida .goreleaser.yaml
goreleaser release --snapshot --clean    # construye en dist/ sin publicar ni tocar tap/bucket
```

**Esperado**: `dist/` contiene los 5 archivos + `*_checksums.txt`; ningún push a GitHub.
