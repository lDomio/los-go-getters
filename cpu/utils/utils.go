package utils

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/sisoputnfrba/tp-golang/cpu/estructurasCPU"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/request"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"

	//"github.com/sisoputnfrba/tp-golang/utils/httpUtils"

	"strconv"
)

type MensajeCPU struct {
	Identificador string `json:"mensaje"`
	IP            string `json:"ip"`
	Puerto        int    `json:"puerto"`
}

type MensajeSolicitarInstruccion struct {
}

var MMU *estructurasCPU.MMU
var TLB *estructurasCPU.TLB
var Cache *estructurasCPU.PageCache

//  -----------------------INSTRUCCIONES ----------------

func NOOP() {
	slog.Debug("Ejecutando NOOP...")
	

}

// WRITE: Escribe datos en memoria
func WRITE(address string, data string) {
	addrInt, err := strconv.Atoi(address)
	if err != nil {
		slog.Debug("Error al convertir la dirección: %v", err)
		return
	}

	hit := MMU.EscribirPaginaEnCache(addrInt, []byte(data))
	if hit { // Simula el delay de la cache
		slog.Debug("Cache hit: Escrito %s en la dirección %d", data, addrInt)
		return
	} //Primero intenta leer desde la cache

	frame, estaEnlaTLB := MMU.EstaEnLaTLB(addrInt)
	if estaEnlaTLB {
		if Cache.Size > 0 { //Si esta en la TLB y hay cache
			paginaCompleta, err := request.RequestPaginaCompleta(frame)
			if err != nil {
				slog.Debug("Error al leer la pagina completa: %v", err)
				return
			}
			MMU.AgregarPaginaCache(addrInt, paginaCompleta)

			slog.Debug("TLB hit: La pagina %d ya estaba en la TLB", frame)
			hit := MMU.EscribirPaginaEnCache(addrInt, []byte(data)) // Intenta escribir en la cache
			if !hit {
				log.Panic("TLB hit pero no se pudo escribir en la cache")
				return
			}
		} else { // Si no hay cache
			physicalAddress := MMU.TraducirDireccionLogica(frame, addrInt)
			err := request.ActualizarMemoriaRequest(physicalAddress, []byte(data))
			if err != nil {
				slog.Debug("Error al escribir en memoria: %v", err)
				return
			}
			slog.Debug("TLB hit: Escrito %s en la dirección %d", data, addrInt)
			return
		}

	} else { // Si no esta en la TLB
		slog.Debug("TLB miss: No se encontró la dirección %d en la TLB", addrInt)
		entradasMultiNivel := MMU.CalculateFrameFromMultiLevelTables(addrInt)
		frame, err := request.ObtenerFramePorEntradasMultinivel(entradasMultiNivel)
		if err != nil {
			slog.Debug("Error al leer memoria multinivel: %v", err)
			return
		}
		MMU.AgregarPaginaTLB(addrInt, frame) // Agrega la pagina a la TLB

		if Cache.Size > 0 {

			slog.Debug("Cache miss y TLB miss")
			paginaCompleta, err := request.RequestPaginaCompleta(frame) // Obtener pagina desde memoria
			if err != nil {
				slog.Debug("Error al leer la pagina completa: %v", err)
				return
			}
			MMU.AgregarPaginaCache(addrInt, paginaCompleta)  // Agrega la pagina a la cache
			MMU.EscribirPaginaEnCache(addrInt, []byte(data)) // Vuelve a intentar escribir en la cache

		} else { // Si no hay cache
			physicalAddress := MMU.TraducirDireccionLogica(frame, addrInt)
			err := request.ActualizarMemoriaRequest(physicalAddress, []byte(data))
			if err != nil {
				slog.Debug("Error al escribir en memoria: %v", err)
				return
			}
			slog.Debug("TLB hit: Escrito %s en la dirección %d", data, addrInt)
			return
		}

	}

}

