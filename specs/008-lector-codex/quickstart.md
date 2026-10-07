# Quickstart · Validación de 008 «Lector de Codex»

**Feature**: `008-lector-codex` | **Fecha**: 2026-10-07 *(E-1, E-2)* | **Forma**: la de `specs/007-coste-fiel/quickstart.md`

Los comandos de las puertas, del contador independiente, de las medidas sobre la copia congelada, del coste de E-1.1 y de los dos ensayos
en Windows. **Nada se ha ejecutado contra el código de 008, porque aún no existe.** El contador se probó el 2026-10-07 sobre la copia
congelada *(`soporte/descubrimiento.md` §FASE 0 (f))*.

## Prerrequisitos

| Herramienta | Versión de la línea base (2026-10-07) |
|---|---|
| Go | `go1.22.2 linux/amd64` |
| `golangci-lint` | **2.12.2**. Con otra versión, se anota |
| `goreleaser` | v2.16.0 |
| `python3` | para el contador y el generador. Se lanza **siempre con `-I`** y desde un directorio que no sea la copia |

## Puertas *(SC-012, FR-022, M-8)*

```sh
gofmt -l .                     # → vacío
go vet ./...                   # → sin hallazgos
golangci-lint run              # → 0 issues
go test -count=1 ./...         # → 9 paquetes ok; 494 + los nuevos, 0 fail
git diff 7b8c77c -- internal/event/ internal/ingest/eventid.go internal/ingest/eventid_test.go \
  internal/ingest/boundary_test.go specs/006-medicion-fiel/contracts/event-id.md      # → vacío
git diff 7b8c77c --name-only --diff-filter=M -- '*_test.go'                            # → vacío (ningún test existente)
git diff 7b8c77c --name-only -- internal/state/                                         # → vacío
grep -rn 'os.Getenv' --include=*.go cmd internal | grep -v _test.go | wc -l             # → 1 (D-008-P7)
md5sum cmd/permea/testdata/codex/referencia-run.stderr                                  # → el transcrito en T030 (SC-011, E-2)
GOOS=windows go build -o /dev/null ./cmd/permea && GOOS=darwin go build -o /dev/null ./cmd/permea
```

## La copia congelada *(disciplina 10)*

8 `.jsonl` *(F1…F8)*. Quien ejecuta recibe su ruta del orquestador y la pone en `COPIA`: es el directorio que contiene `sessions/`.
**No se escribe en ningún documento.** Se lee en su sitio: nada se copia, se escribe ni se borra en ella.

```sh
huella() { (cd "$1" && find . -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum | cut -c1-16; find . -type f | wc -l); }
huella "$COPIA"      # anotar ANTES; repetir al final: debe salir lo mismo
```

Huella del 2026-10-07: **`984d483c12a15601`**, con 8 ficheros.

## El contador independiente *(SC-001 a SC-003)*

Es el de `soporte/descubrimiento.md` §FASE 0 (f), **sin cambios**. No comparte código con el agente y no imprime identificadores.

```sh
T="$(mktemp -d /tmp/permea-008-XXXXXX)"      # scripts, sandbox y salidas; se borra al final tras comprobar el prefijo
# copiar a "$T/contador.py" el bloque PY de descubrimiento.md §FASE 0 (f), literal
(cd "$T" && python3 -I contador.py "$COPIA/sessions") > "$T/contador.txt"
```

**Referencias** *(2026-10-07)*:

| F | Formato | Registros | `tokens_input` | `tokens_cache_creation` | `tokens_cache_read` | `tokens_output` | «tokens used» |
|---|---|---:|---:|---:|---:|---:|---:|
| F1–F4 | anterior | 0 | 0 | 0 | 0 | 0 | — |
| F5 | sin consumo | 0 | 0 | 0 | 0 | 0 | — |
| F6 | actual | 4 | 15 076 | 0 | 51 200 | 88 | 15 164 |
| F7 | actual | 1 | 2 823 | 0 | 11 008 | 5 | **2828** |
| F8 | actual | 2 | 3 880 | 0 | 24 064 | 237 | **4117** |
| **Total** | | **7** | **21 779** | **0** | **86 272** | **330** | |

Si el contador no las reproduce, **se para antes de comparar con el agente**: la copia o el contador no son los de la referencia.

## M · Medidas en sandbox *(SC-001 a SC-005, SC-009)*

