package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	//"utils/globals"   TEMPORAL

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/request"

	//"github.com/sisoputnfrba/tp-golang/cpu/request"
	"github.com/sisoputnfrba/tp-golang/cpu/utils"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/httpUtils"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func main() {

	id := LeerArgumentos()

	// validar que la config este cargada correctamente

	globals.CpuConfig = config.IniciarConfiguracion[globals.CPUConfig]("CPUconfig.json")

	if globals.CpuConfig == nil {
		log.Fatalf("No se pudo cargar la configuracion")
	}

	logger.ConfigurarLogger(fmt.Sprintf("cpu_%d.log", id), globals.CpuConfig.Log_level) //configurar el logger para que escriba en cpu.log

	//======== Preparar el servidor y levantarlo ========

	mux := http.NewServeMux()
	// asignar una funcion

	// HANDSHAKES
	request.HandshakeMemoria()
	request.ConectarCPU(globals.CpuConfig.Ip_kernel, globals.CpuConfig.Port_kernel, id)

	utils.InicializarEstructurasBasicas(globals.DatosMemoria.TamanioPagina, globals.DatosMemoria.NumeroNiveles, globals.DatosMemoria.Cant_Entradas_Tabla)

	// PROCESOS
	mux.HandleFunc("POST /kernel/excec", utils.EjecutarProceso)        //recibe un PCB y lo ejecuta
	mux.HandleFunc("POST /kernel/interrupt", utils.InterrumpirProceso) // recibe una interrupcion

	httpUtils.StartServer(globals.CpuConfig.Port_cpu, mux)

}

// LeerArgumentos lee los argumentos de la línea de comandos y devuelve el identificador de la CPU.
func LeerArgumentos() int {
	// validar que se pasen los argumentos necesarios
	if len(os.Args) < 2 {
		log.Fatalf("Faltan argumentos: ./cpu [identificador]")
	}

	// Leemos los argumentos de entrada
	id, _ := strconv.Atoi(os.Args[1])
	return id
}
