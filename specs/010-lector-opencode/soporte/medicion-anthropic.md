# 010 · Medición con Anthropic *(Q-5)*

**Fecha**: 2026-10-11 · **Fuente**: la **copia congelada** de una sesión de OpenCode 2.0.26 en Windows *(proveedor `anthropic`, Claude Sonnet 5, 3 mensajes,
hacia las 09:12 UTC)*, abierta como `file:<copia>?mode=ro&immutable=1`, con su sha256 comprobado al empezar y al terminar.

**Cómo se lee esto.**
- **MEDIDO**: con `sqlite3` y las cuentas con `decimal` de Python, fuera del repo. La **sonda B** *(el lector propio de la prueba)* da sobre la copia las mismas cifras y el
  mismo sha256 de `data` que `sqlite3`.
- **NO SE PUDO MEDIR**: va con la muestra que haría falta.
- **CONSOLA**: la consola de Anthropic, vista por el dueño *(11-10, 12:02)*.
- **Tarifas de Permea** *(por millón, USD, copiadas de la plataforma)*:
  - `claude-sonnet-5`: entrada 2.00 · salida 10.00 · escritura 5 min 2.50 · escritura 1 h 4.00 · lectura 0.20;
  - `claude-haiku-4-5`: 1.00 · 5.00 · 1.25 · 2.00 · 0.10.

> **Privacidad.** Sólo recuentos, sumas, tokens, costes, horas y nombres de modelo. Los ids no se imprimieron, y sus huellas se compararon sin
> publicarse. De las tablas de cuentas y credenciales no se leyó nada.

## La copia *(MEDIDO)*

- **4 sesiones**:
  - la **1** es la del 09-10;
  - la **2** y la **3**, `google`, del 11-10 a las 08:04 y a las 08:26 UTC;
  - la **4**, `anthropic`, a las 09:12 UTC.
- **`session_message`**: `assistant` 24 · `user` 10 · `idle` 10. Las 24 `assistant` están completadas.
- **Sin `-wal`** *(modo `delete`)*. Páginas de 4 096 B.

## A-1 · Filas `anthropic` *(MEDIDO)*

**5 filas**, todas en la sesión 4, con `providerID` = **`anthropic`**, `model.id` = **`claude-sonnet-5`** y `variant` = `default`.
- `tokens` lleva las mismas rutas que con Google: `input`, `output`, `reasoning`, `cache.read` y `cache.write`. **No lleva `total`.**
- `session_v2.model` de la sesión 4: `anthropic` / `claude-sonnet-5`.

| # | `input` | `output` | `reasoning` | `cache.read` | `cache.write` | `cost` | `finish` |
|---:|---:|---:|---:|---:|---:|---:|---|
| 1 | 2 | 63 | 0 | 0 | 11 518 | 0,029429 | `tool-calls` |
| 2 | 2 | 69 | 0 | 11 518 | 142 | 0,0033526 | `stop` |
| 3 | 2 | 129 | 28 | 11 660 | 88 | 0,004126 | `tool-calls` |
| 4 | 2 | 54 | 0 | 11 748 | 177 | 0,0033361 | `stop` |
| 5 | 2 | 54 | 0 | 11 925 | 70 | 0,003104 | `stop` |
| **Σ** | **10** | **369** | **28** | **46 851** | **11 995** | **0,0433477** | |

La fila 1 **escribe** la caché *(11 518)*, y la 2 la **lee** entera *(11 518)*: `input` es sólo lo que no está en caché.

## A-2 · La fórmula de OpenCode *(MEDIDO)*

**`input × 2.00 + (output + reasoning) × 10.00 + cache.read × 0.20 + cache.write × 2.50`** *(por millón)*. Da el `cost` **exacto** en las 5 filas *(Δ = 0)*.

- **La entrada va sin la caché**, como con Google *(desc. D-3)*.
- **La escritura de caché va a 2.50, la tarifa de 5 min.** A 4.00 *(1 h)*, ninguna fila cuadra.
- **El razonamiento cuenta como salida.** Sólo la fila 3 lo tiene: con él, 0,004126 *(exacto)*; sin él, 0,003846.

| # | `cost` de OpenCode | fórmula *(escritura 5 min)* | Δ | sin razonamiento | escritura a 1 h |
|---:|---:|---:|---:|---:|---:|
| 1 | 0,029429 | 0,029429 | **0** | 0,029429 | 0,046706 |
| 2 | 0,0033526 | 0,0033526 | **0** | 0,0033526 | 0,0035656 |
| 3 | 0,004126 | 0,004126 | **0** | 0,003846 | 0,004258 |
| 4 | 0,0033361 | 0,0033361 | **0** | 0,0033361 | 0,0036016 |
| 5 | 0,003104 | 0,003104 | **0** | 0,003104 | 0,003209 |

Las **19 filas `google`** de esta copia cumplen también su identidad *(desc. D-4)* con Δ = 0.

## A-3 · P-2 con Anthropic *(MEDIDO y CONSOLA: **resuelto**)*

- **Con las partidas de P-2** y la escritura a **5 min**, la identidad de desc. D-4 se cumple en las 5 filas con **Δ = 0**. A 1 h, no *(A-4)*.
- **Sin doble conteo de caché**: `input` = 2 en cada fila, y la caché llega aparte, en `cache.read` y `cache.write`.
- **Sin doble conteo de razonamiento**: Anthropic informa el razonamiento **dentro** de su salida, y OpenCode lo **reparte** en `output` + `reasoning`. La
  consola da **397** = 369 + 28.
- **La escritura es de 5 min**: la consola la rotula «Prompt caching write (5m)».

