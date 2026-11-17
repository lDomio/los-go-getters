package scheduler

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/request"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func SchedulerFactoryCP(algoritmo string) {
	switch algoritmo {
	case "FIFO":
		log.Default().Println("Scheduler FIFO CP")
		PlanificarFifoCP()
	case "SJFSD":
		log.Default().Println("Scheduler SJF Sin Desalojo")
		PlanificarSJFSD()
	case "SJFCD":
		log.Default().Println("Scheduler SJF Con Desalojo")
		PlanificarSJFCD()
	default:
		panic("Algoritmo de scheduler no soportado")
	}
}

func PlanificarFifoCP() {
	for {
		// esperar a que haya un proceso en ready
		globals.Wait(globals.SemReady)

		PCB := globals.ColaReady.Dequeue() //asigna el primer proceso de la cola a la variable PCB y borra el primer elemento de la cola de ready
		//tiene que llegar el proceso en estado ready

		ejecutarProceso(PCB) // Simula la ejecución del proceso

	}

}

func PlanificarSJFSD() {
	for {
		globals.Wait(globals.SemReady)
		//ordenamiento con funcion anonima de menor a mayor
		globals.ColaReady.OrdenarPorEstimacionActual()

		PCB := globals.ColaReady.Dequeue() //asigna el primer proceso de la cola a la variable PCB y borra el primer elemento de la cola de ready
		//tiene que llegar el proceso en estado ready

		ejecutarProceso(PCB) // Simula la ejecución del proceso
	}
}

func PlanificarSJFCD() {
	for {
		globals.Wait(globals.SemReady)
		globals.ColaReady.OrdenarPorEstimacionActual()

		PCB := globals.ColaReady.Dequeue()

		if !hayCpusLibres() {
			log.Println("No hay CPUs libres")
			if debeDesalojar, cpuADesalojar := tieneRafagaMasCortaQueLosProcesosEnExec(PCB); debeDesalojar {
				log.Println("No tiene rafaga mas corta")
				despacharConDesalojo(PCB, cpuADesalojar) // Desaloja el proceso de la CPU con el tiempo restante más alto y asigna el nuevo proceso
			} else {
				log.Printf("tiene rafaga mas corta")
				globals.ColaReady.Enqueue(PCB) // Si no hay CPUs libres, vuelve a encolar el PCB en la cola de ready

				cpu := globals.Consume(globals.HayCPUsLibres) //* espera a que haya una CPU libre (bloqueante para evitar el bucle infinito)
				globals.Product(globals.HayCPUsLibres, cpu)   //* Vuelve a poner la CPU en el canal de CPUs libres

				globals.Signal(globals.SemReady)
			}
		} else {
			log.Println("hay CPUs libres")
			ejecutarProceso(PCB) // Simula la ejecución del proceso
		}

	}
}

func despacharConDesalojo(PCB *structs.PCB, cpuADesalojar *structs.CPU) {
	request.Interrupt(cpuADesalojar)   // Desaloja el proceso de la CPU con el tiempo restante más alto
	globals.Wait(globals.SemDesalojar) // Espera a que se complete el desalojo

	// Asignar el proceso a la CPU
	asignarProcesoACPU(PCB, cpuADesalojar)

	// Enviar el proceso a la CPU para su ejecución
	request.Dispatch(PCB.PID, PCB.PC, cpuADesalojar) // Enviar el proceso a la CPU para su ejecución sin bloquear el scheduler
	slog.Debug(fmt.Sprintf("Ejecutando proceso (ID: %d)\n", PCB.PID))
}

// Verifica si el proceso tiene una estimación de tiempo de ejecución menor que los que estan en exec (en el peor de los casos, el que tiene la estimación más alta)
func tieneRafagaMasCortaQueLosProcesosEnExec(PCB *structs.PCB) (bool, *structs.CPU) {
	pcbConRafagaMasLarga, encontrado := globals.ColaExec.GetPCBConTiempoRestanteMasAlto()

	if !encontrado {
		log.Println("No hay procesos en ejecución para comparar")
		return false, nil // No hay procesos en ejecución, no se puede comparar
	}

	// Obtenemos la CPU en la misma función
	IDCPUADesalojar := globals.ProcesosEnCPU.Get(pcbConRafagaMasLarga.PID)
	log.Printf("IDCPU a desalojar IDCPUADesalojar: %d", IDCPUADesalojar)
	cpuADesalojar := globals.CpusConectadas.Get(IDCPUADesalojar)
	log.Printf("PCB Estimacion actual: %f, rafagaMasLarga: %f", PCB.EstimacionActual, pcbConRafagaMasLarga.EstimacionActual)

	return PCB.EstimacionActual < pcbConRafagaMasLarga.EstimacionActual, cpuADesalojar
}

func hayCpusLibres() bool {
	return len(globals.HayCPUsLibres) > 0 // Verifica si hay CPUs disponibles
}

func desalojarCPU(PCB *structs.PCB) {
	// Buscar CPU que posea al Proceso con el tiempo restante más alto

}

func PlanificarCP(algoritmo string) {
	globals.Wait(globals.SemSyncSchedulerCP)
	SchedulerFactoryCP(algoritmo) // crea el scheduler segun el algoritmo que se le pase por config
}

func ejecutarProceso(PCB *structs.PCB) {
	if PCB.Estado == globals.READY {

		// Buscar una CPU disponible
		cpu := buscarCPUDisponible()

		// Asignar el proceso a la CPU
		asignarProcesoACPU(PCB, cpu)

		// Enviar el proceso a la CPU para su ejecución
		go request.Dispatch(PCB.PID, PCB.PC, cpu) // Enviar el proceso a la CPU para su ejecución sin bloquear el scheduler

		slog.Debug(fmt.Sprintf("Ejecutando proceso (ID: %d)\n", PCB.PID))

	} else {
		slog.Debug(fmt.Sprintf("Proceso (ID: %d) no está listo para ejecutar\n", PCB.PID))
	}
}

func asignarProcesoACPU(PCB *structs.PCB, cpu *structs.CPU) {
	globals.ColaExec.Enqueue(PCB)              // Encola el PCB en la cola de ejecución
	globals.ProcesosEnCPU.Set(PCB.PID, cpu.ID) // Asigna el PCB al mapa de procesos en CPU
	cpu.CambiarEstado(false)                   // Cambia el estado de la CPU a OCUPADA
	PCB.CambiarEstado(globals.EXEC)            // Cambia el estado del proceso a EXEC

	slog.Debug(fmt.Sprintf("Asignando proceso (ID: %d) a CPU (ID: %d)\n", PCB.PID, cpu.ID))
}

func buscarCPUDisponible() *structs.CPU {
	return globals.Consume(globals.HayCPUsLibres) // Consume un valor del canal HayCPUsLibres para obtener una CPU disponible
}
