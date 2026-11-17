package scheduler

import (
	"log/slog"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
)

// TODO implementar planificador de corto plazo
// TODO agregar el manejo de procesos en exit
// TODO implementar planificador de mediano plazo
func Planificar(ready_ingress_algorithm string, scheduler_algorithm string) {
	//planificadorLP := schedulerFactoryLP(scheduler_algorithm)
	planificadorLP := schedulerFactoryLP(ready_ingress_algorithm)

	if inicioPlanificadorLP() {
		globals.Signal(globals.SemSyncSchedulerCP) // Notifica que se ha iniciado la planificación a largo plazo
		go planificadorLP()                        // Inicia la planificación de largo plazo
		//go PlanificarCP(ready_ingress_algorithm)   // Inicia la planificación de corto plazo
		go PlanificarCP(scheduler_algorithm)
		go PlanificarProcesosFinalizados()
	} else {
		slog.Error("No se pudo iniciar la planificacion a largo plazo")
		panic("No se pudo iniciar la planificacion a largo plazo")
	}

	// go planificadorCP() // Inicia la planificación de corto plazo
	// administrarColas Exit, blocked..

}