// READ: Lee datos de memoria
func READ(address string, size string) {
	addrInt, err := strconv.Atoi(address)
	if err != nil {
		slog.Debug("Error al convertir la dirección: %v", err)
		return
	}
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		slog.Debug("Error al convertir el tamaño: %v", err)
		return
	}

	data, hit := MMU.LeerPaginaEnCache(addrInt, sizeInt)
	if hit {
		slog.Debug("Cache hit: Leído %s de la dirección %d", data, addrInt)
		return
	} //Primero intenta leer desde la cache

	frame, estaEnlaTLB := MMU.EstaEnLaTLB(addrInt)
	if estaEnlaTLB { // Log de TLB hit
		if Cache.Size > 0 {
			paginaCompleta, err := request.RequestPaginaCompleta(frame)
			if err != nil {
				slog.Debug("Error al leer la pagina completa: %v", err)
				return
			}
			MMU.AgregarPaginaCache(addrInt, paginaCompleta)

			data, hit := MMU.LeerPaginaEnCache(addrInt, sizeInt) // Intenta leer desde la cache
			if hit {
				slog.Debug("TLB hit y cache miss: Leído %s de la dirección %d", data, addrInt)
				return
			} else {
				slog.Debug("TLB hit pero no se pudo leer")
			}
		} else { // Si no hay cache
			physicalAddress := MMU.TraducirDireccionLogica(frame, addrInt)
			data, err := request.LeerMemoriaRequest(physicalAddress, sizeInt)
			if err != nil {
				slog.Debug("Error al leer de memoria: %v", err)
				return
			}
			slog.Debug("TLB hit: Leído %s de la dirección %d", data, addrInt)
			return
		}

	} else { // Si no esta en la TLB
		slog.Debug("Cache miss y TLB miss: No se encontró la dirección %d en la TLB", addrInt)
		entradasMultiNivel := MMU.CalculateFrameFromMultiLevelTables(addrInt)
		frame, err := request.ObtenerFramePorEntradasMultinivel(entradasMultiNivel)
		if err != nil {
			slog.Debug("Error al leer memoria multinivel: %v", err)
			return
		}
		MMU.AgregarPaginaTLB(addrInt, frame) // Agrega la pagina a la TLB si hay espacio
		if Cache.Size > 0 {                  // Si hay cache
			paginaCompleta, err := request.RequestPaginaCompleta(frame)
			if err != nil {
				slog.Debug("Error al leer la pagina completa: %v", err)
				return
			}
			MMU.AgregarPaginaCache(addrInt, paginaCompleta)      // Agrega la pagina a la cache
			data, hit := MMU.LeerPaginaEnCache(addrInt, sizeInt) // Vuelve a intentar leer desde la cache
			if hit {
				slog.Debug("Cache miss pero TLB hit: Leído %s de la dirección %d", data, addrInt)
				return
			} else {
				slog.Debug("Cache miss y TLB hit pero no se pudo leer de la cache")
			}
		} else { // Si no hay cache
			physicalAddress := MMU.TraducirDireccionLogica(frame, addrInt)
			data, err := request.LeerMemoriaRequest(physicalAddress, sizeInt)
			if err != nil {
				slog.Debug("Error al leer de memoria: %v", err)
				return
			}
			slog.Debug("TLB miss: Leído %s de la dirección %d", data, addrInt)
		}

	}

}

// GOTO: Modifica el Program Counter (PC)
func GOTO(value string) {

	valor, err := strconv.Atoi(value)
	if err != nil {
		fmt.Println("Error converting:", err)
		return
	}

	slog.Debug("Saltando a la dirección %d desde la direccion %d", valor, globals.PCBenEjecucion.PC)

	globals.PCBenEjecucion.PC = valor
}

//  -----------------------SYSCALLS ----------------

// IO: Simula una operación de entrada/salida
func IO(instruccion string, parametros []string) {
	

	slog.Debug("Ejecutando IO en %s por %s ms\n", parametros[0], parametros[1])

	//Llamar al IO
	// ? Se queda esperando la respuesta ?

	MMU.Clear()
	
	request.ReturnProcess("IO", parametros, globals.PCBenEjecucion.PC, globals.PCBenEjecucion.PID )

}

