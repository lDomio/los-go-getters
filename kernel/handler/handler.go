package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/scheduler"
	"github.com/sisoputnfrba/tp-golang/kernel/utils"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// *=====================================| IO |========================================================\\

func RegistrarIO(w http.ResponseWriter, r *http.Request) {
	var io structs.InstanciaIO
	err := json.NewDecoder(r.Body).Decode(&io)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//remoteAddr := r.RemoteAddr Get IP and port
	slog.Debug(fmt.Sprintf("IP: %s \n and Port: %d\n", io.Ip, io.Port))
	slog.Info(fmt.Sprintf("me llego este IO: %s\n", io.Nombre))

	io.Libre = true // La IO está libre al momento de registrarse

	//Se guarda el IO en la variable de IOs conectados
	globals.IOsConectados.SetInstancia(io.Nombre, &io)

	respuesta, err := json.Marshal(fmt.Sprintf("Handshake de %s completado", io.Nombre))
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}
	w.Write(respuesta)

	utils.AtenderProcesosEncoladosEnIO(globals.IOsConectados.Get(io.Nombre)) // Atender procesos encolados en la IO
}

// TODO revisar si funciona
func FinDeIO(w http.ResponseWriter, r *http.Request) {
	var finIO structs.DesconexionOFinDeIO

	// Verifica que el método sea POST
	if r.Method != http.MethodPost {
		http.Error(w, "Solo POST", http.StatusMethodNotAllowed)
		return
	}
	// Decodifica en formato JSON la request.
	err := json.NewDecoder(r.Body).Decode(&finIO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	instanciaIO := globals.IOsConectados.Get(finIO.NombreIO).FindInstanciaByPID(finIO.PID) // Buscar la instancia de IO por PID
	if instanciaIO == nil {
		http.Error(w, fmt.Sprintf("Instancia de IO con PID %d no encontrada",
			finIO.PID), http.StatusNotFound)
		return
	}

	instanciaIO.CambiarEstado(globals.IO_FREE)
	// Cambiar el estado de la instancia de IO a libre

	pcb := globals.ColaBlocked.GetAndRemoveByPID(finIO.PID) // Remover el PCB de la cola de blocked
	if pcb != nil && pcb.Estado == globals.BLOCKED {        // !
		// Si el PCB está en BLOCKED, aviso que la IO terminó y cambio su estado a READY(en el planif MP)

		pcb.FinIO <- struct{}{}

		pcb.CambiarEstado(globals.READY) // Cambiar el estado del PCB a READY
		globals.ColaReady.Enqueue(pcb)   // Encolar el PCB en la cola
		globals.Signal(globals.SemReady) // Notificar que hay procesos en READY
		logger.FinDeIO(pcb.PID)

	} else {
		// Si no está en BLOCKED, puede estar en SUSPENDED_BLOCKED
		pcb = globals.ColaSuspendedBlocked.GetAndRemoveByPID(finIO.PID)
		if pcb != nil {
			scheduler.FinIOEnSuspBlocked(pcb)
		} else {
			slog.Debug(fmt.Sprintf("No se encontró el PCB con PID %d en BLOCKED ni en SUSPENDED_BLOCKED", finIO.PID))
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Fin de IO recibida correctamente"))

	/* verificar si hay proceso encolados para dicha IO y enviarlo a la misma. */
	utils.AtenderProcesosEncoladosEnIO(globals.IOsConectados.Get(finIO.NombreIO)) // Atender procesos encolados en la IO
}

func DesconexionIO(w http.ResponseWriter, r *http.Request) {
	var IOFinalizado structs.DesconexionOFinDeIO

	// Verifica que el método sea POST
	if r.Method != http.MethodPost {
		http.Error(w, "Solo POST", http.StatusMethodNotAllowed)
		return
	}

	// Decodifica en formato JSON la request.
	err := json.NewDecoder(r.Body).Decode(&IOFinalizado)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	/*
		En caso de que el mensaje corresponda a una desconexión de la IO, el proceso que estaba ejecutando en dicha IO, se deberá pasar al estado EXIT. Si hay procesos encolados y no quedan más instancias disponibles, también se deberá pasar a EXIT.
		1.Buscar la IO en la lista de IOs conectados y finalizarla.
		2. Buscar el PCB asociado al PID recibido.
		3. Cambiar el estado del PCB y pasarlo a EXIT.
		4. Si hay procesos encolados, se deberá pasar a EXIT también.
		5. Enviar una señal al planificador de largo plazo para que procese los cambios.
	*/
	// 1. Busco la IO conectada
	ioConectada := globals.IOsConectados.Get(IOFinalizado.NombreIO)
	if ioConectada == nil {
		http.Error(w, fmt.Sprintf("IO %s no encontrada", IOFinalizado.NombreIO), http.StatusNotFound)
		return
	}

	// Busco la instancia de IO por PID y la remuevo de la lista de instancias conectadas
	// * a partir de aca, al remover la instancia el garbage collector se encargará de liberar la memoria
	// TODO revisar que pasa si la io que se desconecta no estaba ejecutando ningun proceso
	ioConectada.GetAndRemoveInstanciaByPID(IOFinalizado.PID)

	// 2. Buscar el PCB asociado al PID recibido.
	PCB := globals.ColaBlocked.GetAndRemoveByPID(IOFinalizado.PID) // Remover el PCB de la cola de blocked
	if PCB != nil {
		// 3. Cambiar el estado del PCB y pasarlo a EXIT.
		PCB.CambiarEstado(globals.EXIT)
		globals.ColaExit.Enqueue(PCB)   // Encolar el PCB en la cola de exit
		globals.Signal(globals.SemExit) // Notificar que hay procesos en exit
	}

	// 4. Si hay procesos encolados, se deberá pasar a EXIT también.
	utils.EnviarProcesosEnEsperaIOAExit(ioConectada.ColaPCBsWaiting)

	slog.Info(fmt.Sprintf("Desconexion de IO: %s\n", IOFinalizado.NombreIO))

	// Buscar el PCB y actualizar su estado, desencolar, etc.
	// pcb := globals.BuscarPCB(finIO.PID)
	// utils.ManejarFinDeIO(pcb, finIO.NombreIO, finIO.IP, finIO.Puerto)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Desconexion de IO recibida correctamente"))
}

// *=====================================| CPU |========================================================\\

// Atiende el handshake de la CPU y la registra en la lista de CPUs conectadas
func RegistrarCpu(w http.ResponseWriter, r *http.Request) {
	var request structs.HandshakeRequestCPU

	// Verifica que el método sea POST
	if r.Method != http.MethodPost {
		http.Error(w, "Solo POST", http.StatusMethodNotAllowed)
		return
	}

	// Decodifica en formato JSON la request.
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("Error al decodificar el JSON:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cpu := &structs.CPU{
		ID:     request.ID,
		IP:     request.IP,
		Puerto: request.Puerto,
		Libre:  true,
	}
	// Guardar la CPU
	globals.CpusConectadas.Set(cpu.ID, cpu)
	globals.Product(globals.HayCPUsLibres, cpu) // Agrega la CPU a la lista de CPUs conectadas

	slog.Info(fmt.Sprintf("CPU registrada: ID=%d, IP=%s, Puerto=%d\n", cpu.ID, cpu.IP, cpu.Puerto))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Handshake: CPU recibido correctamente"))
}

//*=====================================| PROCESOS |========================================================\\

func DevolucionProceso(w http.ResponseWriter, r *http.Request) {
	var devolucion structs.DevolucionProceso

	// Verifica que el método sea POST
	if r.Method != http.MethodPost {
		http.Error(w, "Solo POST", http.StatusMethodNotAllowed)
		return
	}
	// Decodifica en formato JSON la request.
	err := json.NewDecoder(r.Body).Decode(&devolucion)
	if err != nil {
		log.Println("Error al decodificar el JSON:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	slog.Debug(fmt.Sprintf("Devolución de proceso: PID=%d, PC=%d, Motivo=%s\n", devolucion.PID, devolucion.PC, devolucion.MotivoDeDesalojo))

	// log obligatorio (1/8)
	logger.SyscallRecibida(devolucion.PID, devolucion.MotivoDeDesalojo)

	// Buscar el PCB en ejecutando y actualizar su PC
	pcb := globals.ColaExec.GetAndRemoveByPID(devolucion.PID)
	pcb.ActualizarPC(devolucion.PC) // Actualizar el PC del PCB con el valor recibido

	// Administrar el motivo de desalojo
	utils.AdministrarMotivoDesalojo(pcb, devolucion) // Se analiza el motivo de desalojo

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Devolución de proceso recibida correctamente"))
}
