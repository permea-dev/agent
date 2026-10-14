# 011 · Registro de rojos, verdes y mutaciones

**Por qué existe**: como en 007–010, `tasks.md` guarda las casillas y una remisión, y aquí van las transcripciones. Techo: **300** líneas
*(tasks §Presupuesto; dueño, 14-10)*.

**Cómo se corre cada suite** *(tasks §Una suite a la vez)*: antes, **en una orden aparte y en segundo plano**, `bash ~/bin/libre.sh` *(md5 `ab647c01…`)*,
hasta «inicio:». Se anotan `/proc/loadavg` y la duración. **Ninguna orden que espere a `libre.sh` puede llevar en su línea de órdenes el nombre de una
suite** *(§B0, incidencia)*.

**§B0, §B1 y §B2** *(FASE 0; raíz y ficheros; una línea e identidad: 848 → 861 → 902 pass, m1–m20 = declarado)*: archivados íntegros el 2026-10-14 en [`registro-archivo.md`](./registro-archivo.md), que no cuenta para el techo.

**§B3, §B4, §B5 y §B6** *(fichero y tramos; integración; `--scan`; textos: 924 → 940 → 953 pass, m21–m42, M-B3a y M-B4a = declarado, con
la enmienda de m31)*: archivados íntegros el 2026-10-14, al empezar el cierre, en [`registro-archivo.md`](./registro-archivo.md).

## Cierre *(plan §Cierre; tasks T041–T049)*

- **T041 · C1 · Puertas** *(libre 1.48; 20:26:14 → :23, 8,4 s)*:
  - **953 / 0 / 0**, rc 0. **Los 848 de la 0.7.0**, por nombre *(los `--- PASS` del log de la línea base de B0)*: los 848 en PASS, ninguno falta.
  - gofmt vacío, vet limpio, lint **0**. **SC-016**: frontera de FR-027 con 0 B de diff; ningún `*_test.go` de `15c4c47` modificado *(sólo `A`)*.
  - **SC-008**: `go.mod` sin `require`, sin `go.sum`, `go list -m all` → 1 línea. `os.Getenv` → 5. `PENDIENTE` → 1.
  - **SC-007 NORMATIVO** *(`git archive 15c4c47` con `0.7.0` frente al árbol con `0.8.0`; banderas de `.goreleaser.yaml`)*: linux/amd64 **+32 768**;
    linux/arm64, Σ `FileSiz` de los `LOAD` **+26 016** *(fichero 0)*; darwin/amd64 **+29 408**; darwin/arm64 **+33 760**; windows/amd64 **+28 672**.
    **Todas ≤ 102 400 B**: la mayor, el 33,0 %. Compilan Windows y darwin.
  - **SC-020**: `PRIV` → **0** en `git diff 15c4c47`; con `grep -raE`, **0** en `cmd/permea/testdata/copilot/` e `internal/ingest/testdata/copilot/`.
  - **Fuera de 011** *(el encargo pidió todos los testdata)*: 8 líneas en testdata de 006 y 008, sin cambios desde `15c4c47`. 7 son ids sintéticos
    *(con el prefijo `msg_` y los nombres CENTINELA, FIXTUREREF y CODEXB5)*. **Las otras 4**, en `internal/ingest/testdata/claude_code_sample.jsonl` *(desde `ee16067`)*,
    llevan una ruta de hogar con un nombre de pila. **No se toca** *(fuera de plan)*: decisión del dueño.
- **T042 · C2 · Medidas** *(guion `c2.py` y contador `contador.py`, `python3 -I`, fuera del repo)*:
  - **Cómo**: `--run` con el binario de HEAD, en un `mktemp -d`, bajo `unshare -rn`, sin enrolar ni `endpoint`, con `COPILOT_HOME` = una copia de
    `session-state` hecha en el temporal; `--scan` y el contador, sobre los originales.
  - **Cada fichero, por su papel**, por igualdad de contenido entre copias, sin imprimir rutas ni ids: A+B y C *(también en `-14`)*, D y E *(sólo en
    `-14b`)*, G *(sólo en `-14c`)*; y F, `session-state-home-alt` de `-14b`.
  - **SC-001** *(`-14b`)*: 9 eventos, 9 `gpt-6-luna`, 9 `event_id` distintos; `tokens_input` **99**, `tokens_output` **323**, lectura **99 938**,
    escritura **39 945**; el resumen de la spec, byte a byte; 0 con coste, 0 de otras herramientas, **0 avisos** *(la carpeta sin `events.jsonl`, sin
    aviso)*.
  - **F**: `--run` sobre su copia y `--scan`: 1 evento, 3 / 5 / 9 692 / 5 674 *(entrada con caché 15 369)*.
  - **SC-002**: Σ de las 4 líneas de control de `-14b` = **616 390 500 nanoAIU → $0.00616390500**, = el contador, = Σ eventos × la tarifa de `gpt-6-luna`.
  - **SC-003**: la 2.ª `--run` encola **0**, y la cola no cambia. Que ningún fichero se abra lo acreditan (15) y (32): aquí no hay `strace`.
  - **SC-004**: D da **1** evento. **SC-009**: E da **4**, con «compactaciones 1».
  - **SC-023**: G da **4** `gpt-6-luna`, ninguno `gpt-5.6-luna`; 12 / 343 / 26 457 / 26 828; control **$0.00379077000**. Con `--run` sobre `-14c`
    *(13 eventos)*, los 4 de G llevan **el mismo** `session_ref`.
  - **`-14`**: 4 eventos, sus `event_id` ⊂ los de `-14b`, con las partidas de A+B y C.
  - **El contraste independiente**: en cada fichero, eventos, partidas *(input − lectura − escritura, output, lectura, escritura)* y créditos, **IGUAL**.
    **C2: TODO IGUAL.** Las copias del temporal no cambian con `--run` *(SC-006)*.
  - **sha256** de todos los ficheros y carpetas de las 3 copias, antes y después: **iguales** *(10, 24 y 24 ficheros; no se publican)*. Temporal
    borrado.
  - **Incidencia** *(sin efecto en el resultado)*: la 1.ª vuelta dio un DISTINTO en «avisos», porque el filtro contaba «omitido» a secas. Era
    `sync omitido: sin endpoint configurado`, la línea normal sin `endpoint`. El filtro pasa a `copilot_cli: fichero omitido` y `skip`, y la 2.ª vuelta
    sale TODO IGUAL. El temporal de la comprobación se borró.
