package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func Fetch(pid int, pc int) (structs.Instruccion, error) {
	// -----------REQUEST A MEMORIA ----------------
	// Solicitar memoria a la memoria
	var solicitud structs.ProcesoEnEjecucion

	solicitud.PID = pid
	solicitud.PC = pc

	logger.FetchInstrucciones(solicitud.PID, solicitud.PC)

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		panic("Error al codificar la solicitud de memoria")
	}

	slog.Debug("Se solicito la siguiente instruccion a memoria")

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/cpu/solicitar-instruccion", globals.CpuConfig.Ip_memory, globals.CpuConfig.Port_memory)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		return structs.Instruccion{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return structs.Instruccion{}, err
	}

	// -----------RESPONSE MEMORIA ----------------

	// Decodificar la respuesta de memoria
	var respuestaMemoria structs.Instruccion

	err = json.NewDecoder(respuesta.Body).Decode(&respuestaMemoria)
	if err != nil {
		panic("Error al decodificar la respuesta de memoria")
	}

	// TODO arreglar para mostrar los parametros
	slog.Debug("Se recibio la instruccion: %s con parametros: %v", respuestaMemoria.Operacion, respuestaMemoria.Parametros)

	return respuestaMemoria, err
}

func HandshakeMemoria() error {
	// Enviar un mensaje de saludo a memoria para verificar la conexión
	
	
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/cpu/handshake", globals.CpuConfig.Ip_memory, globals.CpuConfig.Port_memory)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)

	if err != nil {
		return err
	}

	var Handshake structs.HandshakeMemoria;

	err = json.NewDecoder(respuesta.Body).Decode(&Handshake)
	if err != nil {
		return fmt.Errorf("error al decodificar la respuesta del handshake: %v", err)
	}
	defer respuesta.Body.Close()

	if respuesta.StatusCode != http.StatusOK {
		return fmt.Errorf("error en el handshake con memoria: %s", respuesta.Status)
	}

	globals.DatosMemoria.TamanioPagina = Handshake.TamanioPagina
	globals.DatosMemoria.NumeroNiveles = Handshake.NumeroNiveles
	globals.DatosMemoria.Cant_Entradas_Tabla = Handshake.CantEntradasTabla

	slog.Debug("Handshake con memoria exitoso")
	return nil
}

func RequestPaginaCompleta(frame int) ([]byte, error) {
	// Crear la solicitud
	solicitud := structs.SolicitudFrameCompleta{
		PID:    globals.PCBenEjecucion.PID,
		Frame: frame,
	}

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		return nil, fmt.Errorf("error al codificar la solicitud: %w", err)
	}

	url := fmt.Sprintf("http://%s:%d/cpu/pagina-completa", globals.CpuConfig.Ip_memory, globals.CpuConfig.Port_memory)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error al crear la solicitud: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al enviar la solicitud: %w", err)
	}
		
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error en la respuesta del servidor: %s", resp.Status)
	}

	datos := make([]byte, globals.DatosMemoria.TamanioPagina)
	n, err := resp.Body.Read(datos)
	if err != nil && err.Error() != "EOF" {
		return nil, fmt.Errorf("error al leer la respuesta de memoria: %w", err)
	}
	datos = datos[:n]


	return datos, nil
}

func ActualizarPaginaCompleta(frame int, datos []byte) error {
	// Crear la solicitud
	solicitud := structs.ActualizarFrameCompletaRequest{
		PID:    globals.PCBenEjecucion.PID,
		Frame:  frame,
		Datos:  datos,
	}

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		return fmt.Errorf("error al codificar la solicitud: %w", err)
	}

	url := fmt.Sprintf("http://%s:%d/cpu/pagina-completa", globals.CpuConfig.Ip_memory, globals.CpuConfig.Port_memory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error al enviar la solicitud: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la respuesta del servidor: %s", resp.Status)
	}

	return nil
}

func LeerMemoriaRequest(physicalAddress int, size int) ([]byte, error) {
	// Crear la solicitud
	solicitud := structs.LeerMemoriaRequest{
		PID:    globals.PCBenEjecucion.PID,
		PhysicalAddress: physicalAddress, // Direccion fisica a leer
		Size:   size,
	}

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		return nil, fmt.Errorf("error al codificar la solicitud: %w", err)
	}

	url := fmt.Sprintf("http://%s:%d/cpu/leer-espacio-usuario", globals.CpuConfig.Ip_memory, globals.CpuConfig.Port_memory)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error al crear la solicitud: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al enviar la solicitud: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error en la respuesta del servidor: %s", resp.Status)
	}

	datos := make([]byte, size)
	n, err := resp.Body.Read(datos)
	if err != nil && err.Error() != "EOF" {
		return nil, fmt.Errorf("error al leer la respuesta: %w", err)
	}
	datos = datos[:n]

	logger.LecturaMemoria(globals.PCBenEjecucion.PID, solicitud.PhysicalAddress, string(datos))

	return datos, nil
}

func ActualizarMemoriaRequest(physicalAddress int, datos []byte) error {
	// Crear la solicitud
	solicitud := structs.EscribirMemoriaRequest{
		PID:    globals.PCBenEjecucion.PID,
		PhysicalAddress: physicalAddress, // Direccion fisica a escribir
		Datos:  datos, // Datos a escribir
	}

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		return fmt.Errorf("error al codificar la solicitud: %w", err)
	}

	url := fmt.Sprintf("http://%s:%d/cpu/escribir-espacio-usuario", globals.CpuConfig.Ip_memory, globals.CpuConfig.Port_memory)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error al enviar la solicitud: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la respuesta del servidor: %s", resp.Status)
	}

	logger.EscrituraMemoria(globals.PCBenEjecucion.PID, solicitud.PhysicalAddress, string(datos))

	return nil
}

func ObtenerFramePorEntradasMultinivel(entradasMultinivel []int) (int, error) {
	// Crear la solicitud
	solicitud := structs.EnvioEntradasMultinivel{
		PID:                globals.PCBenEjecucion.PID,
		Entradas: 			entradasMultinivel,
	}

	// Codificar la solicitud en JSON
	body, err := json.Marshal(solicitud)
	if err != nil {
		return 0, fmt.Errorf("error al codificar la solicitud: %w", err)
	}

	url := fmt.Sprintf("http://%s:%d/cpu/acceder-tabla-de-paginas", globals.CpuConfig.Ip_memory, globals.CpuConfig.Port_memory)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("error al crear la solicitud: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error al enviar la solicitud: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error en la respuesta del servidor: %s", resp.Status)
	}

	var respuesta struct {
		Frame int `json:"frame"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respuesta)
	if err != nil {
		return 0, fmt.Errorf("error al decodificar la respuesta: %w", err)
	}

	return respuesta.Frame, nil
}


