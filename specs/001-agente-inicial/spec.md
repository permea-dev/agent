# FEAT-008-F — Refinamiento del mapa `tipo_gasto_pgc` → subcuentas PGC

## ID

`FEAT-008-F`

## Name

Advanced expense type to PGC account mapping for received invoices

## Status

`approved`

---

## Context

El motor contable determinista de facturas recibidas ya está operativo y resuelve la cuenta de gasto exclusivamente mediante `PgcAccountMapper::cuentaGasto(string $tipoGasto)`. La arquitectura definida en **SPEC-005** (Contabilidad motor IA + PGC) establece que `tipo_gasto_pgc` es una **clasificación funcional previa**, y que la cuenta `6xx` final se deriva de forma **determinista** desde el mapper — sin IA, sin heurísticas.

Actualmente `material_oficina` está agrupado dentro del bloque que devuelve `629` (Otros servicios), lo que provoca asientos `629 / 472 / 410`. Contablemente corresponde a `628` (Suministros), cuenta que ya está prevista en el mapper como destino de `suministros` y `telefonia_internet`.

Esta feature es un cambio **puntual, aislado y de bajo riesgo** sobre el mapper. No toca el motor contable, la persistencia, la UI ni la base de datos. No reclasifica ninguna otra categoría.

## Objective

Mover `material_oficina` de `629` a `628` en `PgcAccountMapper::cuentaGasto()`, manteniendo intacto el resto del comportamiento del mapper y del sistema contable.

---

## User Stories

- **US-01** Como gestor, quiero que una factura clasificada como `material_oficina` genere un asiento contra la cuenta `628` (Suministros), para que la contabilidad refleje correctamente la naturaleza del gasto.
- **US-02** Como equipo técnico, quiero que la corrección se resuelva exclusivamente en el mapper, para preservar la separación entre mapeo determinista y lógica fiscal.
- **US-03** Como sistema contable, quiero que el asiento siga generándose de forma determinista a partir de `tipo_gasto_pgc`, pero con el valor `628` en lugar de `629` para material de oficina.

---

## Scope

### In scope

- Modificar `PgcAccountMapper::cuentaGasto()` para que `material_oficina` devuelva `'628'`.
- Actualizar los tests unitarios de `PgcAccountMapperTest` que cubran el nuevo mapeo y verifiquen que el resto de categorías no cambia.
- Añadir o actualizar un test de regresión en `AccountingEngineServiceTest` que verifique que una factura recibida deducible con `tipo_gasto_pgc = 'material_oficina'` produce línea de gasto en `628`.
- Preservar el soporte de códigos legacy numéricos existentes en el mapper (`600`, `602`, `621`, `622`, `623`, `624`, `625`, `626`, `627`, `628`, `629`, `631`, `640`, `642`).

### Out of scope

- Reclasificación de cualquier otra categoría de `tipo_gasto_pgc`. Explícitamente **permanecen sin cambios** en esta feat:
    - `software_suscripciones` → sigue en `629`
    - `gastos_viaje` → sigue en `629`
    - `dietas_estancia` → sigue en `629`
    - `formacion` → sigue en `629`
    - `otros_servicios` → sigue en `629`
    - `otros_gastos_explotacion` → sigue en `629`
      Cualquier refinamiento de estas categorías se abordará en una futura **FEAT-008-F2** con su propio análisis contable.
- Cambios en `AccountingEngineService`.
- Cambios en `AccountingPersistenceService`.
- Cambios en UI, catálogo de alta de facturas recibidas o frontend.
- Recálculo, regeneración o migración de asientos contables ya persistidos.
- Reclasificación masiva histórica de facturas antiguas.
- Nuevos campos, tablas o migraciones en base de datos.
- Uso de IA para decidir cuentas contables.
- Modificaciones en la cuenta de proveedor (`410` / `410.XXXX`).
- Cambios en `descripcionCuenta()`.

---

