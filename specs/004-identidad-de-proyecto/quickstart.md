# Quickstart · Validación de P-004 Identidad de proyecto

**Fecha**: 2026-08-09 · **Spec**: [spec.md](./spec.md) · **Contratos**:
[project-identity.md](./contracts/project-identity.md) · [cli-config.md](./contracts/cli-config.md)

Guía de validación ejecutable: qué hay que poder demostrar para dar la feature por buena. **No** trae
código de implementación — eso es de `tasks.md`.

---

## Prerrequisitos

```bash
cd ~/dev/agente
go version          # toolchain del repo
make build          # binario en bin/permea
```

Las puertas de calidad de la constitución, que deben quedar en verde al final:

```bash
go vet ./...
golangci-lint run
go test ./...
```

### Aislamiento obligatorio — antes de cualquier validación que toque config o cola

**Ninguna validación de esta guía puede ejecutarse contra la instalación real del desarrollador.**
`config.DataDir()` resuelve por `os.UserConfigDir()` (`internal/config/config.go:51`), que en
Linux/macOS depende de `HOME` (vía `XDG_CONFIG_HOME` o `$HOME/.config`) y en Windows de
`USERPROFILE`/`%AppData%`. Basta con apuntar esas variables a un directorio de pruebas y **todo** el
estado del agente —`config.json`, `state.json`, `queue.jsonl`, `salt`, `machine-id`— aterriza allí.

```bash
# Sandbox de validación: se crea uno por escenario y se descarta al terminar.
export PERMEA_SANDBOX="$(mktemp -d)"
export HOME="$PERMEA_SANDBOX/home"          # Linux / macOS
export USERPROFILE="$PERMEA_SANDBOX/home"   # Windows
export XDG_CONFIG_HOME="$HOME/.config"      # explícito: no depender del valor heredado
mkdir -p "$XDG_CONFIG_HOME"

# Comprobación de que el aislamiento funciona ANTES de seguir:
./bin/permea status      # debe decir "no enrolado" — si dice lo contrario, HOME no se aplicó
```

**Esa última comprobación no es ceremonia.** Si el sandbox no está activo, V9 sobrescribe el
`config.json` real del desarrollador y V0 lee una cola que no es la del escenario. Un `status` que
responde «no enrolado» en una máquina enrolada es la prueba de que el aislamiento está puesto.

> **Regla**: las validaciones que **solo** usan `--scan` sobre un fichero (V1–V6) no tocan config ni
> cola, pero se ejecutan igualmente dentro del sandbox — cuesta cero y elimina la pregunta.

---

## V0 · Línea base — **antes de tocar código**

> **Restricción de orden**: esta validación es la **primera tarea** de la implementación. Si se hace
> después, SC-004 no tiene contra qué medirse.

Hay que capturar las identidades de **sesión** y **máquina** que produce el fixture actual. Existen
dos vías, y **cuál se usa lo decide una verificación, no una suposición**:

### Puerta de decisión (la resuelve la fase de tasks, contra el código)

> **¿Imprime `--scan` los tres refs?**
> ```bash
> grep -n "project_ref\|session_ref\|machine_ref" cmd/permea/main.go
> ```

**Verificado hoy contra `2ba8363`: NO.** `cmd/permea/main.go:333-334` imprime **solo**
`project_ref`, y además **truncado a 8 caracteres** (`main.go:329-331`). Ni `session_ref` ni
`machine_ref` aparecen en ninguna salida.

Luego **la vía aplicable hoy es la B**. La vía A queda escrita porque la puerta debe reevaluarse si
alguien amplía la salida de `--scan` antes de esta tarea.

### Vía A — desde `--scan` *(solo si la puerta da positivo)*

```bash
./bin/permea --scan internal/ingest/testdata/claude_code_sample.jsonl > baseline-refs.txt
```

Requiere que la salida traiga los **tres** refs **completos**, sin truncar. Con el código actual no
se cumple ninguna de las dos condiciones.

