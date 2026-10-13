#!/usr/bin/env bash
# Genera las bases SINTÉTICAS de los tests de OpenCode en `cmd/permea` (spec 010, B6). Nada sale de una base real.
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
PY
mkdir "$T/out"
python3 -I "$T/generar.py" "$T/out"
cp "$T"/out/*.db "$AQUI/"
( cd "$AQUI" && md5sum ./*.db | sed 's|\./||' )
