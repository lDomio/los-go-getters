package request

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// SolicitarIO envía una solicitud de IO a un dispositivo IO específico con un tiempo determinado y un PID.
func SolicitarIO(IOaEnviar *structs.InstanciaIO, milisegundos int, PID int) {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/requestIO/%d", IOaEnviar.Ip, IOaEnviar.Port, milisegundos)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return
	}
	slog.Debug(fmt.Sprintf("Logre enviar la request a: %s", url))

	q := req.URL.Query()
	q.Add("PID", strconv.Itoa(PID))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		panic("Error al enviar la solicitud a IO")
	}
	defer respuesta.Body.Close()

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		panic("status code not ok al enviar al IO")
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return
	}

	fmt.Println(string(bodyBytes))
}
