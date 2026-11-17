package structs

import (
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

// *=====================================| KERNEL |========================================================\\
type KernelConfig struct {
	Ip_memory               string  `json:"ip_memory"`
	Port_memory             int     `json:"port_memory"`
	Scheduler_algorithm     string  `json:"scheduler_algorithm"`
	Ready_ingress_algorithm string  `json:"ready_ingress_algorithm"`
	Alpha                   float64 `json:"alpha"`
	Suspension_time         int     `json:"suspension_time"`
	Log_level               string  `json:"log_level"`
	Port_kernel             int     `json:"port_kernel"`
	Ip_kernel               string  `json:"ip_kernel"`
	Initial_estimate        float64 `json:"initial_estimate"`
}

type PCB struct {
	PID                  int
	PC                   int
	Tamanio              int
	Instrucciones        string
	Estado               string
	ME                   map[string]int     // Cantidad de veces por estado
	MT                   map[string]float64 // Tiempo total por estado
	TInicioEstado        time.Time          // Para calcular cuánto estuvo en un estado
	EstimacionActual     float64            // Inicializado en 1.0 al crear el proceso
	TiempoRafagaAnterior float64            // Tiempo real de la última ráfaga ejecutada
	FinIO                chan struct{}      // Canal para indicar el fin de IO (uso canal para evitar tener que sincronizarlo despues con un mutex)
	mtx                  sync.Mutex         // Mutex para sincronizar el acceso al PCB
}

func (pcb *PCB) CambiarEstado(nuevoEstado string) {
	pcb.mtx.Lock()
	defer pcb.mtx.Unlock()
	// Actualizo el tiempo en las metricas de estados MT
	tiempoTranscurrido := time.Since(pcb.TInicioEstado).Milliseconds() // Calcula el tiempo transcurrido en el estado actual en milisegundos
	pcb.MT[pcb.Estado] += float64(tiempoTranscurrido)

	// Log obligatorio
	logger.CambioDeEstado(pcb.PID, string(pcb.Estado), string(nuevoEstado))

	// Actualiza el estado y reinicia TInicioEstado
	pcb.Estado = nuevoEstado

	// Actualizo la cantidad de veces que estuvo en el estado actual
	pcb.ME[pcb.Estado]++
	pcb.TInicioEstado = time.Now()
}

func (pcb *PCB) CalcularEstimacion(alpha float64) float64 {
	pcb.mtx.Lock()
	defer pcb.mtx.Unlock()
	pcb.TiempoRafagaAnterior = float64(time.Since(pcb.TInicioEstado).Milliseconds()) // Calcula el tiempo transcurrido en el estado actual en milisegundos
	pcb.EstimacionActual = alpha*pcb.TiempoRafagaAnterior + (1-alpha)*pcb.EstimacionActual
	return pcb.EstimacionActual
}

func (pcb *PCB) ActualizarPC(PC int) {
	pcb.mtx.Lock()
	defer pcb.mtx.Unlock()
	// Actualiza el PC del PCB
	pcb.PC = PC
	slog.Debug(fmt.Sprintf("Actualizando PC del proceso %d a %d\n", pcb.PID, pcb.PC))
}

func (pcb *PCB) TiempoEnEstadoActual() int64 {
	//pcb.mtx.Lock()
	//defer pcb.mtx.Unlock()
	// Calcula el tiempo que el PCB ha estado en su estado actual
	return time.Since(pcb.TInicioEstado).Milliseconds()
}

func (pcb *PCB) TiempoRestanteEnCPU() int64 {
	pcb.mtx.Lock()
	defer pcb.mtx.Unlock()
	// Calcula el tiempo restante en la CPU
	// Asumiendo que el tiempo restante es igual a la estimación actual
	return int64(pcb.EstimacionActual) - pcb.TiempoEnEstadoActual()
}

func (pcb *PCB) FinIo() bool {
	pcb.mtx.Lock()
	defer pcb.mtx.Unlock()
	select {
	case <-pcb.FinIO:
		return true
	default:
		return false
	}
}

type RequestMemoria struct {
	PID          int
	Size         int
	Instructions string
}

// ----------------------( LISTA )----------------------\\
type ListaPCBSegura struct {
	Mx   sync.Mutex
	List []*PCB
}

func (sList *ListaPCBSegura) IsEmpty() bool {
	sList.Mx.Lock()
	defer sList.Mx.Unlock()
	return len(sList.List) == 0
}

func (sList *ListaPCBSegura) Dequeue() *PCB {
	sList.Mx.Lock()
	var pcb = sList.List[0]
	sList.List = sList.List[1:]
	sList.Mx.Unlock()

	return pcb
}
func (sList *ListaPCBSegura) Enqueue(value *PCB) {
	sList.Mx.Lock()
	sList.List = append(sList.List, value)
	sList.Mx.Unlock()
}

func (sList *ListaPCBSegura) InsertFirst(value *PCB) {
	sList.Mx.Lock()
	sList.List = append([]*PCB{value}, sList.List...)
	sList.Mx.Unlock()
}

func (sList *ListaPCBSegura) GetFirst() *PCB {
	sList.Mx.Lock()
	var pcb = sList.List[0]
	sList.Mx.Unlock()
	return pcb
}

func (sList *ListaPCBSegura) GetLast() *PCB {
	sList.Mx.Lock()
	var pcb = sList.List[len(sList.List)-1]
	sList.Mx.Unlock()
	return pcb
}

func (sList *ListaPCBSegura) OrdenarPorTamanioAscendente() {
	sList.Mx.Lock()
	defer sList.Mx.Unlock()

	sort.Slice(sList.List, func(i, j int) bool {
		return sList.List[i].Tamanio < sList.List[j].Tamanio
	})

}

func (sList *ListaPCBSegura) OrdenarPorEstimacionActual() {
	sList.Mx.Lock()
	defer sList.Mx.Unlock()

	sort.Slice(sList.List, func(i, j int) bool {
		return sList.List[i].EstimacionActual < sList.List[j].EstimacionActual
	})
}

func (sList *ListaPCBSegura) FindByPID(pcbID int) (*PCB, bool) {
	sList.Mx.Lock()
	defer sList.Mx.Unlock()

	for _, pcb := range sList.List {
		if pcb.PID == pcbID {
			return pcb, true
		}
	}

	return &PCB{}, false
}

func (sList *ListaPCBSegura) GetAndRemoveByPID(pid int) *PCB {
	sList.Mx.Lock()
	defer sList.Mx.Unlock()

	for i, pcb := range sList.List {
		if pcb.PID == pid {
			// Eliminar de la lista y devolver el PCB
			sList.List = append(sList.List[:i], sList.List[i+1:]...)
			return pcb
		}
	}
	return nil
}

func (sList *ListaPCBSegura) GetPCBConTiempoRestanteMasAlto() (*PCB, bool) {
	sList.Mx.Lock()
	defer sList.Mx.Unlock()
	if len(sList.List) == 0 {
		return nil, false // No hay PCBs en la lista
	}
	maxPCB := sList.List[0]
	for _, pcb := range sList.List {
		if pcb.TiempoRestanteEnCPU() > maxPCB.TiempoRestanteEnCPU() {
			maxPCB = pcb
		}
	}
	return maxPCB, true // Retorna el PCB con el tiempo restante más alto
}

// PCBWaitingIO es una estructura que representa un PCB que esta esperando por una IO
type PCBWaitingIO struct {
	PCB    *PCB
	Tiempo int
}

// TODO cambiar a listas seguras
type IOConectada struct {
	Nombre          string
	InstanciasIO    []*InstanciaIO  // lista de instancias de IO por nombre
	ColaPCBsWaiting []*PCBWaitingIO // lista de PCBs que estan esperando por este IO
	Mx              sync.Mutex      // Mutex para sincronizar el acceso a la lista de instancias IO
}

func (io *IOConectada) AgregarInstancia(instancia *InstanciaIO) {
	io.Mx.Lock()
	defer io.Mx.Unlock()
	io.InstanciasIO = append(io.InstanciasIO, instancia)
}

func (io *IOConectada) GetInstanciaIOLibre() *InstanciaIO {
	io.Mx.Lock()
	defer io.Mx.Unlock()
	for _, instancia := range io.InstanciasIO {
		if instancia.Libre {
			return instancia
		}
	}
	return nil
}

func (io *IOConectada) AgregarProcesoAColaDeEspera(pcb *PCB, tiempo int) {
	io.Mx.Lock()
	defer io.Mx.Unlock()
	// Agrega un PCB a la cola de espera del IO
	io.ColaPCBsWaiting = append(io.ColaPCBsWaiting, &PCBWaitingIO{PCB: pcb, Tiempo: tiempo})
	slog.Debug(fmt.Sprintf("Proceso %d agregado a la cola de espera del IO %s por %d milisegundos\n", pcb.PID, io.Nombre, tiempo))
}

func (io *IOConectada) FindInstanciaByPID(pid int) *InstanciaIO {
	io.Mx.Lock()
	defer io.Mx.Unlock()
	// Busca una instancia de IO que tenga el PID en ejecucion
	for _, instancia := range io.InstanciasIO {
		if instancia.PIDEnEjecucion != nil && *instancia.PIDEnEjecucion == pid {
			return instancia
		}
	}
	return nil
}

func (io *IOConectada) GetAndRemoveInstanciaByPID(pid int) *InstanciaIO {
	// Busca y elimina una instancia de IO que tenga el PID en ejecucion
	io.Mx.Lock()
	defer io.Mx.Unlock()
	for i, instancia := range io.InstanciasIO {
		if instancia.PIDEnEjecucion != nil && *instancia.PIDEnEjecucion == pid {
			io.InstanciasIO = append(io.InstanciasIO[:i], io.InstanciasIO[i+1:]...)
			return instancia
		}
	}
	return nil
}

func (io *IOConectada) HayProcesosEnColaDeEspera() bool {
	io.Mx.Lock()
	defer io.Mx.Unlock()
	return len(io.ColaPCBsWaiting) > 0
}

func (io *IOConectada) RemovePCBEnEspera() *PCBWaitingIO {
	io.Mx.Lock()
	defer io.Mx.Unlock()
	if len(io.ColaPCBsWaiting) == 0 {
		return nil // No hay PCBs en espera
	}
	pcbWaiting := io.ColaPCBsWaiting[0]
	io.ColaPCBsWaiting = io.ColaPCBsWaiting[1:] // Elimina el primer PCB de la cola
	slog.Debug(fmt.Sprintf("PCB %d removido de la cola de espera del IO %s\n", pcbWaiting.PCB.PID, io.Nombre))
	return pcbWaiting

}

type CPU struct {
	ID     int
	IP     string
	Puerto int
	Libre  bool
}

func (cpu *CPU) CambiarEstado(nuevoEstado bool) {
	cpu.Libre = nuevoEstado
	if nuevoEstado {
		slog.Debug(fmt.Sprintf("CPU (ID: %d) LIBRE\n", cpu.ID))
	} else {
		slog.Debug(fmt.Sprintf("CPU (ID: %d) OCUPADA\n", cpu.ID))
	}
}

type HandshakeRequestCPU struct {
	ID     int    `json:"id"`
	IP     string `json:"ip"`
	Puerto int    `json:"puerto"`
}
type RequestDispatch struct {
	PID int `json:"pid"`
	PC  int `json:"pc"`
}

type DevolucionProceso struct {
	PID                 int      `json:"pid"`
	PC                  int      `json:"pc"`
	MotivoDeDesalojo    string   `json:"motivo_desalojo"`
	ParametrosDesalojos []string `json:"parametros_desalojo"`
}

// MOTIVOS DE DESALOJO DE PROCESO        KERNEL-CPU
const (
	TERMINADO  string = "TERMINADO" // Fran gil
	BLOQUEADO  string = "BLOQUEADO"
	DESALOJADO string = "DESALOJADO"
	// SUSPENDIDO string = "SUSPENDIDO"
)

// *=====================================| MEMORIA |========================================================\\
type MemoriaConfig struct {
	PortMemory     int    `json:"port_memory"`
	IPMemory       string `json:"ip_memory"`
	MemorySize     int    `json:"memory_size"`
	PageSize       int    `json:"page_size"`
	EntriesPerPage int    `json:"entries_per_page"`
	NumberOfLevels int    `json:"number_of_levels"`
	MemoryDelay    int    `json:"memory_delay"`
	SwapfilePath   string `json:"swapfile_path"`
	SwapDelay      int    `json:"swap_delay"`
	LogLevel       string `json:"log_level"`
	DumpPath       string `json:"dump_path"`
	ScriptsPath    string `json:"scripts_path"`
}

type RespuestaEspacioMemoria struct {
	OK bool `json:"ok"`
}

type ProcesoKernelMemoria struct {
	PID  int    `json:"pid"`
	Size int    `json:"size"`
	Path string `json:"path"`
}

type Instruccion struct {
	Operacion  string   `json:"operacion"`
	Parametros []string `json:"parametros"`
}

type ProcesoMemoria struct {
	PID           int                 `json:"pid"`
	Instrucciones map[int]Instruccion `json:"instructions"`
}

type AccesoEspacioUsuario struct {
	Operacion  string   `json:"operacion"`
	Parametros []string `json:"parametros"`
	PID        int      `json:"pid"`
}

type HandshakeMemoria struct {
	TamanioPagina     int `json:"tamanio_pagina"`
	CantEntradasTabla int `json:"cant_Entradas_Tabla"`
	NumeroNiveles     int `json:"numero_niveles"`
}

type MetricasProceso struct {
	PID                   int
	AccesosTablaPaginas   int
	AccesosMemoriaUsuario int
	SwapIns               int
	SwapOuts              int
	Lecturas              int
	Escrituras            int
}

type EntradaTablaNivel struct {
	SiguienteNivel *TablaNivel
	Marco          int  // base del marco en memoria de usuario
	Presente       bool // si la entrada está presente en memoria
}

type TablaNivel struct {
	Entradas []*EntradaTablaNivel
}

type ProcesoEnMemoriaDeUsuario struct {
	PID          int
	TablaPaginas *TablaNivel
}

type MarcosXProceso struct {
	PID             int
	MarcosAsignados []int
}

type SolicitudFrameCompleta struct {
	PID   int `json:"pid"`
	Frame int `json:"frame"`
}

type ActualizarFrameCompletaRequest struct {
	PID   int    `json:"pid"`
	Frame int    `json:"frame"` // Frame de la pagina a actualizar
	Datos []byte `json:"datos"` // Datos a escribir en la pagina
}

type LeerMemoriaRequest struct {
	PID             int `json:"pid"`
	PhysicalAddress int `json:"physical_address"` // Direccion fisica a leer
	Size            int `json:"size"`             // Tamanio de los datos a leer
}

type EscribirMemoriaRequest struct {
	PID             int    `json:"pid"`
	PhysicalAddress int    `json:"physical_address"` // Direccion fisica a leer
	Datos           []byte `json:"datos"`            // Datos a escribir en la pagina
}

type EnvioEntradasMultinivel struct {
	PID      int   `json:"pid"`
	Entradas []int `json:"entradas"` // Entradas de la tabla multinivel
}

type RespuestaEnvioEntradasMultinivel struct {
	Frame int `json:"frame"` // Frame donde se guardaron las entradas
}

// *=====================================| I/O |========================================================\\

// Instancia de InstanciaIO que se conecta al Kernel
type InstanciaIO struct { // enrealidad ya es un paquete, no un mensaje
	Nombre         string `json:"nombre"`
	Ip             string `json:"ip"`
	Port           int    `json:"port"`
	PIDEnEjecucion *int   // puntero a int para poder ser nil si no hay un proceso en ejecucion
	Libre          bool
	mtx            sync.Mutex // Mutex para sincronizar el acceso a la instancia IO
}
type HandshakeIOKernel struct {
	Nombre string `json:"nombre"`
	Ip     string `json:"ip"`
	Port   int    `json:"port"`
}

// CambiarEstado cambia el estado de la instancia IO a libre u ocupada
// y loggea el cambio de estado
func (instancia *InstanciaIO) CambiarEstado(nuevoEstado bool) {
	instancia.mtx.Lock()
	defer instancia.mtx.Unlock()
	instancia.Libre = nuevoEstado
	if nuevoEstado {
		slog.Debug(fmt.Sprintf("Instancia IO (Nombre: %s) LIBRE\n", instancia.Nombre))
		instancia.PIDEnEjecucion = nil // Limpia el PID en ejecucion al liberar la IO
	} else {
		slog.Debug(fmt.Sprintf("Instancia IO (Nombre: %s) OCUPADA\n", instancia.Nombre))
	}
}

func (instancia *InstanciaIO) AsignarPID(pid int) {
	instancia.mtx.Lock()
	defer instancia.mtx.Unlock()

	if instancia.PIDEnEjecucion == nil {
		instancia.PIDEnEjecucion = &pid // Asigna el PID del proceso a la IO
		slog.Debug(fmt.Sprintf("Instancia IO (Nombre: %s) asigna PID %d\n", instancia.Nombre, pid))
	} else {
		slog.Debug(fmt.Sprintf("Instancia IO (Nombre: %s) ya tiene un PID en ejecucion: %d\n", instancia.Nombre, *instancia.PIDEnEjecucion))
	}
}

type DesconexionOFinDeIO struct {
	NombreIO string `json:"nombre_io"`
	PID      int    `json:"pid"` // PID del proceso que estaba ejecutando en la IO
}

//*=====================================| CPU |========================================================\\

type ProcesoEnEjecucion struct {
	PID int `json:"pid"`
	PC  int `json:"pc"`
}

type RegistroCpu struct {
	IP            string `json:"ip"`
	Puerto        int    `json:"puerto"`
	Identificador int    `json:"identificador"`
}

//*=====================================| Listas con Sincronizacion |========================================================\\

type ListaSegura[T any] struct {
	Mx   sync.Mutex
	List []T
}

// *=====================================| HTTP REQUEST-RESPONSE |========================================================\\

type Init_Process struct {
	Instruccions string `json:"instruccions"`
	Tamanio      int    `json:"tamanio"`
}