## Consola *(CONSOLA, 11-10 09:00 UTC, `claude-sonnet-5`, clave de la medición)*

| | Consola | Σ de las 5 filas | |
|---|---:|---:|:--:|
| entrada | 10 | 10 | = |
| lectura de caché | 46 851 | 46 851 | = |
| escritura de caché | 11 995, «Prompt caching write (5m)» | 11 995 | = |
| salida | 397 | 369 + 28 de razonamiento | = |
| total | 59 253 | 59 253 | = |

**Coincide al token.** Las 800 escrituras de la llamada de fuera **no** están en `claude-sonnet-5`: son de otro modelo, que no se persigue *(D-2)*.

## A-4 · Lo que cobraría Permea *(MEDIDO)*

Con la reparación de la plataforma *(entrada × 2.00 + salida × 10.00 + `tokens_cache_creation` × **4.00** + `tokens_cache_read` × 0.20, `ROUND(·, 6)` por
evento; salida = `output + reasoning`)*:

| # | `cost` de OpenCode | Permea | Δ |
|---:|---:|---:|---:|
| 1 | 0,029429 | 0,046706 | +0,017277 |
| 2 | 0,0033526 | 0,003566 | +0,0002134 |
| 3 | 0,004126 | 0,004258 | +0,000132 |
| 4 | 0,0033361 | 0,003602 | +0,0002659 |
| 5 | 0,003104 | 0,003209 | +0,000105 |
| **Σ** | **0,0433477** | **0,061341** | **+0,0179933 (+41,51 %)** |

- **De la tarifa de escritura** salen **0,0179925**: 11 995 × (4.00 − 2.50) / 10⁶.
- **Del redondeo por evento**, 0,0000008.
- ⇒ **Toda la diferencia es la tarifa de escritura de caché.**

## A-5 · Fuera de los mensajes *(MEDIDO)*

`session_v2` de la sesión 4 menos la Σ de sus 5 filas: **entrada 4 · salida 16 · razonamiento 0 · lectura 0 · escritura 800 · coste 0,0001084**. Es el
**0,249 %** de los 0,0434561 de la sesión.

| Modelo, a la tarifa de Permea | escritura 5 min | escritura 1 h |
|---|---:|---:|
| `claude-sonnet-5` | 0,002168 | 0,003368 |
| `claude-haiku-4-5` | 0,001084 | 0,001684 |

- **Ninguno encaja.**
  - Sonnet 5 queda **descartado** con cualquier precio de escritura: entrada y salida solas ya dan 0,000168, por encima de 0,0001084.
  - Haiku 4.5 necesitaría la escritura a 0,0305 por millón, que no es ninguna de sus tarifas.
- **Observación sin conclusión**: 0,0001084 es **exactamente la décima parte** de Haiku 4.5 con escritura a 5 min *(0.10 / 0.50 / 0.125)*. Con una
  ecuación y tres precios desconocidos, una coincidencia no identifica el modelo.
- **NO SE PUDO MEDIR** el modelo: no queda escrito en la copia *(desc. D-8)*.
- **Las sesiones 2 y 3 *(`google`)*** dejan **545 / 8** y **544 / 5**, con coste 0,0001835 y 0,0001757. Encajan **exactamente** con `gemini-3.5-flash-lite`
  *(0,30 / 2,50)*, como la del 09-10 *(desc. D-8)*.

## A-6 · La barra *(MEDIDO)*

- **«$0.04»**: la Σ de las 5 filas da 0,0433477 y la sesión, 0,0434561. Las dos se muestran como $0.04, así que la barra no distingue entre ellas.
- **«12.1K»** = **`input + output + cache.read + cache.write` de la última llamada**: 2 + 54 + 11 925 + 70 = **12 051**.
  - Es la misma regla que los «8.3K» del 09-10 *(8 244 + 43; aquella llamada no tenía caché; desc. §Contraste con la interfaz)*.
  - **Sin la caché**, serían 56.
- **«1 %»**: la ventana del modelo. **NO SE PUDO MEDIR**.

## A-7 · Lo antiguo no cambió *(MEDIDO)*

- **19 filas `google`**: 8 de la sesión 1 y 6 + 5 de las sesiones 2 y 3.
- **Las 8 de la sesión 1 son idénticas** a las de la copia del 09-10:
  - mismas sumas: **59 321 / 230 / 462 / 3 999 / 0 / 0,047385675**;
  - la **misma huella de todas sus columnas**, en las 14 filas de la sesión;
  - la misma huella de su fila de `session_v2`, con totales **59 859 / 235 / 462 / 3 999 / 0 / 0,047559575**.
- Las huellas no se publican.

## Consecuencias para P-2 y para la plataforma *(sin decidir)*

1. **P-2 cuadra con Anthropic**, sin regla por proveedor: Δ = 0 y la consola, al token *(A-3)*.
2. **La reparación de la plataforma cobra un 41,5 % más** de lo que factura Anthropic, porque lleva toda escritura a 1 h *(P-031 D-2)* y la de OpenCode
   es de **5 min** *(consola)*. El evento sólo tiene `tokens_cache_creation`, sin separar 5 min y 1 h, y su frontera es cerrada *(FR-030)*: la corrección
   es de la plataforma.
3. **La llamada del título** *(D-2)*: con Anthropic, el **0,25 %** del coste de la sesión, y de un modelo **sin identificar**. Con Google, el título lo hace
   `gemini-3.5-flash-lite` en las tres sesiones medidas.
4. **El lector**: la sonda B lee esta copia como `sqlite3` *(cifras y sha256 de `data`)*. Nada de esta medición cambia D-3 ni P-3.
