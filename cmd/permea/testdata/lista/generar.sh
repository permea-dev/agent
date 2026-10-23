#!/usr/bin/env bash
# Genera las plantillas SINTÉTICAS de los tests de `permea project list` (spec 013, plan §9). Nada sale de un registro real:
# ninguna línea, ni un trozo, de una carpeta del dueño. Se construyen desde el esquema de cada lector, con valores inventados.
#
#   bash cmd/permea/testdata/lista/generar.sh
#
# Requiere python3 (sólo para regenerar; ni el binario ni la suite lo necesitan). Se ejecuta con `python3 -I`.
#
# `plantillas/` es un hogar con sus carpetas SIN el punto inicial, que el test pone al copiar: `claude` → `.claude`, `codex` → `.codex`,
# `gemini` → `.gemini`, `copilot` → `.copilot` y `kimi-code` → `.kimi-code`. Cada `cwd` lleva el marcador `@CASA@`, que el test
# sustituye: por el hogar del sandbox en (22), o por `/sintetico/casa` en la referencia de SC-012, cuyos `project_ref` no dependen así
# del temporal. Los `.git` no se generan: los crea el test. OpenCode no se genera: el test copia `cmd/permea/testdata/opencode/sana.db`.
#
# Las identidades que salen, con `@CASA@` = el hogar:
#   @CASA@/dev/proyecto-ejemplo   Claude Code (4 mensajes, desde la raíz y desde `src` y `docs`; 2 repetidos en otro fichero), Codex (2,
#                                 desde `src`) y Kimi Code (2)
#   @CASA@/dev/otro-proyecto      Gemini (la sesión: 3 respuestas, 2 eventos) y Copilot CLI (2)
#   @CASA@/pruebas/tmp-1          Claude Code (1), sin `.git`
#   @CASA@/dev/viejo              Claude Code (2), carpeta que el test no crea
# Ids `m-lista-…`, `r-lista-…`, `s-lista-…`, `l-lista-…` *(nunca `msg_`, `ses_` ni con forma UUID)*, y textos neutros `texto-N`.
set -euo pipefail
AQUI=$(cd "$(dirname "$0")" && pwd)
command -v python3 >/dev/null || { echo "falta python3" >&2; exit 1; }
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
cat > "$T/generar.py" <<'PY'
import json, os, shutil, sys

AQUI = sys.argv[1]
P = os.path.join(AQUI, "plantillas")
shutil.rmtree(P, ignore_errors=True)


def escribir(rel, lineas):
    ruta = os.path.join(P, rel)
    os.makedirs(os.path.dirname(ruta), exist_ok=True)
    with open(ruta, "w", encoding="utf-8", newline="\n") as f:
        for l in lineas:
            f.write((l if isinstance(l, str) else json.dumps(l, ensure_ascii=False, separators=(",", ":"))) + "\n")


# ── Claude Code ──────────────────────────────────────────────────────────────────────────────────────────────────────
def claude(n, cwd, cuando, sesion, salida):
    return {"type": "assistant", "timestamp": cuando, "sessionId": sesion, "cwd": cwd, "requestId": "r-lista-%04d" % n,
            "message": {"id": "m-lista-%04d" % n, "model": "claude-opus-5-5",
                        "usage": {"input_tokens": 10, "output_tokens": salida, "cache_creation_input_tokens": 0,
                                  "cache_read_input_tokens": 0,
                                  "cache_creation": {"ephemeral_5m_input_tokens": 0, "ephemeral_1h_input_tokens": 0}}}}


R = "@CASA@/dev/proyecto-ejemplo"
m1 = claude(1, R + "/src", "2026-10-03T08:00:00Z", "s-lista-claude-1", 5)
m2 = claude(2, R + "/docs", "2026-10-04T09:00:00Z", "s-lista-claude-1", 6)
m3 = claude(3, R, "2026-10-05T10:00:00Z", "s-lista-claude-1", 7)
m4 = claude(4, R, "2026-10-06T11:00:00Z", "s-lista-claude-2", 8)
escribir("claude/projects/proyecto-ejemplo/s-lista-claude-1.jsonl", [m1, m2, m3])
escribir("claude/projects/proyecto-ejemplo/s-lista-claude-2.jsonl", [m1, m2, m4])  # reanudar: copia m1 y m2 (SC-009)
escribir("claude/projects/pruebas-tmp-1/s-lista-claude-3.jsonl",
         [claude(5, "@CASA@/pruebas/tmp-1", "2026-10-20T12:00:00Z", "s-lista-claude-3", 9)])
escribir("claude/projects/viejo/s-lista-claude-4.jsonl",
         [claude(6, "@CASA@/dev/viejo", "2026-10-05T13:00:00Z", "s-lista-claude-4", 3),
          claude(7, "@CASA@/dev/viejo", "2026-10-06T14:00:00Z", "s-lista-claude-4", 4)])

# ── Codex (desde `src`: la identidad es la de la raíz) ───────────────────────────────────────────────────────────────
def codex_uso(n):
    return {"timestamp": "2026-10-07T00:00:%02d.000Z" % n, "type": "token_usage_record",
            "payload": {"response_id": "r-lista-codex-%04d" % n, "session_id": "s-lista-codex-1", "thread_id": "s-lista-codex-1",
                        "turn_id": "t-1", "root_turn_id": "t-1",
                        "usage": {"input_tokens": 100, "cached_input_tokens": 40, "cache_write_input_tokens": 0, "output_tokens": 10,
                                  "reasoning_output_tokens": 0, "total_tokens": 110}}}