### Vía B — desde `queue.jsonl` de una pasada real aislada ✅ *(la aplicable)*

Usa el sandbox de los prerrequisitos, de modo que la pasada no toque la instalación real:

```bash
git rev-parse --short HEAD          # el commit PREVIO a cualquier cambio de comportamiento
make build

# Dentro del sandbox (ver Prerrequisitos). ANTES de la pasada, dos siembras:
#   1) logs_root → un directorio EFÍMERO del sandbox, con UNA copia fresca de
#      internal/ingest/testdata/claude_code_sample.jsonl y nada más.
#   2) `salt` y `machine_id` con las semillas del bloque REPRODUCCIÓN de
#      specs/004-identidad-de-proyecto/baseline-sc004.tsv.
mkdir -p "$PERMEA_SANDBOX/logs"
cp internal/ingest/testdata/claude_code_sample.jsonl "$PERMEA_SANDBOX/logs/"

./bin/permea --run                  # exit != 0 esperado: el envío falla a propósito

# La cola contiene los eventos completos, con los tres refs sin truncar:
jq -r '[.project_ref, .session_ref, .machine_ref] | @tsv' "$XDG_CONFIG_HOME/permea/queue.jsonl" \
  | sort -u   # → filas del .tsv;  y aparte:  wc -l queue.jsonl → events_total
```

**Las semillas no son un detalle de montaje**: `LoadOrCreateSalt` genera un valor **aleatorio** si el
fichero falta (`internal/config/identity.go:13-14`), y como los tres refs son `Ref(salt, valor)`, una
línea base con salt aleatorio **no se puede reproducir** — V8 no tendría con qué comparar. Los valores
viven en el **bloque REPRODUCCIÓN del propio `.tsv`**, que es su única fuente de verdad; aquí no se
copian para que no puedan divergir.

**Y el `logs_root` efímero tampoco es montaje**: `state.FindLogs` recorre **todos** los `.jsonl` bajo
la raíz que se le dé (`internal/state/scan.go:20-31`). Apuntar a `internal/ingest/testdata/` haría que
cualquier fixture que alguien añada ahí entrara en la pasada — **ya pasó** con el fixture de frontera
de T004, que convirtió 2 eventos en 4 y bloqueó una verificación de neutralidad. Congelar el fichero
no bastaba: **hay que congelar el directorio**, y un directorio efímero que solo contiene lo que la
receta copia lo está por construcción.

**Por qué la cola y no la salida por pantalla**: `queue.jsonl` contiene el **evento serializado
completo** —es literalmente lo que cruzaría la frontera—, así que la línea base se toma del mismo
artefacto que V8 comparará después. La salida de `--scan` es un resumen para humanos y no sirve como
referencia byte a byte.

**Resultado esperado**: `baseline-sc004.tsv` **commiteado** antes del primer cambio de
comportamiento, con las identidades previas al cambio **y** el recuento de eventos.

---

## V1 · El proyecto agrupa (US1 · SC-002)

Escenario mínimo reproducible, sobre repositorios desechables:

```bash
T=$(mktemp -d)
git init -q "$T/proy" && mkdir -p "$T/proy/frontend/src"
```

Generar un fixture con dos eventos: uno con `"cwd":"$T/proy"` y otro con
`"cwd":"$T/proy/frontend/src"`.

```bash
./bin/permea --scan "$T/fixture.jsonl"
```

**Esperado**: los dos eventos muestran **el mismo** `project_ref`.

**Contra-prueba obligatoria** (si no, el test pasa por accidente): un tercer evento con `cwd` en un
proyecto **distinto** debe dar un `project_ref` **diferente** (G4).

---

## V2 · Árboles paralelos — el caso que se rompe solo (edge case · G-caso 5)

```bash
cd "$T/proy" && git -c user.email=t@t -c user.name=t commit -q --allow-empty -m base
git worktree add -q "$T/paralelo"
stat -c '%F' "$T/proy/.git"       # directory
stat -c '%F' "$T/paralelo/.git"   # regular file   ← el marcador NO es un directorio
```

