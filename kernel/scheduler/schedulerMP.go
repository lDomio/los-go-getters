package scheduler

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/request"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func manejarTimerSuspension(pcb *structs.PCB) {
	tiempoSuspension := globals.Config.Suspension_time // en milisegundos
	timer := time.NewTimer(time.Duration(tiempoSuspension) * time.Millisecond)

	select {
	case <-timer.C:
		timer.Stop()
		// Si el timer llegó a su fin y el proceso sigue en BLOCKED, se suspende
		if pcb.Estado == globals.BLOCKED {
			pcb = globals.ColaBlocked.GetAndRemoveByPID(pcb.PID) // Remover el PCB de la cola de BLOCKED
			if pcb != nil {
				pcb.CambiarEstado(globals.SUSPENDED_BLOCKED)
				globals.ColaSuspendedBlocked.Enqueue(pcb)
				globals.Signal(globals.SemSuspBlocked) // Notificar que hay un proceso en SUSPENDED_BLOCKED

				slog.Info(fmt.Sprintf("Proceso %d en proceso de swap", pcb.PID))

				request.SwapOut(pcb.PID)

				slog.Info(fmt.Sprintf("Proceso %d swapeado", pcb.PID))

				// Avisar que hay memoria libre (por si hay procesos en NEW o SUSP_READY esperando)
				globals.Signal(globals.SemReintentarConsultaMemoria) // signal para avisar que hay memoria libre
			}
		}
	// si terminó la IO antes de que el timer expirara, se cancela el timer y no se suspende
	case <-pcb.FinIO:
		timer.Stop()

		pcb = globals.ColaBlocked.GetAndRemoveByPID(pcb.PID) // Saco el PCB de la cola de BLOCKED
		if pcb != nil {
			pcb.CambiarEstado(globals.READY)
			globals.ColaReady.Enqueue(pcb)
			globals.Signal(globals.SemReady)
		}
	}
}

func FinIOEnSuspBlocked(pcb *structs.PCB) {
	pcb.CambiarEstado(globals.SUSPENDED_READY)
	globals.ColaSuspendedReady.Enqueue(pcb)
	globals.Signal(globals.SemNew) // Notificar que hay un proceso en SUSPENDED_READY

	globals.Signal(globals.SemReintentarConsultaMemoria)
}

// TODO hacerlo generico para el dump
func BloquearPorIO(pcb *structs.PCB) {
	pcb.CambiarEstado(globals.BLOCKED)
	globals.ColaBlocked.Enqueue(pcb)
	go manejarTimerSuspension(pcb) // Lanza el timer del planifMP
	globals.Signal(globals.SemBlocked)
}
