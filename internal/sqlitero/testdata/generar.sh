#!/usr/bin/env bash
# Genera las bases SINTÉTICAS de los tests de internal/sqlitero (spec 010, plan §6). Nada sale de una base real.
#
#   bash internal/sqlitero/testdata/generar.sh
#
# Requiere sqlite3 >= 3.44 y python3 (sólo para regenerar; ni el binario ni la suite los necesitan). Las sales del WAL son
# aleatorias, así que cada ejecución cambia los bytes de i/ y r/: por eso esperado.json se regenera con ellos, desde el
# módulo sqlite3 de Python, que no comparte código con el lector. Después de regenerar, los md5 van al registro.
set -euo pipefail
AQUI=$(cd "$(dirname "$0")" && pwd)
command -v sqlite3 >/dev/null || { echo "falta sqlite3" >&2; exit 1; }
command -v python3 >/dev/null || { echo "falta python3" >&2; exit 1; }
python3 -I -c 'import sqlite3, sys; sys.exit(0 if tuple(map(int, sqlite3.sqlite_version.split("."))) >= (3, 44, 0) else 1)' ||
  { echo "hace falta SQLite >= 3.44" >&2; exit 1; }

T=$(mktemp -d)
fin() { exec 3>&- 4>&- 2>/dev/null || true; kill $(jobs -p) 2>/dev/null || true; wait 2>/dev/null || true; rm -rf "$T"; }
trap fin EXIT
esperar() { for _ in $(seq 1 300); do [ -e "$1" ] && return 0; sleep 0.1; done; echo "timeout: $1" >&2; exit 1; }

# ═══ I, R y S: el esquema de OpenCode 2.0.26, con filas inventadas ═══════════════════════════════════════════════
mkdir -p "$T/vivo" "$T/inst"
V=$T/vivo/opencode.db
sqlite3 "$V" >/dev/null <<'SQL'
PRAGMA journal_mode=WAL;
CREATE TABLE `session_v2` (`id` text PRIMARY KEY, `project_id` text NOT NULL, `workspace_id` text, `parent_id` text,
  `fork_session_id` text, `fork_boundary` text, `slug` text NOT NULL, `directory` text NOT NULL, `path` text, `title` text,
  `version` text NOT NULL, `share_url` text, `summary_additions` integer, `summary_deletions` integer, `summary_files` integer,
  `summary_diffs` text, `metadata` text, `cost` real DEFAULT 0 NOT NULL, `tokens_input` integer DEFAULT 0 NOT NULL,
  `tokens_output` integer DEFAULT 0 NOT NULL, `tokens_reasoning` integer DEFAULT 0 NOT NULL,
  `tokens_cache_read` integer DEFAULT 0 NOT NULL, `tokens_cache_write` integer DEFAULT 0 NOT NULL, `revert` text,
  `permission` text, `agent` text, `model` text, `time_created` integer NOT NULL, `time_updated` integer NOT NULL,
  `time_idle` integer, `time_viewed` integer, `idle_outcome` text, `time_compacting` integer, `time_archived` integer,
  `time_suspended` integer, `resume_attempts` integer DEFAULT 0 NOT NULL);
CREATE TABLE `session_message` (`id` text PRIMARY KEY, `session_id` text NOT NULL, `type` text NOT NULL,
  `seq` integer NOT NULL, `time_created` integer NOT NULL, `time_updated` integer NOT NULL, `data` text NOT NULL);
