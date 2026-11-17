package utils

import (
	"github.com/sisoputnfrba/tp-golang/memoria/configuracionmemoria"
	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

func ObtenerContenidoEnMarcos(pid int) ([][]byte, error) {

	globals.ProcesosMutex.RLock()
	marcosAsginados := globals.MarcosXProceso[pid]
	globals.ProcesosMutex.RUnlock()
	
	globals.MemoriaMutex.RLock()
	contenidoEnMarcos := make([][]byte, len(marcosAsginados))
	for i, marco := range marcosAsginados {
		tamanioPagina := configuracionmemoria.MemoriaConfig.PageSize
		contenidoFrame := make([]byte, tamanioPagina)

		// recuperamos el contenido del frame con copy - usar el marco real, no el índice
		copy(contenidoFrame, globals.MemoriaPrincipal[marco*tamanioPagina:(marco+1)*tamanioPagina])
		// asignamos el contenido a un array de arrays
		contenidoEnMarcos[i] = contenidoFrame
	}
	globals.MemoriaMutex.RUnlock()

	// Actualizar métricas fuera de la región crítica para evitar deadlock
	for _, marco := range marcosAsginados {
		ActualizarMetricas(pid, marco, "lectura")
	}

	return contenidoEnMarcos, nil
}
