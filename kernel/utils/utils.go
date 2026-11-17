package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/request"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

type Mensaje struct {
	Mensaje string `json:"mensaje"`
}

type MensajeCPU struct {
	Identificador string `json:"id"`
}

func EnviarMensajeKernel(ip string, puerto int, mensajeTxt string) {
	mensaje := MensajeCPU{Identificador: mensajeTxt}
	body, err := json.Marshal(mensaje)
	if err != nil {
		log.Printf("error codificando mensaje: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/kernel", ip, puerto) //%s = ip   %d = puerto (por orden)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando mensaje a ip:%s puerto:%d", ip, puerto)
	}

	slog.Debug(fmt.Sprintf("respuesta del servidor: %s", resp.Status))
}

//*=====================================| PROCESOS |========================================================\\

// CrearPCB crea un nuevo PCB (Process Control Block) con las instrucciones y el tamaño de la memoria
func CrearPCB(instrucciones string, tamanio int) *structs.PCB {

	slog.Debug(fmt.Sprintf("Creando PCB con las siguientes instrucciones: %s y tamaño de memoria: %d", instrucciones, tamanio))

	// Crear un nuevo PCB con el tamaño de la memoria y las instrucciones
	pcb := &structs.PCB{
		PID:                  globals.NextPID(),
		PC:                   0,
		Tamanio:              tamanio,
		Instrucciones:        instrucciones,
		Estado:               "NEW",
		ME:                   map[string]int{"NEW": 1, "READY": 0, "SUSPENDED_READY": 0, "SUSPENDED_BLOCKED": 0, "EXECUTE": 0, "BLOCKED": 0, "EXIT": 0},
		MT:                   map[string]float64{"NEW": 0, "READY": 0, "SUSPENDED_READY": 0, "SUSPENDED_BLOCKED": 0, "EXECUTE": 0, "BLOCKED": 0, "EXIT": 0},
		TInicioEstado:        time.Now(),
		EstimacionActual:     globals.Config.Initial_estimate, // Inicializado en 10000 al crear el proceso
		TiempoRafagaAnterior: 0.0,                             // Tiempo real de la última ráfaga ejecutada
		FinIO:                make(chan struct{}, 1),          // Inicializado en cero al crear el proceso
	}

	// Guardar PCB
	//globals.Procesos[pcb.PID] = *pcb // ! No se usa para nada hasta ahora
	globals.ColaNew.Enqueue(pcb) // Agregar el PCB a la cola de NEW

	globals.Signal(globals.SemNew) // Notificar que hay un proceso en NEW

	// Log obligatorio (3/8)
	logger.CreacionDeProceso(pcb.PID) // Log de creación del proceso
	return pcb
}

// AdministrarMotivoDesalojo maneja el motivo de desalojo del proceso y realiza las acciones correspondientes
func AdministrarMotivoDesalojo(PCB *structs.PCB, devolucion structs.DevolucionProceso) {

	switch devolucion.MotivoDeDesalojo {

	case "EXIT": // replanificacion (se encarga el planificador de largo plazo)
		SyscallExit(PCB) // Llama a la función SyscallExit para manejar la solicitud de salida
	case "IO": // replanificacion (se encarga el planificador de largo plazo)
		nombreIO := devolucion.ParametrosDesalojos[0]                    // Nombre del IO solicitado
		tiempoIO, err := strconv.Atoi(devolucion.ParametrosDesalojos[1]) // Tiempo de espera del IO solicitado
		if err != nil {
			log.Println("Error en la conversión del tiempo de IO:", err)
		}

		SyscallIO(PCB, nombreIO, tiempoIO) // Llama a la función SyscallIO para manejar la solicitud de IO

	case "DUMP_MEMORY":

		PCB.CambiarEstado(globals.BLOCKED)
		globals.ColaBlocked.Enqueue(PCB)
		globals.Signal(globals.SemBlocked) // Notificar que hay procesos en Blocked

		liberarCPUPorPID(PCB.PID)

		exito := request.DumpMemory(PCB.PID)
		if !exito {
			slog.Error(fmt.Sprintf("Error al realizar el dump de memoria para el proceso con PID: %d", PCB.PID))
			PCB.CambiarEstado(globals.EXIT)
			globals.ColaExit.Enqueue(PCB)
			globals.Signal(globals.SemExit) // Notificar que hay procesos en exit
		} else {
			PCB.CambiarEstado(globals.READY)
			globals.ColaReady.Enqueue(PCB)
			globals.Signal(globals.SemReady) // Notificar que hay procesos en exit
			slog.Debug("Dumpeo la memoria")
		}
		//libero la cpu para que los demas procesos puedan seguir ejecutando

	case "INIT_PROC":

		slog.Debug("Iniciando un nuevo proceso")
		instrucciones := devolucion.ParametrosDesalojos[0]
		tamanio, err := strconv.Atoi(devolucion.ParametrosDesalojos[1])
		if err != nil {
			log.Println("Error en la conversión del tamaño:", err)
		}

		CrearPCB(instrucciones, tamanio) // Crea un nuevo PCB con los parámetros recibidos

		// Encola de nuevo el PCB en la cola de exec
		globals.ColaExec.Enqueue(PCB)

	case "DESALOJO":
		liberarCPUPorPID(PCB.PID) // Libera la CPU del proceso actual
		slog.Debug("Desalojo de CPU")

		/*
			La syscall DESALOJO lo vuelve a encolar en ready
			Luego da la señal al planificador de corto plazo para el proceso que estaba esperando para ser replanificado
		*/

		PCB.CambiarEstado(globals.READY) // Cambia el estado del PCB a READY
		globals.ColaReady.Enqueue(PCB)   // Agrega el PCB a la cola de ready
		globals.Signal(globals.SemDesalojar)

	default:
		slog.Debug("No se entendió la interrupción")

	}
}

// LeerArgumentos lee los argumentos de la línea de comandos y devuelve el nombre del archivo y el tamaño del proceso
func LeerArgumentos() (string, int) {
	// validar que se pasen los argumentos necesarios
	if len(os.Args) < 3 {
		log.Fatalf("Faltan argumentos: ./kernel [archivo_pseudocodigo] [tamanio_proceso]")
	}

	// Leemos los argumentos de entrada
	nombreArchivo := os.Args[1]
	tamanio, _ := strconv.Atoi(os.Args[2])

	return nombreArchivo, tamanio
}

// BuscarIOSolicitado busca una IO conectada por su nombre y devuelve la instancia de IO conectada
func BuscarIOSolicitado(IOaBuscar string) *structs.IOConectada {
	IO := globals.IOsConectados.Get(IOaBuscar)
	return IO
}

func liberarCPUPorPID(PID int) {
	idCpu, exist := globals.ProcesosEnCPU.GetAndRemove(PID) // Remueve el proceso de la CPU
	// La syscall EXIT saca el proceso de la cola de exec
	if !exist {
		log.Printf("No se pudo encontrar la CPU para el proceso %d", PID)
	}
	cpu := globals.CpusConectadas.Get(idCpu)
	cpu.CambiarEstado(true)
	globals.Product(globals.HayCPUsLibres, cpu) // da señal al planificador de corto plazo que hay una CPU libre
}
