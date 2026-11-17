package utils

import (
	"fmt"
	"log/slog"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/request"
	"github.com/sisoputnfrba/tp-golang/kernel/scheduler"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func SyscallExit(PCB *structs.PCB) {

	liberarCPUPorPID(PCB.PID) // libera la CPU asignada al proceso

	// lo paso a exit
	PCB.CambiarEstado(globals.EXIT)
	globals.ColaExit.Enqueue(PCB)
	globals.Signal(globals.SemExit) // Notifica que hay procesos en exit

	// liberar la cpu asignada al proceso

}

func SyscallIO(PCB *structs.PCB, nombreIO string, tiempoIO int) {

	liberarCPUPorPID(PCB.PID) // libera la CPU asignada al proceso

	PCB.CalcularEstimacion(globals.Config.Alpha) //calculo el nuevo tiempo de estimación

	IOBuscado := BuscarIOSolicitado(nombreIO)
	//* si existe la IO se transiciona a BLOCKED y se logea el IO (cpu queda esperando)
	//* si no existe la IO se transiciona a EXIT y se logea el error (largo plazo se encarga de exit)
	if IOBuscado != nil {
		slog.Debug(fmt.Sprintf("IO Encontrado: %s", nombreIO))

		/*
			"En caso de que si exista al menos una instancia de IO, aun si la misma se encuentre ocupada,
			el kernel deberá pasar el proceso al estado BLOCKED y agregarlo a la cola de bloqueados por la IO solicitada"
		*/
		scheduler.BloquearPorIO(PCB) // Bloquea el proceso por IO, inicia el timer etc etc

		// log obligatorio (5/8)
		logger.MotivoDeBloqueo(PCB.PID, IOBuscado.Nombre) // Log del motivo de bloqueo

		/*
			"en caso de que el kernel tenga una instancia de esta IO libre,
			deberá enviarle al módulo correspondiente el PID y el tiempo por el cual debe bloquearse"
		*/
		// Se busca una IO libre para realizar la solicitud
		IOLibre := IOBuscado.GetInstanciaIOLibre()
		if IOLibre != nil {
			slog.Debug(fmt.Sprintf("IO %s encontrado y libre", nombreIO))

			IOLibre.CambiarEstado(globals.IO_BUSY)
			IOLibre.AsignarPID(PCB.PID) // Asigna el PID del proceso a la IO
			request.SolicitarIO(IOLibre, tiempoIO, PCB.PID)
		} else {
			slog.Debug("La IO %s libre para el proceso %d, se bloqueará\n", nombreIO, PCB.PID)
			// Agrego el proceso a la cola de espera de IO
			IOBuscado.AgregarProcesoAColaDeEspera(PCB, tiempoIO) // Agrega el PCB a la cola de espera del IO
		}

	} else {
		slog.Debug(fmt.Sprintf("IO NO Encontrado: %s", nombreIO))

		PCB.CambiarEstado(globals.EXIT) // Cambia el estado a EXIT si no se encuentra el IO
		globals.ColaExit.Enqueue(PCB)   // Agrega el PCB a la cola de exit

		// Da señal al planificador de largo plazo que hay un proceso en exit
		globals.Signal(globals.SemExit) // Notifica que el proceso ha finalizado
	}
}
