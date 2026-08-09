# Contrato · Configuración y parada por modo retirado

**Feature**: P-004 · **Fecha**: 2026-08-09 · **Estado**: propuesto

Fija el comportamiento observable del agente ante una configuración que menciona el ajuste retirado.
**Cumple** `specs/003-enrolamiento/contracts/cli.md`; no lo redefine.

---

## Superficie de configuración: antes y después

`config.json` — el fichero legible por el usuario en el directorio de datos por SO.

| Clave | Antes | Después |
|---|---|---|
| `endpoint` | ✅ | ✅ sin cambio |
| `device_token` | ✅ | ✅ sin cambio |
| `org_id` | ✅ | ✅ sin cambio |
| `dev_id` | ✅ | ✅ sin cambio |
| **`project_ref_mode`** | ✅ declarada, **inerte** | ❌ **retirada** de la superficie con significado (FR-015) |
| `tools` | ✅ | ✅ sin cambio |
| `sync_interval` | ✅ | ✅ sin cambio |
| `logs_root` | ✅ | ✅ sin cambio |

**Tras la retirada, la clave desaparece del fichero en la siguiente escritura** (`config.Save`), sin
que nadie la borre a mano: al no estar en el struct, no se serializa.

---

## Comportamiento ante la clave obsoleta

| Valor en `config.json` | Comportamiento | Requisito |
|---|---|---|
| `"project_ref_mode": "plain"` | **PARADA** con error visible; exit ≠ 0; **cero eventos** procesados o emitidos | FR-013 |
| `"project_ref_mode": "hash"` | **Silencio total**: sin error, sin aviso, funcionamiento normal | FR-013a |
| `"project_ref_mode": <cualquier otro>` | **Silencio total** | FR-013b |
| clave ausente | **Silencio total** | — |

El criterio, escrito para que no se pierda: **la parada se dispara por lo que el usuario pedía**, no
por la clave que escribió. `plain` pedía algo que el producto ya no promete; `hash` pedía exactamente
lo que el agente hace siempre.

---

## Alcance de la parada — por subcomando

Fijado por **D-004-5** ([plan.md](../plan.md) §Decisiones de plan): «arranque», en FR-013/SC-007, son
los caminos que procesan **hacia emisión**. La garantía protege la frontera, no el binario.

| Invocación | Carga config | Procesa/emite | Se detiene ante `plain` |
|---|---|---|---|
| `permea --run` | `setup()` | **sí** | **SÍ** |
| `permea --daemon` | `setup()` | **sí** | **SÍ** |
| `permea --scan <fichero>` | **no** | **procesa, no emite** | **NO** — excepción razonada de **D-004-5**: procesamiento diagnóstico con salt fijo, sin cola y sin transporte |
| `permea status` | sí | no | **NO** — D-004-5: diagnóstico; informa solo ante `"plain"` |
| `permea enroll …` | sí | no | **NO** — D-004-5: vía de reparación (`Load`+`Save` limpia la clave) |
| `permea --version` | no | no | no |

**Sobre `--scan`, que es la fila que más se presta a error**: sí **procesa** líneas, y la garantía de
FR-013 nombra «procesados **o** emitidos». No se detiene porque su procesamiento **no puede alcanzar
la frontera**, y no por una barrera sino por tres a la vez: usa un salt de dry-run —las identidades
que imprime no son las del usuario—, no escribe en `queue.jsonl` y no abre ninguna conexión. Que
además **no lea `config.json`** es una cuarta razón, no la principal: si mañana la leyera, la
respuesta seguiría siendo la misma.

**Por qué `status` y `enroll` tampoco**: los tres caminos que procesan **hacia emisión** pasan todos
por `setup()`, así que cortar ahí cumple la garantía entera. Detener además el diagnóstico y la
reparación no añade seguridad y deja al usuario encerrado. `enroll` reescribe `config.json` y por
tanto **limpia la clave obsoleta** por sí solo.

---

## Forma del error de parada

En `stderr`, con el prefijo `error:` que ya usa el resto del arranque (`cmd/permea/main.go:46`):

```
$ permea --run
Permea 0.1.0
error: el modo `project_ref_mode: "plain"` fue retirado y ya no existe.
       La identidad de proyecto cruza siempre de forma irreversible.
       Elimina la clave "project_ref_mode" de <dataDir>/config.json para continuar.
$ echo $?
1
```

Requisitos de forma:

- **Nombra clave y valor exactos** — el usuario no debe adivinar cuál de sus ajustes es.
- **Da la ruta real** del fichero, resuelta por SO, no «tu config.json».
- **Dice qué hace el producto ahora**: quien puso `plain` creía tener otra cosa; el mensaje corrige la
  creencia, no solo el fichero.
- **Exit 1** — el mismo del resto de errores de arranque, sin código especial.
- **NUNCA** vuelca la configuración completa ni ningún secreto (P-003 FR-007).

---

## Comportamiento de `status` ante la clave obsoleta

`status` **informa y termina con éxito**:

```
$ permea status
enrolado: sí (https://…)
aviso: la clave "project_ref_mode" de config.json fue retirada; su valor "plain" ya no
       tiene efecto y detendrá `permea --run`. Elimínala de <dataDir>/config.json.
$ echo $?
0
```

Este aviso **no contradice FR-013a**: solo aparece cuando el valor es `plain` —el caso que sí es un
problema—, nunca para `hash` ni para otros valores.

---

## Verificación

| Criterio | Cómo se comprueba |
|---|---|
| SC-007 (parada) | Test de proceso `os/exec` + `ExitCode() != 0`; `queue.jsonl` sin líneas nuevas |
| SC-007 (limpio) | `ExitCode() == 0` **y** stderr sin aviso, para `hash` y para clave ausente |
| FR-015 | La clave no aparece en `config.json` tras un `Save` |
| SC-008 | Barrido documental con **cuatro** términos: `plain`, `opt-in`, `project_ref_mode`, `en claro` |

Los tests de proceso comparan **código de salida**, no texto, por el puente Windows/WSL.
