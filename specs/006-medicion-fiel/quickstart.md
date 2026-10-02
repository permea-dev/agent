# Quickstart · Validación de 006 Medición fiel y publicable

**Feature**: `006-medicion-fiel` | **Fecha**: 2026-10-02 | **Forma**: la de `specs/005-adhesion-a-proyecto/quickstart.md`

Dos partes:
- **A**: validaciones automatizables o locales, V1–V18, una por criterio o grupo de criterios.
- **B**: los dos ensayos en Windows (W1, W2) y la publicación (P1). Los nombres W/P evitan confundirlos
  con los bloques B0–B5 del plan.

Cada validación nombra el criterio que acredita. **Nada de este documento se ha ejecutado todavía**:
los comandos son la receta para cuando exista el código, y donde no se ha podido comprobar algo se
dice «sin comprobar».

---

## Prerrequisitos

| Herramienta | Versión con la que se midió la línea base (2026-10-02) |
|---|---|
| Go | `go1.22.2 linux/amd64` (CI usa `1.22`) |
| `golangci-lint` | **2.12.2**. El «0 avisos» de SC-018 se mide con **esta** versión; con otra, se anota cuál |
| `goreleaser` | v2.16.0 |
| `gh` | autenticado contra `permea-dev` (sólo lectura, salvo la PR y la etiqueta del Cierre) |
| `python3` | para el contador independiente. Lee **sólo** identificadores y `usage` |

### Aislamiento obligatorio — antes de cualquier validación que toque config, estado o cola

Igual que en 004 y 005 (disciplina 6): ningún test ni validación escribe en la instalación real.

```sh
export PERMEA_SANDBOX="$(mktemp -d)"
mkdir -p "$PERMEA_SANDBOX/home"
export HOME="$PERMEA_SANDBOX/home"
export XDG_CONFIG_HOME="$PERMEA_SANDBOX/home/.config"
export USERPROFILE="$PERMEA_SANDBOX/home"   # Windows
# Comprobación ANTES de seguir: no debe existir nada todavía.
ls -A "$XDG_CONFIG_HOME/permea" 2>/dev/null && echo "⛔ el sandbox no está vacío" || echo "sandbox limpio"
```

### La copia congelada — el sujeto de los criterios sobre logs reales

Los criterios sobre logs reales (SC-001 a SC-004, SC-006, SC-010 y SC-021) se miden sobre **una copia
congelada y fechada**, nunca sobre el historial vivo, que cambia mientras se mide (spec, §Success
Criteria).

```sh
# Desde un shell SIN el sandbox (necesita el HOME real para leer los logs):
COPIA="$(mktemp -d)/claude-projects-$(date +%F)"
cp -a "$HOME/.claude/projects" "$COPIA"
find "$COPIA" -name '*.jsonl' | wc -l        # anotar: nº de ficheros de la copia
```

⛔ **La copia contiene conversaciones.** Vive en un temporal fuera del repositorio, nunca se commitea
y se borra al terminar (`rm -rf "$COPIA"`). Ningún comando de este documento imprime su contenido:
sólo recuentos.

### El contador independiente — la referencia de SC-001, SC-002 y SC-021

Lee los mismos seis campos que el descubrimiento (spec, §Contexto de partida) y **no imprime ningún
identificador**:

```sh
python3 - "$COPIA" <<'EOF'
import json, glob, os, sys
raiz = sys.argv[1]; facturables = sinteticas = sin_id = 0
mensajes = {}
for f in glob.glob(os.path.join(raiz, '**', '*.jsonl'), recursive=True):
    for linea in open(f, encoding='utf-8', errors='replace'):
        try: r = json.loads(linea)
        except Exception: continue
        if r.get('type') != 'assistant': continue
        m = r.get('message') or {}
        if not m.get('model'): continue
        facturables += 1
        if m['model'] == '<synthetic>': sinteticas += 1; continue
        mid, req = m.get('id'), r.get('requestId')
        if not mid and not req: sin_id += 1; continue
        u = m.get('usage') or {}
        t = sum(u.get(k) or 0 for k in ('input_tokens', 'output_tokens',
                'cache_creation_input_tokens', 'cache_read_input_tokens'))
        mensajes.setdefault((mid, req), t)          # la PRIMERA línea manda (FR-005)
print(f"facturables={facturables} sinteticas={sinteticas} sin_identificador={sin_id}")
print(f"mensajes_distintos={len(mensajes)} tokens_una_vez_por_mensaje={sum(mensajes.values())}")
EOF
```

