package main

import (
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/handler"
	"github.com/sisoputnfrba/tp-golang/kernel/scheduler"
	"github.com/sisoputnfrba/tp-golang/kernel/utils"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/httpUtils"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func main() {
	// initMock()

	// validar que la config tenga los campos necesarios y leer los valores
	nombreArchivo, tamanio := utils.LeerArgumentos()

	// ======== Creacion del primer proceso ========
	utils.CrearPCB(nombreArchivo, tamanio) //crea el proceso y lo agrega a la cola de new

	// ======== Inicia la planificacion ========

	go scheduler.Planificar(globals.Config.Ready_ingress_algorithm, globals.Config.Scheduler_algorithm) //inicia la planificacion del kernel

	//======== Preparar el servidor y levantarlo ========
	iniciarServidor()

}

func init() {
	globals.Config = config.IniciarConfiguracion[structs.KernelConfig]("KERNELconfig.json")
	logger.ConfigurarLogger("kernel.log", globals.Config.Log_level) //configurar el logger para que escriba en kernel.log

	// validar que la config este cargada correctamente
	if globals.Config == nil {
		log.Fatalf("No se pudo cargar la configuracion")
	}

	// globals.InicializarEstructuras() //inicializa las estructuras de datos del kernel
}

func iniciarServidor() {
	// Inicializar el servidor HTTP
	log.Println("Iniciando el servidor HTTP en el puerto:", globals.Config.Port_kernel)
	mux := http.NewServeMux()

	// Asignar las rutas y sus manejadores
	mux.HandleFunc("POST /io", handler.RegistrarIO)             // Registrar IO
	mux.HandleFunc("POST /cpu", handler.RegistrarCpu)           // Registrar CPU
	mux.HandleFunc("POST /cpu/handshake", handler.RegistrarCpu) // Registrar handshake de CPU

	// procesos
	mux.HandleFunc("POST /process/devolucion", handler.DevolucionProceso) // syscall INIT_PROCESS
	mux.HandleFunc("POST /process/finDeIO", handler.FinDeIO)              // TODO syscall FIN_IO

	mux.HandleFunc("POST /io/desconexion", handler.DesconexionIO) // syscall FIN_CPU
	// Iniciar el servidor en un goroutine para que no bloquee
	httpUtils.StartServer(globals.Config.Port_kernel, mux)
}
