package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/httpUtils"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func ConsultarEspacioSuficiente(tamanio int) bool {
	// Solicitar memoria a la memoria
	solicitud := structs.ProcesoKernelMemoria{
		Size: tamanio,
	}

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		panic("Error al codificar la solicitud de memoria")
	}

	// Enviar la solicitud a memoria
	//respuesta, err := httpUtils.RequestBody(globals.Config.Port_memory, globals.Config.Ip_memory, "POST", "kernel/cargar-pseudocodigo", body)

	// resp, err := httpUtils.EnviarPOST(ip, puerto, "cpu/handshake", body)
	url := fmt.Sprintf("http://%s:%d/kernel/consultar-espacio", globals.Config.Ip_memory, globals.Config.Port_memory) //%s = ip   %d = puerto (por orden)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	//respuesta, err := EnviarPOST(globals.Config.Ip_memory, globals.Config.Port_memory, "/kernel/reservar", body)
	if err != nil {
		panic("Error al enviar la solicitud a memoria")
	}
	defer resp.Body.Close()

	// Decodificar la respuesta de memoria
	var respuestaMemoria structs.RespuestaEspacioMemoria
	err = json.NewDecoder(resp.Body).Decode(&respuestaMemoria)
	if err != nil {
		panic("Error al decodificar la respuesta de memoria")
	}

	slog.Debug("Se recibio ok en la respuesta de memoria")
	return respuestaMemoria.OK
}

func EscribirProcesoEnMemoria(pid int, instrucciones string, tamanio int) bool {
	// Solicitar memoria a la memoria
	solicitud := structs.ProcesoKernelMemoria{
		PID:  pid,
		Size: tamanio,
		Path: instrucciones,
	}

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		panic("Error al codificar la solicitud de memoria")
	}

	// Enviar la solicitud a memoria
	//respuesta, err := httpUtils.RequestBody(globals.Config.Port_memory, globals.Config.Ip_memory, "POST", "kernel/cargar-pseudocodigo", body)

	// resp, err := httpUtils.EnviarPOST(ip, puerto, "cpu/handshake", body)
	url := fmt.Sprintf("http://%s:%d/kernel/cargar-pseudocodigo", globals.Config.Ip_memory, globals.Config.Port_memory) //%s = ip   %d = puerto (por orden)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	//respuesta, err := EnviarPOST(globals.Config.Ip_memory, globals.Config.Port_memory, "/kernel/reservar", body)
	if err != nil {
		panic("Error al enviar la solicitud a memoria")
	}
	defer resp.Body.Close()

	// Decodificar la respuesta de memoria
	var respuestaMemoria structs.RespuestaEspacioMemoria
	err = json.NewDecoder(resp.Body).Decode(&respuestaMemoria)
	if err != nil {
		panic("Error al decodificar la respuesta de memoria")
	}

	slog.Debug("Se recibio ok en la respuesta de memoria")
	return respuestaMemoria.OK

}

// TODO ver con luca
func LiberarMemoria(pid int) {
	// Solicitar liberar memoria a la memoria
	solicitud := structs.ProcesoKernelMemoria{
		PID: pid,
	}

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		panic("Error al codificar la solicitud de liberación de memoria")
	}

	// Enviar la solicitud a memoria
	respuesta, err := httpUtils.RequestBody(globals.Config.Port_memory, globals.Config.Ip_memory, "POST", "kernel/finalizar-proceso", body)
	if err != nil {
		panic("Error al enviar la solicitud de liberación de memoria")
	}
	defer respuesta.Body.Close()

	slog.Debug(fmt.Sprintf("Memoria liberada para el proceso con PID:", pid))
}

func SwapOut(pid int) {
	// Solicitar swap out a la memoria
	solicitud := structs.ProcesoKernelMemoria{
		PID: pid,
	}

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		panic("Error al codificar la solicitud de swap out")
	}

	// Enviar la solicitud a memoria
	respuesta, err := httpUtils.RequestBody(globals.Config.Port_memory, globals.Config.Ip_memory, "POST", "kernel/suspender-proceso", body)
	if err != nil {
		panic("Error al enviar la solicitud de swap out")
	}
	defer respuesta.Body.Close()
}

func SwapIn(pid int) {
	// Solicitar swap in a la memoria
	solicitud := structs.ProcesoKernelMemoria{
		PID: pid,
	}

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		panic("Error al codificar la solicitud de swap in")
	}

	// Enviar la solicitud a memoria
	respuesta, err := httpUtils.RequestBody(globals.Config.Port_memory, globals.Config.Ip_memory, "POST", "kernel/des-suspender-proceso", body)
	if err != nil {
		panic("Error al enviar la solicitud de swap in")
	}
	defer respuesta.Body.Close()
}

func DumpMemory(pid int) bool{
	// Solicitar dump de memoria
	var solicitud structs.ProcesoKernelMemoria

	solicitud.PID = pid

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		panic("Error al codificar la solicitud de dump de memoria")
	}

	// Enviar la solicitud a memoria
	respuesta, err := httpUtils.RequestBody(globals.Config.Port_memory, globals.Config.Ip_memory, "POST", "kernel/dump-memory", body)
	if err != nil {
		panic("Error al enviar la solicitud de dump de memoria")
	}
	defer respuesta.Body.Close()

	slog.Debug("Dump de memoria realizado correctamente")

	var respuestaMemoria structs.RespuestaEspacioMemoria
	err = json.NewDecoder(respuesta.Body).Decode(&respuestaMemoria)
	if err != nil {
		panic("Error al decodificar la respuesta de memoria")
	}

	slog.Debug("Se recibio ok en la respuesta de memoria")
	return respuestaMemoria.OK
}