Fixture con un evento en `$T/proy/…` y otro en `$T/paralelo/…`.

**Esperado**: `project_ref` **distintos**, uno por árbol.

**Por qué esta validación es la más importante de la feature**: una implementación que compruebe
«existe un **directorio** `.git`» falla aquí — el worktree caería al fallback. Y si el worktree está
fuera de otro repositorio, el resultado seguiría siendo «dos identidades distintas» **por el motivo
equivocado**, así que la prueba debe verificar además que la identidad del worktree es la de **su
propia raíz** (dos eventos dentro del worktree, en raíz y en subdirectorio, comparten identidad).

---

## V3 · Directorio suelto y variaciones sintácticas (US2 · G6)

```bash
mkdir -p "$T/suelto"
```

Fixture con `"cwd":"$T/suelto/"`, `"cwd":"$T/suelto"` y `"cwd":"$T/suelto/./"`.

**Esperado**: **una sola** identidad para los tres.

---

## V4 · Enlace hacia el interior de un proyecto (US1 esc. 6 · G5)

```bash
ln -s "$T/proy/frontend" "$T/enlace"
```

Fixture con `"cwd":"$T/enlace/src"` y `"cwd":"$T/proy/frontend/src"`.

**Esperado**: **misma** identidad, y además **igual a la de V1** — porque la resolución de enlaces
precede al reconocimiento de raíz (FR-006a), así que ambos aterrizan en la raíz `$T/proy`.

---

## V5 · Mejor esfuerzo: nada interrumpe la emisión (SC-003 · G8)

Tres fixtures, cada uno con un `cwd` imposible:

| Caso | `cwd` |
|---|---|
| Inexistente | `/no/existe/jamas` |
| Enlace roto | un symlink cuyo destino se borró |
| Sin permisos | un directorio con permisos denegados (omitir si se ejecuta como root) |

```bash
./bin/permea --scan "$T/degradado.jsonl"; echo "exit=$?"
```

**Esperado**: **exit 0**, un evento por cada línea facturable, `project_ref` no vacío en los tres.
Cero eventos perdidos — que es literalmente SC-003.

---

## V6 · Home y raíz del sistema no son proyectos (FR-004a)

**La técnica es un HOME falso que SÍ contiene un `.git`** — el caso realista que el requisito existe
para cubrir (quien versiona su configuración personal). Como el home se resuelve por
`os.UserHomeDir()`, basta con apuntar la variable de entorno: nunca se toca el directorio personal
real.

```bash
FAKEHOME="$PERMEA_SANDBOX/fakehome"
mkdir -p "$FAKEHOME/dev/proy/sub" "$FAKEHOME/suelto-a" "$FAKEHOME/suelto-b"
git init -q "$FAKEHOME"              # ← el HOME es él mismo un repositorio
git init -q "$FAKEHOME/dev/proy"     # ← y hay un proyecto genuino debajo

HOME="$FAKEHOME" USERPROFILE="$FAKEHOME" ./bin/permea --scan "$PERMEA_SANDBOX/v6.jsonl"
```

Fixture `v6.jsonl` con cuatro eventos:

| `cwd` | Esperado | Por qué |
|---|---|---|
| `$FAKEHOME/suelto-a` | identidad **propia** | techo alcanzado sin marcador → fallback normalizado |
| `$FAKEHOME/suelto-b` | identidad **propia y distinta de la anterior** | **la contra-prueba**: si coincidieran, el home se estaría tratando como proyecto |
| `$FAKEHOME/dev/proy/sub` | identidad **de `$FAKEHOME/dev/proy`** | **techo, no salto**: un marcador bajo el home cuenta con total normalidad |
| `$FAKEHOME` | identidad **propia** (del propio directorio, normalizado) | el techo no se examina ni siendo él mismo el `cwd` |