CREATE UNIQUE INDEX `session_message_session_seq_idx` ON `session_message` (`session_id`,`seq`);
CREATE INDEX `session_message_session_type_seq_idx` ON `session_message` (`session_id`,`type`,`seq`);
SQL
# filas sesion desde hasta relleno: `relleno` bytes de ceros en hex (x2) cada 5 filas, para que desborden.
filas() {
cat <<SQL
WITH RECURSIVE n(i) AS (SELECT $2 UNION ALL SELECT i+1 FROM n WHERE i<$3)
INSERT INTO session_message(id, session_id, type, seq, time_created, time_updated, data)
SELECT printf('o-%s-%05d', '$1', i), '$1', CASE WHEN i%5=0 THEN 'user' ELSE 'assistant' END, i, 1000+i, 1000+i,
  json_object('time', CASE WHEN i%7=3 THEN json_object('created', 1000+i) ELSE json_object('created', 1000+i, 'completed', 2000+i) END,
    'tokens', json_object('input', 100*i+7, 'output', 10*i+1, 'reasoning', i%3,
       'cache', json_object('read', CASE WHEN i%4=0 THEN 50*i ELSE 0 END, 'write', i%6)),
    'cost', (100*i+7)*0.00000075 + (10*i+1+i%3)*0.00000375,
    'pad', CASE WHEN i%5=2 THEN hex(zeroblob($4)) ELSE '' END)
FROM n;
SQL
}
sesion() {
echo "INSERT INTO session_v2(id, project_id, slug, directory, version, cost, tokens_input, tokens_output, tokens_reasoning,
  tokens_cache_read, tokens_cache_write, time_created, time_updated)
