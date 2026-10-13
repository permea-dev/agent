#!/usr/bin/env bash
# Genera las bases SINTÉTICAS de los tests de OpenCode en `cmd/permea` (spec 010, B6 y B7). Nada sale de una base real.
#
#   bash cmd/permea/testdata/opencode/generar.sh
#
# Requiere python3 con SQLite >= 3.44 (sólo para regenerar). El esquema es el de internal/ingest/testdata/opencode/generar.sh.
# NINGUNA fila en curso (PARA 11, D-010-P9): su resultado cambiaría 24 h después de generarlas. El guion lo comprueba.
# Identificadores `o-…` y `s-…`, centinelas `CENTINELA-…` y rutas `/sintetico/…`: nada con la forma de un dato real.
set -euo pipefail
AQUI=$(cd "$(dirname "$0")" && pwd)
command -v python3 >/dev/null || { echo "falta python3" >&2; exit 1; }
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
cat > "$T/generar.py" <<'PY'
import json, os, sqlite3, struct, sys
assert tuple(map(int, sqlite3.sqlite_version.split("."))) >= (3, 44, 0), "hace falta SQLite >= 3.44"
OUT = sys.argv[1]
T0 = 1760000000000

SESION_V2 = """CREATE TABLE `session_v2` (`id` text PRIMARY KEY, `project_id` text NOT NULL, `slug` text NOT NULL,
  `directory` text NOT NULL, `path` text, `title` text, `version` text NOT NULL, `time_created` integer NOT NULL,
  `time_updated` integer NOT NULL)"""
MENSAJE = """CREATE TABLE `session_message` (`id` text PRIMARY KEY, `session_id` text NOT NULL, `type` text NOT NULL,
  `seq` integer NOT NULL, `time_created` integer NOT NULL, `time_updated` integer NOT NULL, `data` text NOT NULL)"""

def base(nombre):
    con = sqlite3.connect(os.path.join(OUT, nombre), isolation_level=None)
    con.execute(SESION_V2)
    con.execute(MENSAJE)
    con.execute("CREATE UNIQUE INDEX `session_message_session_seq_idx` ON `session_message` (`session_id`,`seq`)")
    return con

def sesion(con, sid, directorio, path=None, title="titulo-sintetico", slug="slug-sintetico", proyecto="prj-sintetico"):
    con.execute("INSERT INTO session_v2 VALUES (?,?,?,?,?,?,'2.0.26',?,?)", (sid, proyecto, slug, directorio, path, title, T0, T0))

def fila(con, mid, sid, seq, tipo, data):
    con.execute("INSERT INTO session_message VALUES (?,?,?,?,?,?,?)",
                (mid, sid, tipo, seq, T0, T0, json.dumps(data, separators=(",", ":"))))

def llamada(n):
    return {"role": "assistant", "time": {"created": T0 + n, "completed": T0 + n + 5},
            "tokens": {"input": 11 + n, "output": 5, "reasoning": 2, "cache": {"read": 13, "write": 3}}, "cost": 0.0001,
            "model": {"providerID": "proveedor-sintetico", "id": "modelo-sintetico-opencode", "variant": "default"}}

def usuario(texto="texto-sintetico"):
    return {"role": "user", "time": {"created": T0}, "text": texto}

def cerrar(con):
    en_curso = con.execute("SELECT count(*) FROM session_message WHERE type='assistant'"
                           " AND json_extract(data,'$.time.completed') IS NULL").fetchone()[0]
    assert en_curso == 0, "PARA 11: una fila en curso en un fixture de cmd"
    con.close()

# sana.db: una sesión, una fila `user` y dos llamadas completadas.
con = base("sana.db")
sesion(con, "s-sana", "/sintetico/sana")
fila(con, "o-sana-u", "s-sana", 1, "user", usuario())
fila(con, "o-sana-1", "s-sana", 2, "assistant", llamada(1))
fila(con, "o-sana-2", "s-sana", 3, "assistant", llamada(2))
cerrar(con)

# otra.db: una llamada completada; sustituye a la corrupta cuando «vuelve a estar sana» (SC-020).
con = base("otra.db")
sesion(con, "s-otra", "/sintetico/otra")
fila(con, "o-otra-1", "s-otra", 1, "assistant", llamada(3))
cerrar(con)

# centinelas.db: centinelas en `id`, `session_id`, `directory`, `path`, `title`, `slug`, `project_id` y el texto de una `user`
# (SC-014). El nombre de la base lo pone el test.
con = base("centinelas.db")
sesion(con, "s-CENTINELA-sesion", "/sintetico/CENTINELA-directorio", path="/sintetico/CENTINELA-ruta", title="CENTINELA-titulo",
       slug="CENTINELA-slug", proyecto="CENTINELA-proyecto")
