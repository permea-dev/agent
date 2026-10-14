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
# B2 (T011): `lineas/`, una línea (o tres) por fichero: los registros de `RegistroCopilot`, cada clase, y las líneas que NO son
# registro (el falso marcador escapado dentro de un texto, y tres `model.*`). Ids `u-…` (usageId), `r-…` (requestId), `c-…`, `q-…`.
# B3 (T019): `ficheros/`, sesiones enteras para `LeerFicheroCopilot`: tramos, el corte en dos pasadas, la línea larga (≥ 64 KB),
# la línea final sin LF, el truncado, la repetida, las clases de SC-010, un subagente, una sesión matada, la corrupta, sin `start`,
# sin `cwd` y una carpeta para la sólo lectura. `esperado.json` guarda la línea de resumen de una pasada de cada fichero.
# B5 (T033): `ficheros/control*.jsonl`, con créditos (`totalNanoAiu`) fijados a mano para la línea de control de `--scan`, y
# `sin_copilotversion.jsonl`, un `session.start` sin `copilotVersion` (no es Copilot CLI para `--scan`).
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
                       "timestamp": f"2026-01-01T00:{n // 60:02d}:{n % 60:02d}.000Z", "parentId": None},
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

# --- B2 · lineas/ ---------------------------------------------------------------------------------------------------
L = os.path.join(AQUI, "lineas")
shutil.rmtree(L, ignore_errors=True)
TS = "2026-01-01T00:00:10.000Z"


