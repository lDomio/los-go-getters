package globals

import "github.com/sisoputnfrba/tp-golang/utils/structs"

const MAXIMO_PROCESOS_BUFFER = 300

// Canales de Sincro
var (
	SemReady       chan struct{} // Canal para notificar que hay procesos en la cola de ready
	SemNew         chan struct{} // Canal para notificar que hay un proceso en la cola de new
	SemExit        chan struct{} // Canal para notificar que hay procesos en la cola de exit
	SemBlocked     chan struct{} // Canal para notificar que hay procesos en la cola de blocked
	SemSuspBlocked chan struct{} // Canal para notificar que hay procesos en la cola de suspended blocked

	SemSuspendedReady chan struct{} // Canal para notificar que hay procesos en la cola de suspended ready

	SemSyncSchedulerCP chan struct{} // Canal para notificar que se ha iniciado el planificador de LP

	HayCPUsLibres chan *structs.CPU // Canal para notificar que hay CPUs disponibles

	SemDesalojar                 chan struct{} // Canal para notificar que se ha producido un desalojo de CPU
	SemReintentarConsultaMemoria chan struct{} // Canal para notificar que se debe reintentar la consulta de memoria
)

func init() {
	// Inicializar los canales
	SemReady = make(chan struct{}, MAXIMO_PROCESOS_BUFFER)   // Canal para notificar que hay procesos en la cola de ready
	SemNew = make(chan struct{}, MAXIMO_PROCESOS_BUFFER)     // Canal para notificar que hay un proceso en la cola de new
	SemExit = make(chan struct{}, MAXIMO_PROCESOS_BUFFER)    // Canal para notificar que hay procesos en la cola de exit
	SemBlocked = make(chan struct{}, MAXIMO_PROCESOS_BUFFER) // Canal para notificar que hay procesos en la cola de blocked

	SemSuspBlocked = make(chan struct{}, MAXIMO_PROCESOS_BUFFER) // Canal para notificar que hay procesos en la cola de suspended blocked

	SemSuspendedReady = make(chan struct{}, MAXIMO_PROCESOS_BUFFER) // Canal para notificar que hay procesos en la cola de suspended ready

	SemSyncSchedulerCP = make(chan struct{}, 1) // Canal para notificar que se ha iniciado el planificador de LP

	HayCPUsLibres = make(chan *structs.CPU, 300) // Canal para notificar que hay CPUs disponibles

	SemDesalojar = make(chan struct{}, 20)                                     // Canal para notificar que se ha producido un desalojo de CPU
	SemReintentarConsultaMemoria = make(chan struct{}, MAXIMO_PROCESOS_BUFFER) // Canal para notificar que se debe reintentar la consulta de memoria
}

func Wait(canal chan struct{}) {
	// Espera a que se envíe un valor al canal
	<-canal
}

func Signal(canal chan struct{}) {
	// Envía un valor al canal para notificar que se ha producido un evento
	canal <- struct{}{}
}

func Product[T any](canal chan T, valor T) {
	// Envía un valor al canal
	canal <- valor
}
func Consume[T any](canal chan T) T {
	// Consume un valor del canal
	valor := <-canal
	return valor
}
