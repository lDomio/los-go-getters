package request

import (
	"encoding/json"
	"log/slog"

	"github.com/sisoputnfrba/tp-golang/utils/httpUtils"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Dispatch(pid int, pc int, cpu *structs.CPU) {
	request := structs.RequestDispatch{
		PID: pid,
		PC:  pc,
	}

	// Codificar la solicitud en JSON
	body, err := json.Marshal(request)
	if err != nil {
		panic("Error al codificar la solicitud de CPU")
	}

	// Enviar la solicitud a CPU
	respuesta, err := httpUtils.RequestBody(cpu.Puerto, cpu.IP, "POST", "kernel/excec", body)
	if err != nil {
		panic("Error al enviar la solicitud a CPU")
	}
	defer respuesta.Body.Close()

	slog.Debug("Solicitud de despacho enviada a CPU", "PID", pid, "PC", pc)

}

// Envia una señal de interrupción a la CPU (esta ya sabe que proceso esta ejecutando)
func Interrupt(cpu *structs.CPU) {
	// Enviar la solicitud a CPU
	respuesta, err := httpUtils.RequestBody(cpu.Puerto, cpu.IP, "POST", "kernel/interrupt", nil)
	if err != nil {
		panic("Error al enviar la solicitud de interrupción a CPU")
	}
	defer respuesta.Body.Close()
}