**Las dos primeras filas son la prueba de fuego y las dos últimas son las que impiden pasarla por el
motivo equivocado.** Una implementación que simplemente ignorase todo lo que cuelga del home pasaría
las dos primeras y fallaría la tercera — y esa tercera es la que distingue «techo» de «zona
prohibida».

**Raíz del sistema**: no se valida por entorno (no hay variable que la falsee sin privilegios). Se
cubre con test unitario del resolutor, inyectando el techo como parámetro.

---

## V7 · La frontera no se mueve (SC-005 · G9)

```bash
go test ./internal/ingest -run TestBoundary -v
go test ./internal/event -run TestEvent_OnlyAllowlistKeys -v
```

**Esperado**: verde, con la denylist **ampliada** con fragmentos de la raíz resuelta y del
subdirectorio de lanzamiento, y con el alcance extendido al evento serializado, `queue.jsonl` y el
cuerpo transmitido (`httptest.NewTLSServer`).

**La prueba que importa**: que la **ruta de la raíz** —un valor que antes no circulaba por el
proceso— no aparezca por ningún camino.

---

## V8 · Regresión cero en sesión y máquina (SC-004) y recuento de eventos (SC-003)

**Se repite la pasada de V0, no un `--scan`**: `--scan` solo imprime `project_ref`, y truncado a 8
caracteres (`cmd/permea/main.go:329-334`), así que no sirve para comparar byte a byte. Misma receta,
mismo sandbox y **las mismas semillas** del bloque REPRODUCCIÓN de `baseline-sc004.tsv` — con otras,
los refs no comparan y una discrepancia no significaría nada.

```bash
# Sandbox + config + semillas idénticas a V0, y después:
./bin/permea --run                  # exit != 0 esperado (envío fallido a propósito)

jq -r '[.project_ref, .session_ref, .machine_ref] | @tsv' "$XDG_CONFIG_HOME/permea/queue.jsonl" | sort -u
wc -l < "$XDG_CONFIG_HOME/permea/queue.jsonl"
```

**Esperado, en las dos dimensiones que el artefacto guarda**:

- **Conjunto** (SC-004 / FR-019): las identidades de **sesión** y **máquina** coinciden **byte a
  byte** con las filas de `baseline-sc004.tsv`. Si alguna cambió, se ha tocado algo que FR-019
  prohíbe tocar.
- **Recuento** (SC-003): el número de líneas de `queue.jsonl` coincide con `events_total` del
  artefacto. Si bajó, se han perdido eventos — y el `sort -u` de la primera comprobación **no lo
  habría detectado**, porque colapsa los eventos que comparten refs.

*(El `project_ref` de la línea base es para la neutralidad de T007, no para V8: aquí ya cambió a
propósito, y esa es toda la feature.)*

---

## V9 · La parada del modo retirado (SC-007)

**Se ejecuta entero dentro del sandbox de los prerrequisitos.** `DATADIR` es el que resuelve el
agente bajo el `HOME` de pruebas — nunca se edita el `config.json` real:

```bash
DATADIR="$XDG_CONFIG_HOME/permea"       # el mismo que resuelve config.DataDir() con el HOME falso
mkdir -p "$DATADIR"

# Caso parada
echo '{"project_ref_mode":"plain"}' > "$DATADIR/config.json"
./bin/permea --run; echo "exit=$?"      # esperado: exit != 0, mensaje con clave+valor+ruta
wc -l "$DATADIR/queue.jsonl" 2>/dev/null || echo "0"   # esperado: sin líneas nuevas

# Caso limpio (valor ya satisfecho)
echo '{"project_ref_mode":"hash"}' > "$DATADIR/config.json"
./bin/permea --run; echo "exit=$?"      # esperado: exit 0, SIN aviso ni error

# Caso limpio (clave ausente)
echo '{}' > "$DATADIR/config.json"
./bin/permea --run; echo "exit=$?"      # esperado: exit 0
```

