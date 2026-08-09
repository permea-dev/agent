package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/permea-dev/agent/internal/config"
)

// runStatus informa el estado de enrolamiento. Es una operación **local**: deriva el
// estado de la config persistida (endpoint + token + https) y NUNCA contacta al backend
// (a diferencia de enroll). NUNCA imprime el device_token (SC-005): a lo sumo un indicador
// de presencia.
func runStatus(stdout io.Writer) error {
	dir, err := config.DataDir()
	if err != nil {
		return err
	}
	rutaCfg := filepath.Join(dir, "config.json")

	// ═══ P-004 T028 · EL AVISO DEL MODO RETIRADO — INFORMA, NO ABORTA (D-004-5) ═══════════
	//
	// `status` es DIAGNÓSTICO: su función es explicar el estado, y una configuración que pide
	// un modo que ya no existe **es parte del estado**. Por eso lo cuenta y sigue: detener
	// aquí no protegería la frontera —`status` no procesa ni emite nada— y le quitaría al
	// usuario la herramienta con la que averiguaría qué le pasa.
	//
	// Es la MISMA función de detección que consume `setup()` para detener (T027), escrita una
	// sola vez en `internal/config`: las dos lecturas tienen que decir lo mismo sobre la misma
	// configuración, y con dos implementaciones acabarían divergiendo.
	//
	// Solo aparece ante `"plain"`. Para `"hash"`, cualquier otro valor y la clave ausente no se
	// imprime **ni una línea** (FR-013a/FR-013b): advertir de una petición ya satisfecha es
	// ruido, y el ruido enseña a ignorar los avisos.
	//
	// Va a **stderr** y no al `stdout` que recibe la función: `stdout` es la respuesta a la
	// pregunta «¿estoy enrolado?», y quien la canalice a un fichero o a otro programa no debe
	// encontrarse un aviso mezclado con el dato.
	if errRetirado := config.CheckRetiredProjectRefMode(rutaCfg); errRetirado != nil {
		_, _ = fmt.Fprintln(os.Stderr, "aviso:", errRetirado)
	}

	cfg, err := config.Load(rutaCfg)
	if err != nil {
		return err
	}

	if config.IsEnrolled(cfg) {
		// dev_id es identidad (no secreto): se muestra. El device_token NUNCA (SC-005): a lo
		// sumo un indicador de presencia.
		_, _ = fmt.Fprintf(stdout, "enrolado contra %s (dev_id: %s, token: configurado)\n", cfg.Endpoint, cfg.DevID)
	} else {
		_, _ = fmt.Fprintln(stdout, "no enrolado")
	}
	return nil
}