## Functional Requirements

- **FR-01** `PgcAccountMapper::cuentaGasto('material_oficina')` debe devolver exactamente la cadena `'628'`.
- **FR-02** `PgcAccountMapper::cuentaGasto('suministros')` debe seguir devolviendo `'628'` (comportamiento existente, no cambia).
- **FR-03** `PgcAccountMapper::cuentaGasto('telefonia_internet')` debe seguir devolviendo `'628'` (comportamiento existente, no cambia).
- **FR-04** `PgcAccountMapper::cuentaGasto('software_suscripciones')` debe seguir devolviendo `'629'`.
- **FR-05** `PgcAccountMapper::cuentaGasto('gastos_viaje')` debe seguir devolviendo `'629'`.
- **FR-06** `PgcAccountMapper::cuentaGasto('dietas_estancia')` debe seguir devolviendo `'629'`.
- **FR-07** `PgcAccountMapper::cuentaGasto('formacion')` debe seguir devolviendo `'629'`.
- **FR-08** `PgcAccountMapper::cuentaGasto('otros_servicios')` debe seguir devolviendo `'629'`.
- **FR-09** `PgcAccountMapper::cuentaGasto('otros_gastos_explotacion')` debe seguir devolviendo `'629'`.
- **FR-10** Para cada código legacy numérico del listado `['600','602','621','622','623','624','625','626','627','628','629','631','640','642']`, `cuentaGasto($code)` debe devolver ese mismo código sin transformación.
- **FR-11** `PgcAccountMapper::descripcionCuenta('628')` debe devolver `'Suministros'`.
- **FR-12** `PgcAccountMapper::descripcionCuenta('629')` debe devolver `'Otros servicios'`.
- **FR-13** La resolución de la cuenta de gasto en `AccountingEngineService` debe seguir delegándose íntegramente en `PgcAccountMapper::cuentaGasto()`. No debe introducirse lógica condicional adicional en el engine.
- **FR-14** El mapa resultante debe ser puramente declarativo: sin heurísticas, sin condicionales sobre importes, fechas, cliente o proveedor.

---

## Non-Functional Requirements

- **Determinismo**: misma entrada devuelve siempre la misma cuenta. Sin aleatoriedad, sin dependencias externas, sin IA.
- **Auditabilidad**: el criterio debe quedar explícito en el código del mapper, con un comentario que documente que `628` agrupa suministros y consumos recurrentes (incluido material de oficina) y que `629` queda reservado para servicios genéricos sin cuenta específica.
- **Aislamiento del cambio**: la modificación se limita a `app/Services/PgcAccountMapper.php` y a sus tests. Cero modificaciones en controllers, engine, persistencia, UI, migraciones o seeders.
- **Compatibilidad hacia atrás**: asientos ya persistidos con `629` para material de oficina no se regeneran ni se tocan. La feat no incluye lógica de migración histórica.
- **Testabilidad**: toda la lógica afectada queda cubierta por tests unitarios y un test de integración del engine.

---

## Technical Design

### Architecture overview

Cambio puntual en el mapper determinista. La cadena de llamada no cambia:

```
AccountingEngineService
  └─> PgcAccountMapper::cuentaGasto($expense->tipo_gasto_pgc)
        └─> devuelve '628' (antes '629') para 'material_oficina'
```

`AccountingPersistenceService` persiste la línea tal cual la recibe del engine.

### Main file

`app/Services/PgcAccountMapper.php`

### Current problematic block

```php
'material_oficina',
'software_suscripciones',
'gastos_viaje',
'dietas_estancia',
'formacion',
'otros_servicios',
'otros_gastos_explotacion' => '629',
```

`material_oficina` termina resolviendo a `629`.

### Proposed block