```sh
go build -o "$T/permea" ./cmd/permea                     # el binario de la rama; ruta ABSOLUTA: env -i vacía PATH
mkdir -p "$T/hogar"
E="env -i HOME=$T/hogar XDG_CONFIG_HOME=$T/hogar/.config CODEX_HOME=$COPIA"
$E "$T/permea" status                                    # → «no enrolado». Si no, se para
```

**M1 · `--scan` fichero a fichero** *(SC-001, SC-003, SC-004)*:
```sh
find "$COPIA/sessions" -name '*.jsonl' -print0 | sort -z | while IFS= read -r -d '' f; do
  printf '== %s\n' "$(basename "$(dirname "$f")")"; $E "$T/permea" --scan "$f" 2>&1
done > "$T/scan.txt"
awk '/^evento: tool=codex/{n++; for(i=1;i<=NF;i++){split($i,a,"="); s[a[1]]+=a[2]}; if($0 !~ /cost_avail=false/) mal++}
     END{printf "eventos=%d in=%d cw=%d cr=%d out=%d cost_avail_true=%d\n", n, s["in"], s["cw"], s["cr"], s["out"], mal}' "$T/scan.txt"
```
**Esperado**: `eventos=7 in=21779 cw=0 cr=86272 out=330 cost_avail_true=0`. Fichero a fichero, igual a la tabla. En F7, `in + cw + out` = 2828;
en F8, 4117 *(SC-003)*.

**M2 · `--run` dos veces** *(SC-002, SC-005, SC-009)*. Sin endpoint, nada se transmite:
```sh
$E "$T/permea" --run 2> "$T/run1.err"; $E "$T/permea" --run 2> "$T/run2.err"
grep '^codex:' "$T/run1.err" "$T/run2.err"
grep -c -e 'codex: fichero omitido' -e 'skip (línea corrupta)' "$T/run1.err" "$T/run2.err"   # → 0 y 0 (FR-028, FR-029)
python3 -I -c 'import json,sys,collections
q=[json.loads(l) for l in open(sys.argv[1]) if l.strip()]
c=[e for e in q if e.get("tool")=="codex"]
print("codex", len(c), collections.Counter(e["model"] for e in c), "sin_coste", all(not e["cost_available"] and e["cost_usd"]==0 for e in c))' \
  "$T/hogar/.config/permea/queue.jsonl"
```
**Esperado**:
- `run1.err`: `codex: respuestas 7 · eventos 7 · repetidas 0 · sin identificador 0 · incoherentes 0 · sin modelo 0 · ficheros en formato anterior 4 · ficheros comprimidos 0`;
- `run2.err`: `codex: respuestas 0 · eventos 0 · repetidas 0 · sin identificador 0 · incoherentes 0 · sin modelo 0 · ficheros en formato anterior 0 · ficheros comprimidos 0`;
- la cola: `codex 7 Counter({'gpt-6-luna': 7}) sin_coste True`.

## Coste del contexto *(SC-014, E-1.1)*

Un fichero sintético de **≥ 100 MB** en `$T`, **nunca en el repo**. Las líneas de relleno imitan la forma de Codex, con texto
inventado. Se ejecuta una pasada inicial y después **tres** medidas; en cada una se añade **1** registro nuevo y se cronometra `--run`.

```sh
mkdir -p "$T/codex/sessions/2026/10/07" "$T/hogar2"
F="$T/codex/sessions/2026/10/07/rollout-sintetico.jsonl"
cat > "$T/gen.py" <<'PY'
import json, sys
f, n = sys.argv[1], int(sys.argv[2])
def l(t, p, ts='2026-10-07T00:00:00.000Z'): return json.dumps({'timestamp': ts, 'type': t, 'payload': p}) + '\n'
def reg(k): return l('token_usage_record', {'response_id': f'r-{k:024d}', 'session_id': 's-0', 'thread_id': 's-0',
    'turn_id': 't-1', 'root_turn_id': 't-1', 'usage': {'input_tokens': 100, 'cached_input_tokens': 40,
    'cache_write_input_tokens': 0, 'output_tokens': 10, 'reasoning_output_tokens': 0, 'total_tokens': 110}})
with open(f, 'a') as o:
    if n == 0:
        o.write(l('session_meta', {'id': 's-0', 'cwd': '/tmp/proyecto-sintetico', 'cli_version': '0.160.1'}))
        o.write(l('turn_context', {'turn_id': 't-1', 'model': 'modelo-sintetico', 'cwd': '/tmp/proyecto-sintetico'}))
        relleno = l('response_item', {'type': 'message', 'role': 'assistant', 'content': [{'text': 'x' * 900}]})
        for _ in range(110_000): o.write(relleno)          # ≈ 100 MB
        o.write(reg(0))
    else:
        o.write(reg(n))
PY
python3 -I "$T/gen.py" "$F" 0; ls -l "$F" | awk '{print $5}'          # → ≥ 100 000 000
C="env -i HOME=$T/hogar2 XDG_CONFIG_HOME=$T/hogar2/.config CODEX_HOME=$T/codex"
$C "$T/permea" --run 2>/dev/null                                      # pasada inicial: lee todo
for k in 1 2 3; do
  python3 -I "$T/gen.py" "$F" "$k"
  ( TIMEFORMAT=%R; time $C "$T/permea" --run 2> "$T/coste$k.err" )   # → ≤ 3,000 s
  grep '^codex:' "$T/coste$k.err"                                     # → respuestas 1 · eventos 1 · …
done
```
**Se anotan** los tres tiempos y la máquina *(CPU y disco, sin nombre)*. **Si alguno pasa de 3 s, se para**: el plan cambia al plan B de
D-008-P1, y el tope no se mueve.