// INIT_PROC: Simula la inicialización de un proceso
func INIT_PROC(filename string, size string) {
	
	request.ReturnProcess("INIT_PROC", []string{filename, size}, globals.PCBenEjecucion.PC, globals.PCBenEjecucion.PID)
	
	slog.Debug("Inicializando proceso desde %s con tamaño %s\n", filename, size)

}

// DUMP_MEMORY: Muestra el contenido de la memoria
func DUMP_MEMORY() {

	MMU.Clear()

	request.ReturnProcess("DUMP_MEMORY", nil, globals.PCBenEjecucion.PC, globals.PCBenEjecucion.PID )
	log.Println("Se pidio dumpear la memoria")
}

// EXIT: Finaliza la ejecución
func EXIT() {
	MMU.Clear()
	
	request.ReturnProcess("EXIT", nil,  globals.PCBenEjecucion.PC, globals.PCBenEjecucion.PID )
	slog.Info("Finalizando ejecución...")
}

// -- -----------------------FUNCIONES ----------------

func DecodeAndExecute(instruccion structs.Instruccion, cicloFinalizado *bool) {
	globals.PCBenEjecucion.PC++
	
	// Log obligatorio (2/11)
	logger.InstruccionEjecutada(globals.PCBenEjecucion.PID, instruccion.Operacion, instruccion.Parametros)
	switch instruccion.Operacion {
	case "NOOP":
		NOOP()
	case "WRITE":
		WRITE(instruccion.Parametros[0], instruccion.Parametros[1])
	case "READ":
		READ(instruccion.Parametros[0], instruccion.Parametros[1])
	case "GOTO":
		GOTO(instruccion.Parametros[0])
	case "IO":
		*cicloFinalizado = true
		IO(instruccion.Operacion, instruccion.Parametros)
	case "INIT_PROC":
		INIT_PROC(instruccion.Parametros[0], instruccion.Parametros[1])
	case "DUMP_MEMORY":
		DUMP_MEMORY()
		*cicloFinalizado = true
	case "EXIT":
		*cicloFinalizado = true
		EXIT()
	default:
		log.Println("No se entendio la operacion recibida")
	}
	
	

	
}


func CicloDeInstrucciones() {
	// Ciclo de instrucciones
	var instruccion structs.Instruccion
	for !globals.CicloFinalizado {
		log.Println("EMpezo el ciclo de instrucciones")
		instruccion, _ = request.Fetch(globals.PCBenEjecucion.PID, globals.PCBenEjecucion.PC)

		if instruccion.Operacion == "" {
			slog.Info("No hay mas instrucciones para ejecutar")
			globals.CicloFinalizado = true
			continue
		}

		DecodeAndExecute(instruccion, &globals.CicloFinalizado)

		CheckInterrupt()
	}

	globals.CicloFinalizado = false
	globals.CheckInterrupt = false


	slog.Info("La CPU esta libre para utilizarse")

}

func CheckInterrupt() {
	if globals.CheckInterrupt {
		globals.CicloFinalizado = true
		
		request.ReturnProcess("DESALOJO", nil, globals.PCBenEjecucion.PC, globals.PCBenEjecucion.PID)
	}
}

func InicializarEstructurasBasicas(pageSize int, levels int, tableEntries int) {
	Cache = estructurasCPU.NewPageCache(globals.CpuConfig.Cache_entries, globals.CpuConfig.Cache_replacement)
	TLB = estructurasCPU.NewTLB(globals.CpuConfig.Tlb_entries, globals.CpuConfig.Tlb_replacement)
	MMU = estructurasCPU.NewMMU(TLB, Cache, pageSize, tableEntries, levels)
	Cache.SetTLB(TLB)
	Cache.SetMMU(MMU)

	//Aca guardo el Cache, la TLB y la MMU en las variables globales
	log.Println("Estructuras inicializadas: TLB, Cache y MMU")
}
