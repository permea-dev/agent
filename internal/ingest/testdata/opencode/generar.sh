#!/usr/bin/env bash
# Genera las bases SINTÉTICAS de los tests de la pasada de OpenCode (spec 010, plan §6, T023). Nada sale de una base real.
#
#   bash internal/ingest/testdata/opencode/generar.sh
#
# Requiere python3 con SQLite >= 3.44 (sólo para regenerar; ni el binario ni la suite lo necesitan). Las bases se escriben
# con el módulo sqlite3 de Python, que ES SQLite. Una conexión retenida evita que el -wal se vuelque, como en
# internal/sqlitero/testdata/generar.sh. Las sales del WAL son aleatorias: esperado.json y los md5 se regeneran con ellas.
#
# Las secuencias entre pasadas son INSTANTÁNEAS (`paso1/`, `paso2/`…): el test las copia, una tras otra, sobre la misma ruta.
# Identificadores `o-…` y `s-…`, centinelas `CENTINELA-…` y rutas `/sintetico/…`: nada con la forma de un dato real.
set -euo pipefail
AQUI=$(cd "$(dirname "$0")" && pwd)
command -v python3 >/dev/null || { echo "falta python3" >&2; exit 1; }
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
cat > "$T/generar.py" <<'PY'
import hashlib, json, os, shutil, sqlite3, struct, sys
assert tuple(map(int, sqlite3.sqlite_version.split("."))) >= (3, 44, 0), "hace falta SQLite >= 3.44"
SAL = sys.argv[1]
T0 = 1760000000000  # ms; los tests inyectan `Ahora` relativo a esto
H = 3600 * 1000

SESION_V2 = """CREATE TABLE `session_v2` (`id` text PRIMARY KEY, `project_id` text NOT NULL, `workspace_id` text, `parent_id` text,
  `fork_session_id` text, `fork_boundary` text, `slug` text NOT NULL, `directory` text NOT NULL, `path` text, `title` text,
  `version` text NOT NULL, `share_url` text, `summary_additions` integer, `summary_deletions` integer, `summary_files` integer,
  `summary_diffs` text, `metadata` text, `cost` real DEFAULT 0 NOT NULL, `tokens_input` integer DEFAULT 0 NOT NULL,
  `tokens_output` integer DEFAULT 0 NOT NULL, `tokens_reasoning` integer DEFAULT 0 NOT NULL,
  `tokens_cache_read` integer DEFAULT 0 NOT NULL, `tokens_cache_write` integer DEFAULT 0 NOT NULL, `revert` text,
  `permission` text, `agent` text, `model` text, `time_created` integer NOT NULL, `time_updated` integer NOT NULL,
  `time_idle` integer, `time_viewed` integer, `idle_outcome` text, `time_compacting` integer, `time_archived` integer,
  `time_suspended` integer, `resume_attempts` integer DEFAULT 0 NOT NULL)"""
MENSAJE = """CREATE TABLE `session_message` (`id` text PRIMARY KEY, `session_id` text NOT NULL, `type` text NOT NULL,
  `seq` integer NOT NULL, `time_created` integer NOT NULL, `time_updated` integer NOT NULL, `data` text NOT NULL)"""
INDICES = ["CREATE UNIQUE INDEX `session_message_session_seq_idx` ON `session_message` (`session_id`,`seq`)",
           "CREATE INDEX `session_message_session_type_seq_idx` ON `session_message` (`session_id`,`type`,`seq`)"]

def base(ruta, pragmas=(), mensaje=MENSAJE, indices=INDICES):
    con = sqlite3.connect(ruta, isolation_level=None)
    for p in pragmas:
        con.execute(p)
    con.execute(SESION_V2)
    con.execute(mensaje)
    for i in indices:
        con.execute(i)
    return con

