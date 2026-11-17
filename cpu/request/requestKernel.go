package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func ConectarCPU(ip string, puerto int, identificador int) {
	mensaje := structs.HandshakeRequestCPU{
		IP:     globals.CpuConfig.Ip_cpu,
		ID:     identificador,
		Puerto: globals.CpuConfig.Port_cpu,
	}

	body, err := json.Marshal(mensaje)
	if err != nil {
		log.Printf("error codificando mensaje: %s", err.Error())
	}

	// resp, err := httpUtils.EnviarPOST(ip, puerto, "cpu/handshake", body)
	url := fmt.Sprintf("http://%s:%d/cpu/handshake", ip, puerto) //%s = ip   %d = puerto (por orden)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando mensaje a ip:%s puerto:%d", ip, puerto)
	}

	slog.Debug(fmt.Sprintf("respuesta del servidor: %s", resp.Status))
}

func ReturnProcess(motivoDesalojo string, parametros []string, PC int, PID int) {
	// Armo la estructura de devolución
	procesoADevolver := structs.DevolucionProceso{
		PID:                 PID,
		PC:                  PC,
		MotivoDeDesalojo:    motivoDesalojo,
		ParametrosDesalojos: parametros,
	}

	body, err := json.Marshal(procesoADevolver)
	if err != nil {
		log.Printf("error codificando mensaje: %s", err.Error())
		return
	}

	url := fmt.Sprintf("http://%s:%d/process/devolucion", globals.CpuConfig.Ip_kernel, globals.CpuConfig.Port_kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando mensaje a ip:%s puerto:%v", globals.CpuConfig.Ip_kernel, globals.CpuConfig.Port_kernel)
		panic(err)
	}
	defer resp.Body.Close()

	// TODO recibir el código de estado de la respuesta y manejarlo adecuadamente. Si
	// Leer la respuesta del servidor
	responseBody := new(bytes.Buffer)
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		log.Printf("error leyendo la respuesta del servidor: %s", err.Error())
		return
	}

	slog.Debug(fmt.Sprintf("respuesta del servidor: %s - %s", resp.Status, responseBody.String()))

}
