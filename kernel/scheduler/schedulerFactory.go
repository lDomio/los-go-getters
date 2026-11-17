package scheduler

import (
	"log"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
)

func schedulerFactoryLP(algoritmo string) func() {
	switch algoritmo {
	case "FIFO":
		return PlanificarLPFifoLP
	case "PMCP":
		return PlanificarLPPMCP
	default:
		panic("Algoritmo de scheduler no soportado")
	}
}

// TODO completar schedulerFactoryCP
func schedulerFactoryCP(algoritmo string) func() {
	switch algoritmo {
	case "FIFO":
		return func() {}
	case "SJFSD":
		return func() {}
	case "SJFCD":
		return func() {}
	default:
		panic("Algoritmo de scheduler no soportado")
	}
}

func PlanificarLPFifoLP() {
	log.Default().Println("Scheduler FIFO LP")
	planificarLargoPlazo(noOrdenar)
}

func PlanificarLPPMCP() {
	log.Default().Println("Scheduler PMCP LP")
	planificarLargoPlazo(ordenarPorTamanio)
}

// 1. Define funciones con nombre
func noOrdenar() {
	// No hace nada, para FIFO
}

// 2. Define una función que ordena por tamaño
func ordenarPorTamanio() {
	globals.ColaNew.OrdenarPorTamanioAscendente()
	globals.ColaSuspendedReady.OrdenarPorTamanioAscendente()
}
