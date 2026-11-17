package utils

import (
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func EnviarProcesosEnEsperaIOAExit(colaPCBWaiting []*structs.PCBWaitingIO) {
	for _, pcbWaiting := range colaPCBWaiting {
		pcb := pcbWaiting.PCB

		pcb.CambiarEstado(globals.EXIT)
		globals.ColaExit.Enqueue(pcb)
		globals.Signal(globals.SemExit) // Notificar que hay procesos en exit
	}

}
