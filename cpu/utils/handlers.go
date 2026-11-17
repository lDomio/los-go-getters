package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"net/http"
	"sync"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var mutexHandler sync.Mutex

func EjecutarProceso(w http.ResponseWriter, r *http.Request) {
	var request structs.ProcesoEnEjecucion
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutexHandler.Lock()

	slog.Debug(fmt.Sprintf("me llego este PC: %d y este PID: %d", request.PC, request.PID))

	globals.PCBenEjecucion.PID = request.PID
	globals.PCBenEjecucion.PC = request.PC

	slog.Debug(fmt.Sprintf("Se guardo el PCB recibido en el dato de CPU"))

	CicloDeInstrucciones()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensaje":"Proceso ejecutado correctamente"}`))

	mutexHandler.Unlock()
}

func InterrumpirProceso(w http.ResponseWriter, r *http.Request) {
	logger.InterrupcionRecibida()

	// Responder al kernel que la interrupción fue recibida
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensaje":"Interrupción recibida correctamente"}`))
}