def sesion(con, sid, directorio, path=None, title="titulo-sintetico", slug="slug-sintetico", proyecto="prj-sintetico"):
    con.execute("INSERT INTO session_v2(id, project_id, slug, directory, path, title, version, time_created, time_updated)"
                " VALUES (?,?,?,?,?,?,'2.0.26',?,?)", (sid, proyecto, slug, directorio, path, title, T0, T0))

SEQ = {}
def fila(con, mid, sid, tipo, data, creada, columna_creada=None):
    """Inserta una fila de session_message; `data` es un dict (se serializa) o un texto tal cual."""
    SEQ[sid] = SEQ.get(sid, 0) + 1
    texto = data if isinstance(data, str) else json.dumps(data, separators=(",", ":"))
    con.execute("INSERT INTO session_message(id, session_id, type, seq, time_created, time_updated, data) VALUES (?,?,?,?,?,?,?)",
                (mid, sid, tipo, SEQ[sid], columna_creada if columna_creada is not None else creada, creada, texto))

def llamada(creada, completada=None, entrada=11, salida=5, razon=2, lectura=13, escritura=3, modelo="modelo-sintetico-opencode",
            relleno=0, sin_creada=False):
    t = {} if sin_creada else {"created": creada}
    if completada is not None:
        t["completed"] = completada
    d = {"role": "assistant", "time": t, "tokens": {"input": entrada, "output": salida, "reasoning": razon,
         "cache": {"read": lectura, "write": escritura}}, "cost": 0.0001}
    if modelo is not None:
        d["model"] = {"providerID": "proveedor-sintetico", "id": modelo, "variant": "default"}
    if relleno:
        d["pad"] = "0" * relleno
    return d

def usuario(creada, texto="texto-sintetico", relleno=0):
    d = {"role": "user", "time": {"created": creada}, "text": texto}
    if relleno:
        d["pad"] = "0" * relleno
    return d

def copiar(con, ruta, destino):
    """Copia la base y su -wal, con la conexión abierta: el -wal no se vuelca."""
    os.makedirs(destino, exist_ok=True)
    for suf in ("", "-wal"):
        if os.path.exists(ruta + suf):
            shutil.copy(ruta + suf, os.path.join(destino, "opencode.db" + suf))

def cerrar(con, ruta):
    con.close()
    for suf in ("-wal", "-shm"):
        assert not os.path.exists(ruta + suf), f"quedó {ruta + suf}"

def rowids(ruta):
    con = sqlite3.connect("file:" + ruta + "?mode=ro", uri=True)
    r = {i: rid for rid, i in con.execute("SELECT rowid, id FROM session_message")}
    con.close()
    return r

def niveles(ruta_db, tabla):
    """Niveles del árbol de una tabla, sobre una copia volcada (el -wal, dentro)."""
    tmp = os.path.join(SAL, "niveles")
    shutil.rmtree(tmp, ignore_errors=True)
    os.makedirs(tmp)
    for suf in ("", "-wal"):
        if os.path.exists(ruta_db + suf):
            shutil.copy(ruta_db + suf, os.path.join(tmp, "x.db" + suf))
    con = sqlite3.connect(os.path.join(tmp, "x.db"))
    con.execute("PRAGMA wal_checkpoint(TRUNCATE)")
    raiz = con.execute("SELECT rootpage FROM sqlite_schema WHERE name=?", (tabla,)).fetchone()[0]
    con.close()
    b = open(os.path.join(tmp, "x.db"), "rb").read()
    P = struct.unpack(">H", b[16:18])[0]
    P = 65536 if P == 1 else P
    n, k = raiz, 1
    while True:
        p = b[(n - 1) * P:n * P]
        h = 100 if n == 1 else 0
        if p[h] != 0x05:
            break
        n, k = struct.unpack(">I", p[h + 8:h + 12])[0], k + 1
    shutil.rmtree(tmp)
    return k

OUT = os.path.join(SAL, "out")
os.makedirs(OUT)
esp = {"sqlite": sqlite3.sqlite_version, "t0": T0}