Los dos últimos casos parecen redundantes y no lo son: separan «ignoro el valor» (FR-013a) de
«ignoro la clave» (FR-013b).

**Y la diferencia entre invocaciones**, que es lo que fija **D-004-5**:

```bash
echo '{"project_ref_mode":"plain"}' > "$DATADIR/config.json"

./bin/permea status; echo "exit=$?"     # esperado: exit 0 + aviso visible (diagnóstico)
./bin/permea --scan "$PERMEA_SANDBOX/v1.jsonl"; echo "exit=$?"
                                        # esperado: exit 0, SIN parada — procesamiento diagnóstico
```

**El caso de `--scan` con `"plain"` presente es el que ancla D-004-5** y conviene no saltárselo: la
invocación **procesa** líneas con una configuración que contiene el valor retirado, y aun así no se
detiene, porque su procesamiento no alcanza la frontera (salt de dry-run, sin cola, sin transporte).
Si alguna vez esta comprobación empezara a fallar, la pregunta correcta no es «arreglar el test» sino
«¿ha dejado `--scan` de ser diagnóstico?».

---

## V10 · El barrido documental (SC-008 · FR-014)

**Con cuatro términos, no uno** — con solo `plain` se escapan **3 de los 6 sitios en alcance**.

**Alcance: la documentación VERSIONADA del repositorio.** El barrido se hace sobre ficheros que están
en el índice de git, porque «documentación del repositorio» es lo que un lector del repositorio puede
leer. Se excluye `Roadmap.md`, que está **gitignoreado** (`.gitignore:17`) y no forma parte de lo
publicado (ver nota en `research.md`).

```bash
# El alcance sale de git, no de una lista escrita a mano:
DOCS=$(git ls-files '*.md' | grep -v '^\.specify/\|^\.github/')

grep -rni "plain"            $DOCS
grep -rni "opt-in"           $DOCS
grep -rn  "project_ref_mode" $DOCS $(git ls-files '*.json')
grep -rn  "en claro"         $DOCS
```

**Esperado**: cero menciones de un modo de envío en claro de la identidad de proyecto, **incluido el
README** (que hoy ya es correcto: `README.md:13-14` promete solo hash salado). Los sitios que el
barrido debe haber tocado están enumerados en `research.md` §«Alcance documental real».

**Control de que el barrido funciona**: ejecutado **antes** del cambio, los cuatro comandos deben
reencontrar los seis sitios versionados conocidos. Si no aparecen, el barrido está mal construido y
su «cero» posterior no significaría nada.

**Exclusiones y por qué**: `.specify/` y `.github/` son plantillas de la herramienta —sus «plain
language» y «Explain» son falsos positivos, no promesas del producto—; `Roadmap.md` no está
versionado.

---

## V11 · Coste de la resolución (SC-006 — orientativo, NUNCA puerta de CI)

```bash
time ./bin/permea --scan <log-grande>
```

**Esperado**: sin incremento perceptible respecto de la línea base. **Este número no bloquea nada**:
SC-006 se garantiza **por diseño** (una resolución por directorio distinto, no por evento), y los
tests de tiempo son la familia de flakes ya fichada. La verificación real es leer que la caché existe
y que su clave es el `cwd` declarado.

---

## Validación bloqueada

**SC-001** no se puede validar hasta que exista la **enumeración de verdad-terreno** de los 12
directorios de origen (spec, Assumptions; formato propuesto en `research.md` §P7). Bloquea la
validación, no la implementación.

---

## Checklist de cierre

- [ ] V0 capturado **antes** del primer cambio de código
- [ ] V1–V6 en verde (derivación)
- [ ] V7, V8 en verde (frontera y regresión cero)
- [ ] V9 en verde (parada y silencios)
- [ ] V10: barrido con cuatro términos, cero menciones
- [ ] `go vet` / `golangci-lint run` / `go test ./...` en verde
- [ ] SC-001 pendiente de la verdad-terreno — **declarado, no olvidado**
