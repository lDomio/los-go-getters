// Package utils proporciona funciones auxiliares para el manejo de solicitudes y respuestas HTTP.
package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func EnviarInstruccion(w http.ResponseWriter, r *http.Request) {
	var request structs.ProcesoEnEjecucion
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Se envian las instrucciones")

	respuesta, err := json.Marshal(obtenerInstruccion(request.PID, request.PC))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func obtenerInstruccion(PID int, PC int) structs.Instruccion {

	//log.Printf(" \nInstruccion a devolver: %s", globals.ProcesosEnMemoriaDeSistema[PID].Instrucciones[PC])

	globals.ProcesosMutex.RLock()
	defer globals.ProcesosMutex.RUnlock()

	logger.ObtenerInstruccion(PID, PC, globals.ProcesosEnMemoriaDeSistema[PID].Instrucciones[PC].Operacion, globals.ProcesosEnMemoriaDeSistema[PID].Instrucciones[PC].Parametros)
	response := globals.ProcesosEnMemoriaDeSistema[PID].Instrucciones[PC] // Retorna la instrucción correspondiente al PID y PC
	

	return response
}
