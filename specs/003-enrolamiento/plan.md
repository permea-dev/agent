# Implementation Plan: Enrolamiento de dispositivo — emparejar el agente `permea` con su backend

**Branch**: `003-enrolamiento` | **Date**: 2026-07-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-enrolamiento/spec.md`

## Summary

Añadir al binario `permea` dos comandos de usuario —`permea enroll <enrollment-string>` y
`permea status`— que materializan el emparejamiento agente↔backend por el lado del agente.
`enroll` recibe un **enrollment string** (envoltorio que empaqueta la URL del backend y el
device token en claro, emitido y revelado una sola vez por P-002), lo **decodifica** en URL +
token, y **verifica** el token con un **ping de ingesta de lote vacío** contra el `/ingest`
existente (2xx → confirmado; 401/403 → rechazado). Solo un token verificado se **persiste** en
`config.json` (permisos `0600`, ruta resuelta por SO). `status` informa si el agente está
enrolado y contra qué backend (URL), **nunca** el token.

La feature **reutiliza** infraestructura ya construida en P-001: los campos `Endpoint` y
`DeviceToken` ya existen en `internal/config.Config`; `config.Save` ya escribe `0600` de forma
atómica; `transport.Client.Send([]event.Event{})` ya produce un lote vacío `[]` y clasifica
`2xx`/`401`/`403` según `contracts/transport.md`. Por tanto 003 **no** inventa endpoint nuevo,
**no** toca la frontera (Principio I) y **no** añade dependencias. El único artefacto de datos
nuevo es el **formato del enrollment string** (`contracts/enrollment-string.md`), que 003 define
por ser el lado que lo decodifica y que P-002 consumirá para emitirlo.

## Technical Context

**Language/Version**: Go 1.22 (módulo `github.com/permea-dev/agent`). Binario único vía `cmd/permea`.

**Primary Dependencies**: **Solo librería estándar** (`encoding/json`, `encoding/base64`,
`net/url`, `os`, `bufio`, `path/filepath`). La lectura del enrollment string por stdin usa
`os.Stdin` + `bufio`. Se reutilizan `internal/config` y `internal/transport` ya existentes. Cero
dependencias externas nuevas (Principio III).

**Storage**: `config.json` bajo el directorio de configuración por-usuario por SO (ya resuelto por
`config.DataDir` vía `os.UserConfigDir`). Enrolar escribe dos campos ya presentes en el esquema de
config —`endpoint` y `device_token`— con reescritura atómica temp+rename ya implementada. La
restricción de acceso a **solo el usuario propietario** se materializa distinto por plataforma
(coherente con SC-002): en POSIX, `Chmod(0600)`; en Windows nativo, el directorio por-usuario
(ACL heredada que lo restringe al usuario) más el atributo de solo-lectura que `Chmod` sí fija —no
se promete semántica POSIX literal en Windows. No se crea ningún fichero nuevo ni cambia la
implementación (se reutiliza `config.Save`).

**Testing**: `go test ./...`. El **golden test de frontera** (`internal/ingest/boundary_test.go`)
permanece en verde y sin cambios. Se añaden tests: decodificación del enrollment string
(válido/malformado), enrolamiento feliz (2xx→persiste), rechazo (401/403→no persiste, estado
indistinguible), verificación de que el ping es un lote vacío `[]` sin metadato, permisos `0600`,
redacción del secreto en toda la salida de `enroll`/`status`, y re-enrolamiento sin residuos.

**Target Platform**: macOS, Linux y Windows nativo. Ruta de config resuelta por SO
(`os.UserConfigDir` vía `config.DataDir`); nunca hardcodeada.

**Project Type**: Agente/CLI de escritorio de proyecto único (un binario estático).

**Performance Goals**: No crítico. `enroll` hace un único intento de verificación (una petición
HTTP); `status` es lectura local. Sin bucles ni backoff en el camino de enrolamiento.

**Constraints**: frontera inviolable (el ping es lote vacío, cero eventos/metadato); higiene de
secreto del mismo calibre que el `salt` (el token y el enrollment string nunca se loguean, muestran
ni filtran en errores); HTTPS obligatorio (una URL `http://` se rechaza); cero dependencias nuevas;
contract-driven (cumple `transport.md`, define `enrollment-string.md`, no redefine token ni endpoint).

