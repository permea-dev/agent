# Quickstart · Validación de 005 Adhesión a proyecto

**Fecha**: 2026-08-18 · **Spec**: [spec.md](./spec.md) · **Contratos**:
[adhesion.md](./contracts/adhesion.md) · [cli.md](./contracts/cli.md)

Guía de validación ejecutable: qué hay que poder demostrar para dar la feature por buena. **No** trae
código de implementación — eso es de `tasks.md`.

> ## ⚠️ ESTA GUÍA TIENE DOS MITADES, Y NO SE VALIDAN IGUAL
>
> | | Qué es | Cómo se ejecuta |
> |---|---|---|
> | **V1 – V9** | Comportamiento **del agente** | **Automatizable en este repositorio.** Van a la suite |
> | **C1 – C4** | **La ceremonia**: efecto sobre **la plataforma** | **Validación manual contra una instalación real.** **NO son tests, y fingirlos está prohibido** — ver §Ceremonia |
>
> Mezclarlas es el error que esta feature tiene declarado en su propia spec: **SC-006 prohíbe
> expresamente el test que finja** lo que solo la plataforma puede demostrar.

---

## Prerrequisitos

```bash
cd ~/dev/agente
go version          # toolchain del repo
make build          # binario en bin/permea
```

Las puertas de calidad de la constitución, que deben quedar en verde al final:

```bash
go vet ./...
golangci-lint run
go test ./...       # línea base a preservar: 9 paquetes ok, 0 [no test files], 0 FAIL
```

> ⚠️ **`which permea` NO es el binario que compilas.** Resuelve a una copia instalada en
> `~/.local/bin`, que es **otro fichero**. Toda esta guía usa **`./bin/permea` por ruta explícita**. Un
> `project join` que responde «uso: permea …» no es un fallo de la feature: es el binario equivocado.

### Aislamiento obligatorio — antes de cualquier validación que toque config o cola

**Ninguna validación de esta guía puede ejecutarse contra la instalación real del desarrollador.**
`config.DataDir()` resuelve por `os.UserConfigDir()` (`internal/config/config.go:51`), así que basta
apuntar las variables de entorno del hogar a un directorio de pruebas y **todo** el estado del agente
—`config.json`, `state.json`, `queue.jsonl`, `salt`, `machine_id`— aterriza allí.

```bash
export PERMEA_SANDBOX="$(mktemp -d)"
export HOME="$PERMEA_SANDBOX/home"          # Linux / macOS
export USERPROFILE="$PERMEA_SANDBOX/home"   # Windows
export XDG_CONFIG_HOME="$HOME/.config"      # explícito: no depender del valor heredado
mkdir -p "$XDG_CONFIG_HOME"

# Comprobación de que el aislamiento funciona ANTES de seguir:
./bin/permea status      # debe decir "no enrolado" — si dice lo contrario, HOME no se aplicó
```

**Esa última comprobación no es ceremonia**: sin el sandbox, las validaciones que enrolan sobrescriben
el `config.json` real.

---

## El banco TLS local — **sin esto no hay forma de ejercer la feature**

**FR-017 no tiene exención**: la adhesión exige transporte seguro con la misma dureza que la emisión de
eventos, y `php artisan serve` sirve en claro. Así que **el banco TLS local no es un accesorio de
comodidad: es la única forma de ejercer la feature contra una plataforma real.**

Montado el 2026-08-18 y ya disponible:

| Pieza | Dónde |
|---|---|
| Terminador TLS | `~/tls-local/proxy/tlsproxy` — escucha en **`:8443`** y reenvía a `127.0.0.1:8000` |
| CA local | `~/tls-local/ca.crt` |
| Certificado de servidor | `~/tls-local/localhost.crt` + `.key`, con SAN `DNS:localhost, IP:127.0.0.1, IP:0:0:0:0:0:0:0:1` |

```bash
# Terminal 1 — el backend, en claro
cd ~/dev/permea-platform/backend && php artisan serve --host=127.0.0.1 --port=8000

# Terminal 2 — el terminador TLS
cd ~/tls-local/proxy && ./tlsproxy

# Comprobar el canal ANTES de meter al agente (sin -k: la CA debe bastar)
curl --cacert ~/tls-local/ca.crt https://localhost:8443/api/v1/ingest -i
```