def detalle(m, inp, cr, cw, out):
    """copilotUsage con la forma del descubrimiento (desc. Q3), precios por lote inventados."""
    filas = [("input", inp - cr - cw, 1000), ("cache_read", cr, 100), ("cache_write", cw, 1250), ("output", out, 5000)]
    return {"tokenDetails": [{"batchSize": 1000000, "costPerBatch": p, "tokenCount": n, "tokenType": t, "model": m}
                             for t, n, p in filas],
            "totalNanoAiu": sum(n * p for _, n, p in filas) // 1000000}


def usage(n, inp=100, out=7, cr=40, cw=2, reas=0, modelo="modelo-sintetico-1", usage_id=None, quitar=(), cambiar=None):
    uid = f"u-{n:024d}" if usage_id is None or usage_id is ... else usage_id
    u = {"model": modelo, "provider": {"id": f"p-{n:024d}", "kind": "copilot", "label": "texto-1"},
         "modelDisplayName": "texto-2", "inputTokens": inp, "outputTokens": out, "cacheReadTokens": cr,
         "cacheWriteTokens": cw, "reasoningTokens": reas, "cost": 1.0, "duration": 100, "initiator": "user",
         "interactionType": "conversation-agent", "isAuto": True, "apiCallId": uid, "providerCallId": f"c-{n:024d}",
         "serviceRequestId": f"q-{n:024d}", "accounting": {"sourceSessionId": f"s-{1:024d}", "sequence": n, "usageId": uid},
         "copilotUsage": detalle(modelo, inp, cr, cw, out), "toolCounts": {}, "finishReason": "stop"}
    for k in quitar:
        u.pop(k, None)
    if usage_id is ...:
        u["accounting"].pop("usageId")
    for k, v in (cambiar or {}).items():
        u[k] = v
    return {"usage": u}


def compactacion(n, ctu=True, request_id=None, success=True, inp=1000, out=50, cr=600, cw=100):
    d = {"success": success, "preCompactionTokens": 300, "postCompactionTokens": 290, "messagesRemoved": 4,
         "tokensRemoved": 10, "summaryContent": "texto-3", "behaviorModelId": "modelo-sintetico-1",
         "checkpointNumber": 1, "checkpointPath": "/sintetico/checkpoints/001-texto.md", "trigger": "manual",
         "requestId": f"r-{n:024d}" if request_id is None or request_id is ... else request_id, "serviceRequestId": f"q-{n:024d}"}
    if request_id is ...:
        d.pop("requestId")
    if not success:
        d = {"success": False, "error": "texto-4", "trigger": "manual"}
    if ctu is True:
        d["compactionTokensUsed"] = {"requestCount": 1, "model": "modelo-sintetico-1", "modelDisplayName": "texto-2",
                                     "inputTokens": inp, "outputTokens": out, "cacheReadTokens": cr, "cacheWriteTokens": cw,
                                     "copilotUsage": detalle("modelo-sintetico-1", inp, cr, cw, out),
                                     "aiCreditsStatus": "complete", "duration": 100}
    elif ctu is not False:
        d["compactionTokensUsed"] = ctu
    return d


def reg(nombre, tipo, datos, n=10, ts=TS):
    linea_ = json.loads(linea(tipo, datos, n))
    linea_["timestamp"] = ts
    escribir(os.path.join(L, nombre + ".jsonl"), json.dumps(linea_, ensure_ascii=False, separators=(",", ":")) + "\n")


U, C = "session.usage_record", "session.compaction_complete"
reg("evento", U, usage(1))
reg("sc005", U, usage(2, inp=100, out=7, cr=40, cw=2, reas=3))                    # SC-005: 58/7/40/2
reg("razonamiento", U, usage(3, reas=999))                                         # (10/razonamiento)
reg("desfase", U, usage(4), ts="2026-01-01T02:00:10.000+02:00")                    # (11/desfase): 00:00:10Z
reg("sin_modelo", U, usage(5, quitar=("model",)))                                  # (12)
reg("sin_usageid_ausente", U, usage(6, usage_id=...))                              # (9)
reg("sin_usageid_vacio", U, usage(7, usage_id=""))
reg("sin_usageid_no_textual", U, usage(8, usage_id=7))
reg("incoherente_cuatro_ausentes", U, usage(9, quitar=("inputTokens", "outputTokens", "cacheReadTokens", "cacheWriteTokens")))
reg("incoherente_no_numerica", U, usage(10, cambiar={"outputTokens": "texto-5"}))
reg("incoherente_no_entera", U, usage(11, cambiar={"outputTokens": 7.5}))
reg("incoherente_negativa", U, usage(12, cambiar={"cacheReadTokens": -1}))
reg("incoherente_no_cabe", U, usage(13, inp=100, cr=90, cw=20))                   # cr + cw > input
reg("incoherente_timestamp_malo", U, usage(14), ts="texto-6")
reg("compactacion", C, compactacion(1))                                            # (7/partidas): 300/50/600/100
reg("compactacion_fallida", C, compactacion(2, success=False))
reg("compactacion_sin_ctu", C, compactacion(3, ctu=False))
reg("compactacion_sin_requestid", C, compactacion(4, request_id=...))
reg("incoherente_ctu_no_objeto", C, compactacion(5, ctu=5))                         # (8/usage_no_objeto)
# (14) centinelas en todos los campos de texto que no deben viajar
cent = usage(15, usage_id="CENTINELA-usageId")
cent["usage"]["accounting"]["sourceSessionId"] = "CENTINELA-sourceSessionId"
cent["usage"]["providerCallId"] = "CENTINELA-providerCallId"
cent["usage"]["serviceRequestId"] = "CENTINELA-serviceRequestId"
cent["usage"]["provider"]["label"] = "CENTINELA-label"
cent["usage"]["modelDisplayName"] = "CENTINELA-modelDisplayName"
cent["usage"]["toolCounts"] = {"CENTINELA-herramienta": 1}
reg("centinelas", U, cent)
centc = compactacion(6, request_id="CENTINELA-requestId")
centc["summaryContent"] = "CENTINELA-summaryContent"
centc["checkpointPath"] = "/sintetico/CENTINELA-checkpointPath.md"
reg("centinelas_compactacion", C, centc)
# (13) líneas que NO son registro: el marcador dentro de un texto va escapado, y no casa con el prefiltro
falso = 'texto-7 "type":"session.usage_record" texto-8 "type":"session.compaction_complete"'
escribir(os.path.join(L, "falso_user.jsonl"),
         linea("user.message", {"content": falso, "transformedContent": falso, "messageId": "m-1"}, 20))
escribir(os.path.join(L, "falso_system.jsonl"), linea("system.message", {"role": "system", "content": falso}, 21))
escribir(os.path.join(L, "model_aux.jsonl"),
         linea("model.turn_started", {"kind": "turn_started", "model": "modelo-sintetico-2"}, 22)
         + linea("model.model_call_success", {"kind": "model_call_success", "requestMessages": falso,
                                              "modelCall": {"model": "modelo-sintetico-2", "initiator": "agent"}}, 23)
         + linea("model.response", {"kind": "response", "response": {"content": falso}}, 24))

# --- B3 · ficheros/ -------------------------------------------------------------------------------------------------
F = os.path.join(AQUI, "ficheros")
shutil.rmtree(F, ignore_errors=True)
X, Y = "/sintetico/x", "/sintetico/y"


def l(tipo, datos, n, **sobre):
    """Una línea como `linea`, con campos de la envoltura de más (p. ej. `agentId`)."""
    d = json.loads(linea(tipo, datos, n))
    d.update(sobre)
    return json.dumps(d, ensure_ascii=False, separators=(",", ":")) + "\n"


def u(n, **kw):
    return l(U, usage(n, **kw), n)


def reanudar(n, cwd):
    return l("session.resume", {"resumeTime": "2026-01-01T00:01:00.000Z", "eventCount": 3, "selectedModel": "auto",
                                "context": {"cwd": cwd}}, n)


def sistema(n, tam=66000):
    return l("system.message", {"role": "system", "content": "texto-9 " * (tam // 8)}, n)


def fichero(nombre, *lineas, lf=True):
    contenido = "".join(lineas)
    escribir(os.path.join(F, nombre), contenido if lf else contenido.rstrip("\n"))


ESPERADO = {}


def resumen(nombre, reg, ev, comp=0, rep=0, sid=0, inc=0, smod=0):
    ESPERADO[nombre] = (f"copilot_cli: registros {reg} · eventos {ev} · compactaciones {comp} · repetidas {rep} · "
                        f"sin identificador {sid} · incoherentes {inc} · sin modelo {smod}")


ini = inicio(s1, 1, X)
fichero("dos_llamadas.jsonl", ini, u(21), u(22)); resumen("dos_llamadas.jsonl", 2, 2)
fichero("linea_larga.jsonl", ini, sistema(2), u(23), u(24)); resumen("linea_larga.jsonl", 2, 2)
fichero("tramos.jsonl", ini, u(25), reanudar(3, Y), u(26)); resumen("tramos.jsonl", 2, 2)
paso1 = [ini, u(27), sistema(4)]
fichero("cortado/paso1.jsonl", *paso1); resumen("cortado/paso1.jsonl", 1, 1)
fichero("cortado/paso2.jsonl", *paso1, reanudar(5, Y), u(28)); resumen("cortado/paso2.jsonl", 2, 2)
fichero("sin_lf.jsonl", ini, u(29), lf=False); resumen("sin_lf.jsonl", 0, 0)
fichero("sin_lf_completo.jsonl", ini, u(29)); resumen("sin_lf_completo.jsonl", 1, 1)
fichero("truncado/largo.jsonl", ini, u(30), u(31)); resumen("truncado/largo.jsonl", 2, 2)
fichero("truncado/corto.jsonl", ini, u(30)); resumen("truncado/corto.jsonl", 1, 1)
fichero("repetida.jsonl", ini, u(32), l(U, usage(32), 33)); resumen("repetida.jsonl", 2, 1, rep=1)
# SC-010: evento sin modelo · compactación · el mismo usageId repetido · sin usageId · cr + cw > input · partida no entera
sin_modelo = l(U, usage(33, quitar=("model",)), 34)
fichero("clases.jsonl", ini, sin_modelo, l(C, compactacion(7), 35), l(U, usage(33, quitar=("model",)), 36),
        u(34, usage_id=""), u(35, inp=100, cr=90, cw=20), u(36, cambiar={"outputTokens": 7.5}))
resumen("clases.jsonl", 6, 2, comp=1, rep=1, sid=1, inc=2, smod=1)
# Un subagente síncrono (muestra 3): sus registros, con `initiator: sub-agent`, `agentId`, otro `sourceSessionId` y
# `parentToolCallId`; declara otro modelo, pero el registro trae el cobrado.
sub = []
for n in (38, 39):
    d = usage(n)
    d["usage"].update({"initiator": "sub-agent", "interactionType": "conversation-subagent", "parentToolCallId": "t-1"})
    d["usage"]["accounting"]["sourceSessionId"] = f"s-{9:024d}"
    sub.append(l(U, d, n, agentId="a-1"))
fichero("subagente.jsonl", ini, u(37),
        l("tool.execution_start", {"toolCallId": "t-1", "toolName": "task", "model": "modelo-sintetico-1",
                                   "arguments": {"prompt": "texto-10"}}, 40),
        l("subagent.started", {"toolCallId": "t-1", "agentName": "task", "model": "modelo-subagente"}, 41, agentId="a-1"),
        l("subagent.configured", {"model": "modelo-subagente", "multiTurn": True}, 42, agentId="a-1"),
        l("session.model_change", {"source": "agent", "newModel": "modelo-subagente"}, 43),
        sub[0], sub[1],
        l("subagent.completed", {"toolCallId": "t-1", "model": "modelo-subagente", "firstDispatchedModel": "modelo-sintetico-1",
                                 "totalTokens": 214}, 44, agentId="a-1"),
        u(45))
resumen("subagente.jsonl", 4, 4)
# Una sesión matada a mitad del 2.º prompt (muestra 2, M1): del 2.º sólo quedan su inicio y una auxiliar.
fichero("matada.jsonl", ini, u(46), l("assistant.message", {"content": "texto-11"}, 47),
        l("session.usage_checkpoint", {"totalNanoAiu": 1, "totalPremiumRequests": 1}, 48),
        l("user.message", {"content": "texto-12"}, 49), l("assistant.turn_start", {"turnId": "0"}, 50),
        l("model.turn_started", {"kind": "turn_started", "model": "modelo-sintetico-2"}, 51),
        l("model.model_call_success", {"kind": "model_call_success", "requestMessages": "texto-13"}, 52))
resumen("matada.jsonl", 1, 1)
fichero("corrupta.jsonl", ini, u(53), '{"type":"session.usage_record","data":{"usage":{texto-14\n', u(54))
resumen("corrupta.jsonl", 2, 2)
fichero("sin_start.jsonl", u(55), u(56)); resumen("sin_start.jsonl", 2, 2)
sin_cwd = json.loads(inicio(s1, 1)); sin_cwd["data"].pop("context")
fichero("sin_cwd.jsonl", json.dumps(sin_cwd, separators=(",", ":")) + "\n", u(57)); resumen("sin_cwd.jsonl", 1, 1)
# Para la sólo lectura: una carpeta con sus `.lock` y `checkpoints/`.
escribir(os.path.join(F, "solo_lectura", ".session-operation-locks", s1 + ".lock"))
escribir(os.path.join(F, "solo_lectura", s1, ".workspace-fork.lock"))
escribir(os.path.join(F, "solo_lectura", s1, "checkpoints", "index.md"), "# texto-15\n")
escribir(os.path.join(F, "solo_lectura", s1, "events.jsonl"), ini + u(58) + u(59))
resumen("solo_lectura/" + s1 + "/events.jsonl", 2, 2)
# --- B5 · la línea de control de `--scan` ---------------------------------------------------------------------------
def con_credito(datos, nano, clave="usage"):
    datos[clave]["copilotUsage"]["totalNanoAiu"] = nano
    return datos


# Σ de los que dan evento = 123 456 789 (llamada) + 1 000 (compactación) = 123 457 789: ni la incoherente ni la repetida suman.
llamada = con_credito(usage(60), 123456789)
fichero("control.jsonl", ini, l(U, llamada, 60), l(C, con_credito(compactacion(9), 1000, "compactionTokensUsed"), 61),
        l(U, con_credito(usage(62, inp=100, cr=90, cw=20), 999999), 62), l(U, llamada, 63))
resumen("control.jsonl", 4, 2, comp=1, rep=1, inc=1)
fichero("control_uno.jsonl", ini, l(U, con_credito(usage(64), 1), 64)); resumen("control_uno.jsonl", 1, 1)
sin_version = json.loads(inicio(s1, 1, X)); sin_version["data"].pop("copilotVersion")
fichero("sin_copilotversion.jsonl", json.dumps(sin_version, separators=(",", ":")) + "\n", u(65))
escribir(os.path.join(F, "esperado.json"), json.dumps(ESPERADO, ensure_ascii=False, indent=1, sort_keys=True) + "\n")
PY
python3 -I "$T/generar.py" "$AQUI"
( cd "$AQUI" && find sesiones lineas ficheros -type f | LC_ALL=C sort | xargs md5sum )
