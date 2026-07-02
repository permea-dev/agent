# Contrato de frontera — Event

Materialización del **Principio I** de la constitución y del *Contrato de frontera de datos*
del [spec.md](../spec.md). Este es el **único** contrato de datos que cruza hacia el backend.
Punto de auditoría único: `internal/event/event.go`.

## Allowlist — lo que cruza (esquema del evento)

El evento se serializa como JSON. Esquema (informativo, JSON-Schema-like):

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "schema_version", "agent_version", "event_id", "occurred_at",
    "tool", "model",
    "tokens_input", "tokens_output", "tokens_cache_creation", "tokens_cache_read",
    "cost_usd", "cost_available",
    "project_ref", "session_ref", "machine_ref", "dev_id", "org_id"
  ],
  "properties": {
    "schema_version":        { "type": "integer", "const": 1 },
    "agent_version":         { "type": "string" },
    "event_id":              { "type": "string", "description": "hex; clave de deduplicación" },
    "occurred_at":           { "type": "string", "format": "date-time" },
    "tool":                  { "type": "string", "enum": ["claude_code"] },
    "model":                 { "type": "string" },
    "tokens_input":          { "type": "integer", "minimum": 0 },
    "tokens_output":         { "type": "integer", "minimum": 0 },
    "tokens_cache_creation": { "type": "integer", "minimum": 0 },
    "tokens_cache_read":     { "type": "integer", "minimum": 0 },
    "cost_usd":              { "type": "number", "description": "válido solo si cost_available" },
    "cost_available":        { "type": "boolean", "description": "false si el modelo no está en la tabla local" },
    "project_ref":           { "type": "string", "description": "hash salado; ruta en claro solo con opt-in plain" },
    "session_ref":           { "type": "string", "description": "hash salado" },
    "machine_ref":           { "type": "string", "description": "hash salado" },
    "dev_id":                { "type": "string" },
    "org_id":                { "type": "string" }
  }
}
```

`additionalProperties: false` es la regla clave: **ningún** campo fuera de esta lista puede
aparecer. En Go esto se garantiza por construcción (struct cerrado), no por validación.

### Ejemplo de evento válido

```json
{
  "schema_version": 1,
  "agent_version": "0.1.0",
  "event_id": "a3f1c9d2e4b60718293a4b5c6d7e8f90",
  "occurred_at": "2026-06-20T10:15:30Z",
  "tool": "claude_code",
  "model": "claude-opus-4-6",
  "tokens_input": 1200,
  "tokens_output": 800,
  "tokens_cache_creation": 300,
  "tokens_cache_read": 5000,
  "cost_usd": 0.0765,
  "cost_available": true,
  "project_ref": "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
  "session_ref": "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
  "machine_ref": "d4735e3a265e16eee03f59718b9b5d03019c07d8b6c51f90da3a666eec13ab35",
  "dev_id": "dev-42",
  "org_id": "org-1"
}
```

### Modelo desconocido (coste no disponible)

```json
{ "...": "...", "model": "modelo-futuro-x", "tokens_input": 500,
  "cost_usd": 0, "cost_available": false }
```
`cost_available: false` distingue "no sabemos el coste" de "coste 0". Los tokens sí se
contabilizan (edge case del spec).

## Denylist — lo que NUNCA cruza

| Categoría | Ejemplos | Cómo se impide |
|---|---|---|
| Texto de mensajes | prompts, respuestas | `rawRecord` no decodifica `message.content`. |
| Código | ficheros, diffs, fragmentos | Sin campo en el struct. |
| Identificadores en claro | rutas, nombres de proyecto/rama | Solo `*_ref` hasheados (salvo opt-in `plain`). |
| Llamadas a herramientas | argumentos, resultados | Sin campo en el struct. |
| Secretos | env vars, claves de API | Sin campo en el struct. |
| Cualquier dato no listado en la allowlist | — | deny-by-default: no existe el campo. |

## Invariantes verificables (golden test)

`internal/ingest/boundary_test.go` inyecta contenido sensible en el fixture y falla si
cualquier término de la denylist aparece en el JSON del evento. **Cualquier ampliación de
este contrato con contenido está prohibida** (Principio I, no negociable). Ampliaciones con
metadato derivado (p. ej. `cost_available`) son admisibles y se documentan aquí con su
justificación en [research.md](../research.md#r5).