**Y el agente tiene que confiar en esa CA.** Medido el 2026-08-18: en Go, `SSL_CERT_FILE` es
**aditivo**, no sustitutivo —añade la CA **sin perder ninguna raíz pública**— y **no requiere `sudo`
ni modificar la máquina**:

```bash
export SSL_CERT_FILE="$HOME/tls-local/ca.crt"     # ← ojo: HOME está redefinido por el sandbox;
                                                  #   usar la ruta real, p. ej. /home/bfgnet/tls-local/ca.crt
```

> ⚠️ **El sandbox redefine `HOME`, y la CA vive en el HOME real.** Es el choque de las dos secciones
> anteriores y muerde en cuanto se combinan: `SSL_CERT_FILE` debe apuntar a la **ruta absoluta real**
> de `ca.crt`, no a `$HOME/tls-local/ca.crt`.

---

# PARTE A · Validaciones automatizables (V1 – V9)

## V1 · La identidad presentada es la que se estampa (SC-001 · FR-004, FR-005)

**La validación que sostiene la feature.** Se demuestra **origen compartido**, no coincidencia caso a
caso:

1. **Punto único**: la identidad que presenta el comando y la que se estampa en los eventos salen de
   la misma derivación. Alterar ese punto **cambia las dos a la vez**.
2. **Comparación** sobre cuatro clases de árbol —raíz de un proyecto, subdirectorio profundo, árbol
   paralelo, directorio sin raíz—: iguales carácter a carácter en las cuatro.
3. **La comparación sabe fallar**: una alteración deliberada en el punto único **pone (2) en rojo**.

> ⚠️ **NO usar `--scan` para obtener las identidades**: usa un salt literal (`cmd/permea/main.go:342`),
> así que sus refs **están en otro espacio de valores** y no comparan con nada. Ver `research.md` §R5.

## V2 · La ausencia de raíz rehúsa **antes de hablar** (SC-004 · FR-006)

Lanzar el comando desde un directorio sin raíz de proyecto por encima. **Cero peticiones emitidas.**

**Con su observador declarado**: un destino instrumentado que **cuenta** peticiones. Y **con su caso
positivo**: el **mismo** destino, con el comando lanzado **dentro** de un árbol, **registra exactamente
una**. Sin la segunda mitad, «no se emitió» no se distingue de «no se miró».

## V3 · Los dos desenlaces indistinguibles (SC-002 · FR-010)

Presentar el mismo código dos veces desde la misma instalación. Las **tres** piezas:
**(a)** cada salida no vacía y del tipo que le toca —la de éxito **nombra un Proyecto**, comprobado
porque contiene la denominación **que existe en ese momento**, no contra un texto fijado—;
**(b)** idénticas **comparadas entre sí**; **(c)** una diferencia deliberada **pone (b) en rojo**.

## V4 · Los rechazos no filtran (SC-003 · FR-011, FR-012)

Dos códigos no utilizables **por causas distintas** → mensajes **byte a byte idénticos**, con las
mismas tres piezas. Y un código utilizable desde una instalación ya unida a **otro** Proyecto → se
informa **sin nombrarlo**.

## V5 · El código no aparece en ninguna salida (SC-005 · FR-020)

Para **cada uno** de los ocho desenlaces de [`cli.md`](./contracts/cli.md): generar las subcadenas de
**longitud ocho** del valor presentado y buscarlas todas en la salida completa —ambos canales—.
**Cero apariciones.**

## V6 · Nada se escribe en local (SC-007 · FR-019)

Capturar **íntegro** el conjunto de artefactos locales —configuración, estado de lectura, cola,
secretos— **antes** de la ejecución, y comparar byte a byte **contra esa captura** tras cada uno de los
ocho desenlaces. **Con su caso positivo**: una operación que **sí** modifica el estado local **hace
fallar la misma comparación**.

## V7 · Sin transporte seguro no se completa (SC-008 · FR-017)

Cuatro clases, **enumeradas y reproducibles**: (a) destino en claro · (b) destino en claro sobre la
máquina local · (c) destino sin transporte seguro **y con un código utilizable**, para que el rechazo
no pueda atribuirse al código · (d) los tres anteriores con **cada** ajuste de configuración que la
instalación admita.

