# Especificación: Distribución — release multiplataforma del binario `permea`

**Feature Branch**: `002-distribucion`
**Created**: 2026-07-03
**Status**: Draft
**Input**: Empaquetar y distribuir el binario `permea` (producido por la spec 001) como binario estático único, instalable con un solo comando en macOS, Linux y Windows, publicado de forma automática y reproducible a partir de una etiqueta de versión.

## Clarifications

- El alcance es **la distribución del binario ya existente**, no su funcionalidad. Esta feature no modifica la frontera de datos ni la ingesta; solo empaqueta, firma con checksum y publica el artefacto que la spec 001 ya compila.
- **No incluye** ningún servicio en segundo plano, autoarranque, demonio ni gestor de actualizaciones automáticas: eso es objeto de la **spec 003**. Aquí el usuario instala un binario y lo ejecuta manualmente.
- La materialización del **Principio III** de la constitución (binario único, instalable con un comando en los tres SO) es el objetivo central de esta especificación.
- El detalle técnico (herramienta de release, formato de los ficheros de workflow, plantillas de fórmula/manifiesto) corresponde al plan; la especificación describe **qué** se publica y **qué garantías** cumple.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Publicar una versión con una etiqueta (Priority: P1)

El responsable del proyecto marca un commit con una etiqueta de versión semántica. Sin pasos manuales adicionales, se construyen todos los binarios de la matriz y se publica una release con los artefactos y sus checksums.

**Acceptance Scenarios:**

1. **Given** un commit en la rama principal, **When** se empuja una etiqueta `vX.Y.Z`, **Then** se construyen los binarios de los cinco objetivos de la matriz y se publica una release con esos artefactos, un fichero de checksums y las notas de la versión.
2. **Given** una release publicada, **When** se inspecciona cualquiera de los binarios, **Then** su versión reportada coincide exactamente con la etiqueta (`X.Y.Z`), inyectada en tiempo de compilación.
3. **Given** una etiqueta que no cumple el formato semántico, **When** se empuja, **Then** el proceso de release no se dispara (o falla de forma explícita) y no se publica nada.

### User Story 2 — Instalar en macOS/Linux con un comando (Priority: P1)

Un desarrollador en macOS o Linux instala `permea` con un único comando y obtiene un binario funcional cuya integridad ha sido verificada.

**Acceptance Scenarios:**

1. **Given** una release publicada, **When** el usuario instala mediante el gestor de su plataforma (tap de Homebrew) o el script de instalación, **Then** obtiene el binario correcto para su sistema operativo y arquitectura y `permea` arranca sin dependencias previas.
2. **Given** el script de instalación descarga un artefacto, **When** el checksum del artefacto no coincide con el publicado, **Then** la instalación se aborta y no se coloca ningún binario en el sistema.

### User Story 3 — Instalar en Windows con un comando (Priority: P1)

Un desarrollador en Windows instala `permea` mediante el gestor Scoop desde el bucket del proyecto y obtiene un binario funcional.

**Acceptance Scenarios:**

1. **Given** una release publicada y el bucket del proyecto añadido, **When** el usuario ejecuta la instalación por Scoop, **Then** obtiene el binario de Windows y `permea` arranca sin dependencias previas.
2. **Given** el manifiesto de Scoop, **When** Scoop instala, **Then** verifica el checksum del artefacto contra el valor publicado antes de instalar.

### Edge Cases

- Una arquitectura o sistema operativo no soportados por la matriz: el script de instalación termina con un mensaje claro y sin instalar nada.
- Una release con un artefacto faltante o un checksum ausente: la verificación falla de forma visible; el usuario no obtiene un binario a medias.
- Re-empujar una etiqueta existente: el proceso no debe publicar dos releases contradictorias para la misma versión (idempotencia de release).
- Un binario manipulado tras la publicación: la verificación de checksum en la instalación lo detecta y aborta.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001:** El proceso de release DEBE dispararse automáticamente al empujar una etiqueta de versión semántica (`vX.Y.Z`) y NUNCA requerir pasos manuales de compilación por plataforma.
- **FR-002:** Se DEBEN producir binarios para toda la matriz de objetivos: macOS (amd64, arm64), Linux (amd64, arm64) y Windows (amd64).
- **FR-003:** Cada binario DEBE ser **estático**, compilado sin CGO, sin dependencias de runtime externas, coherente con el Principio III.
- **FR-004:** La versión del binario DEBE inyectarse desde la etiqueta en tiempo de compilación; el binario DEBE poder reportar su versión, y esta DEBE coincidir con la etiqueta.
- **FR-005:** Para cada artefacto publicado se DEBE generar y publicar un checksum **SHA256**; el conjunto de checksums acompaña a la release.
- **FR-006:** Se DEBE publicar una **fórmula de Homebrew** en un tap propio del proyecto que instale el binario correcto en macOS (y Linux via Homebrew) con un solo comando.
- **FR-007:** Se DEBE publicar un **manifiesto de Scoop** en un bucket propio del proyecto que instale el binario de Windows con un solo comando.
- **FR-008:** Se DEBE proporcionar un **script de instalación** para macOS/Linux que detecte SO y arquitectura, descargue el artefacto correcto, **verifique su checksum SHA256 antes de instalar** y aborte ante cualquier discrepancia.
- **FR-009:** Todos los canales de instalación (Homebrew, Scoop, script) DEBEN verificar la integridad del artefacto mediante su checksum publicado antes de dejar el binario en el sistema.
- **FR-010:** La release DEBE ser reproducible a partir de la etiqueta: la misma etiqueta sobre el mismo commit produce los mismos binarios (sin entradas no versionadas).
- **FR-011:** Esta feature NUNCA DEBE instalar un servicio en segundo plano, demonio, tarea programada ni mecanismo de autoarranque (reservado a la spec 003).
- **FR-012:** La distribución NUNCA DEBE alterar la frontera de datos ni la funcionalidad del binario; solo lo empaqueta y publica.

### Key Entities

- **Release**: la unidad publicada asociada a una etiqueta `vX.Y.Z`; agrupa los artefactos, sus checksums y las notas.
- **Artefacto**: un archivo comprimido por objetivo (SO+arquitectura) que contiene el binario y metadatos mínimos.
- **Fichero de checksums**: la lista SHA256 de todos los artefactos de la release; base de la verificación de integridad.
- **Canal de instalación**: tap de Homebrew, bucket de Scoop o script de instalación; cada uno instala el artefacto correcto verificando su integridad.

## Success Criteria *(mandatory)*

- **SC-001:** Empujar una etiqueta `vX.Y.Z` produce, sin intervención manual, una release con los **cinco** artefactos de la matriz y un fichero de checksums SHA256.
- **SC-002:** La versión reportada por cada binario publicado es **exactamente** `X.Y.Z`, igual a la etiqueta.
- **SC-003:** Cada binario publicado es estático y arranca en un sistema limpio de su plataforma **sin instalar dependencias previas**.
- **SC-004:** En macOS/Linux, `permea` se instala con **un único comando** (tap de Homebrew o script de instalación) y queda funcional; en Windows, con un único comando vía Scoop.
- **SC-005:** El script de instalación **aborta y no instala nada** cuando el checksum del artefacto no coincide con el publicado (detección de manipulación).
- **SC-006:** La release es reproducible desde la etiqueta: reconstruir la misma etiqueta produce binarios equivalentes.
- **SC-007:** La instalación **no deja ningún servicio en segundo plano** ni mecanismo de autoarranque en el sistema (verificable tras instalar).