Referencia del 2026-10-02, sobre el historial de ese día: 10 698 mensajes y 4 835 167 368 tokens (M2).
**No es el valor esperado en otra copia.**

---

# PARTE A · Validaciones (V1 – V18)

## V1 · Puertas (SC-018 · FR-031, D-006-13)

```sh
gofmt -l .                 # → vacío
go vet ./...               # → sin hallazgos
go test -count=1 ./...     # → 9 paquetes ok, 0 FAIL
golangci-lint run          # → 0 issues (versión 2.12.2)
grep -rn "nolint" --include=*.go .   # → exactamente UNA, la de SA1007 (research.md R9 #7)
```

## V2 · Un mensaje, un evento, sobre la copia (SC-001 · FR-001, FR-010)

```sh
permea-dev() { go run ./cmd/permea "$@"; }   # o el binario de `make build`
for f in $(find "$COPIA" -name '*.jsonl'); do permea-dev --scan "$f" 2>/dev/null; done > /tmp/scan-006.txt
grep -c '^evento:' /tmp/scan-006.txt         # → == mensajes_distintos del contador
```

**Falsable**: copiar a un temporal un fichero de la copia, duplicar en él una línea `assistant`, y
pasar `--scan` sobre él. El recuento **no** cambia.

## V3 · Tokens contados una vez (SC-002 · FR-001, FR-005)

Suma de las cuatro partidas que imprime cada línea `evento:` (el nombre de los campos lo fija B2; aquí
`in=`, `out=`, `cw=` y `cr=`):

```sh
awk '/^evento:/{for(i=1;i<=NF;i++){split($i,a,"="); if(a[1]~/^(in|out|cw|cr)$/) s+=a[2]}} END{print s}' /tmp/scan-006.txt
# → == tokens_una_vez_por_mensaje del contador
```

## V4 · Determinismo (SC-003 · FR-002, FR-004)

- **(a) Independencia de la sal y del estado** (automatizado): `go test ./internal/ingest -run 'EventID'`.
  La misma línea con dos `Context` de sal, máquina y desarrollador distintos da el mismo `event_id`, y
  los vectores de `contracts/event-id.md` se reproducen.
- **(b) Sobre la copia**: dos ejecuciones con HOME de sandbox distintos dan el mismo conjunto, todo de
  32 hex.

```sh
grep -o 'event_id=[0-9a-f]*' /tmp/scan-006.txt | sort > /tmp/ids-1.txt
# repetir V2 con otro PERMEA_SANDBOX → /tmp/ids-2.txt
cmp /tmp/ids-1.txt /tmp/ids-2.txt && echo "idénticos"
grep -vcE '^event_id=[0-9a-f]{32}$' /tmp/ids-1.txt     # → 0
sort -u /tmp/ids-1.txt | wc -l                          # → == mensajes_distintos (sin colisiones)
```

> **La mitad «secretos locales distintos» de SC-003 no se puede ver con `--scan`**, que usa una sal
> literal (`research.md` de 005, R5). Se acredita con (a). Si el encargo de implementación autoriza
> `--run` **en sandbox y sin endpoint** —no transmite nada: «sync omitido: sin endpoint configurado»—,
> la comparación de las dos colas con sales distintas es la medida directa. **Va a §Dudas.**

### V4-c · Dos colas con sales distintas *(añadido el 2026-10-02, E-006-P2: `--run` autorizado SÓLO así)*

**Condiciones, todas a la vez**, o no se ejecuta:
- `env -i`, con `HOME` y `XDG_CONFIG_HOME` temporales;
- **sin enrolar**;
- `logs_root` apuntando a la copia congelada.

Fuera de este caso, `--run`, `--daemon` y `enroll` siguen prohibidos.

