package globals

import (
	"sync/atomic"

	"github.com/sisoputnfrba/tp-golang/utils/safe"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var Config *structs.KernelConfig

// Estados de los procesos
const (
	NEW     string = "NEW"
	READY   string = "READY"
	EXEC    string = "EXECUTE"
	BLOCKED string = "BLOCKED"
	EXIT    string = "EXIT"

	SUSPENDED_BLOCKED string = "SUSPENDED_BLOCKED" // Estado de un proceso suspendido en BLOCKED
	SUSPENDED_READY   string = "SUSPENDED_READY"   // Estado de un proceso suspendido en READY

	IO_BUSY bool = false // Indica que hay IO ocupada
	IO_FREE bool = true  // Indica que hay IO libre
)

// Cola de procesos
var (

	// Estructura para guardar los procesos
	Procesos map[int]structs.PCB

	// pidCounter es un contador para asignar PID a los procesos
	pidCounter int32
)

// CPUs registrados
var CpusConectadas safe.MapCPUSeguro        // HashMap de CPUs registradas
var ProcesosEnCPU *safe.MapSeguro[int, int] // map[pid, cpuID] Mapa de procesos ejecutando por CPU
// IOs registrados
var IOsConectados *safe.MapIOConectadaSeguro // HashMap de IOs registradas

func InicializarEstructuras() {

	CpusConectadas = safe.InicializarMapCPUSeguro()       // Inicializar el mapa de CPUs conectadas
	ProcesosEnCPU = safe.InicializarMapSeguro[int, int]() // Inicializar el mapa de procesos en CPU

	IOsConectados = safe.InicializarMapIOConectadaSeguro() // HashMap de IOs registradas

	Procesos = make(map[int]structs.PCB)
	pidCounter = 0

	// HayProcesosEnReady = make(chan struct{}, 10) // Canal para notificar que hay procesos en la cola de ready
	// HayProcesoEnNew = make(chan struct{}, 10)    // Canal para notificar que hay un proceso en la cola de new

}

func init() {
	InicializarEstructuras() // Inicializa las estructuras de datos del kernel
}

func NextPID() int {
	return int(atomic.AddInt32(&pidCounter, 1) - 1)
}
