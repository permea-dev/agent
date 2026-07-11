# Data Model — Enrolamiento de dispositivo (003)

Entidades que intervienen en el enrolamiento por el lado del agente. 003 **no** introduce
ningún campo en la frontera (`event.Event`); reutiliza el esquema de `config.Config` de P-001.

## Enrollment string (transitorio, nunca persistido tal cual)

Envoltorio de entrega que empaqueta la URL del backend y el device token. El agente lo recibe
como argumento, lo decodifica, lo usa para verificar y **descarta la forma envuelta**; solo
persiste sus dos campos por separado en `config.json`.

| Campo (decodificado) | Tipo | Reglas de validación |
|---|---|---|
| `endpoint` | string | Requerido. URL absoluta con esquema **`https`** (una `http://` se rechaza, FR-011). |
| `token` | string | Requerido. No vacío. Secreto opaco para 003; su esquema es de P-002. |

**Formato de superficie**: `pmea1.<base64url(json)>` — ver
[contracts/enrollment-string.md](./contracts/enrollment-string.md). Estructura Go de decodificación
(cerrada; punto único en `internal/config/enrollment.go`):

```go
// payload es la estructura interna del enrollment string (cerrada).
type payload struct {
    Endpoint string `json:"endpoint"`
    Token    string `json:"token"`
}
```

**Clasificación de seguridad**: secreto del mismo calibre que el `salt` (contiene el token en
claro). NUNCA se loguea, muestra ni filtra en errores (FR-007).

## Config (persistida — sin cambios de esquema)

`internal/config.Config` ya existe (P-001). El enrolamiento fija **dos campos ya presentes**;
no añade ninguno:

| Campo | Tipo | Papel en 003 |
|---|---|---|
| `endpoint` (`Endpoint`) | string | URL del backend, decodificada del enrollment string. Base de `status`. |
| `device_token` (`DeviceToken`) | string | Token verificado. Secreto; `status` NUNCA lo muestra. |

Resto de campos (`org_id`, `dev_id`, `project_ref_mode`, `tools`, `sync_interval`, `logs_root`)
quedan **intactos** al enrolar (se preservan los valores previos vía `Load` → mutar 2 campos → `Save`).

**Persistencia**: `config.json` bajo el directorio de datos por SO (`config.DataDir`), permisos
**`0600`**, reescritura atómica temp+rename (ya implementada en `config.Save`/`atomicWrite`).

## Estado de enrolamiento (derivado — no persistido como tal)

Propiedad calculada localmente, sin red:

```
enrolado ⇔ Endpoint != "" ∧ DeviceToken != "" ∧ Validate() == nil (endpoint https)
```

| Estado | Condición | Salida de `status` |
|---|---|---|
| Enrolado | ambos campos presentes y endpoint https | `enrolado` + la URL del backend (sin token) |
| No enrolado | falta token y/o endpoint | `no enrolado` (sin error, sin secreto) |

## Verification ping (lote vacío)

Petición efímera de verificación; no es una entidad persistida.

| Propiedad | Valor |
|---|---|
| Cuerpo | `[]` (JSON array vacío) — `[]event.Event{}` no-nil, cero eventos, cero metadato |
| Método/cabeceras | POST HTTPS + `Authorization: Bearer <token>` (contrato de transporte existente) |
| Interpretación | `2xx` → confirmado · `401/403` → rechazado · otro/red → no verificable |
| Efecto en backend | Ninguno más allá de confirmar auth (no aporta `event_id`, no crea eventos) — SC-009 |

## Transiciones (flujo `enroll`)

```
[sin enrolar]
   │  permea enroll <enrollment-string>
   ▼
decodificar ──(malformado / endpoint no-https)──▶ ABORTA (no persiste, no filtra el argumento)
   │ ok (endpoint https + token)
   ▼
Verify() = ping lote vacío
   ├─ 2xx ─────────▶ persistir {endpoint, token} 0600 ─▶ [ENROLADO]
   ├─ 401/403 ─────▶ RECHAZA  (no persiste; estado = sin enrolar)
   └─ 5xx/red ─────▶ NO VERIFICABLE (no persiste; estado = sin enrolar)
```

El re-enrolamiento repite el flujo; al persistir, sobrescribe `config.json` por rename atómico sin
dejar el token viejo en disco (FR-014).
