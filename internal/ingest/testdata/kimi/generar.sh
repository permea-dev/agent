#!/usr/bin/env bash
# Genera los ficheros SINTÉTICOS de los tests del lector de Kimi Code (spec 012, plan §6). Nada sale de una sesión real:
# ninguna línea, ni un trozo, de una copia congelada. Se construyen desde el esquema del descubrimiento, con valores inventados.
#
#   bash internal/ingest/testdata/kimi/generar.sh
#
# Requiere python3 (sólo para regenerar; ni el binario ni la suite lo necesitan). Se ejecuta con `python3 -I`.
# Sesiones `session_sintetica-NNNN` *(nunca con forma UUID)*, carpeta `wd_proyecto-a_000000000001`, textos neutros `texto-N`.
# Ningún `state.json`: `.gitignore` lo ignora en cualquier carpeta, y B1 no lo necesita.
#
# B1 (T004): `raiz/`, el árbol de `ListarKimi`: dos sesiones en una `wd_*`, un subagente, `sessions/.index-cache/` y `.index-dirty/`,
# carpetas ocultas con forma de sesión en los tres niveles, un `.bak` y un `.tmp.*` junto a un `wire.jsonl`, `raiz/imported/`, un
# `wire.jsonl` un nivel más hondo y una sesión sin `agents/`.
set -euo pipefail
AQUI=$(cd "$(dirname "$0")" && pwd)
command -v python3 >/dev/null || { echo "falta python3" >&2; exit 1; }
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
cat > "$T/generar.py" <<'PY'
import json, os, shutil, sys

AQUI = sys.argv[1]


def metadata(n):
    """La primera línea de un wire.jsonl (desc. Q2): type primero; created_at inventado."""
    return json.dumps({"type": "metadata", "protocol_version": "1.5", "created_at": n}, separators=(",", ":")) + "\n"


def escribir(ruta, contenido=""):
    os.makedirs(os.path.dirname(ruta), exist_ok=True)
    with open(ruta, "w", encoding="utf-8", newline="\n") as f:
        f.write(contenido)


# --- B1 · raiz/ -----------------------------------------------------------------------------------------------------
R = os.path.join(AQUI, "raiz")
shutil.rmtree(R, ignore_errors=True)
S = os.path.join(R, "sessions")
WD = os.path.join(S, "wd_proyecto-a_000000000001")
s1, s2, s3, s9 = (f"session_sintetica-{i:04d}" for i in (1, 2, 3, 9))
escribir(os.path.join(S, ".index-cache", "scan.json"), '{"version":1,"sessions":{}}\n')
escribir(os.path.join(S, ".index-dirty", s2 + ".1"))                                   # 0 B, como los de Kimi
escribir(os.path.join(S, ".oculta", s9, "agents", "main", "wire.jsonl"), metadata(9))   # oculta en el nivel de la wd
escribir(os.path.join(WD, s1, "agents", "main", "wire.jsonl"), metadata(1))
escribir(os.path.join(WD, s1, "agents", "main", "wire.jsonl.bak"), metadata(1))
escribir(os.path.join(WD, s1, "agents", "main", "wire.jsonl.tmp.1.ab"), metadata(1))
escribir(os.path.join(WD, s1, "agents", "main", "sub", "wire.jsonl"), metadata(1))      # un nivel más hondo
escribir(os.path.join(WD, s1, "agents", "agent-0", "wire.jsonl"), metadata(2))
escribir(os.path.join(WD, s1, "agents", ".oculto", "wire.jsonl"), metadata(3))         # oculta en el nivel del agente
escribir(os.path.join(WD, s2, "agents", "main", "wire.jsonl"), metadata(4))
escribir(os.path.join(WD, s3, "logs", "texto-1.log"), "texto-1\n")                      # una sesión sin agents/
escribir(os.path.join(WD, ".oculta", "agents", "main", "wire.jsonl"), metadata(5))      # oculta en el nivel de la sesión
escribir(os.path.join(R, "imported", "imp-0001", "agents", "main", "wire.jsonl"), metadata(6))
PY
python3 -I "$T/generar.py" "$AQUI"
