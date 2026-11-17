package utils

import (
	"log/slog"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/request"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// TODO revisar si esto esta bien y anda
func AtenderProcesosEncoladosEnIO(ioConectada *structs.IOConectada) {
	// Buscar una instancia libre
	instanciaLibre := ioConectada.GetInstanciaIOLibre()
	if instanciaLibre == nil {
		slog.Debug("No hay instancias libres de IO para atender")
		return // No hay instancia libre, no se puede atender
	}

	// Si hay procesos esperando
	if ioConectada.HayProcesosEnColaDeEspera() {
		pcbWaiting := ioConectada.RemovePCBEnEspera()

		pcb := pcbWaiting.PCB
		tiempoIO := pcbWaiting.Tiempo

		// Asignar la instancia y cambiar estado
		instanciaLibre.CambiarEstado(globals.IO_BUSY)
		instanciaLibre.AsignarPID(pcb.PID)

		// Notificar a la IO
		request.SolicitarIO(instanciaLibre, tiempoIO, pcb.PID)
	}
}
