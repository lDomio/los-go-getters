package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sisoputnfrba/tp-golang/io/globals"
	"github.com/sisoputnfrba/tp-golang/io/request"
	"github.com/sisoputnfrba/tp-golang/io/utils"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/httpUtils"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	nombre := LeerArgumentos()

	// Cargar la configuración de la IO
	globals.IoConfig = config.IniciarConfiguracion[globals.IOConfig]("IOconfig.json")
	if globals.IoConfig == nil {
		log.Fatalf("No se pudo cargar la configuracion")
	}
	slog.Debug("Se obtuvo estos datos del config: ", "IOConfig", globals.IoConfig)
	globals.IoConfig.Nombre = nombre // Asignar el nombre de la IO desde los argumentoss

	// Configurar el logger
	logger.ConfigurarLogger(fmt.Sprintf("io_%s.log", nombre), globals.IoConfig.Log_level) //configurar el logger para que escriba en io.log

	// Capturar señales de interrupción o terminación
	go request.DesconexionIO(sigs, nombre)

	go request.HandshakeKernel(nombre) // Enviar el mensaje de handshake al kernel

	//Abrir server para
	mux := http.NewServeMux()
	// asignar una funcion
	mux.HandleFunc("/requestIO/{tiempo}", utils.TrabajarIO) //recibir mensaje de la io

	httpUtils.StartServer(globals.IoConfig.Port_io, mux) //! sigue escuchando mientras se siguen ejecutando las demás lineas de código

	
}

func LeerArgumentos() string {
	// validar que se pasen los argumentos necesarios
	if len(os.Args) < 2 {
		log.Fatalf("Faltan argumentos: ./io [nombre]")
	}

	// Leemos los argumentos de entrada
	nombreIO := os.Args[1]

	return nombreIO
}