```sh
BIN="$(pwd)/bin/permea"            # make build; ruta ABSOLUTA, porque env -i vacía PATH
for n in 1 2; do
  S="$(mktemp -d)"; mkdir -p "$S/home/.config/permea"
  printf '{"logs_root": "%s"}\n' "$COPIA" > "$S/home/.config/permea/config.json"
  E="env -i HOME=$S/home XDG_CONFIG_HOME=$S/home/.config"
  # 1 · PRECONDICIÓN: no enrolado. Si dice otra cosa, se PARA.
  $E "$BIN" status | grep -qx 'no enrolado' || { echo "⛔ sandbox $n no dice «no enrolado»"; exit 1; }
  # 2 · la pasada: sin endpoint no transmite («sync omitido: sin endpoint configurado»)
  $E "$BIN" --run 2>"/tmp/run-$n.err"
  grep -q 'sync omitido' "/tmp/run-$n.err" || { echo "⛔ sandbox $n intentó transmitir"; exit 1; }
  python3 -c "import json,sys;[print(json.loads(l)['event_id']) for l in open(sys.argv[1])]" \
    "$S/home/.config/permea/queue.jsonl" | sort > "/tmp/cola-$n.txt"
  cat "$S/home/.config/permea/salt" | sha256sum >> /tmp/sales.txt   # sólo el hash, para ver que DIFIEREN
done
sort -u /tmp/sales.txt | wc -l           # → 2 (sales distintas)
cmp /tmp/cola-1.txt /tmp/cola-2.txt && echo "mismo conjunto de event_id"
sort -u /tmp/cola-1.txt | wc -l          # → == mensajes_distintos del contador
```

Lo que se mide son **recuentos y comparaciones**: ningún `event_id` se imprime a pantalla
(disciplina 9). Al terminar se borran los dos sandboxes y `/tmp/cola-*`, `/tmp/run-*` y
`/tmp/sales.txt`.

## V5 · Nada del proveedor en la salida (SC-004 · FR-003, FR-013)

```sh
go test ./internal/ingest -run 'Boundary'     # golden con los centinelas de identificador
python3 - "$COPIA" /tmp/scan-006.txt <<'EOF'
import json, glob, os, sys
ids = set()
for f in glob.glob(os.path.join(sys.argv[1], '**', '*.jsonl'), recursive=True):
    for l in open(f, encoding='utf-8', errors='replace'):
        try: r = json.loads(l)
        except Exception: continue
        m = r.get('message') or {}
        for v in (m.get('id'), r.get('requestId')):
            if v: ids.add(v); ids.add(v[4:])     # entero y sin prefijo
salida = open(sys.argv[2]).read()
print("apariciones =", sum(1 for i in ids if i and i in salida))   # → 0
EOF
```

## V6 · La frontera no cambia (SC-005 · FR-012)

```sh
git diff 0311fa1 -- internal/event   # → vacío
go test ./internal/event             # TestEvent_OnlyAllowlistKeys verde, sin modificar
grep -rn "event.NewID" --include=*.go . | grep -v '^./internal/event/'   # → vacío (research.md R1.5)
```

## V7 · `<synthetic>` no se emite (SC-006 · FR-007)

```sh
grep -c 'model=<synthetic>' /tmp/scan-006.txt   # → 0
```

## V8 · Actualizar no reenvía (SC-007 · FR-009)

Automatizado en `cmd/permea`, en sandbox:
- `state.json` a mitad de un log;
- una cola con un evento de `event_id` aleatorio;
- una pasada de `generate()`.

La pasada encola sólo lo posterior al offset, y el evento previo queda byte a byte.
`go test ./cmd/permea -run 'Actualizar'`.

## V9 · Reglas de los casos límite (SC-008 · FR-005, FR-006)

`go test ./internal/ingest -run 'CasoLimite'`: tres casos, cada uno en rojo antes de su implementación.

## V10 · Deduplicación en la pasada (SC-021 · FR-033)

```sh
go test ./internal/ingest ./cmd/permea -run 'Pasada'
grep -o 'event_id=[0-9a-f]*' /tmp/scan-006.txt | sort | uniq -d | wc -l   # → 0
```

