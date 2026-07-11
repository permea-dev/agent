# Quickstart — Validar el enrolamiento (003)

Guía de validación de extremo a extremo del enrolamiento por el lado del agente. Prueba que
`permea enroll` y `permea status` cumplen el spec sin filtrar el secreto ni tocar la frontera.
Detalles de formato y comportamiento en [contracts/enrollment-string.md](./contracts/enrollment-string.md)
y [contracts/cli.md](./contracts/cli.md).

> **Nota de entorno**: el toolchain de Go se ejecuta en WSL/Linux; el binario también corre en
> Windows nativo. Las rutas de config se resuelven por SO (`os.UserConfigDir`).

## Prerrequisitos

- Go 1.22 (`go version`).
- Repo del agente; frontera y transporte de P-001 ya presentes.
- `python3` disponible (solo para fabricar un enrollment string de prueba a mano).

## 0. Compilar

```bash
go build -o /tmp/permea ./cmd/permea
```

## 1. Fabricar un enrollment string de prueba

El backend (P-002) lo emite en producción; para validar a mano, constrúyelo con el mismo formato
`pmea1.<base64url(json)>`:

```bash
mkenroll() { # uso: mkenroll <endpoint> <token>
  local json; json=$(printf '{"endpoint":"%s","token":"%s"}' "$1" "$2")
  printf 'pmea1.%s' "$(printf '%s' "$json" | python3 -c 'import sys,base64; sys.stdout.write(base64.urlsafe_b64encode(sys.stdin.buffer.read()).rstrip(b"=").decode())')"
}
ENROLL_OK=$(mkenroll "https://api.permea.example/ingest" "dev_tok_demo")
```

## 2. Rechazo de entradas inválidas (offline, sin backend)

Estos casos no requieren red: fallan en la decodificación/validación, **antes** del ping.

```bash
/tmp/permea enroll "no-es-un-enrollment-string"   # prefijo desconocido
/tmp/permea enroll "pmea1.###"                     # base64 inválido
/tmp/permea enroll "$(mkenroll http://inseguro.example/ingest tok)"  # endpoint http:// → rechazado
/tmp/permea enroll                                 # falta el argumento
```

**Esperado**: cada uno termina con exit ≠ 0 y un mensaje de error que **no** reproduce el argumento
ni contiene el token. No se escribe/crea `config.json`. (Verifica: `grep` del token en la salida no
devuelve nada.)

## 3. Estado sin enrolar

```bash
/tmp/permea status
```

**Esperado**: informa **no enrolado**, exit 0, sin error y sin secreto en pantalla.

## 4. Enrolar con un token válido (2xx)

Requiere un `/ingest` HTTPS que devuelva `2xx` al lote vacío `[]` con el token dado (el backend real
de P-001/P-002, o el arnés de tests que usa `httptest.NewTLSServer`).

```bash
/tmp/permea enroll "$ENROLL_OK"
```

**Esperado**:
- Exit 0 y confirmación que incluye la **URL** del backend, **nunca** el token.
- `config.json` contiene `endpoint` y `device_token` y tiene permisos **0600**:

```bash
CFG="$(python3 -c 'import os,platform; print(os.path.join(os.environ.get("XDG_CONFIG_HOME", os.path.expanduser("~/.config")),"permea","config.json"))')"
stat -c '%a %n' "$CFG"      # → 600 …/permea/config.json   (Linux)
```

## 5. `status` tras enrolar + redacción del token

```bash
/tmp/permea status
/tmp/permea status | grep -F "dev_tok_demo" && echo "FALLO: token filtrado" || echo "OK: sin token"
```

**Esperado**: `status` dice **enrolado** y muestra la URL; el `grep` del token **no** encuentra nada
(OK). Repite el `grep` sobre toda la salida de `enroll` del paso 4 para confirmar SC-003.

## 6. Rechazo de token inválido (401/403)

Con un backend que responde `401` al lote vacío (o el arnés de tests):

```bash
/tmp/permea enroll "$(mkenroll https://api.permea.example/ingest tok_revocado)"
```

**Esperado**: exit ≠ 0, mensaje "token rechazado", **no** se modifica/crea `config.json` (estado
indistinguible de no haber enrolado, SC-004), token no filtrado.

## 7. Re-enrolamiento sin residuos

```bash
/tmp/permea enroll "$ENROLL_OK"          # segundo enrolamiento válido
ls -la "$(dirname "$CFG")" | grep -E '\.bak|\.tmp'   # no debe quedar residuo con el secreto
```

**Esperado**: el token se reemplaza; no quedan ficheros `.bak`/temporales con el secreto (SC-010).

## 8. Vía stdin: enrolar sin exponer el secreto

El enrollment string también se acepta por **stdin**, la vía **recomendada** para no dejarlo en la
lista de procesos (argv) ni en el historial del shell. Produce el **mismo flujo** que por argumento.

### 8a. Enrolar por stdin (mismo resultado que por argv)

```bash
# Con el convenio '-' (stdin explícito):
echo "$ENROLL_OK" | /tmp/permea enroll -
# Equivalente, sin argumento:
echo "$ENROLL_OK" | /tmp/permea enroll
```

**Esperado**: idéntico al paso 4 — exit 0, confirmación con la **URL** (nunca el token) y
`config.json` con `endpoint` + `device_token` a **0600**. El agente recorta el salto de línea final
que añade `echo`. (FR-001, misma US1; no repite la verificación 2xx, ya cubierta en el paso 4.)

### 8b. El secreto no aparece en argv ni deja residuo (SC-011)

Para observar el argv real del proceso mientras lee de stdin, aliméntalo por una FIFO:

```bash
mkfifo /tmp/enrfifo
/tmp/permea enroll - < /tmp/enrfifo &           # bloquea esperando stdin
PID=$!
tr '\0' ' ' < /proc/$PID/cmdline; echo          # argv: '…/permea enroll -'  (SIN el token)
tr '\0' ' ' < /proc/$PID/cmdline | grep -F "dev_tok_demo" \
  && echo "FALLO: token en argv" || echo "OK: token no en argv"
printf '%s' "$ENROLL_OK" > /tmp/enrfifo         # suministra el string; el proceso termina
wait $PID; rm -f /tmp/enrfifo
```

**Esperado**: `cmdline` contiene solo `permea enroll -`; el `grep` del token **no** encuentra nada
(**SC-011**). Como en el paso 7, tampoco quedan residuos con el secreto
(`ls -la "$(dirname "$CFG")" | grep -E '\.bak|\.tmp'` vacío). En SO sin `/proc` el principio se
mantiene: el secreto viaja solo por stdin, nunca por argv.

### 8c. La ayuda de `enroll` recomienda stdin

```bash
/tmp/permea enroll --help 2>&1 | grep -iE 'stdin' && echo "OK: la ayuda menciona stdin"
```

**Esperado**: la ayuda de `enroll` menciona la vía **stdin** como la **recomendada** para no dejar el
secreto en el historial del shell (contrato en [contracts/cli.md](./contracts/cli.md)).

## 9. Suite automática (fuente de verdad de la verificación)

Los caminos con red (2xx / 401 / lote vacío sin metadato) se prueban de forma determinista con
`httptest.NewTLSServer` (HTTPS, nunca en claro):

```bash
go test ./...
go vet ./...
golangci-lint run
```

**Esperado**: todo en verde, incluido el **golden test de frontera** (sin cambios) y el test que
afirma que el ping de verificación transmite `[]` (cero eventos, cero metadato, SC-006/SC-009).