**Al terminar**: `huella "$COPIA"`, igual que al empezar, y `rm -rf "$T"` tras comprobar que empieza por `/tmp/permea-008-`.

## W1 · Ensayo en Windows ANTES de la etiqueta *(C4 · ✋ el dueño · en sandbox, sin enrolar: 007 E-7)*

Con el `permea_*_windows_amd64.zip` de `goreleaser release --snapshot --clean`, que no publica nada. Un directorio de datos aparte, y
la carpeta real de Codex en **sólo lectura**: el agente no escribe en ella.

| Paso | Comando (PowerShell) | Esperado |
|---|---|---|
| 1 | `.\permea.exe --version` | la del snapshot, **ni** `0.0.1-dev` **ni** `0.4.0`. **Si no, se para** |
| 2 | `$env:APPDATA = "<un directorio aparte>"` y `.\permea.exe status` | «no enrolado». **Si no, no se lanza ningún `--run`** |
| 3 | `$env:CODEX_HOME`; `Test-Path "$env:USERPROFILE\.codex\sessions"` | vacía; `True` *(Q-6: la raíz del agente es la de Codex)* |
| 4 | `Measure-Command { .\permea.exe --run 2> run1.txt }` y `Get-Content run1.txt` | la línea `codex:` con `respuestas` = los registros del contador sobre la misma carpeta *(el dueño lo lanza desde WSL contra `<la carpeta de Codex de Windows>`)*, y `sync omitido`; ninguna línea `codex: fichero omitido` *(FR-028)*. Se anota el tiempo *(R-1)* |
| 5 | Sin usar Codex, `.\permea.exe --run 2> run2.txt` | `codex: respuestas 0 · eventos 0 · …` *(SC-005)* |

Se anotan la fecha, el commit del snapshot, la huella del zip y la salida de cada paso. Si falla un paso, no hay etiqueta.

## W2 · Ensayo final *(C9 · ✋ el dueño, en la instalación real)*

1. `scoop update permea`, y `permea --version` → `0.5.0`.
2. `permea status` → enrolado.
3. `permea --run`: se anotan la línea `codex:` y «N eventos transmitidos y confirmados».
4. En la plataforma, los eventos `tool = codex` de esa instalación = `eventos` de la línea `codex:`, y = los registros del contador sobre la
   misma carpeta menos las repetidas. Todos con el coste ciego y señalizado *(D-3)*.
5. Tras usar Codex, otro `--run`: salen sólo las respuestas nuevas.

## Checklist de cierre

- [x] Puertas en verde *(C1, 2026-10-07: 574 pass, 0 SKIP)*.
- [x] M1 y M2 con las referencias, y la huella de la copia igual antes y después *(C2: `984d483c12a15601`)*.
- [x] SC-014: tres medidas ≤ 3 s, anotadas *(C2: 0,340 · 0,340 · 0,342 s sobre 115 MB)*.
- [x] Snapshot y `strings` con los textos aprobados *(C3: `0.4.0-SNAPSHOT-9516a08`)*.
- [ ] W1 anotado sin fallos *(C4)*.
- [ ] `PENDIENTE` → 0 antes de fusionar *(C5)*.
- [ ] Release y canales en `0.5.0` *(C8)*.
- [ ] W2 anotado *(C9)*.
- [ ] Temporales borrados.
- [ ] Nota para la plataforma: la fila de `gpt-6-luna` *(Dependencias)*.
