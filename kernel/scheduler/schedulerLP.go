package scheduler

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"log/slog"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/request"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// ------------------------------------------------------------------------------
// Obtiene un proceso con prioridad para SUSP-READY
func obtenerProcesoAPlanificar(ordenarCola func()) *structs.PCB {

	globals.Wait(globals.SemNew)

	if !globals.ColaSuspendedReady.IsEmpty() {
		ordenarCola()
		return globals.ColaSuspendedReady.Dequeue()
	}

	ordenarCola()
	return globals.ColaNew.Dequeue() // Obtiene el primer proceso de la cola de New
}

// Reinserta un proceso en su cola original
func reinsertarProceso(proceso *structs.PCB) {
	if proceso.Estado == globals.SUSPENDED_READY {
		globals.ColaSuspendedReady.InsertFirst(proceso)
		globals.Signal(globals.SemNew)
		slog.Info(fmt.Sprintf("No hay memoria suficiente para el proceso %d (SUSP-READY)", proceso.PID))
	} else {
		globals.ColaNew.InsertFirst(proceso)
		globals.Signal(globals.SemNew)
		slog.Info(fmt.Sprintf("No hay memoria suficiente para el proceso %d (NEW)", proceso.PID))
	}
}

func planificarLargoPlazo(ordenarCola func()) {
	for {
		proceso := obtenerProcesoAPlanificar(ordenarCola)

		if request.ConsultarEspacioSuficiente(proceso.Tamanio) {

			if proceso.Estado == globals.SUSPENDED_READY {
				// Si el proceso estaba en Suspended Ready, se cambia su estado a Ready
				request.SwapIn(proceso.PID)

				slog.Info("Proceso ", strconv.Itoa(proceso.PID), "des-swapeado")
			} else {
				request.EscribirProcesoEnMemoria(proceso.PID, proceso.Instrucciones, proceso.Tamanio)
			}

			// Se agrega el proceso a la cola de Ready
			proceso.CambiarEstado(globals.READY)
			globals.ColaReady.Enqueue(proceso)
			globals.Signal(globals.SemReady) // Se libera el canal para que pueda haber más procesos en New

			slog.Debug("Proceso agregado a la cola de Ready")
		} else {
			reinsertarProceso(proceso)
			// Esperar hasta que un proceso finalice
			//<-globals.ProcesoFinalizado
			globals.Wait(globals.SemReintentarConsultaMemoria) // Espera a que se notifique que hay memoria suficiente
		}
	}
}

// TODO iniciarlo en algun lado
// FinalizarProceso es la función que se encarga de finalizar un proceso cuando haya un proceso en la cola de Exit.
func PlanificarProcesosFinalizados() {
	for {
		// Espera a que haya un proceso en Exit y lo elimina del sistema
		globals.Wait(globals.SemExit)     // Espera a que haya un proceso en Exit
		pcb := globals.ColaExit.Dequeue() // Obtiene el primer proceso de la cola de Exit
		//pcb.CambiarEstado(globals.EXIT)   // Cambia el estado del proceso a EXIT

		// el garbage collector se encarga de liberar la memoria del proceso

		request.LiberarMemoria(pcb.PID) // Llama a la función para finalizar el proceso

		// Log obligatorio: (7/8)
		logger.FinDeProceso(pcb.PID) // Log obligatorio: Fin de proceso

		// log obligatorio (8/8)
		logger.MetricasDeEstado(pcb.PID, pcb.ME, pcb.MT) // Registra las métricas del estado del proceso

		globals.Signal(globals.SemReintentarConsultaMemoria)
	}

}

func inicioPlanificadorLP() bool {
	reader := bufio.NewReader(os.Stdin)
	log.Println("Presione ENTER para iniciar la planificación a largo plazo")

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error al leer la entrada:", err)
			continue
		}

		text = strings.TrimSpace(text)
		if text == "" {
			slog.Debug("Planificación a largo plazo iniciada")
			return true
		} else {
			slog.Debug("Entrada incorrecta. Presione ENTER para iniciar la planificación a largo plazo")
		}
	}
}