## V8 · La petición nunca se encola (SC-010 · FR-018)

Con el servidor **inalcanzable**: inspeccionar la cola **antes y después**. **No crece.** Y con su caso
positivo: una emisión ordinaria de eventos con el destino igualmente caído **sí la hace crecer**. Si no
crece en ninguno de los dos casos, el observador no está mirando.

## V9 · Regresión cero en la derivación (SC-009 · FR-014, FR-015, FR-016)

**La puerta que demuestra que exponer «hubo raíz» no cambió el camino de ingesta.** Repetir la pasada
de referencia **reutilizando las semillas del bloque REPRODUCCIÓN** de
`../004-identidad-de-proyecto/baseline-sc004.tsv` y comparar **las tres columnas** de ese fichero.

> **Con otras semillas los refs no comparan y un «fallo» no significaría nada.** Es el mismo
> procedimiento que 004 usó en su T007 para demostrar neutralidad.

## V-canales · El reparto de stdout/stderr (SC-011)

Transversal a V1–V9: en cada desenlace, **capturar los dos canales POR SEPARADO** y comprobar que el
éxito deja **stdout no vacío y stderr sin el desenlace**, y que el rehúse deja **stderr no vacío y
stdout VACÍO**. **Capturados mezclados no cuenta como pasado.**

Y **las dos vías de entrada** (SC-011 A): el mismo código por argumento y por entrada estándar produce
desenlaces idénticos, con las tres piezas de V3.

---

# PARTE B · CEREMONIA — validación manual contra plataforma real

> ## ⚠️ ESTO NO SON TESTS, Y AUTOMATIZARLOS ESTÁ PROHIBIDO
>
> **SC-006** lo declara en la spec: lo que estas comprobaciones observan es comportamiento **de la
> plataforma** —la agrupación resuelta en lectura—, y la suite de este repositorio **tendría que
> fabricar la agrupación para probarlo, con lo que estaría comprobando su propio simulacro**. La spec
> dice literalmente que **un test automático que lo finja es peor que no tenerlo**: daría verde
> permanente sobre la propiedad que justifica la feature entera.
>
> **Se ejecutan a mano, una vez, y se anota el resultado.**

## Sujeto de la ceremonia — elegido, y con motivo

**`~/dev/test/RecetApp`**, sin mapear, **1 895 eventos reales** (221 del 12-ago + 1 674 del 13-ago),
identificado en el descubrimiento del 2026-08-18 por coincidencia exacta de recuentos.

**Por qué éste y no uno de laboratorio**: tiene volumen real, está **sin agrupar ahora mismo**, y su
consumo es **anterior** a la unión — que es exactamente lo que SC-006 mide, el efecto retroactivo. Con
un proyecto recién creado **no habría histórico que apareciera**.

## Preparación

Banco TLS levantado (arriba), agente enrolado contra `https://localhost:8443/...`, y un código de
adhesión acuñado desde el panel para el Proyecto destino.

### ⛔ EL MODAL DE EMISIÓN NO SE CAPTURA — y esto salió de ejecutar la ceremonia

**El código es una credencial y el modal que lo revela lo enseña UNA VEZ**, así que la reacción
natural es hacerle una captura de pantalla para no perderlo. **No se hace.** Una captura vive en el
carrete, en el portapapeles, en el historial de la herramienta con la que se tomó y a veces en una
copia en la nube — sitios donde una credencial no se puede retirar.

**Medido en la ceremonia del 2026-08-19**: pasó. El código acabó dentro de una captura de pantalla.
**Quedó neutralizado por una propiedad del diseño** —emitir un código revoca el anterior, así que
para cuando la captura existía el valor ya no servía—, pero **el procedimiento lo permitía y ahora no
debe**: apoyarse en que la revocación llegue a tiempo no es una defensa, es una carambola.

**Qué hacer en su lugar**: llevar el código del modal al comando **sin pasar por una imagen** —copiar
al portapapeles y consumirlo por la entrada estándar— y **volver a emitir** si se pierde, que es
barato y revoca el anterior. Si alguna captura ya existe, **se borra y se emite un código nuevo**.