# ═══ frontera/: 4 completadas y 1 en curso, en el -wal; páginas de 1 KB y una fila por hoja (SC-011) ═══════════════
r = os.path.join(SAL, "frontera.db")
con = base(r, ["PRAGMA page_size=1024", "PRAGMA journal_mode=WAL", "PRAGMA wal_autocheckpoint=0"])
sesion(con, "s-frontera", "/sintetico/frontera")
for i in range(1, 21):
    fila(con, f"o-frontera-{i:02d}", "s-frontera", "user", usuario(T0 + i, relleno=500), T0 + i)
con.execute("PRAGMA wal_checkpoint(TRUNCATE)")
con.execute("BEGIN")
for i in range(21, 25):
    fila(con, f"o-frontera-{i:02d}", "s-frontera", "assistant", llamada(T0 + i, T0 + 1000 + i, relleno=500), T0 + i)
fila(con, "o-frontera-25", "s-frontera", "assistant", llamada(T0 + 25, relleno=500), T0 + 25)
con.execute("COMMIT")
# Una fila por hoja y SIN desbordamiento: la carga local máxima de una página de 1 KB es de 989 B (SC-011 cuenta páginas).
assert con.execute("SELECT max(length(CAST(data AS BLOB))) FROM session_message").fetchone()[0] < 850
copiar(con, r, os.path.join(OUT, "frontera", "paso1"))
copiar(con, r, os.path.join(OUT, "frontera", "paso2"))
con.execute("UPDATE session_message SET data=?, time_updated=? WHERE id='o-frontera-25'",
            (json.dumps(llamada(T0 + 25, T0 + 2000, relleno=500), separators=(",", ":")), T0 + 2000))
copiar(con, r, os.path.join(OUT, "frontera", "paso3"))
assert os.path.getsize(os.path.join(OUT, "frontera", "paso1", "opencode.db-wal")) > 0
esp["frontera"] = {"completadas": [f"o-frontera-{i:02d}" for i in range(21, 25)], "en_curso": "o-frontera-25",
                   "h": 25, "niveles": niveles(os.path.join(OUT, "frontera", "paso1", "opencode.db"), "session_message")}
assert esp["frontera"]["niveles"] >= 2
cerrar(con, r)

def secuencia(nombre, pasos, sesiones=(("s-" + "x", "/sintetico/x"),)):
    """Una base sin WAL que pasa por varios estados: cada paso es (lista de sentencias o funciones)."""
    r = os.path.join(SAL, nombre + ".db")
    con = base(r)
    for sid, d in sesiones:
        sesion(con, sid, d)
    for k, paso in enumerate(pasos, 1):
        paso(con)
        copiar(con, r, os.path.join(OUT, nombre, f"paso{k}"))
    cerrar(con, r)
    return r

# ═══ inverso/: dos en curso que se completan al revés de su rowid (SC-011) ═════════════════════════════════════
def inv1(con):
    fila(con, "o-inverso-1", "s-inverso", "user", usuario(T0), T0)
    fila(con, "o-inverso-2", "s-inverso", "assistant", llamada(T0 + 1), T0 + 1)
    fila(con, "o-inverso-3", "s-inverso", "assistant", llamada(T0 + 2), T0 + 2)
def completar(mid, creada, cuando):
    def f(con):
        con.execute("UPDATE session_message SET data=? WHERE id=?",
                    (json.dumps(llamada(creada, cuando), separators=(",", ":")), mid))
    return f
secuencia("inverso", [inv1, completar("o-inverso-3", T0 + 2, T0 + 100), completar("o-inverso-2", T0 + 1, T0 + 200)],
          [("s-inverso", "/sintetico/inverso")])
esp["inverso"] = {"orden": ["o-inverso-3", "o-inverso-2"]}

