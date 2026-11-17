package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

type SolicitudMemoria struct {
	PID          int
	Size         int
	Instructions []string
}

func SolicitarMemoria(pcb structs.PCB) {
	// Solicitar memoria a la memoria
	solicitud := structs.RequestMemoria{
		PID:  pcb.PID,
		Size: pcb.Tamanio,
	}

	body, err := json.Marshal(solicitud)
	if err != nil {
		panic("Error al codificar la solicitud de memoria")
	}

	//url := "http://" + globals.KernelConfig.Ip_memory + ":" + strconv.Itoa(globals.KernelConfig.Port_memory) + "/memory/reservar"
	url := fmt.Sprintf("http://%s:%d/reservar", globals.Config.Ip_memory, globals.Config.Port_memory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic("Error al enviar la solicitud de memoria")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic("Error al solicitar memoria")
	}

}
