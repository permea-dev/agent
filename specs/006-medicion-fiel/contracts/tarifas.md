# Contrato — La tabla de tarifas como espejo del catálogo de la plataforma (P-006)

*Sustituido el 2026-10-06 por `specs/007-coste-fiel/contracts/tarifas.md` (cinco cifras, `8f147d1`).*

**Feature**: `006-medicion-fiel` | **Fecha**: 2026-10-02 | **Requisitos**: P-006 FR-014 a FR-020 · D-006-4 · Q-006-1 (resuelta el 2026-10-02)

**Relación con contratos existentes.** Cumple la semántica de 001: `Cost(model, …) → (coste,
disponible)`, y `disponible = false` si el modelo no está en la tabla
(`specs/001-agente-inicial/data-model.md:114-120`). Lo nuevo es **de dónde salen las cifras** y **cómo
se vigilan**.

---

## La regla

**La tabla empaquetada es una copia exacta del catálogo de la plataforma en un commit concreto**:
- las mismas claves, ni una más ni una menos;
- las mismas cuatro cifras por clave (entrada, salida, escritura de caché, lectura de caché), en USD
  por millón de tokens.

**Referencia vigente**: `permea-dev/permea-platform` · `backend/config/pricing.php` · **`e50d0a5`** (16 claves;
spec, M4). *(Enmendado 2026-10-02, Q-006-1 resuelta: antes `865bba0`. La plataforma corrigió
`claude-sonnet-5` a 2.00 / 10.00 / 2.50 / 0.20 en `e50d0a5`, PR #78.)*

## La cabecera (FR-015, FR-019)

La tabla **DEBE** llevar, en el comentario que la precede:

| Elemento | Valor en `e50d0a5` |
|---|---|
| Fuente | `https://platform.claude.com/docs/en/about-claude/pricing` |
| Verificación | la fecha en que se comprobó la fuente |
| Aprobación | quién y cuándo |
| Catálogo replicado | repositorio · fichero · commit |
| Casamiento | exacto, sin normalizar sufijos ni prefijos (FR-018) |
| Limitación 1 | escritura de caché a la tarifa de **5 minutos**; la de 1 hora quedaría infravalorada |
| Limitación 2 | **«modo rápido»** no distinguido; quedaría infravalorado |
| ~~Limitación 3~~ | ~~`claude-sonnet-5` a precio estándar (depende de Q-006-1)~~ *(Enmendado 2026-10-02, Q-006-1 resuelta: retirada; `e50d0a5` ya no la declara. **Quedan dos.**)* |

## La vigilancia (FR-020)

El test de la tabla compara contra una **tabla esperada escrita aparte**, en el propio test, con su
comentario de procedencia (el mismo commit). Tres aserciones **independientes** (`t.Errorf`):
1. el **número de claves** es el esperado (16 en `e50d0a5`) *(Enmendado 2026-10-02, Q-006-1 resuelta: también eran 16 en `865bba0`.)*;
2. cada clave esperada existe y sus **cuatro** cifras son exactamente las esperadas;
3. **ninguna clave sobra**.

El test **no lee** el repositorio de la plataforma. La comparación entre repositorios es una
validación manual del quickstart.

## Cómo se absorbe un cambio del catálogo (Q-006-1 y cualquier otro)

*(Enmendado 2026-10-02, Q-006-1 resuelta: Q-006-1 ya se absorbió así, antes de B3. La spec se
enmendó y B3 replica `e50d0a5` desde el principio, así que los cinco sitios nacen ya con la fila nueva.)*

Un cambio de fila en la plataforma es, en el agente, **un cambio en cinco sitios, todos en el mismo
commit**:

1. la fila en la tabla;
2. la fila en la tabla esperada del test;
3. el commit replicado en la cabecera;
4. la limitación afectada, si la hay;
5. M4 y las referencias de la spec (enmienda fechada).

**Si se hacen 1 y no 2, el test se pone rojo, y ése es su trabajo.** Ningún otro componente del agente
depende de una cifra concreta.

## Lo que este contrato NO hace

- **No recalcula** el coste de eventos ya emitidos: el coste se fija al ingerir.
- **No normaliza** identificadores de modelo: casamiento exacto, como el catálogo, cuya clave es el
  identificador «tal como llega en el evento» (`pricing.php`, cabecera de `models`).
- **No distingue** la caché de 1 hora ni el modo rápido: el evento no trae esa información (D-006-3).