fila(con, "o-CENTINELA-usuario", "s-CENTINELA-sesion", 1, "user", usuario("CENTINELA-texto"))
fila(con, "o-CENTINELA-llamada", "s-CENTINELA-sesion", 2, "assistant", llamada(4))
cerrar(con)

# corrupta.db: una llamada, y la raíz de session_message con un tipo de página imposible (FR-029, SC-020).
con = base("corrupta.db")
sesion(con, "s-corrupta", "/sintetico/corrupta")
fila(con, "o-corrupta-1", "s-corrupta", 1, "assistant", llamada(5))
raiz = con.execute("SELECT rootpage FROM sqlite_schema WHERE name='session_message'").fetchone()[0]
cerrar(con)
ruta = os.path.join(OUT, "corrupta.db")
b = bytearray(open(ruta, "rb").read())
P = struct.unpack(">H", b[16:18])[0]
b[(raiz - 1) * P] = 0x07
open(ruta, "wb").write(b)

# contraste.db (B7, `--scan`): el `session_v2` completo, el de internal/ingest/testdata/opencode/generar.sh, y dos sesiones con
# una llamada cada una. La sesión A lleva en sus totales la llamada del título, que no está en los mensajes (D-2): entrada +38,
# salida +3, razonamiento +4 (Δ`reasoning` ≠ 0, H13) y coste +0.000374.
SESION_V2_COMPLETA = """CREATE TABLE `session_v2` (`id` text PRIMARY KEY, `project_id` text NOT NULL, `workspace_id` text, `parent_id` text,
  `fork_session_id` text, `fork_boundary` text, `slug` text NOT NULL, `directory` text NOT NULL, `path` text, `title` text,
  `version` text NOT NULL, `share_url` text, `summary_additions` integer, `summary_deletions` integer, `summary_files` integer,
  `summary_diffs` text, `metadata` text, `cost` real DEFAULT 0 NOT NULL, `tokens_input` integer DEFAULT 0 NOT NULL,
  `tokens_output` integer DEFAULT 0 NOT NULL, `tokens_reasoning` integer DEFAULT 0 NOT NULL,
  `tokens_cache_read` integer DEFAULT 0 NOT NULL, `tokens_cache_write` integer DEFAULT 0 NOT NULL, `revert` text,
  `permission` text, `agent` text, `model` text, `time_created` integer NOT NULL, `time_updated` integer NOT NULL,
  `time_idle` integer, `time_viewed` integer, `idle_outcome` text, `time_compacting` integer, `time_archived` integer,
  `time_suspended` integer, `resume_attempts` integer DEFAULT 0 NOT NULL)"""
con = sqlite3.connect(os.path.join(OUT, "contraste.db"), isolation_level=None)
con.execute(SESION_V2_COMPLETA)
con.execute(MENSAJE)
con.execute("CREATE UNIQUE INDEX `session_message_session_seq_idx` ON `session_message` (`session_id`,`seq`)")
for sid, directorio, totales in (("s-contraste-a", "/sintetico/contraste-a", (0.012674, 138, 23, 11, 30, 4)),
                                 ("s-contraste-b", "/sintetico/contraste-b", (0.0045, 200, 10, 3, 0, 0))):
    con.execute("INSERT INTO session_v2(id, project_id, slug, directory, title, version, cost, tokens_input, tokens_output,"
                " tokens_reasoning, tokens_cache_read, tokens_cache_write, time_created, time_updated)"
                " VALUES (?,'prj-sintetico','slug-sintetico',?,'titulo-sintetico','2.0.26',?,?,?,?,?,?,?,?)",
                (sid, directorio) + totales + (T0, T0))
fila(con, "o-contraste-u", "s-contraste-a", 1, "user", usuario())
for mid, sid, n, k, coste in (("o-contraste-1", "s-contraste-a", 1, (100, 20, 7, 30, 4), 0.0123),
                              ("o-contraste-2", "s-contraste-b", 2, (200, 10, 3, 0, 0), 0.0045)):
    d = llamada(n)
    d["tokens"] = {"input": k[0], "output": k[1], "reasoning": k[2], "cache": {"read": k[3], "write": k[4]}}
    d["cost"] = coste
    fila(con, mid, sid, n + 1, "assistant", d)
cerrar(con)
PY
mkdir "$T/out"
python3 -I "$T/generar.py" "$T/out"
cp "$T"/out/*.db "$AQUI/"
( cd "$AQUI" && md5sum ./*.db | sed 's|\./||' )
