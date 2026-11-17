package logger

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

func ConfigurarLogger(nombre string, nivel string) { // Habria que ver de usar slog

	parseLogLevel(nivel)
	//logFile, err := os.OpenFile(nombre, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	//logFile, err := os.Create(nombre)
	logFile, err := os.OpenFile(nombre, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	//log.Println("Soy un log")
	log.SetOutput(mw)

	// Crear el logger con slog
	// Crear el logger con slog en formato texto
	//handler := slog.NewTextHandler(mw, &slog.HandlerOptions{
	//	Level: parseLogLevel(nivel), // Configurar el nivel de log
	//})

	//slog.SetDefault(slog.New(handler)) // Establecer el logger por defecto
}

// Función auxiliar para convertir el nivel de log a slog.Level
func parseLogLevel(nivel string) {
	switch nivel {
	case "DEBUG":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "INFO":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "WARNING":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "ERROR":
		slog.SetLogLoggerLevel(slog.LevelError)
	}
}

// *=====================================| KERNEL LOG'S |========================================================\\

// Log obligatorio (1/8)
func SyscallRecibida(pid int, nombre_syscall string) {
	//fmt.SPrintf("## (<%d>) - Solicitó syscall: <%s>", pid, nombre_syscall)
	slog.Info(fmt.Sprintf("## (<%d>) - Solicitó syscall: <%s>", pid, nombre_syscall))
}

// Log obligatorio (2/8)
func CambioDeEstado(pid int, estadoAnterior string, nuevoEstado string) {
	//log.Printf("## (%d) Pasa del estado <%s> al estado <%s>", idProceso, estadoAnterior, nuevoEstado)
	slog.Info(fmt.Sprintf("## (<%d>) Pasa del estado <%s> al estado <%s>", pid, estadoAnterior, nuevoEstado))
}

// Log obligatorio (3/8)
func CreacionDeProceso(pid int) {
	//fmt.Sprintf("## (<%d>) Se crea el proceso - Estado: NEW", pid)
	slog.Info(fmt.Sprintf("## (<%d>) Se crea el proceso - Estado: NEW", pid))
}

// Log obligatorio (4/8)
func MotivoDeBloqueo(pid int, dispositivo_io string) {
	//fmt.Sprintf("## (<%d>) - Bloqueado por IO: <%s>", pid, dispositivo_io)
	slog.Info(fmt.Sprintf("## (<%d>) - Bloqueado por IO: <%s>", pid, dispositivo_io))
}

// Log obligatorio (5/8)
func FinDeIO(pid int) {
	//fmt.Sprintf("## (<%d>) finalizó IO y pasa a READY", pid)
	slog.Info(fmt.Sprintf("## (<%d>) finalizó IO y pasa a READY", pid))
}

// Log obligatorio (6/8) SJF/SRT
func Desalojo(pid int, algoritmo string) {
	//log.Printf("## (<%d>) - Desalojado por el algoritmo %s", pid, algoritmo)
	slog.Info(fmt.Sprintf("## (<%d>) - Desalojado por el algoritmo %s", pid, algoritmo))
}

// Log obligatorio (7/8)
func FinDeProceso(pid int) {
	//log.Printf("## (<%d>) - Finalizó el proceso", pid)
	slog.Info(fmt.Sprintf("## (<%d>) - Finalizó el proceso", pid))
}

// Log obligatorio (8/8)
func MetricasDeEstado(pid int, ME map[string]int, MT map[string]float64) {
	//log.Printf("## (<%d>) - Métricas de estados: %v", pid, state_count_time(ME, MT))
	slog.Info(fmt.Sprintf("## (<%d>) - Métricas de estados: %s", pid, state_count_time(ME, MT)))
}

// "NEW (NEW_COUNT) (NEW_TIME), READY (READY_COUNT) (READY_TIME), …"
func state_count_time(ME map[string]int, MT map[string]float64) string {
	result := ""
	for estado, cantidad := range ME {
		tiempo := MT[estado]
		result += estado + " (" + strconv.Itoa(cantidad) + ") (" + fmt.Sprintf("%.2f", tiempo) + "), "
	}
	if len(result) > 0 {
		result = result[:len(result)-2] // Elimina la última coma y espacio
	}
	return result
}

// *=====================================| MEMORIA LOG'S |========================================================\\
// Log obligatorio (1/5)
func MetricasProceso(
	pid int,
	accesosTablaPaginas int,
	accesosMemoriaUsuario int,
	swapIns int,
	swapOuts int,
	lecturas int,
	escrituras int,
) {
	slog.Info(fmt.Sprintf("## PID: <%d> - Proceso Destruido - Métricas - Acc.T.Pag: %d; Mem.Prin.: %d; SWAPIns: %d; SWAPOuts: %d; Lec.Mem.: %d; Esc.Mem.: %d",
		pid,
		accesosTablaPaginas,
		accesosMemoriaUsuario,
		swapIns,
		swapOuts,
		lecturas,
		escrituras,
	))
}

// Log obligatorio (2/5)
func CreacionProceso(pid int, tamanio int) {
	slog.Info(fmt.Sprintf("## PID: <%d> - Proceso Creado - Tamaño: %d", pid, tamanio))
}

// Log obligatorio (3/5)
func ObtenerInstruccion(pid int, pc int, instruccion string, args []string) {
	argsStr := ""
	if len(args) > 0 {
		argsStr = " " + strings.Join(args, " ")
	}
	slog.Info(fmt.Sprintf("## PID: <%d> - Obtener instrucción: %d - Instrucción: %s %s", pid, pc, instruccion, argsStr))
}

// Log obligatorio (4/5)
func EscrituraLecturaUsuario(pid int, operacion string, direccionFisica int, tamanio int) {
	slog.Info(fmt.Sprintf("## PID: <%d> - %s - Dir. Física: %d - Tamaño: %d", pid, operacion, direccionFisica, tamanio))
}

// Log obligatorio (5/5)
func MemoryDump(pid int) {
	slog.Info(fmt.Sprintf("## PID: <%d> - Memory Dump solicitado", pid))
}

// *=====================================| I/O LOG'S |========================================================\\
// Log obligatorio (1/2)
func InicioIO(pid int, tiempo int) {
	slog.Info(fmt.Sprintf("## PID: <%d> - Inicio de IO - Tiempo: %d", pid, tiempo))
}

// Log obligatorio (2/2)
func FinalizacionDeIO(pid int) {
	slog.Info(fmt.Sprintf("## PID: <%d> - Fin de IO", pid))
}

// *=====================================| CPU LOG'S |========================================================\\

// Log obligatorio (1/11)
func FetchInstrucciones(pid int, pc int) {
	slog.Info(fmt.Sprintf("## PID: <%d> - FETCH - Program Counter: %d", pid, pc))
}

// Log obligatorio (2/11)
func InstruccionEjecutada(pid int, instruccion string, parametros []string) {
	slog.Info(fmt.Sprintf("## PID: <%d> - Ejecutando: %s - %s", pid, instruccion, strings.Join(parametros, " ")))
}

// Log obligatorio (3/11)
func InterrupcionRecibida() {
	slog.Info("## Llega interrupción al puerto Interrup")
}

// Log obligatorio (4/11)
func LecturaMemoria(pid int, direccionFisica int, valor string) {
	slog.Info(fmt.Sprintf("PID: <%d> - LEER - Dirección Física: %d - Valor: %s", pid, direccionFisica, valor))
}

// Log obligatorio (4/11) SON MISMO LOGS PERO DIFERENCIE LAS FUNCIONES
func EscrituraMemoria(pid int, direccionFisica int, valor string) {
	slog.Info(fmt.Sprintf("PID: <%d> - ESCRIBIR - Dirección Física: %d - Valor: %s", pid, direccionFisica, valor))
}

// Log obligatorio (5/11)
func ObtenerMarcoDePagina(pid int, pagina int, marco int) {
	slog.Info(fmt.Sprintf("PID: <%d> - OBTENER MARCO - Página: %d - Marco: %d", pid, pagina, marco))
}

// Log obligatorio (6/11)
func TlbHit(pid int, pagina int) {
	slog.Info(fmt.Sprintf("PID: <%d> - TLB HIT - Página: %d", pid, pagina))
}

// Log obligatorio (7/11)
func TlbMiss(pid int, pagina int) {
	slog.Info(fmt.Sprintf("PID: <%d> - TLB MISS - Página: %d", pid, pagina))
}

// Log obligatorio (8/11)
func PaginaEnCache(pid int, pagina int) {
	slog.Info(fmt.Sprintf("PID: <%d> - Cache Hit - Página: %d", pid, pagina))
}

// Log obligatorio (9/11)
func PaginaFaltanteCache(pid int, pagina int) {
	slog.Info(fmt.Sprintf("PID: <%d> - Cache Miss - Página: %d", pid, pagina))
}

// Log obligatorio (10/11)
func PaginaIngresoCache(pid int, pagina int) {
	slog.Info(fmt.Sprintf("PID: <%d> - Cache Add - Página: %d", pid, pagina))
}

// Log obligatorio (11/11)
func PaginaDesalojoCache(pid int, pagina int, marco int) {
	slog.Info(fmt.Sprintf("PID: <%d> - Memory Update - Página: %d - Frame: %d", pid, pagina, marco))
}