# ═══ lista_borrada/: la fila en curso desaparece; la H, completada, sigue (SC-011) ════════════════════════════════
def lb1(con):
    fila(con, "o-borrada-1", "s-borrada", "assistant", llamada(T0), T0)
    fila(con, "o-queda-2", "s-queda", "assistant", llamada(T0 + 1, T0 + 50), T0 + 1)
def lb2(con):
    con.execute("DELETE FROM session_message WHERE session_id='s-borrada'")
    con.execute("DELETE FROM session_v2 WHERE id='s-borrada'")
secuencia("lista_borrada", [lb1, lb2], [("s-borrada", "/sintetico/borrada"), ("s-queda", "/sintetico/queda")])

# ═══ vacuum/: 2 borradas antes de H, VACUUM y 3 nuevas (SC-024) ═══════════════════════════════════════════════════
# Con `id text PRIMARY KEY` (el esquema de OpenCode), el VACUUM de SQLite 3.45 CONSERVA los rowid: la tabla tiene un
# índice. Sin índices, los renumera. Esta base no los lleva, para que el VACUUM renumere de verdad (medido el 12-10).
r = os.path.join(SAL, "vacuum.db")
con = base(r, mensaje=MENSAJE.replace("`id` text PRIMARY KEY", "`id` text NOT NULL"), indices=())
sesion(con, "s-vacuum", "/sintetico/vacuum")
for i in range(1, 9):
    fila(con, f"o-vacuum-{i:02d}", "s-vacuum", "assistant", llamada(T0 + i, T0 + 100 + i), T0 + i)
copiar(con, r, os.path.join(OUT, "vacuum", "antes"))
antes = rowids(r)
h_antes = max(antes.values())
id_h = next(i for i, rid in antes.items() if rid == h_antes)
con.execute("DELETE FROM session_message WHERE rowid IN (2, 3)")
con.execute("VACUUM")
for i in range(9, 12):
    fila(con, f"o-vacuum-{i:02d}", "s-vacuum", "assistant", llamada(T0 + i, T0 + 100 + i), T0 + i)
copiar(con, r, os.path.join(OUT, "vacuum", "despues"))
despues = rowids(r)
cerrar(con, r)
en_h = next((i for i, rid in despues.items() if rid == h_antes), None)
assert despues[id_h] != h_antes and en_h not in (None, id_h), "el VACUUM no renumeró"
esp["vacuum"] = {"h_antes": h_antes, "id_h_antes": id_h, "rowid_de_id_h_despues": despues[id_h], "id_en_h_despues": en_h,
                 "completadas_despues": sorted(despues)}

# ═══ borrada_h/: se borran las de rowid máximo y entran 2 que reutilizan rowids ≤ H (H10) ═════════════════════════
def bh1(con):
    for i in range(1, 6):
        fila(con, f"o-borradah-{i}", "s-borradah", "assistant", llamada(T0 + i, T0 + 100 + i), T0 + i)
def bh2(con):
    con.execute("DELETE FROM session_message WHERE id IN ('o-borradah-3', 'o-borradah-4', 'o-borradah-5')")
    for i in (6, 7):
        fila(con, f"o-borradah-{i}", "s-borradah", "assistant", llamada(T0 + i, T0 + 100 + i), T0 + i)
r = secuencia("borrada_h", [bh1, bh2], [("s-borradah", "/sintetico/borradah")])
nuevos = {i: rid for i, rid in rowids(os.path.join(OUT, "borrada_h", "paso2", "opencode.db")).items()
          if i in ("o-borradah-6", "o-borradah-7")}
assert sorted(nuevos.values()) == [3, 4], nuevos
esp["borrada_h"] = {"h_antes": 5, "nuevas": sorted(nuevos)}

# ═══ abandono/: Ahora +1 h, +25 h y +26 h (SC-025) ═══════════════════════════════════════════════════════════════
def ab1(con):
    fila(con, "o-abandono-c", "s-abandono", "assistant", llamada(T0, T0 + 5000), T0)
    fila(con, "o-abandono-a", "s-abandono", "assistant", llamada(T0), T0)
    fila(con, "o-abandono-b", "s-abandono", "assistant", llamada(T0), T0)
