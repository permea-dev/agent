# Quickstart · Validación de 007 «Coste fiel»

**Feature**: `007-coste-fiel` | **Fecha**: 2026-10-06 | **Forma**: la de `specs/006-medicion-fiel/quickstart.md`

Los comandos de las puertas, del contador independiente y de las medidas sobre las copias del dueño, más los dos ensayos en Windows.
**Nada se ha ejecutado contra el código de 007, porque aún no existe.** El contador y la comprobación del coste se probaron el
2026-10-06 sobre datos **sintéticos**, no sobre las copias.

## Prerrequisitos

| Herramienta | Versión de la línea base (2026-10-06) |
|---|---|
| Go | `go1.22.2 linux/amd64` |
| `golangci-lint` | **2.12.2**. Con otra versión, se anota |
| `goreleaser` | v2.16.0 |
| `python3` | para el contador. Se lanza **siempre con `-I`** y desde un directorio que no sea la copia |

## Puertas *(SC-012, SC-009, FR-018)*

```sh
gofmt -l .                     # → vacío
go vet ./...                   # → sin hallazgos
golangci-lint run              # → 0 issues
go test -count=1 ./...         # → 9 paquetes ok; 432 + los nuevos, 0 fail
git diff 222c824 -- internal/event/ internal/ingest/eventid.go internal/ingest/eventid_test.go \
  internal/ingest/boundary_test.go specs/006-medicion-fiel/contracts/event-id.md      # → vacío
git diff 222c824 -- internal/state/state_test.go                                       # → vacío (D-007-P3)
grep -rn nolint --include=*.go . | wc -l                                               # → 1 (la de SA1007)
grep -n '8f147d1' internal/pricing/pricing.go                                          # → la cabecera
GOOS=windows go build -o /dev/null ./cmd/permea && GOOS=darwin go build -o /dev/null ./cmd/permea
```

**Q-T · El espejo, a mano** *(contrato §La vigilancia)*. Desde el repositorio de la plataforma:
`git show 8f147d1:backend/config/pricing.php`. Las 17 filas, con sus cinco cifras, deben coincidir con `contracts/tarifas.md` y con
`esperadaDelCatalogo`.

## Las copias del dueño *(disciplina 10)*

Son la **copia del dueño 2026-10-06-wsl** *(31 `.jsonl`)* y la **-windows** *(29)*. Quien ejecuta recibe sus rutas del orquestador y las
pone en `COPIA_WSL` y `COPIA_WIN`. **No se escriben en ningún documento.** Se leen en su sitio: nada se copia, se escribe ni se borra en
ellas.

```sh
huella() { (cd "$1" && find . -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum | cut -c1-16; find . -name '*.jsonl' | wc -l); }
huella "$COPIA_WSL"; huella "$COPIA_WIN"      # anotar ANTES; repetir al final: debe salir lo mismo
```

Huellas del 2026-10-06 *(E-1)*: -wsl `82d79809411f61a3` con 31 `.jsonl`; -windows `d65f53cf7c56325c` con 29.

## El contador independiente *(SC-001, SC-002, SC-003 y SC-005)*

Corrige al de 006, que aplicaba «la primera manda» y no era independiente en ese punto *(Hallazgo de W2)*. Éste aplica el **máximo
por partida** *(FR-009)*. El desglose lo toma de la línea que da el máximo de la escritura, la última si empatan. Sin desglose, o con
uno que no suma, va todo a 1 hora *(P-1, Q-4)*. **No comparte código con el agente** y **no imprime ningún identificador**.

