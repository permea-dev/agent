# agente — medidor de coste de IA (local-first, multi-herramienta)

Agente local que lee los logs de uso de herramientas de IA, calcula coste **en local**
y transmite al backend de equipo **únicamente metadato derivado** — nunca contenido.

## Garantía de frontera
`internal/event/event.go` define un struct CERRADO: el único dato que puede salir de
la máquina. `internal/ingest` mapea explícitamente solo los campos permitidos; lo que
no se mapea, se descarta (deny-by-default). El test `internal/ingest/boundary_test.go`
lo verifica sobre un fixture con contenido sensible inyectado a propósito.

## Primeros pasos
    make test    # test de frontera en verde (empezar por aquí)
    make run     # dry-run: imprime eventos desde el fixture, sin transmitir
    make build   # binario en bin/agente

## Estructura
    cmd/agente        punto de entrada
    internal/event    LA FRONTERA (struct cerrado del evento)
    internal/ingest   lectores por herramienta (claude_code) + tests de frontera
    internal/pricing  cálculo de coste local (tabla empaquetada)
    internal/state    escaneo incremental
    internal/transport cliente HTTPS al backend
    internal/config   configuración local

Renombrar el módulo en `go.mod` (`github.com/tu-org/agente`) al repo real.