```php
'publicidad_marketing' => '627',

// 628 — Suministros y consumos recurrentes
'suministros',
'telefonia_internet',
'material_oficina' => '628',

'tributos_no_estatales' => '631',
'gastos_personal' => '640',
'seguridad_social_empresa' => '642',

// 629 — Otros servicios (residual: sin cuenta más específica)
'software_suscripciones',
'gastos_viaje',
'dietas_estancia',
'formacion',
'otros_servicios',
'otros_gastos_explotacion' => '629',
```

### New models / tables

Ninguno.

### Modified models / tables

Ninguno.

### Key services / actions

| Class                                                | Responsibility                                                    |
| ---------------------------------------------------- | ----------------------------------------------------------------- |
| `App\Services\PgcAccountMapper::cuentaGasto()`       | Único punto modificado: `material_oficina` pasa de `629` a `628`. |
| `App\Services\PgcAccountMapper::descripcionCuenta()` | Sin cambios.                                                      |
| `App\Services\AccountingEngineService`               | Sin cambios.                                                      |
| `App\Services\AccountingPersistenceService`          | Sin cambios.                                                      |

### Queue jobs

Ninguno.

---

## API

No aplica. Cambio interno en capa de servicios. Ningún endpoint HTTP se modifica.

---

## Data Model

Sin cambios de esquema, sin migraciones. El campo `client_expenses.tipo_gasto_pgc` ya existe desde FEAT-008-H.

---

## Mapping Change Summary

| `tipo_gasto_pgc`   | Antes | Después   |
| ------------------ | ----- | --------- |
| `material_oficina` | `629` | **`628`** |

Todas las demás categorías mantienen su cuenta actual.

---

## Domain Rules

- `tipo_gasto_pgc` es una clasificación funcional; la cuenta PGC final la resuelve siempre `PgcAccountMapper`.
- **El cambio se limita al mapper**. No se modifica `AccountingEngineService`, `AccountingPersistenceService` ni ningún otro servicio.
- **No se recalculan asientos históricos**. Los asientos contables ya persistidos permanecen con la cuenta con la que se generaron originalmente.
- **No se usa IA para decidir cuentas contables**. El mapeo es puramente declarativo.
- No se modifica la cuenta de proveedor (`410` / `410.XXXX`) ni su lógica de generación.
- No se modifican reglas fiscales del engine (cálculo de IVA, rama de deducibilidad, tratamiento de retenciones).
- Los importes siguen en céntimos (`_cents`); esta feat no toca importes.

---

## Acceptance Criteria

- **AC-01** Dado `tipo_gasto_pgc = 'material_oficina'`, cuando se invoca `PgcAccountMapper::cuentaGasto()`, entonces devuelve exactamente `'628'`. _(FR-01)_
- **AC-02** Dada una factura recibida deducible con `tipo_gasto_pgc = 'material_oficina'`, cuando el motor contable genera el asiento, entonces la línea de gasto usa cuenta `628`. _(FR-01, FR-13)_
- **AC-03** Dada esa misma factura, cuando se genera el asiento, entonces la línea del proveedor utiliza una cuenta cuyo código **empieza por** `'410'`, registrada en el **haber**, sin cambios respecto al comportamiento previo. _(FR-13)_
- **AC-04** Dada una factura recibida deducible estándar, cuando se genera el asiento tras el cambio, entonces la línea de IVA soportado (`472`) se mantiene intacta y el asiento cuadra (`∑debe = ∑haber`). _(FR-13)_
- **AC-05** Dados los valores `'suministros'` y `'telefonia_internet'` como `tipo_gasto_pgc`, cuando se invoca `cuentaGasto()`, entonces devuelve `'628'`. _(FR-02, FR-03)_
- **AC-06** Dados los valores `'software_suscripciones'`, `'gastos_viaje'`, `'dietas_estancia'`, `'formacion'`, `'otros_servicios'` y `'otros_gastos_explotacion'` como `tipo_gasto_pgc`, cuando se invoca `cuentaGasto()` para cada uno, entonces devuelve `'629'`. _(FR-04 a FR-09)_
- **AC-07** Dado cualquier código legacy numérico del listado `['600','602','621','622','623','624','625','626','627','628','629','631','640','642']`, cuando se invoca `cuentaGasto($code)`, entonces devuelve ese mismo código sin transformación. _(FR-10)_
- **AC-08** Dadas las llamadas `descripcionCuenta('628')` y `descripcionCuenta('629')`, cuando se invocan, entonces devuelven `'Suministros'` y `'Otros servicios'` respectivamente. _(FR-11, FR-12)_
- **AC-09** Dado un asiento contable ya persistido antes del despliegue de esta feat con cuenta `629` para material de oficina, cuando la feat se despliega, entonces ese asiento no se regenera, no se modifica y sigue existiendo tal cual. _(Out of scope confirmado)_

