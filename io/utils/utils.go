package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/io/globals"
	"github.com/sisoputnfrba/tp-golang/io/request"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

var mutexIO sync.Mutex

type MensajeIo struct { // enrealidad ya es un paquete, no un mensaje
	Nombre string `json:"nombre"`
	Ip     string `json:"ip"`
	Port   int    `json:"port"`
}

func EnviarMensajeIO(ip string, puerto int, mensaje MensajeIo) {

	body, err := json.Marshal(mensaje)
	if err != nil {
		slog.Error("error codificando mensaje:", "error", err.Error())
	}

	// http://ip:8080/io

	url := fmt.Sprintf("http://%s:%d/io", ip, puerto) //%s = ip   %d = puerto (por orden)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		slog.Error("error enviando mensaje a ip al puerto:", "IP", ip, "Puerto", puerto)
	}

	slog.Info("respuesta del servidor:", "Status", resp.Status)
}

func TrabajarIO(w http.ResponseWriter, r *http.Request) {

	queryParams := r.URL.Query()
	tiempoString := r.PathValue("tiempo")
	pid := queryParams.Get("PID")
	intValue, _ := strconv.Atoi(pid)
	

	tiempo, err := strconv.Atoi(tiempoString)
	if err != nil {
		slog.Error("Error al pasar el tiempo a Int:", "error", err)
		return
	}

	// Iniciar un goroutine para simular el tiempo de IO
	go IniciarTimerIO(tiempo, intValue)

	w.WriteHeader(http.StatusOK)                                   // Responder con 200 OK
	w.Write([]byte("Se recibio la solicitud de IO correctamente")) // Enviar un mensaje de confirmación
}

func IniciarTimerIO(tiempo int, pid int) {

	mutexIO.Lock()

	globals.PIDEnEjecucion = pid // Guardar el PID de la IO en ejecución

	logger.InicioIO(pid, tiempo) // Log de inicio de IO

	time.Sleep(time.Duration(tiempo) * time.Millisecond)
	request.FinDeIO(pid) // Notificar al kernel que se finalizó la IO

	logger.FinalizacionDeIO(pid) // Log de finalización de IO

	globals.PIDEnEjecucion = 0 // Limpiar el PID de la IO en ejecución

	mutexIO.Unlock()
}