secuencia("abandono", [ab1, completar("o-abandono-b", T0, T0 + 24 * H + 30 * 60 * 1000),
                       completar("o-abandono-a", T0, T0 + 25 * H + 30 * 60 * 1000)], [("s-abandono", "/sintetico/abandono")])
esp["abandono"] = {"tarde": "o-abandono-b"}

# abandono_1h.db: una en curso de hace 25 h y otra de hace 1 h, ésta sin `$.time.created` (vale la columna).
r = os.path.join(SAL, "abandono_1h.db")
con = base(r)
sesion(con, "s-unahora", "/sintetico/unahora")
fila(con, "o-unahora-c", "s-unahora", "assistant", llamada(T0, T0 + 5000), T0)
fila(con, "o-unahora-25", "s-unahora", "assistant", llamada(T0), T0)
fila(con, "o-unahora-1", "s-unahora", "assistant", llamada(0, sin_creada=True), T0 + 24 * H, columna_creada=T0 + 24 * H)
cerrar(con, r)
shutil.copy(r, os.path.join(OUT, "abandono_1h.db"))

# ═══ clases.db: una fila de cada clase (SC-013); Ahora = T0 + 25 h ═══════════════════════════════════════════════
r = os.path.join(SAL, "clases.db")
con = base(r)
sesion(con, "s-clases", "/sintetico/clases")
fila(con, "o-clases-usuario", "s-clases", "user", usuario(T0), T0)
fila(con, "o-clases-evento", "s-clases", "assistant", llamada(T0, T0 + 10), T0)
fila(con, "o-clases-sinmodelo", "s-clases", "assistant", llamada(T0, T0 + 20, modelo=None), T0)
fila(con, "", "s-clases", "assistant", llamada(T0, T0 + 30), T0)
fila(con, "o-clases-nojson", "s-clases", "assistant", '{"role":"assistant","time":{"completed":', T0)
fila(con, "o-clases-negativa", "s-clases", "assistant", llamada(T0, T0 + 40, salida=-1), T0)
fila(con, "o-clases-encurso", "s-clases", "assistant", llamada(T0 + 24 * H), T0 + 24 * H)
fila(con, "o-clases-vieja", "s-clases", "assistant", llamada(T0), T0)
fila(con, "o-clases-idle", "s-clases", "idle", {"outcome": "succeeded", "time": {"created": T0}}, T0)
cerrar(con, r)
shutil.copy(r, os.path.join(OUT, "clases.db"))

# ═══ dos_sesiones.db: dos `directory`, y una fila sin su sesión en session_v2 (P-6) ═══════════════════════════════
r = os.path.join(SAL, "dos_sesiones.db")
con = base(r)
sesion(con, "s-uno", "/sintetico/uno")
sesion(con, "s-dos", "/sintetico/dos")
fila(con, "o-dos-1", "s-uno", "assistant", llamada(T0, T0 + 10), T0)
fila(con, "o-dos-2", "s-dos", "assistant", llamada(T0, T0 + 20), T0)
fila(con, "o-dos-3", "s-huerfana", "assistant", llamada(T0, T0 + 30), T0)
cerrar(con, r)
shutil.copy(r, os.path.join(OUT, "dos_sesiones.db"))
esp["dos_sesiones"] = {"o-dos-1": ["s-uno", "/sintetico/uno"], "o-dos-2": ["s-dos", "/sintetico/dos"], "o-dos-3": ["", ""]}

# ═══ centinelas.db: centinelas en `id`, `session_id`, `directory`, `path`, `title` y el texto de una `user` (H23) ════
r = os.path.join(SAL, "centinelas.db")
con = base(r)
sesion(con, "s-CENTINELA-sesion", "/sintetico/CENTINELA-directorio", path="/sintetico/CENTINELA-ruta", title="CENTINELA-titulo",
       slug="CENTINELA-slug", proyecto="CENTINELA-proyecto")