### ⚠️ Y LAS DOS CAPTURAS PREVIAS — antes de tocar nada

**C2 y C4 comparan contra un "antes", y ese "antes" hay que tomarlo aquí.** Sin estas dos capturas, las
dos comprobaciones se hacen contra un recuerdo:

| Captura | Para qué | Por qué antes |
|---|---|---|
| **El recuento de eventos de `~/dev/test/RecetApp`** | **C2** — que **no cambie** durante la unión | Lo que se comprueba **no es una cifra absoluta**: es que **no se mueva**. Si alguien ha trabajado ahí desde que se eligió el sujeto, la cifra es otra — y sigue valiendo, mientras se tome antes |
| **El directorio de datos del agente, íntegro** | **C4** — «byte a byte igual» | «Igual **a qué**». Sin la captura, C4 no tiene contra qué comparar |

```bash
# 1 · el recuento del día (contra la plataforma, antes de unir)
# 2 · el estado íntegro del directorio de datos del agente
cp -a "$XDG_CONFIG_HOME/permea" "$PERMEA_SANDBOX/datadir-antes"
```

**Las dos se anotan y se conservan hasta C4.**

## C1 · La unión, desde el árbol correcto

```bash
cd ~/dev/test/RecetApp
SSL_CERT_FILE=/home/bfgnet/tls-local/ca.crt ~/dev/agente/bin/permea project join -   # código por stdin
```

**Se espera**: éxito por **stdout**, **nombrando el Proyecto**, y **exit 0**.

## C2 · El efecto retroactivo — **la razón de ser de la feature** (SC-006)

En el panel, la vista de gasto de la organización:

- El consumo **anterior** a la unión de esa instalación **aparece ahora bajo su Proyecto**.
- **El número total de eventos NO ha cambiado.** Es la comprobación que distingue «se agrupó en
  lectura» de «se reprocesó algo»: si el recuento se moviera, algo escribió, y FR-003 lo prohíbe.

## C3 · La repetición es indistinguible (US2 escenario 3)

Ejecutar C1 otra vez, sin cambiar nada.

- La salida es **la misma**, y el **código de salida también**.
- En el panel: la instalación sigue unida **una sola vez** al mismo Proyecto.

## C4 · Nada quedó en local

Comparar el directorio de datos del agente **contra la captura tomada en §Preparación**
(`datadir-antes`): **byte a byte igual**. La unión no dejó rastro.

> **El «estado previo» sale de §Preparación, no de aquí.** C4 no toma ninguna captura: si se intentara
> tomar ahora, ya estaría contaminada por C1–C3.

---

## Validación bloqueada

Ninguna. **La ceremonia depende de la plataforma en marcha y del banco TLS**, y las dos piezas están
disponibles. Si alguna vez no lo estuvieran, **C1–C4 quedan pendientes y se dice**, en vez de
sustituirlas por un test.

---

## Checklist de cierre

**Automatizable:**

- [ ] V1 · identidad presentada == estampada, con las tres piezas y **(c) demostrando que sabe fallar**
- [ ] V2 · cero peticiones fuera de árbol, **con caso positivo**
- [ ] V3 · los dos desenlaces indistinguibles, con las tres piezas
- [ ] V4 · rechazos idénticos entre sí y sin nombrar el Proyecto ajeno
- [ ] V5 · ninguna subcadena de ocho del código en ninguna salida
- [ ] V6 · nada escrito en local, **con caso positivo**
- [ ] V7 · sin transporte seguro no se completa, en las cuatro clases
- [ ] V8 · nunca se encola, **con caso positivo**
- [ ] V9 · regresión cero contra `baseline-sc004.tsv`, **con sus semillas**
- [ ] V-canales · reparto de stdout/stderr **capturados por separado**, y las dos vías de entrada
- [ ] `go vet` · `golangci-lint run` · `go test ./...` → **9 paquetes ok, 0 sin ficheros de test**

**Manual, contra plataforma real** *(no automatizar)*:

- [ ] C1 · la unión nombra el Proyecto
- [ ] C2 · el consumo previo cuenta bajo él **y el número de eventos no cambia**
- [ ] C3 · repetir es indistinguible, en salida y en código de salida
- [ ] C4 · el directorio de datos, byte a byte igual