---

## Testing Requirements

### Unit tests — `PgcAccountMapperTest`

Cobertura obligatoria (un test o dataset por aserción):

```
cuentaGasto('material_oficina')         === '628'   // cambio principal
cuentaGasto('suministros')              === '628'   // regresión
cuentaGasto('telefonia_internet')       === '628'   // regresión
cuentaGasto('software_suscripciones')   === '629'   // regresión
cuentaGasto('gastos_viaje')             === '629'   // regresión
cuentaGasto('dietas_estancia')          === '629'   // regresión
cuentaGasto('formacion')                === '629'   // regresión
cuentaGasto('otros_servicios')          === '629'   // regresión
cuentaGasto('otros_gastos_explotacion') === '629'   // regresión

// legacy numéricos
cuentaGasto('628') === '628'
cuentaGasto('629') === '629'

// descripciones
descripcionCuenta('628') === 'Suministros'
descripcionCuenta('629') === 'Otros servicios'
```

### Integration test — `AccountingEngineServiceTest`

Caso:

- Factura recibida deducible
- `tipo_gasto_pgc = 'material_oficina'`
- Base imponible: `98000` céntimos (980,00 €)
- IVA: `20580` céntimos (205,80 €, 21%)
- Total: `118580` céntimos (1.185,80 €)

Aserciones sobre las líneas del asiento generado:

- Existe una línea con cuenta **igual a** `'628'`, en el **debe**, por `98000` céntimos.
- Existe una línea con cuenta **igual a** `'472'`, en el **debe**, por `20580` céntimos.
- Existe una línea con cuenta que **empieza por** `'410'` (comprobación por prefijo, no por valor exacto — la cuenta puede ser `410`, `410.0001`, `410.XXXX` o cualquier subcuenta por proveedor), en el **haber**, por `118580` céntimos.
- El asiento cuadra: `∑debe === ∑haber === 118580`.

La verificación del proveedor por prefijo `410` permite que el esquema de subcuentas por proveedor evolucione sin romper este test.

---

## Success Metrics

| Metric                                                      | Target                                                                   |
| ----------------------------------------------------------- | ------------------------------------------------------------------------ |
| Suite de tests Pest completa                                | En verde (`docker exec -it asesoria_app php artisan test`)               |
| Validación manual en dev con factura de material de oficina | Asiento generado con cuenta de gasto `628`, IVA `472` y proveedor `410*` |
| Regresiones en otras categorías o ramas fiscales            | 0                                                                        |

---

## Risks

| Risk                                                                                                  | Likelihood | Impact | Mitigation                                                                                                                      |
| ----------------------------------------------------------------------------------------------------- | ---------- | ------ | ------------------------------------------------------------------------------------------------------------------------------- |
| Existen tests actuales que asumen `629` para `material_oficina`                                       | Alta       | Bajo   | `grep -r "material_oficina" tests/` antes del cambio. Actualizar assertions a `628` en el mismo PR.                             |
| Existe código o documentación hardcoded con `629` asociado a material de oficina fuera del mapper     | Baja       | Medio  | `grep -r "material_oficina" app/` y `grep -rn "'629'" app/Services/` antes del PR.                                              |
| Presión para ampliar el alcance y reclasificar también `software_suscripciones`, `gastos_viaje`, etc. | Media      | Medio  | **Scope blindado**: esas categorías están explícitamente fuera de scope. Cualquier refinamiento adicional va a **FEAT-008-F2**. |
| Asientos históricos con `629` conviven con nuevos `628` para el mismo `tipo_gasto_pgc`                | Baja       | Bajo   | Comportamiento intencional y documentado. No se reclasifica histórico para preservar integridad de ejercicios cerrados.         |

