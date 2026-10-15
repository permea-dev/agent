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
# B2 (T011): `lineas/`, una línea (o dos) por fichero para `RegistroKimi`: un registro por clase, las partidas de SC-005, los seis
# modelos de SC-008, una línea con centinelas, y tres líneas que NO son registro aunque nombren `usage.record` (un texto que lo
# contiene, una `forked` y un `subagent.completed`).
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

# --- B2 · lineas/ ---------------------------------------------------------------------------------------------------
L = os.path.join(AQUI, "lineas")
shutil.rmtree(L, ignore_errors=True)


def registro(usage=None, modelo="moonshot-ai/kimi-k2.7-code", alcance="turn", momento=1000, quitar=(), agente="main", extra=None):
    """Un `usage.record` con la forma medida (m1 1): type, agentId, model, usage, usageScope, time; type primero."""
    if usage is None:
        usage = {"inputOther": 11, "output": 22, "inputCacheRead": 33, "inputCacheCreation": 0}
    r = {"type": "usage.record", "agentId": agente, "model": modelo, "usage": usage, "usageScope": alcance, "time": momento}
    r.update(extra or {})
    for k in quitar:
        r.pop(k)
    return json.dumps(r, ensure_ascii=False, separators=(",", ":")) + "\n"


def fichero(nombre, *lineas):
    escribir(os.path.join(L, nombre + ".jsonl"), "".join(lineas))


def u(o, s, r, c):
    return {"inputOther": o, "output": s, "inputCacheRead": r, "inputCacheCreation": c}


fichero("evento", registro())
fichero("compactacion", registro(alcance="session"))
fichero("sc005", registro(u(100, 7, 40, 2)))
fichero("modelo_prefijo", registro(modelo="moonshot-ai/kimi-k2.7-code"))
fichero("modelo_sin_barra", registro(modelo="kimi-k2.7-code"))
fichero("modelo_varias", registro(modelo="a/b/c"))
fichero("modelo_inicial", registro(modelo="/x"))
fichero("modelo_acabada", registro(modelo="x/"))
fichero("modelo_vacio_o_no_textual", registro(modelo=""), registro(modelo=5))
fichero("incoherente_usage_no_objeto", registro([1, 2]))
fichero("incoherente_cuatro_ausentes", registro({}))
fichero("incoherente_no_numerica", registro(u("texto-1", 22, 33, 0)))
fichero("incoherente_no_entera", registro(u(1.5, 22, 33, 0)))
fichero("incoherente_negativa", registro(u(-1, 22, 33, 0)))
fichero("incoherente_time_textual", registro(momento="1000"))
fichero("incoherente_time_ausente", registro(quitar=("time",)))
fichero("a_cero", registro(u(0, 0, 0, 0)))
fichero("ms", registro(momento=1))
fichero("centinelas", registro(agente="CENTINELA-agentId", alcance="CENTINELA-usageScope",
                               extra={"raw": {"CENTINELA-raw": "CENTINELA-texto"}}))
# Tres líneas que NO son registro: el marcador dentro de un texto (escapado), una `forked` y un `subagent.completed`.
fichero("falso_texto", json.dumps({"type": "context.append_message", "agentId": "main",
                                   "message": {"role": "user", "content": [{"type": "text",
                                   "text": '{"type":"usage.record","model":"texto-1"}'}]}, "time": 1},
                                  ensure_ascii=False, separators=(",", ":")) + "\n")
fichero("forked", json.dumps({"type": "forked", "agentId": "main", "time": 1}, separators=(",", ":")) + "\n")
fichero("subagent_completed", json.dumps({"type": "subagent.completed", "subagentId": "agent-0", "resultSummary": "texto-2",
                                          "usage": u(5, 6, 7, 0), "contextTokens": 8, "time": 1},
                                         separators=(",", ":")) + "\n")
PY
python3 -I "$T/generar.py" "$AQUI"
