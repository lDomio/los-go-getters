package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/io/globals"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func DesconexionIO(sigs chan os.Signal, nombre string) {
	<-sigs
	// Al recibir la señal, notificar al Kernel
	data := structs.DesconexionOFinDeIO{
		NombreIO: nombre,
		PID:      globals.PIDEnEjecucion,
	}

	body, _ := json.Marshal(data)
	url := fmt.Sprintf("http://%s:%d/io/desconexion", globals.IoConfig.Ip_kernel, globals.IoConfig.Port_kernel)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error notificando desconexion de IO al kernel: %v\n", err)
	} else {
		resp.Body.Close()
	}

	os.Exit(0)
}

func FinDeIO(pid int) {
	// Al finalizar la IO, notificar al Kernel
	data := structs.DesconexionOFinDeIO{
		NombreIO: globals.IoConfig.Nombre,
		PID:      pid,
	}

	body, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error serializando datos para fin de IO: %v\n", err)
		return
	}

	url := fmt.Sprintf("http://%s:%d/process/finDeIO", globals.IoConfig.Ip_kernel, globals.IoConfig.Port_kernel)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error notificando fin de IO al kernel: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if err != nil {
		fmt.Printf("Error cerrando el cuerpo de la respuesta: %v\n", err)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("Error en la respuesta del kernel: Código de estado %d\n", resp.StatusCode)
		return
	}
}

func HandshakeKernel(nombre string) {
	// Enviar el mensaje de handshake al kernel
	data := structs.HandshakeIOKernel{
		Nombre: nombre,
		Ip:     globals.IoConfig.Ip_io,
		Port:   globals.IoConfig.Port_io,
	}

	body, err := json.Marshal(data)
	if err != nil {
		slog.Error("Error serializando datos para handshake de IO", "error", err)
		return
	}

	url := fmt.Sprintf("http://%s:%d/io", globals.IoConfig.Ip_kernel, globals.IoConfig.Port_kernel)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error notificando fin de IO al kernel: %v\n", err)
		return
	}

	// Cerrar el cuerpo de la respuesta explícitamente
	err = resp.Body.Close()
	if err != nil {
		fmt.Printf("Error cerrando el cuerpo de la respuesta: %v\n", err)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("Error en la respuesta del kernel: Código de estado %d\n", resp.StatusCode)
		return
	}

	slog.Debug("Handshake exitoso con el kernel", "Status", resp.Status)
}
