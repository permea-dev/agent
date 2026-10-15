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
# B3 (T020): `ficheros/`, `wire.jsonl` enteros para `LeerFicheroKimi`: dos registros, el fork *(una, dos y ninguna `forked`; cortado en
# dos pasadas)*, una corrupta que cuenta para N, la cabeza *(antes, cambiada, encogido)*, sin LF final, una sesión matada, un subagente con
# `subagent.completed` y las clases de SC-010; un `session_index.jsonl` con las dos barras de Windows y una sesión con dos líneas. Las líneas de
# relleno *(`context.append_message`, `llm.request`)* llevan textos neutros. `esperado.json` guarda la línea de resumen de una pasada de cada
# fichero, calculada aquí por un clasificador propio, que no comparte código con el lector.
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

# --- B3 · ficheros/ -------------------------------------------------------------------------------------------------
F = os.path.join(AQUI, "ficheros")
shutil.rmtree(F, ignore_errors=True)
MD = metadata(1)
FORKED = json.dumps({"type": "forked", "agentId": "main", "time": 5}, separators=(",", ":")) + "\n"
RELLENO = (json.dumps({"type": "context.append_message", "agentId": "main", "message": {"role": "user", "content": [
           {"type": "text", "text": "texto-3"}]}, "time": 2}, separators=(",", ":")) + "\n"
           + json.dumps({"type": "llm.request", "agentId": "main", "kind": "loop", "provider": "proveedor-sintetico",
                         "model": "modelo-sintetico-1", "messageCount": 1, "time": 3}, separators=(",", ":")) + "\n")
CORRUPTA = '{"type":"usage.record","model":\n'


def r(n, **kw):
    """El registro n: partidas y hora distintas para cada n."""
    kw.setdefault("usage", u(10 * n, n, 100 * n, 0))
    kw.setdefault("momento", 1000 * n)
    return registro(**kw)


def f(nombre, *lineas, final=True):
    t = "".join(lineas)
    escribir(os.path.join(F, nombre), t if final else t.rstrip("\n"))


f("dos_registros.jsonl", MD, RELLENO, r(1), RELLENO, r(2))
f("fork/un_forked.jsonl", MD, r(1), r(2), FORKED, RELLENO, r(3))
f("fork/dos_forked.jsonl", MD, r(1), FORKED, r(2), FORKED, r(3))
f("fork/sin_forked.jsonl", MD, r(1), r(2), r(3))
f("fork_cortado/paso1.jsonl", MD, r(1), r(2), FORKED)
f("fork_cortado/paso2.jsonl", MD, r(1), r(2), FORKED, r(3))
f("corrupta_cuenta.jsonl", MD, r(1), CORRUPTA, r(2))
f("corrupta_cuenta_paso1.jsonl", MD, r(1), CORRUPTA)
f("cabeza/antes.jsonl", r(1), r(2))                          # sin metadata, como un fichero anterior a la 1.5
f("cabeza/cambiada.jsonl", MD, r(1), r(2))                   # la migración: la cabeza cambia y el fichero crece
f("cabeza/cambiada_mas.jsonl", MD, r(1), r(2), r(3))
f("cabeza/largo.jsonl", MD, r(1), r(2), r(3))
f("cabeza/encogido.jsonl", MD, r(1))                          # la reparación: el prefijo válido, más corto
f("sin_lf.jsonl", MD, r(1), r(2), final=False)
f("sin_lf_completo.jsonl", MD, r(1), r(2))
f("matada.jsonl", MD, RELLENO)                                # el arranque y una llm.request, sin registro
SUB = [r(1, agente="agent-0"), r(2, agente="agent-0"), r(3, agente="agent-0")]
f("subagente/main.jsonl", MD, r(4), json.dumps({"type": "subagent.completed", "subagentId": "agent-0", "resultSummary": "texto-4",
  "usage": u(60, 6, 600, 0), "contextTokens": 8, "time": 4000}, separators=(",", ":")) + "\n", r(5))
f("subagente/agent-0.jsonl", MD, *SUB)
f("clases.jsonl", MD, r(1), FORKED, r(2), r(3, alcance="session"), r(4, modelo="x/"), r(5, usage=u(0, 0, 0, 0)),
  r(6, usage=u(-1, 6, 600, 0)), r(7, momento="7000"))
escribir(os.path.join(F, "indice", "session_index.jsonl"), "".join(
    json.dumps({"sessionId": sid, "sessionDir": "/sintetico/kimi/sessions/" + sid, "workDir": w}, ensure_ascii=False,
               separators=(",", ":")) + "\n"
    for sid, w in [("session_sintetica-0001", "C:\\sintetico\\proyecto-a"), ("session_sintetica-0002", "C:/sintetico/proyecto-a"),
                   ("session_sintetica-0003", "/sintetico/proyecto-a"), ("session_sintetica-0003", "/sintetico/proyecto-b")]))


def clasificar(texto):
    """El clasificador del guion (spec FR-020, D-3, P-8, P-9), sin código del lector: sólo líneas completas."""
    lineas = texto.split("\n")[:-1]
    ultima = max([i for i, l in enumerate(lineas) if l.startswith('{"type":"forked"')], default=-1)
    c = dict(registros=0, eventos=0, heredados=0, repetidas=0, acero=0, incoherentes=0, compactaciones=0, sinmodelo=0)
    for i, l in enumerate(lineas):
        if not l.startswith('{"type":"usage.record"'):
            continue
        try:
            o = json.loads(l)
        except ValueError:
            continue                                          # corrupta: aviso, no es registro
        c["registros"] += 1
        if i < ultima:
            c["heredados"] += 1
            continue
        us, t = o.get("usage"), o.get("time")
        ps = [us.get(k) for k in ("inputOther", "output", "inputCacheRead", "inputCacheCreation")] if isinstance(us, dict) else None
        if (ps is None or all(v is None for v in ps) or any(v is not None and (type(v) is not int or v < 0) for v in ps)
                or type(t) is not int or t <= 0):
            c["incoherentes"] += 1
        elif all((v or 0) == 0 for v in ps):
            c["acero"] += 1
        else:
            c["eventos"] += 1
            c["compactaciones"] += o.get("usageScope") == "session"
            m = o.get("model") if isinstance(o.get("model"), str) else ""
            c["sinmodelo"] += m.rsplit("/", 1)[-1] == ""
    return ("kimi_code: registros {registros} · eventos {eventos} · heredados {heredados} · repetidas {repetidas} · a cero {acero} · "
            "incoherentes {incoherentes} · compactaciones {compactaciones} · sin modelo {sinmodelo}").format(**c)


esperado = {}
for raiz, _, nombres in sorted(os.walk(F)):
    for n in sorted(nombres):
        if n.endswith(".jsonl") and n != "session_index.jsonl":
            rel = os.path.relpath(os.path.join(raiz, n), F).replace(os.sep, "/")
            with open(os.path.join(raiz, n), encoding="utf-8") as fh:
                esperado[rel] = clasificar(fh.read())
with open(os.path.join(F, "esperado.json"), "w", encoding="utf-8", newline="\n") as fh:
    json.dump(esperado, fh, ensure_ascii=False, indent=1, sort_keys=True)
    fh.write("\n")
PY
python3 -I "$T/generar.py" "$AQUI"