## V11 · Espejo de tarifas (SC-009 · FR-014, FR-020)

```sh
go test ./internal/pricing -v   # 16 claves, 64 cifras, ninguna de más
# Comparación ENTRE REPOSITORIOS, manual (el test no lee el otro repo):
git -C ~/dev/permea-platform show e50d0a5:backend/config/pricing.php \
  | grep -E "^\s*'(claude-[a-z0-9-]+|input|output|cache_write|cache_read)'" | sed 's/[ ,]//g'
# y comparar fila a fila con internal/pricing/pricing.go. (Enmendado 2026-10-02, Q-006-1 resuelta:
# el commit de referencia es e50d0a5; antes, 865bba0.)
```

**Falsable**: alterar una cifra de `pricing.go` → `go test ./internal/pricing` en rojo, y revertir por
edición inversa (disciplina 3).

## V12 · Coste disponible para lo que se usa (SC-010 · FR-014, FR-017)

```sh
grep '^evento:' /tmp/scan-006.txt | grep -E 'model=claude-(opus-5-5|opus-5|sonnet-5) ' | grep -c 'cost_avail=false'   # → 0
```

Y un evento de referencia por modelo, a mano, con las cifras de M4 sobre `e50d0a5` (Q-006-1 resuelta el
2026-10-02: `claude-sonnet-5` a 2.00 / 10.00 / 2.50 / 0.20).

## V13 · La cabecera está completa (SC-011 · FR-015, FR-019)

Revisión con una casilla por elemento de `contracts/tarifas.md` §La cabecera: fuente · verificación ·
aprobación · catálogo replicado · casamiento · ~~3~~ **2** limitaciones. *(Enmendado 2026-10-02, Q-006-1 resuelta: la de Sonnet 5
desaparece.)*

## V14 · La ayuda general (SC-012 · FR-021)

```sh
go test ./cmd/permea -run 'Ayuda'
# a mano, con los canales capturados POR SEPARADO (disciplina 7):
for a in "" help -h --help; do permea-dev $a >/tmp/o-"${a:-vacio}" 2>/tmp/e-"${a:-vacio}"; echo "[$a] exit=$?"; done
cmp /tmp/o-vacio /tmp/o-help && cmp /tmp/o-help /tmp/o--h && cmp /tmp/o--h /tmp/o---help && echo "idénticas"
cat /tmp/e-* | wc -c    # → 0
```

**Opción desconocida** *(añadido el 2026-10-02, E-006-P3; `contracts/cli.md` §Opción desconocida)*:

```sh
for a in --bogus --scan; do permea-dev $a >/tmp/oo 2>/tmp/eo; echo "[$a] exit=$?"   # → 2
  wc -c </tmp/oo                                       # → 0: nada por stdout
  grep -c -- "${a#--}" /tmp/eo; grep -c 'permea help' /tmp/eo   # → ≥1 y ≥1
done
```

## V15 · Las ayudas de subcomando no hacen nada (SC-013 · FR-023)

`go test ./cmd/permea -run 'AyudaSubcomando'`:
- sandbox vacío;
- agente de prueba enrolado contra `httptest`;
- `cwd` dentro de un árbol con raíz;
- para las 8 invocaciones: 0 peticiones, el árbol del sandbox idéntico, exit 0.

## V16 · Subcomando inexistente y secretos (SC-014 · FR-022)

`go test ./cmd/permea -run 'Desconocido'`:
- `enrol` → exit 1, nombrado;
- `pmea2.<centinela>`, `pmeaj1.<centinela>` y `pmea1.<centinela>` → exit 1, sin el centinela en
  ninguno de los dos canales.

## V17 · El token nunca sale (SC-015 · FR-024)

`go test ./cmd/permea -run 'TokenCentinela'`: `config.json` de prueba con un token centinela, todas las
invocaciones de `contracts/cli.md`, y búsqueda literal en stdout y en stderr.

## V18 · README, CHANGELOG y comentarios (SC-016, SC-017 · FR-025 a FR-029)