**Scale/Scope**: una máquina de desarrollador; dos comandos nuevos; un contrato de formato nuevo.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Puerta | Estado |
|---|---|---|
| **I. Frontera de datos inviolable** | No se amplía la allowlist; el ping de verificación no transporta contenido ni metadato. | ✅ PASS — la verificación reutiliza `transport.Client.Send` con `[]event.Event{}` (JSON `[]`, cero campos). No se añade ningún campo al `event.Event` ni passthrough alguno. El golden test de frontera no cambia. |
| **II. Privacidad auditable, local-first** | Token en local; verificación reutiliza el transporte existente; código abierto y legible. | ✅ PASS — el token y la URL se guardan solo en `config.json` local; la lógica de enrolamiento es leíble en `cmd/permea` + `internal/config`; no depende del backend salvo el ping de auth. |
| **III. Binario único y auditable** | Binario estático; se favorece stdlib; ruta por SO; toda dependencia se justifica. | ✅ PASS — **cero dependencias nuevas**; formato del enrollment string con `encoding/base64` + `encoding/json` (stdlib); ruta de config por SO ya existente. |
| **IV. Test-first en la frontera** | Golden test antes que nada; Given/When/Then; DEBE/NUNCA; nada se cierra con la frontera en rojo. | ✅ PASS — el golden test sigue verde; los nuevos tests (incl. "el ping es lote vacío sin metadato" y "el secreto no aparece en la salida") se escriben junto a/antes de la implementación. |
| **V. Desarrollo dirigido por especificaciones** | Ciclo specify→plan→tasks→implement; artefactos en `specs/NNN-slug/`; consistencia con la constitución. | ✅ PASS — este plan deriva del `spec.md`; el contrato nuevo (`enrollment-string.md`) no redefine los contratos de P-001/P-002. |

**Resultado del gate**: sin violaciones. La sección *Complexity Tracking* queda vacía.

## Project Structure

### Documentation (this feature)

```text
specs/003-enrolamiento/
├── plan.md                      # Este fichero (/speckit.plan)
├── research.md                  # Fase 0 (/speckit.plan)
├── data-model.md                # Fase 1 (/speckit.plan)
├── quickstart.md                # Fase 1 (/speckit.plan)
├── contracts/                   # Fase 1 (/speckit.plan)
│   ├── enrollment-string.md        # NUEVO: formato del envoltorio URL+token (FR-013). 003 lo define.
│   └── cli.md                       # Contrato de los comandos enroll/status (E/S, exit codes, redacción)
├── checklists/
│   └── requirements.md          # Checklist de calidad del spec (ya existe)
└── tasks.md                     # Fase 2 (/speckit.tasks — NO lo crea /speckit.plan)
```

El contrato de transporte que 003 **consume sin redefinir** vive en
`specs/001-agente-inicial/contracts/transport.md` (fuente de verdad del `/ingest` y del token).

### Source Code (repository root)

Layout `cmd/` + `internal/` fijado por la constitución (ya existente). 003 **reutiliza** paquetes;
no crea paquetes nuevos ni mueve la frontera.

```text
cmd/
└── permea/
    ├── main.go          # Se añade despacho de subcomandos: `enroll`, `status` (además de los flags
    │                    #   existentes --scan/--run/--daemon/--version, que no cambian).
    ├── enroll.go        # NUEVO: flujo `permea enroll` (decodificar→verificar→persistir). Lee el
    │                    #   enrollment string por dos vías: argv si se pasa un argumento, stdin si no
    │                    #   (convenio `-` para stdin explícito). Solo stdlib (os.Stdin, bufio); sin eco.
    └── status.go        # NUEVO: flujo `permea status` (lee config, informa estado + URL, redacta token).

internal/
├── config/
│   ├── config.go        # SIN cambios de esquema: Endpoint y DeviceToken ya existen; Save ya escribe 0600.
│   └── enrollment.go    # NUEVO: ParseEnrollmentString(s) → (endpoint, token, error); validación y
│                        #   helpers de estado (IsEnrolled). Punto único de manejo del secreto de enrolamiento.
├── transport/
│   └── transport.go     # Se añade `Verify()` (wrapper de Send([]event.Event{})): ping de lote vacío,
│                        #   un solo intento, sin backoff. Documenta "sin endpoint nuevo".
└── event/               # SIN cambios (la frontera no se toca).
```

**Structure Decision**: Proyecto único en Go con el layout existente. El manejo del enrollment
string (decodificación y validación) se concentra en `internal/config/enrollment.go`, junto al resto
de la configuración local y su disciplina de secreto (como el `salt`). La verificación se apoya en
`internal/transport` reutilizando `Send`; el nuevo `Verify()` solo documenta la intención "lote vacío,
un intento". Los comandos viven en `cmd/permea` como subcomandos, sin alterar los flags de P-001/P-002.

## Complexity Tracking

> Sin violaciones de la Constitution Check. No hay complejidad que justificar.