```sh
T="$(mktemp -d)"                       # aquí viven los scripts y las salidas; se borra al final
cat > "$T/contador.py" <<'PY'
import json, os, sys
raiz = sys.argv[1]
ficheros = fact = sint = sin_id = corr = sin_des = no_suma = 0
msgs = {}
for dp, _, fs in os.walk(raiz):
    for n in sorted(fs):
        if not n.endswith('.jsonl'):
            continue
        ficheros += 1
        with open(os.path.join(dp, n), 'rb') as fh:
            for raw in fh:                      # como `--scan`: también la última línea sin salto
                try:
                    r = json.loads(raw)
                except Exception:
                    corr += 1
                    continue
                if not isinstance(r, dict) or r.get('type') != 'assistant':
                    continue
                m = r.get('message') if isinstance(r.get('message'), dict) else {}
                if not m.get('model'):
                    continue
                fact += 1
                if m['model'] == '<synthetic>':
                    sint += 1
                    continue
                clave = (m.get('id') or '', r.get('requestId') or '')
                if clave == ('', ''):
                    sin_id += 1
                    continue
                u = m.get('usage') or {}
                p = [u.get(k) or 0 for k in ('input_tokens', 'output_tokens',
                                             'cache_creation_input_tokens', 'cache_read_input_tokens')]
                cc = u.get('cache_creation')
                if isinstance(cc, dict) and 'ephemeral_5m_input_tokens' in cc and 'ephemeral_1h_input_tokens' in cc:
                    d = (cc['ephemeral_5m_input_tokens'] or 0, cc['ephemeral_1h_input_tokens'] or 0)
                    if sum(d) != p[2]:
                        no_suma += 1
                        d = (0, p[2])                   # Q-4 (a): como sin desglose
                else:
                    sin_des += 1
                    d = (0, p[2])                       # P-1: todo a 1 hora
                e = msgs.get(clave)
                if e is None:
                    msgs[clave] = {'primera': p, 'max': list(p), 'des': d}
                    continue
                if p[2] >= e['max'][2]:
                    e['des'] = d                        # la línea del máximo de escritura; la última si empatan
                e['max'] = [max(a, b) for a, b in zip(e['max'], p)]
crecen = sum(1 for e in msgs.values() if e['max'] != e['primera'])
smax = [sum(e['max'][i] for e in msgs.values()) for i in range(4)]
print(f"ficheros={ficheros} facturables={fact} sinteticas={sint} sin_identificador={sin_id} corruptas={corr}")
print(f"mensajes={len(msgs)} crecen={crecen} lineas_sin_desglose={sin_des} lineas_desglose_no_suma={no_suma}")
print(f"maximo: in={smax[0]} out={smax[1]} cw={smax[2]} cr={smax[3]}")
print(f"primera: out={sum(e['primera'][1] for e in msgs.values())}")
print(f"desglose: cw5m={sum(e['des'][0] for e in msgs.values())} cw1h={sum(e['des'][1] for e in msgs.values())}")
PY
(cd "$T" && python3 -I contador.py "$COPIA_WSL") > "$T/contador-wsl.txt"
(cd "$T" && python3 -I contador.py "$COPIA_WIN") > "$T/contador-win.txt"
```

**Referencias de la spec** *(E-1)*:

| Copia | mensajes | in | out *(máximo)* | out *(primera)* | cw | cw5m | cw1h | cr | crecen |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| -wsl | 10 121 | 26 392 | 11 713 955 | 11 713 955 | 32 765 802 | 247 506 | 32 518 296 | 4 499 158 833 | 0 |
| -windows | 6 074 | 12 206 | 6 723 801 | 6 635 290 | 22 307 249 | 971 559 | 21 335 690 | 2 423 738 410 | 143 |

Si el contador no las reproduce, **se para antes de comparar con el agente**: la copia o el contador no son los de la referencia.

## M · Medidas con `--scan` en un sandbox