```sh
gh api repos/permea-dev/homebrew-permea --jq .full_name   # → permea-dev/homebrew-permea
gh api repos/permea-dev/scoop-permea --jq .full_name      # → permea-dev/scoop-permea
curl -fsSI https://raw.githubusercontent.com/permea-dev/agent/main/install.sh | head -1   # → 200
grep -c bfgnet README.md .goreleaser.yaml .github/workflows/release.yml                    # → 0 en los tres
git diff 0311fa1 -- .goreleaser.yaml .github/workflows/release.yml \
  | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | grep -vE '^[+-]\s*#'                      # → vacío
```

CHANGELOG: comprobar por separado los cuatro elementos de SC-017. README: el aviso del historial va
**antes** del paso `-run`/`-daemon` (SC-016).

---

# PARTE B · Ensayos en Windows y publicación

## W1 · Ensayo en Windows ANTES de la etiqueta (SC-022 · FR-034)

**Quién**: el dueño. **Con qué**: el `permea_*_windows_amd64.zip` de
`goreleaser release --snapshot --clean`, que no publica nada. La versión del snapshot **no será
`0.3.0`** (`research.md` R10).

**Secreto de enrolamiento** *(añadido el 2026-10-02, E-006-P6)*: lo prepara **el dueño** desde la
plataforma, para la instalación de ensayo, y lo deja en `enroll.txt` en la máquina Windows. Es tarea
✋ suya. Claude no lo ve, no lo pide y no lo transcribe a ningún artefacto.

| Paso | Comando (PowerShell) | Esperado | Anotado |
|---|---|---|---|
| 1 | `.\permea.exe --version` | la versión del snapshot, **no** `0.0.1-dev` | ☐ |
| 2 | `Get-Content enroll.txt \| .\permea.exe enroll -` | «enrolado contra …», sin el token | ☐ |
| 3 | `.\permea.exe status` | «enrolado contra … (token: configurado)» | ☐ |
| 4 | `.\permea.exe --scan <log de %USERPROFILE%\.claude\projects\…>` | tantos `evento:` como mensajes distintos de ese fichero | ☐ |

El paso 4 se cuenta desde WSL leyendo el mismo fichero en `/mnt/c/Users/<usuario>/.claude/projects/…`
con el contador independiente. **Sin comprobar** que esa ruta sea accesible desde este WSL.

**Si algún paso falla, la etiqueta no se crea.** Se corrige, se rehace el snapshot y se repite el
ensayo entero. Se anota: fecha, commit del snapshot y resultado de cada paso.

## P1 · Publicación (SC-019 · FR-030)

1. V1 en verde sobre la rama.
2. PR `006-medicion-fiel` → `main`. Se fusiona con merge commit, como #1 y #2.
3. En `main` actualizado y limpio:
   `git tag -a v0.3.0 -m "…"` y `git push origin v0.3.0`.
4. Verificar:
   - `gh release view v0.3.0` (5 archivos y checksums);
   - `gh api repos/permea-dev/scoop-permea/contents/permea.json --jq .content | base64 -d` →
     `"version": "0.3.0"`;
   - el cask en `version "0.3.0"`.

## W2 · Ensayo final en Windows (SC-020 · FR-032)

`scoop update` → `permea --version` = `0.3.0` → `enroll` por stdin → `status` → `-run`. En la
plataforma, el recuento de eventos de esa instalación en la ventana del ensayo es igual al de mensajes
distintos de sus logs en esa ventana, contados como en el contador independiente. Se anota con fecha.

---

## Checklist de cierre

- [ ] V1–V18 en verde, con sus salidas anotadas en `tasks.md`.
- [ ] Q-006-1 decidida (**sí, el 2026-10-02**), y la cabecera de tarifas cita el commit vigente del
  catálogo (`e50d0a5`).
- [ ] W1 (ensayo antes de la etiqueta) anotado, sin fallos.
- [ ] P1: release publicada y canales en `0.3.0`.
- [ ] W2 (ensayo final) anotado.
- [ ] `rm -rf "$COPIA"`, y los temporales de `/tmp/*-006*` y `/tmp/ids-*` borrados.
