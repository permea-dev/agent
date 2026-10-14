#!/usr/bin/env bash
# Genera los ficheros SINTÉTICOS de los tests del lector de Copilot CLI (spec 011, plan §6). Nada sale de una sesión real:
# ninguna línea, ni un trozo, de una copia congelada. Se construyen desde el esquema del descubrimiento, con valores inventados.
#
#   bash internal/ingest/testdata/copilot/generar.sh
#
# Requiere python3 (sólo para regenerar; ni el binario ni la suite lo necesitan). Se ejecuta con `python3 -I`.
# Identificadores `s-…` y `l-…` *(nunca con forma UUID)*, textos neutros `texto-N`, centinelas `CENTINELA-…` y rutas `/sintetico/…`.
#
# B1 (T004): `sesiones/`, el árbol de `ListarCopilot`. El `session.db` de 0 B NO se genera aquí: `.gitignore` ignora `*.db`, y
# el test lo crea en su copia de `t.TempDir()`.
set -euo pipefail
AQUI=$(cd "$(dirname "$0")" && pwd)
command -v python3 >/dev/null || { echo "falta python3" >&2; exit 1; }
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
cat > "$T/generar.py" <<'PY'
import json, os, shutil, sys

AQUI = sys.argv[1]


def linea(tipo, datos, n):
    """Una línea de events.jsonl con la envoltura del descubrimiento (desc. Q2): type, data, id, timestamp, parentId."""
    return json.dumps({"type": tipo, "data": datos, "id": f"l-{n:024d}",
                       "timestamp": f"2026-01-01T00:00:{n:02d}.000Z", "parentId": None},
                      ensure_ascii=False, separators=(",", ":")) + "\n"


def inicio(sesion, n, cwd="/sintetico/proyecto-a"):
    return linea("session.start", {"sessionId": sesion, "version": 1, "producer": "productor-sintetico",
                                    "copilotVersion": "0.0.0-sintetico", "startTime": "2026-01-01T00:00:00.000Z",
                                    "context": {"cwd": cwd}}, n)


def escribir(ruta, contenido=""):
    os.makedirs(os.path.dirname(ruta), exist_ok=True)
    with open(ruta, "w", encoding="utf-8", newline="\n") as f:
        f.write(contenido)


# --- B1 · sesiones/ -------------------------------------------------------------------------------------------------
S = os.path.join(AQUI, "sesiones")
shutil.rmtree(S, ignore_errors=True)
s1, s2, s3 = (f"s-{i:024d}" for i in (1, 2, 3))
escribir(os.path.join(S, ".session-operation-locks", s1 + ".lock"))           # 0 B, como los de Copilot
escribir(os.path.join(S, s1, "events.jsonl"), inicio(s1, 1))
escribir(os.path.join(S, s1, "checkpoints", "index.md"), "# texto-1\n")
escribir(os.path.join(S, s2, "events.jsonl"), inicio(s2, 2))
escribir(os.path.join(S, s2, "sub", "events.jsonl"), inicio(s2, 3))          # a dos niveles: no se lista
escribir(os.path.join(S, s3, "checkpoints", "index.md"), "# texto-2\n")      # una carpeta sin events.jsonl
PY
python3 -I "$T/generar.py" "$AQUI"
( cd "$AQUI" && find sesiones -type f | LC_ALL=C sort | xargs md5sum )