VALUES('$1', 'prj-sintetico', 'slug', '/sintetico/proyecto', '2.0.26', $2*0.000001234567, $2*3, $2, $2/7, $2*2, $2%5, 1, 1);"
}
# Estado 0: se vuelca a la base al cerrar.
{ sesion s-1 101; filas s-1 1 5 10; } | sqlite3 "$V"
[ ! -e "$V-wal" ] || { echo "el estado 0 no se volcó" >&2; exit 1; }
# Una conexión que se queda abierta, para que nadie vuelque el WAL.
mkfifo "$T/h.fifo"; sqlite3 "$V" < "$T/h.fifo" >/dev/null 2>&1 &
exec 3>"$T/h.fifo"; printf 'SELECT count(*) FROM session_v2;\n.system touch %s/h.listo\n' "$T" >&3; esperar "$T/h.listo"
# T1, T2 y T3, confirmadas en el WAL.
{ echo "PRAGMA wal_autocheckpoint=0; BEGIN;"; sesion s-2 2002; filas s-2 1 20 4500
  echo "UPDATE session_v2 SET tokens_input=tokens_input+1000, cost=cost+0.001234567 WHERE id='s-1'; COMMIT;"; } | sqlite3 "$V" >/dev/null
{ echo "PRAGMA wal_autocheckpoint=0; BEGIN;
UPDATE session_message SET data=json_set(data,'\$.time.completed',3000), time_updated=3000
  WHERE session_id='s-2' AND json_extract(data,'\$.time.completed') IS NULL AND seq<=10;
UPDATE session_message SET data=json_set(data,'\$.pad',hex(zeroblob(6000))) WHERE session_id='s-2' AND seq=1;
UPDATE session_v2 SET tokens_output=tokens_output+77, cost=cost+0.000000321 WHERE id='s-2'; COMMIT;"; } | sqlite3 "$V" >/dev/null
{ echo "PRAGMA wal_autocheckpoint=0; BEGIN;"; sesion s-3 30303; filas s-3 1 30 5000; echo "COMMIT;"; } | sqlite3 "$V" >/dev/null
# Una transacción a medias, con caché mínima para que vuelque frames sin commit.
mkfifo "$T/m.fifo"; sqlite3 "$V" < "$T/m.fifo" >/dev/null 2>&1 &
exec 4>"$T/m.fifo"
{ echo "PRAGMA wal_autocheckpoint=0; PRAGMA cache_size=5; BEGIN;"; sesion s-amedias 999999; filas s-amedias 1 150 500
  echo "UPDATE session_v2 SET tokens_input=tokens_input+999999 WHERE id='s-1';"; echo ".system touch $T/m.listo"; } >&4
esperar "$T/m.listo"
# I: la instantánea. El -shm no se versiona: el lector no lo abre nunca (plan §1).
cp "$V" "$V-wal" "$T/inst/"
echo "ROLLBACK;" >&4; exec 4>&-; exec 3>&-; wait
[ ! -e "$V-wal" ] || { echo "quedó -wal en la base viva" >&2; exit 1; }

# ═══ Los raros y profundo.db ══════════════════════════════════════════════════════════════════════════════════
mkdir -p "$T/raros"
raro() { # raro nombre "pragmas previos" filas relleno
  sqlite3 "$T/raros/$1.db" >/dev/null <<SQL
$2
CREATE TABLE t (\`id\` text, "data" text);
WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i<$3)
INSERT INTO t SELECT printf('o-%s-%03d', '$1', i), printf('dato-%03d-%s', i, hex(zeroblob($4))) FROM n;
SQL
}
raro p512 "PRAGMA page_size=512;" 40 60
raro p65536 "PRAGMA page_size=65536;" 40 60
raro utf16 "PRAGMA encoding='UTF-16le';" 10 10
sqlite3 "$T/raros/reserva.db" ".filectrl reserve_bytes 16" "CREATE TABLE t (\`id\` text, \"data\" text);
  WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i<40)
  INSERT INTO t SELECT printf('o-reserva-%03d', i), printf('dato-%03d-%s', i, hex(zeroblob(60))) FROM n;" >/dev/null
sqlite3 "$T/raros/without_rowid.db" "CREATE TABLE t (\`id\` text PRIMARY KEY, \"data\" text) WITHOUT ROWID;
  INSERT INTO t VALUES ('o-wr-001', 'dato'), ('o-wr-002', 'dato');" >/dev/null
sqlite3 "$T/raros/desborda.db" "CREATE TABLE t (\`id\` text, \"data\" text);
  INSERT INTO t VALUES ('o-desborda-1', printf('corta-%s', hex(zeroblob(50)))),
    ('o-desborda-8k', printf('8k-%s', hex(zeroblob(4500)))), ('o-desborda-64k', printf('64k-%s', hex(zeroblob(40000))));" >/dev/null
sqlite3 "$T/profundo.db" "PRAGMA page_size=512; CREATE TABLE fila (\`id\` text, \`n\` integer, \`data\` text);
  WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i<2000)
  INSERT INTO fila SELECT printf('o-p-%05d', i), i, printf('dato-%05d-abcdefghij', i) FROM n;
  DELETE FROM fila WHERE rowid = 777;" >/dev/null

# ═══ R, las comprobaciones y esperado.json ═════════════════════════════════════════════════════════════════════
cat > "$T/esperado.py" <<'PY'
import hashlib, json, os, shutil, sqlite3, struct, sys
T = sys.argv[1]

def varint(b, i):
    v = 0
    for k in range(8):
        v = (v << 7) | (b[i + k] & 0x7F)
        if b[i + k] < 0x80:
            return v, k + 1
    return (v << 8) | b[i + 8], 9

def frames(wal):
    """Frames en cadena (sales y sumas) y los índices de commit, como en la prueba."""
    magia, _, tp, _, s1, s2, c1, c2 = struct.unpack(">8I", wal[:32])
    fmt = ">" if magia & 1 else "<"
    def suma(d, a, b):
        for i in range(0, len(d), 8):
            x, y = struct.unpack(fmt + "2I", d[i:i + 8])
            a = (a + x + b) & 0xFFFFFFFF
            b = (b + y + a) & 0xFFFFFFFF
        return a, b
    a, b = suma(wal[:24], 0, 0)
    assert (a, b) == (c1, c2), "cabecera del WAL inválida"
    cadena, commits, off = [], [], 32
    while off + 24 + tp <= len(wal):
        pag, commit, x1, x2, f1, f2 = struct.unpack(">6I", wal[off:off + 24])
        if (x1, x2) != (s1, s2):
            break
        a, b = suma(wal[off:off + 8], a, b)
        a, b = suma(wal[off + 24:off + 24 + tp], a, b)
        if (a, b) != (f1, f2):
            break
        cadena.append((off, pag))
        if commit:
            commits.append(len(cadena))
        off += 24 + tp
    return tp, cadena, commits

def fin_de_data(p, U):
    """Desplazamiento del último byte del `data` (texto, última columna) de una fila de session_message sin
    desbordamiento de la hoja p, o None. Es la garantía de H9: romperlo cambia el sha256 de `data` y rompe su JSON.
    No se usa el último byte de la PÁGINA: en las hojas de T3, la celda pegada al final es una fila que desborda, y sus
    últimos 4 bytes son el puntero de desbordamiento (medido al generar, 2026-10-11)."""
    if p[0] != 0x0D:
        return None
    for k in range(struct.unpack(">H", p[3:5])[0]):
        c = struct.unpack(">H", p[8 + 2 * k:10 + 2 * k])[0]
        P, n1 = varint(p, c)
        _, n2 = varint(p, c + n1)
        ini = c + n1 + n2
        if P > U - 35:
            continue
        hdr, h = varint(p, ini)
        tipos, j = [], ini + h
        while j < ini + hdr:
            st, k2 = varint(p, j)
            tipos.append(st)
            j += k2
        fin = ini + P - 1
        if len(tipos) == 7 and tipos[-1] >= 13 and tipos[-1] % 2 == 1 and p[fin] == 0x7D:
            return fin
    return None

def resumen(dirbase):
    """Lo que sqlite3 ve en una copia: recuentos, sumas y sha256 de `data` en orden de session_id, seq."""
    tmp = os.path.join(T, "lee")
    shutil.rmtree(tmp, ignore_errors=True)
    shutil.copytree(dirbase, tmp)
    con = sqlite3.connect(os.path.join(tmp, "opencode.db"))
    h = hashlib.sha256()
    filas = 0
    for (d,) in con.execute("SELECT data FROM session_message ORDER BY session_id, seq"):
        h.update(d.encode())
        filas += 1
    ses, ti, to = con.execute("SELECT count(*), sum(tokens_input), sum(tokens_output) FROM session_v2").fetchone()
    ids = hashlib.sha256("".join(r[0] for r in con.execute("SELECT id FROM session_v2 ORDER BY id")).encode()).hexdigest()
    con.close()
    shutil.rmtree(tmp)
    return {"session_message": {"filas": filas, "sha256_data": h.hexdigest()},
            "session_v2": {"filas": ses, "tokens_input": ti, "tokens_output": to, "sha256_ids": ids}}

def tabla_t(ruta):
    con = sqlite3.connect("file:" + ruta + "?mode=ro", uri=True)
    h = hashlib.sha256()
    n = 0
    for (d,) in con.execute("SELECT data FROM t ORDER BY rowid"):
        h.update(d.encode())
        n += 1
    con.close()
    return {"filas": n, "sha256_data": h.hexdigest()}

esp = {"sqlite": sqlite3.sqlite_version}
inst = os.path.join(T, "inst")
wal = bytearray(open(os.path.join(inst, "opencode.db-wal"), "rb").read())
tp, cadena, commits = frames(bytes(wal))
assert len(commits) == 3, f"se esperaban 3 commits (T1–T3) y hay {len(commits)}"
assert len(cadena) > commits[-1], "la transacción a medias no volcó frames al WAL"
# R: romper el último byte del `data` de una fila de session_message en una hoja de T3, dentro de su celda (H9).
roto = None
for idx in range(commits[-2], commits[-1]):
    off, pag = cadena[idx]
    fin = fin_de_data(wal[off + 24:off + 24 + tp], tp) if pag != 1 else None
    if fin is not None:
        roto = (idx, off + 24 + fin)
        break
assert roto, "ninguna hoja de session_message de T3 tiene una celda de `data` sin desbordamiento"
r = bytearray(wal)
r[roto[1]] ^= 0x01
os.makedirs(os.path.join(T, "r"))
shutil.copy(os.path.join(inst, "opencode.db"), os.path.join(T, "r", "opencode.db"))
open(os.path.join(T, "r", "opencode.db-wal"), "wb").write(r)
os.makedirs(os.path.join(T, "s"))
shutil.copy(os.path.join(T, "vivo", "opencode.db"), os.path.join(T, "s", "opencode.db"))
shutil.copytree(inst, os.path.join(T, "i"))

esp["i"] = resumen(os.path.join(T, "i"))
esp["i"]["frames_confirmados"] = commits[-1]
esp["r"] = resumen(os.path.join(T, "r"))
esp["r"]["frames_confirmados"] = commits[-2]
esp["r"]["frame_roto"] = roto[0] + 1
esp["s"] = resumen(os.path.join(T, "s"))
esp["s"]["frames_confirmados"] = 0
assert esp["i"]["session_v2"]["filas"] == 3 and esp["r"]["session_v2"]["filas"] == 2, "R debe perder T3 entera"
assert esp["s"]["session_message"] == esp["i"]["session_message"], "S no es I sin la transacción a medias"

raros = os.path.join(T, "raros")
esp["raros"] = {n: tabla_t(os.path.join(raros, n + ".db")) for n in ("p512", "p65536", "reserva", "utf16")}
con = sqlite3.connect("file:" + os.path.join(raros, "desborda.db") + "?mode=ro", uri=True)
esp["desborda"] = {i: {"bytes": len(d.encode()), "sha256": hashlib.sha256(d.encode()).hexdigest()}
                   for i, d in con.execute("SELECT id, data FROM t")}
con.close()
assert esp["desborda"]["o-desborda-8k"]["bytes"] > 8192 and esp["desborda"]["o-desborda-64k"]["bytes"] > 65536

# profundo.db: niveles, una clave de página interior que existe, una fila de hoja que no es clave, la ausente y el máximo.
ruta = os.path.join(T, "profundo.db")
b = open(ruta, "rb").read()
P = struct.unpack(">H", b[16:18])[0]
con = sqlite3.connect("file:" + ruta + "?mode=ro", uri=True)
raiz = con.execute("SELECT rootpage FROM sqlite_schema WHERE name='fila'").fetchone()[0]
pagina = lambda n: b[(n - 1) * P:n * P]
claves, niveles, n = set(), 1, raiz
while pagina(n)[0] == 0x05:
    niveles += 1
    c = struct.unpack(">H", pagina(n)[12:14])[0]  # la primera celda: su hijo izquierdo
    n = struct.unpack(">I", pagina(n)[c:c + 4])[0]
def interiores(n):
    p = pagina(n)
    if p[0] != 0x05:
        return
    for k in range(struct.unpack(">H", p[3:5])[0]):
        c = struct.unpack(">H", p[12 + 2 * k:14 + 2 * k])[0]
        claves.add(varint(p, c + 4)[0])
        interiores(struct.unpack(">I", p[c:c + 4])[0])
    interiores(struct.unpack(">I", p[8:12])[0])
interiores(raiz)
existe = lambda r: con.execute("SELECT count(*) FROM fila WHERE rowid=?", (r,)).fetchone()[0] == 1
clave = min(k for k in claves if existe(k))
en_hoja = min(r for r in range(1, 2001) if r not in claves and existe(r))
maximo = con.execute("SELECT max(rowid), count(*) FROM fila").fetchone()
assert niveles >= 3 and not existe(777)
esp["profundo"] = {"niveles": niveles, "clave_interior": clave, "en_hoja": en_hoja, "ausente": 777,
                   "max_rowid": maximo[0], "filas": maximo[1],
                   "datos_en_hoja": con.execute("SELECT id, n, data FROM fila WHERE rowid=?", (en_hoja,)).fetchone()}
con.close()
json.dump(esp, open(os.path.join(T, "esperado.json"), "w"), indent=1, sort_keys=True)
open(os.path.join(T, "esperado.json"), "a").write("\n")
PY
python3 -I "$T/esperado.py" "$T"

# ═══ Copiar al repo ═══════════════════════════════════════════════════════════════════════════════════════════
for d in i r s raros; do rm -rf "${AQUI:?}/$d"; cp -r "$T/$d" "$AQUI/$d"; done
cp "$T/profundo.db" "$T/esperado.json" "$AQUI/"
( cd "$AQUI" && find i r s raros profundo.db esperado.json -type f | sort | xargs md5sum )