- **T043 · C3 · Snapshot**: `goreleaser release --snapshot --clean` → rc 0, `0.7.0-SNAPSHOT-22c9cf0`.
  - Zip de Windows: sha256 `8be731d5…` *(= `checksums.txt`)*. `permea.exe` `acfb3e20…`, el mismo dentro y fuera del zip.
  - **SC-017**: los 3 textos de spec §Textos, **1 vez cada uno**, byte a byte, en `permea.exe` *(y con `strings -e S`)*.
  - `dist/` borrado; el árbol, limpio *(el `go mod tidy` no cambió nada)*.
- **Material de W1** *(para T044 ✋, fuera del repo, en la carpeta `w1` de la prueba)*:
  - `permea-w1.exe` = el `permea.exe` del snapshot;
  - `contador-w1.exe` y su fuente `contador-w1.go`: Go con sólo la biblioteca estándar, en su propio módulo, sin código del agente. Validado contra
    C2 en las 3 copias y en F;
  - `w1.ps1` *(ASCII, PowerShell 7; parsea sin errores)*, `LEEME-W1.md` y `SHA256SUMS`.
  - **Probado de punta a punta** con una `session-state` sintética *(`COPILOT_HOME` de prueba; `Read-Host` desde stdin, sólo en la copia de prueba)*:
    **W1: TODO IGUAL**. Carpeta de prueba borrada.
  - **Lo que enseñó la prueba**: en el equipo del dueño corre un `copilot-runtime`, y la detección por el nombre `copilot*` daba Copilot CLI por abierto.
    Ahora el script sólo informa de los nombres y pide confirmación.
- **T044 · C4 · W1** *(✋ dueño, 2026-10-14)*: Windows, Copilot CLI 1.0.97, plan Free *(Auto → `gpt-6-luna`)*, sin enrolar y en sandbox; el contador
  independiente, `contador-w1.exe`. **W1: TODO IGUAL**, 0 DISTINTO.
  - **Paso 1** *(Copilot cerrado, sobre una copia de `session-state`)*: **16 eventos = 16 registros** del contador *(14 llamadas + 2 compactaciones)*,
    16 `event_id` distintos. Partidas **192 · 968 · 157 235 · 83 366**, iguales. Créditos **1 249 630 000 nanoAIU** *($0.01249630000)*: iguales en
    `--scan`, en el contador y en tokens × la tarifa de `gpt-6-luna`. **SC-006**: la copia, intacta tras `--run` y tras `--scan`. La 2.ª `--run`, **0**.
  - **Paso 2** *(Copilot abierto, `--daemon`)*: 2 prompts + `/compact` → **3 eventos nuevos = 3 registros nuevos**, con 1 compactación. **19 = 19**,
    cada uno una vez. Partidas **273 · 1 410 · 198 581 · 90 572**, iguales. Créditos **1 403 961 000 nanoAIU** *($0.01403961000)*, iguales en los tres.
    **0 ficheros nuevos del agente** en la carpeta real: el único nuevo, de Copilot.
  - 0 eventos con coste, 0 de otras herramientas, 0 avisos, 0 líneas ilegibles. **SC-022 cumple en Windows.**
- **T045 · C5**: `## 0.8.0 — PENDIENTE` → **`## 0.8.0 — 2026-10-14`** *(dueño, 14-10, como en la 0.7.0)*. `PENDIENTE` → 0; ningún `.go` lee el CHANGELOG.
  La fecha ya estaba en C5 en tasks: no hace falta enmendar C7. El cuerpo del PR, en `tmp/pr-body.md`, fuera del repo *(ignorado por `.gitignore:24`)*.