fila(con, "o-CENTINELA-usuario", "s-CENTINELA-sesion", "user", usuario(T0, texto="CENTINELA-texto"), T0)
fila(con, "o-CENTINELA-llamada", "s-CENTINELA-sesion", "assistant", llamada(T0, T0 + 10), T0)
fila(con, "o-CENTINELA-encurso", "s-CENTINELA-sesion", "assistant", llamada(T0 + 1), T0 + 1)
cerrar(con, r)
shutil.copy(r, os.path.join(OUT, "centinelas.db"))
esp["centinelas"] = {"h": "o-CENTINELA-encurso", "en_curso": ["o-CENTINELA-encurso"]}

# ═══ formatos/ y corrupta.db (P-9, FR-023, FR-029) ═══════════════════════════════════════════════════════════════
F = os.path.join(OUT, "formatos")
os.makedirs(F)
con = sqlite3.connect(os.path.join(F, "v1.db"))
con.executescript("CREATE TABLE session (id text PRIMARY KEY, directory text); CREATE TABLE message (id text PRIMARY KEY,"
                  " session_id text, data text); CREATE TABLE part (id text PRIMARY KEY, message_id text, data text);"
                  " INSERT INTO session VALUES ('s-v1', '/sintetico/v1'); INSERT INTO message VALUES ('o-v1-1', 's-v1', '{}');")
con.close()
for nombre, pragmas, mensaje in [("without_rowid", (), MENSAJE + " WITHOUT ROWID"),
                                 ("utf16", ("PRAGMA encoding='UTF-16le'",), MENSAJE),
                                 ("sin_columna", (), MENSAJE.replace(" `time_updated` integer NOT NULL,", ""))]:
    con = base(os.path.join(F, nombre + ".db"), pragmas, mensaje=mensaje, indices=())
    sesion(con, "s-" + nombre, "/sintetico/" + nombre)
    cols = "id, session_id, type, seq, time_created" + ("" if nombre == "sin_columna" else ", time_updated") + ", data"
    vals = ("o-" + nombre, "s-" + nombre, "assistant", 1, T0) + (() if nombre == "sin_columna" else (T0,))
    con.execute(f"INSERT INTO session_message({cols}) VALUES ({','.join('?' * (len(vals) + 1))})",
                vals + (json.dumps(llamada(T0, T0 + 10)),))
    con.close()
r = os.path.join(OUT, "corrupta.db")
con = base(r)
sesion(con, "s-corrupta", "/sintetico/corrupta")
fila(con, "o-corrupta-1", "s-corrupta", "assistant", llamada(T0, T0 + 10), T0)
raiz = con.execute("SELECT rootpage FROM sqlite_schema WHERE name='session_message'").fetchone()[0]
con.close()
b = bytearray(open(r, "rb").read())
P = struct.unpack(">H", b[16:18])[0]
b[(raiz - 1) * P] = 0x07  # tipo de página imposible en la raíz de session_message
open(r, "wb").write(b)
esp["corrupta"] = {"pagina": raiz}

json.dump(esp, open(os.path.join(OUT, "esperado.json"), "w"), indent=1, sort_keys=True)
open(os.path.join(OUT, "esperado.json"), "a").write("\n")
PY
python3 -I "$T/generar.py" "$T"

# ═══ Copiar al repo (sin tocar filas/, de B3) ════════════════════════════════════════════════════════════════════
cd "$T/out"
for d in *; do rm -rf "${AQUI:?}/$d"; cp -r "$d" "$AQUI/$d"; done
( cd "$AQUI" && find . -path ./filas -prune -o -type f ! -name generar.sh -print | sed 's|^\./||' | sort | xargs md5sum )