---

## Dependencies

- **Spec previa**: FEAT-008-H (catálogo `tipo_gasto_pgc`) — completada.
- **Spec relacionada**: FEAT-008-I (edición inline de `tipo_gasto_pgc`) — recomendada, no bloqueante.
- **Spec marco**: SPEC-005 — Contabilidad motor IA + PGC.
- **Externa**: Ninguna.

---

## Definition of Done

- [ ] `PgcAccountMapper::cuentaGasto('material_oficina')` devuelve `'628'`.
- [ ] Todas las demás categorías de `tipo_gasto_pgc` mantienen la cuenta que devolvían antes de esta feat (verificado por tests de regresión).
- [ ] `AccountingEngineService` no ha sido modificado.
- [ ] `AccountingPersistenceService` no ha sido modificado.
- [ ] Tests unitarios de `PgcAccountMapperTest` cubren AC-01, AC-05, AC-06, AC-07 y AC-08, y pasan.
- [ ] Test de integración en `AccountingEngineServiceTest` verifica el asiento con `material_oficina` descrito arriba, comprobando la cuenta de proveedor por prefijo `410`, y pasa.
- [ ] Suite Pest completa en verde (`docker exec -it asesoria_app php artisan test`).
- [ ] Validación manual en entorno dev (agencia demo): factura recibida con `tipo_gasto_pgc = 'material_oficina'` contabilizada genera asiento con cuenta de gasto `628`.
- [ ] `grep -r "material_oficina" app/ tests/` confirma que no queda ninguna referencia hardcoded a `629` asociada a material de oficina fuera de los sitios actualizados.
- [ ] El mapper incluye un comentario que documenta la agrupación: `628` para suministros y consumos recurrentes (incluido material de oficina), `629` para servicios residuales.
- [ ] PR referencia `FEAT-008-F`.
- [ ] `changelog.md` actualizado con entrada explícita del cambio de mapeo.
- [ ] Revisado por al menos un miembro del equipo.

---

## Implementation Plan

1. **Auditar referencias**: `grep -r "material_oficina" app/ tests/` y `grep -rn "'629'" app/Services/ tests/` para localizar puntos afectados.
2. **Editar `app/Services/PgcAccountMapper.php`**: mover la clave `'material_oficina'` del bloque que resuelve a `'629'` al bloque que resuelve a `'628'`. Añadir comentario documentando la semántica de cada bloque.
3. **Actualizar `tests/Unit/Services/PgcAccountMapperTest.php`** con los casos listados en Testing Requirements (incluidos los de regresión para todas las categorías que NO cambian).
4. **Actualizar o añadir** el caso en `tests/Feature/Services/AccountingEngineServiceTest.php` para factura recibida deducible con `material_oficina` (base `98000`, IVA `20580`, total `118580`), verificando cuenta de proveedor por prefijo `410`.
5. **Ejecutar suite completa** (`docker exec -it asesoria_app php artisan test`) y resolver cualquier regresión detectada.
6. **Validación manual** en entorno dev con la agencia demo: crear factura recibida con `tipo_gasto_pgc = 'material_oficina'`, contabilizar e inspeccionar las líneas de `accounting_entry_lines`.
7. **Actualizar `changelog.md`** con entrada explícita del cambio de mapeo.
8. **Abrir PR** referenciando `FEAT-008-F`.