```sh
go build -o "$T/permea" ./cmd/permea                     # el binario de la rama; ruta ABSOLUTA: env -i vacía PATH
mkdir -p "$T/sandbox/home"
E="env -i HOME=$T/sandbox/home XDG_CONFIG_HOME=$T/sandbox/home/.config"
$E "$T/permea" status                                   # → «no enrolado». Si no, se para
for n in wsl win; do
  [ $n = wsl ] && C="$COPIA_WSL" || C="$COPIA_WIN"
  find "$C" -name '*.jsonl' -print0 | sort -z | while IFS= read -r -d '' f; do
    $E "$T/permea" --scan "$f" 2>>"$T/scan-$n.err"
  done > "$T/scan-$n.txt"
done
find "$T/sandbox" -type f | wc -l                       # → 0: `--scan` no escribe nada
```

**M1 · SC-001** *(-wsl)*: las sumas de los `evento:` = las del contador en las cuatro partidas.
```sh
awk '/^evento:/{n++; for(i=1;i<=NF;i++){split($i,a,"="); s[a[1]]+=a[2]}}
     END{printf "eventos=%d in=%d out=%d cw=%d cw5m=%d cw1h=%d cr=%d\n", n, s["in"], s["out"], s["cw"], s["cw5m"], s["cw1h"], s["cr"]}' "$T/scan-wsl.txt"
```
**M2 · SC-002** *(-windows)*: el mismo `awk` sobre `scan-win.txt`. Salida = **6 723 801**, no 6 635 290, y las otras tres partidas
iguales al contador.

**M3 · SC-003**: `eventos=` = `mensajes=` del contador, y ningún `event_id` repetido. El recuento no se imprime con identificadores:
```sh
grep -o 'event_id=[0-9a-f]*' "$T/scan-wsl.txt" | sort | uniq -d | wc -l    # → 0
```
Falsable, sobre un fichero **sintético**, nunca sobre la copia: duplicar una línea `assistant` no cambia el recuento *(lo cubre T021)*.

**M4 · SC-005** *(-wsl)*: el coste de cada evento, recalculado con la tabla del contrato **en aritmética decimal exacta** *(E-6)*. El
coste exacto se calcula con `Decimal` a partir de los tokens y la tabla, sin pasar por `float`. El `--scan` imprime 4 decimales, así que un
evento es **distinto** sólo si |impreso − exacto| > 0,00005. La igualdad, |impreso − exacto| = 0,00005, es un **empate de redondeo** y
vale: se cuenta aparte.
```sh
cat > "$T/coste.py" <<'PY'
import re, sys
from decimal import Decimal
T = {  # contracts/tarifas.md, 8f147d1: input, output, cache_write (5 min), cache_write_1h, cache_read — como CADENAS, sin float
    'claude-fable-5': ('10', '50', '12.5', '20', '1'), 'claude-fable-5-1': ('10', '50', '12.5', '20', '0.25'),
    'claude-mythos-5': ('10', '50', '12.5', '20', '1'), 'claude-opus-5-5': ('4', '20', '5', '8', '0.2'),
    'claude-opus-5': ('5', '25', '6.25', '10', '0.5'), 'claude-opus-4-8': ('5', '25', '6.25', '10', '0.5'),
    'claude-opus-4-7': ('5', '25', '6.25', '10', '0.5'), 'claude-opus-4-6': ('5', '25', '6.25', '10', '0.5'),
    'claude-opus-4-5': ('5', '25', '6.25', '10', '0.5'), 'claude-opus-4-1': ('15', '75', '18.75', '30', '1.5'),
    'claude-opus-4': ('15', '75', '18.75', '30', '1.5'), 'claude-sonnet-5': ('2', '10', '2.5', '4', '0.2'),
    'claude-sonnet-4-6': ('3', '15', '3.75', '6', '0.3'), 'claude-sonnet-4-5': ('3', '15', '3.75', '6', '0.3'),
    'claude-sonnet-4': ('3', '15', '3.75', '6', '0.3'), 'claude-haiku-4-5': ('1', '5', '1.25', '2', '0.1'),
    'claude-haiku-3-5': ('0.8', '4', '1', '1.6', '0.08')}
TOLERANCIA = Decimal('0.00005')
eventos = distintos = empates = sin_tarifa = 0
for linea in open(sys.argv[1], encoding='utf-8'):
    if not linea.startswith('evento:'):
        continue
    eventos += 1
    f = dict(re.findall(r'(\w+)=(\S+)', linea))
    tokens = [int(f[k]) for k in ('in', 'out', 'cw5m', 'cw1h', 'cr')]
    if f['model'] not in T:
        sin_tarifa += 1
        exacto = Decimal(0)
    else:
        exacto = sum(Decimal(t) * Decimal(r) for t, r in zip(tokens, T[f['model']])) / Decimal(10**6)
    diferencia = abs(Decimal(f['cost'].lstrip('$')) - exacto)
    if diferencia > TOLERANCIA:
        distintos += 1
    elif diferencia == TOLERANCIA:
        empates += 1                     # el exacto acaba en 5 en la quinta cifra: vale cualquiera de los dos lados
print(f"eventos={eventos} coste_distinto={distintos} empates={empates} sin_tarifa={sin_tarifa}")
PY
(cd "$T" && python3 -I coste.py "$T/scan-wsl.txt")      # → coste_distinto=0 (empates: 31 en -wsl y 18 en -windows, medidos el 2026-10-06)
```