escribir("codex/sessions/2026/10/07/rollout-lista-1.jsonl", [
    {"timestamp": "2026-10-07T00:00:00.000Z", "type": "session_meta",
     "payload": {"id": "s-lista-codex-1", "session_id": "s-lista-codex-1", "cwd": R + "/src", "cli_version": "0.160.1", "source": "exec"}},
    {"timestamp": "2026-10-07T00:00:00.000Z", "type": "turn_context",
     "payload": {"turn_id": "t-1", "model": "modelo-sintetico", "cwd": R + "/src"}},
    codex_uso(1), codex_uso(2)])

# ── Gemini CLI (la raíz, por `.project_root`) ────────────────────────────────────────────────────────────────────────
O = "@CASA@/dev/otro-proyecto"
def gemini(n, salida, extra=None):
    l = {"id": "m-lista-gemini-%04d" % n, "timestamp": "2026-10-08T00:00:%02d.000Z" % n, "type": "gemini", "content": "texto-1",
         "thoughts": [], "tokens": {"input": 100, "output": salida, "cached": 0, "thoughts": 3, "tool": 0, "total": 100 + salida + 3},
         "model": "modelo-sintetico-gemini"}
    if extra:
        l.update(extra)
    return l


escribir("gemini/tmp/proyecto-lista/chats/s-lista-gemini.jsonl", [
    {"sessionId": "s-lista-gemini-1", "projectHash": "h-lista-1", "startTime": "2026-10-08T00:00:00.000Z",
     "lastUpdated": "2026-10-08T00:00:00.000Z", "kind": "main"},
    gemini(1, 7),
    gemini(1, 7, {"toolCalls": [{"id": "c-1", "name": "texto-2", "args": {}, "status": "success"}]}),  # la misma respuesta, repetida
    {"id": "u-lista-gemini-0001", "timestamp": "2026-10-08T00:00:02.000Z", "type": "user", "content": [{"text": "texto-3"}]},
    gemini(2, 8)])
with open(os.path.join(P, "gemini/tmp/proyecto-lista/.project_root"), "w", encoding="utf-8", newline="\n") as f:
    f.write(O + "\n")

# ── Copilot CLI ──────────────────────────────────────────────────────────────────────────────────────────────────────
def copilot_uso(n):
    detalle = lambda cuenta, tipo, coste: {"batchSize": 1000000, "costPerBatch": coste, "tokenCount": cuenta, "tokenType": tipo,
                                           "model": "modelo-sintetico-1"}
    return {"type": "session.usage_record", "data": {"usage": {
        "model": "modelo-sintetico-1", "provider": {"id": "p-lista-%04d" % n, "kind": "copilot", "label": "texto-1"},
        "modelDisplayName": "texto-2", "inputTokens": 100, "outputTokens": 7, "cacheReadTokens": 40, "cacheWriteTokens": 2,
        "reasoningTokens": 0, "cost": 1.0, "duration": 100, "initiator": "user", "interactionType": "conversation-agent", "isAuto": True,
        "apiCallId": "u-lista-%04d" % n, "providerCallId": "c-lista-%04d" % n, "serviceRequestId": "q-lista-%04d" % n,
        "accounting": {"sourceSessionId": "s-lista-copilot-1", "sequence": n, "usageId": "u-lista-%04d" % n},
        "copilotUsage": {"tokenDetails": [detalle(58, "input", 1000), detalle(40, "cache_read", 100), detalle(2, "cache_write", 1250),
                                          detalle(7, "output", 5000)], "totalNanoAiu": 0},
        "toolCounts": {}, "finishReason": "stop"}},
        "id": "l-lista-%04d" % n, "timestamp": "2026-10-09T00:00:%02d.000Z" % n, "parentId": None}


escribir("copilot/session-state/s-lista-copilot-1/events.jsonl", [
    {"type": "session.start", "data": {"sessionId": "s-lista-copilot-1", "version": 1, "producer": "productor-sintetico",
                                       "copilotVersion": "0.0.0-sintetico", "startTime": "2026-10-09T00:00:00.000Z",
                                       "context": {"cwd": O}},
     "id": "l-lista-0000", "timestamp": "2026-10-09T00:00:00.000Z", "parentId": None},
    copilot_uso(1), copilot_uso(2)])

# ── Kimi Code (el proyecto, por el índice) ───────────────────────────────────────────────────────────────────────────
def kimi(entrada, salida, ms):
    return {"type": "usage.record", "agentId": "main", "model": "moonshot-ai/kimi-k2.7-code",
            "usage": {"inputOther": entrada, "output": salida, "inputCacheRead": 100, "inputCacheCreation": 0}, "usageScope": "turn",
            "time": ms}


escribir("kimi-code/sessions/wd_lista_000000000001/session_lista-0001/agents/main/wire.jsonl", [
    {"type": "metadata", "protocol_version": "1.5", "created_at": 1},
    {"type": "context.append_message", "agentId": "main", "message": {"role": "user", "content": [{"type": "text", "text": "texto-3"}]},
     "time": 2},
    kimi(10, 1, 1760000000000), kimi(20, 2, 1760000060000)])
escribir("kimi-code/session_index.jsonl", [
    {"sessionId": "session_lista-0001", "sessionDir": "@CASA@/.kimi-code/sessions/wd_lista_000000000001/session_lista-0001",
     "workDir": R}])
PY
python3 -I "$T/generar.py" "$AQUI"
( cd "$AQUI/plantillas" && find . -type f | LC_ALL=C sort | xargs md5sum )
