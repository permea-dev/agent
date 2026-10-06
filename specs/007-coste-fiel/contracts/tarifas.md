# Contrato — La tabla de tarifas, espejo de cinco cifras del catálogo de la plataforma (007)

**Feature**: `007-coste-fiel` | **Fecha**: 2026-10-06 | **Requisitos**: 007 FR-002, FR-003, FR-006 a FR-008 · D-3 · P-1 ✅

**Sustituye a** `specs/006-medicion-fiel/contracts/tarifas.md` *(espejo de cuatro cifras de `e50d0a5`)*. Conserva su semántica y la de 001:
`Cost(model, …) → (coste, disponible)`, con `disponible = false` si el modelo no está en la tabla. Lo nuevo es **la quinta cifra**
*(escritura de caché a 1 hora)*, **la fila 17** y **la regla de lo que no trae desglose**.

---

## La regla

**La tabla empaquetada es una copia exacta del catálogo de la plataforma en un commit concreto**:
- las mismas claves, ni una más ni una menos;
- las mismas **cinco** cifras por clave, en USD por millón de tokens: entrada, salida, escritura de caché a **5 minutos**
  (`cache_write`), escritura de caché a **1 hora** (`cache_write_1h`) y lectura de caché.

**Referencia vigente**: `permea-dev/permea-platform` · `backend/config/pricing.php` · **`8f147d1`**
*(«P-031: la escritura de cache de 1 hora, a su precio (#82)», 2026-10-06 11:12 +0200)*.

## La tabla literal *(leída de `pricing.php@8f147d1` con `git show` el 2026-10-06; 17 filas, `currency` = `USD` en todas)*

| Clave | `input` | `output` | `cache_write` *(5 min)* | `cache_write_1h` | `cache_read` |
|---|---:|---:|---:|---:|---:|
| `claude-fable-5` | 10.00 | 50.00 | 12.50 | 20.00 | 1.00 |
| `claude-fable-5-1` | 10.00 | 50.00 | 12.50 | 20.00 | **0.25** |
| `claude-mythos-5` | 10.00 | 50.00 | 12.50 | 20.00 | 1.00 |
| `claude-opus-5-5` | 4.00 | 20.00 | 5.00 | 8.00 | 0.20 |
| `claude-opus-5` | 5.00 | 25.00 | 6.25 | 10.00 | 0.50 |
| `claude-opus-4-8` | 5.00 | 25.00 | 6.25 | 10.00 | 0.50 |
| `claude-opus-4-7` | 5.00 | 25.00 | 6.25 | 10.00 | 0.50 |
| `claude-opus-4-6` | 5.00 | 25.00 | 6.25 | 10.00 | 0.50 |
| `claude-opus-4-5` | 5.00 | 25.00 | 6.25 | 10.00 | 0.50 |
| `claude-opus-4-1` | 15.00 | 75.00 | 18.75 | 30.00 | 1.50 |
| `claude-opus-4` | 15.00 | 75.00 | 18.75 | 30.00 | 1.50 |
| `claude-sonnet-5` | 2.00 | 10.00 | 2.50 | 4.00 | 0.20 |
| `claude-sonnet-4-6` | 3.00 | 15.00 | 3.75 | 6.00 | 0.30 |
| `claude-sonnet-4-5` | 3.00 | 15.00 | 3.75 | 6.00 | 0.30 |
| `claude-sonnet-4` | 3.00 | 15.00 | 3.75 | 6.00 | 0.30 |
| `claude-haiku-4-5` | 1.00 | 5.00 | 1.25 | 2.00 | 0.10 |
| `claude-haiku-3-5` | 0.80 | 4.00 | 1.00 | 1.60 | 0.08 |

**Comprobaciones hechas sobre la tabla**:
- `cache_write_1h` = **2 × `input`** en las 17;
- `cache_write` = **1,25 × `input`** en las 17;
- las 16 claves de `e50d0a5` conservan sus cuatro cifras. La fila nueva es `claude-fable-5-1`, cuya lectura de caché es 0,025 × la
  entrada: la cabecera de la plataforma dice que **no es una errata**.

## El cálculo *(FR-002 a FR-005)*

```
coste = in × input + out × output + 5m × cache_write + 1h × cache_write_1h + cr × cache_read      (por millón)
```

| Caso de la línea | `5m` | `1h` | `tokens_cache_creation` del evento |
|---|---|---|---|
| Trae las dos claves y `5m + 1h = cache_creation_input_tokens` | el del log | el del log | el total del log |
| **No trae el desglose** *(P-1 ✅)* | 0 | **el total** | el total del log |
| **El desglose no suma el total** *(Q-4 ✅ (a))* | 0 | **el total** | el total del log |

El desglose **nunca** cruza la frontera *(D-1, FR-005)*: el evento lleva el total y el coste ya calculado.

## La cabecera *(FR-007)*

La tabla **DEBE** llevar, en el comentario que la precede:

| Elemento | Valor |
|---|---|
| Fuente | `https://platform.claude.com/docs/en/about-claude/pricing` |
| Verificación | **2026-10-04**: las 17 filas, con las cinco cifras, contra la fuente *(la de `pricing.php@8f147d1`)* |
| Aprobación | el dueño, 2026-08-07 *(catálogo)*; la quinta cifra y la fila `claude-fable-5-1`, 2026-10-04 *(P-031)* |
| Catálogo replicado | `permea-dev/permea-platform` · `backend/config/pricing.php` · `8f147d1` |
| Casamiento | exacto, sin normalizar sufijos, prefijos ni mayúsculas *(006 FR-018)* |
| **Hipótesis P-1** *(sustituye a la «Limitación 1»)* | una línea **sin desglose**, o con un desglose que no suma, tarifa **toda** su escritura de caché a 1 hora. Es la misma hipótesis que declara la cabecera del catálogo |
| Limitación 2 *(sigue)* | el **«modo rápido»** no se distingue; un evento en modo rápido queda infravalorado |

⚠️ **La hipótesis P-1 sólo es verdad cuando `Cost` tarifa por duración.** La cabecera cambia el commit, la verificación y la aprobación
en **B1**, junto con la tabla, pero sustituye la «Limitación 1» en **B2**, junto con `Cost` *(plan, B1 y B2)*.

## La vigilancia *(FR-008)*

El test compara contra una **tabla esperada escrita aparte**, en el propio test, con su comentario de procedencia *(el mismo commit)*.
Tres aserciones **independientes** (`t.Errorf`):
1. el **número de claves** es el esperado: **17**;
2. cada clave esperada existe y sus **cinco** cifras son exactamente las esperadas, con un subtest por clave;
3. **ninguna clave sobra**.

**Mutación que lo valida** *(SC-010)*: poner a cero `CacheWrite1h` de `claude-fable-5-1`. Debe caer sólo el subtest
`claude-fable-5-1` de la aserción 2.

El test **no lee** el repositorio de la plataforma. La comparación entre repositorios es una validación manual del quickstart *(Q-T)*.

## Cómo se absorbe un cambio del catálogo

Un cambio de fila en la plataforma es, en el agente, **un cambio en cinco sitios, todos en el mismo commit**:

1. la fila en `Table` (`internal/pricing/pricing.go`);
2. la fila en la tabla esperada del test (`internal/pricing/pricing_test.go`);
3. el commit replicado en la cabecera *(y la verificación y la aprobación, si cambian)*;
4. la hipótesis o la limitación afectada, si la hay;
5. este contrato *(su tabla literal y su referencia vigente)* y D-3 de la spec, por enmienda fechada.

**Si se hace el 1 y no el 2, el test se pone rojo, y ése es su trabajo.** Ningún otro componente del agente depende de una cifra concreta.

## Lo que este contrato NO hace

- **No recalcula** el coste de eventos ya emitidos: el coste se fija al ingerir *(P-4 ✅)*.
- **No normaliza** identificadores de modelo.
- **No distingue** el modo rápido *(N-3)*.
- **No toca la plataforma.** Su cabecera *(`pricing.php@8f147d1:32-35`)* dice que el agente «0.3.0» calcula la caché a 5 minutos. Tras
  publicar la 0.4.0 necesita una enmienda, que es un encargo aparte de la plataforma *(N-6)*.