**Al terminar**: `huella` de las dos copias, igual que al empezar, y después `rm -rf "$T"`. Sólo se transcriben a `tasks.md` los recuentos
y las sumas.

## W1 · Ensayo en Windows ANTES de la etiqueta *(C4 · ✋ el dueño)*

Con el `permea_*_windows_amd64.zip` de `goreleaser release --snapshot --clean`, que no publica nada.

| Paso | Comando (PowerShell) | Esperado |
|---|---|---|
| 1 | `.\permea.exe --version` | la del snapshot, **ni** `0.0.1-dev` **ni** `0.3.0`. **Si no, se para y no se lanza nada más** |
| 2 | `.\permea.exe status` | enrolado, token configurado |
| 3 | `.\permea.exe --scan <un log de subagente>` | tantos `evento:` como mensajes del contador, y `cw5m=`/`cw1h=` en cada línea |
| 4 | `.\permea.exe --run` | primera línea del resumen como la de 0.3.0; segunda con «N en espera de cerrarse»; y el aviso de `--run` si N > 0 |
| 5 | > 10 min **sin usar Claude Code**, y `.\permea.exe --run` | «en espera» baja a 0 y salen esos N. Si no baja, el `mtime` de NTFS **no** se comporta como se espera *(R-1, R-2)*, y se para |

Se anotan la fecha, el commit del snapshot, la huella del zip y el resultado de cada paso. Si falla un paso, no hay etiqueta.

## W2 · Ensayo final *(C9 · ✋ el dueño)*

1. `scoop update permea`, y `permea --version` → `0.4.0`.
2. `permea status`.
3. `permea --run`, anotando **E** = «en espera» de la segunda línea del resumen.
4. En la plataforma, los eventos `0.4.0` de esa instalación en la ventana = los mensajes distintos de la ventana, contados con el
   contador **menos E**.
5. Tras > 10 min sin usar Claude Code, otro `--run`. El recuento debe cuadrar con todos, y la salida de la ventana coincidir con
   el **máximo** del contador.

## Checklist de cierre

- [ ] Puertas en verde *(C1)*.
- [ ] M1–M4 con las referencias, y las huellas de las copias iguales antes y después *(C2)*.
- [ ] Snapshot y `strings` con los textos aprobados *(C3)*.
- [ ] W1 anotado sin fallos *(C4)*.
- [ ] `PENDIENTE` → 0 antes de fusionar *(C5)*.
- [ ] Release y canales en `0.4.0` *(C8)*.
- [ ] W2 anotado *(C9)*.
- [ ] Temporales borrados.
- [ ] Nota para la plataforma: la cabecera de `pricing.php` *(N-6)*.
