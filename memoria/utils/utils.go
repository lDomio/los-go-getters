package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/sisoputnfrba/tp-golang/memoria/configuracionmemoria"
	configuracionMemoria "github.com/sisoputnfrba/tp-golang/memoria/configuracionmemoria"
	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	"github.com/sisoputnfrba/tp-golang/memoria/swap"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

type Mensaje struct {
	Mensaje string `json:"mensaje"`
}

// * ========================================= Memoria =========================================* \\

func InicializarEstructuras(configuracionMemoria *structs.MemoriaConfig) {
	globals.MemoriaPrincipal = make([]byte, configuracionMemoria.MemorySize)
	globals.ProcesosEnMemoriaDeSistema = make(map[int]structs.ProcesoMemoria) // [pid]DatoDelProceso (instrucciones)

	globals.CantidadDeMarcos = configuracionMemoria.MemorySize / configuracionMemoria.PageSize
	globals.BitMap = make([]bool, globals.CantidadDeMarcos)
	// Definir procesos en memoria de usuario
	globals.ProcesosEnMemoriaDeUsuario = make(map[int]structs.ProcesoEnMemoriaDeUsuario) // [pid]DatoDelProceso
	globals.MetricasXProceso = make(map[int]structs.MetricasProceso)                     // [pid]metricas
	globals.MarcosXProceso = make(map[int][]int)                                         // [pid]marcosAsignados
	swap.InicializarSwap()
}

// * ========================================= Primera entrega =========================================* \\

func RecibirCPU(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var mensaje Mensaje
	err := decoder.Decode(&mensaje)
	if err != nil {
		log.Printf("Error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	slog.Debug("Me llego un mensaje de la CPU")
	slog.Debug(fmt.Sprintf("%+v\n", mensaje))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func RecibirKernel(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var mensaje Mensaje
	err := decoder.Decode(&mensaje)
	if err != nil {
		log.Printf("Error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	slog.Debug("Me llego un mensaje del Kernel")
	slog.Debug(fmt.Sprintf("%+v\n", mensaje))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// * ========================================= Cargar proceso de Kernel =========================================* \\

func CargarProceso(proceso structs.ProcesoKernelMemoria, cantidadDePaginas int, marcos *[]int) {

	globals.ProcesosMutex.Lock()
	globals.BitMapMutex.Lock()
	defer globals.ProcesosMutex.Unlock()
	defer globals.BitMapMutex.Unlock()

	tablaRaiz := crearTablasDePaginasYMarcos(
		configuracionMemoria.MemoriaConfig.NumberOfLevels,
		configuracionMemoria.MemoriaConfig.EntriesPerPage,
		proceso,
		marcos,
	)

	globals.MetricasXProceso[proceso.PID] = structs.MetricasProceso{}

	if tablaRaiz == nil {
		fmt.Println("Error: No se pudo crear la tabla de páginas")
		return
	}

	procesoSerializadoMemoriaPrincipal := structs.ProcesoEnMemoriaDeUsuario{
		PID:          proceso.PID,
		TablaPaginas: tablaRaiz,
	}

	// Obtener la ruta absoluta
	rutaAbsoluta, err := filepath.Abs(configuracionMemoria.MemoriaConfig.ScriptsPath)
	if err != nil {
		fmt.Println("Error al obtener la ruta absoluta:", err)
		return
	}
	ruta := rutaAbsoluta + "/" + proceso.Path
	slog.Debug(fmt.Sprintf("Ruta del archivo: %s", ruta))

	contenido, err := os.ReadFile(ruta)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	lineasDePseudocodigo := strings.Split(string(contenido), "\n")
	pseudocodigoMapeado := make(map[int]structs.Instruccion)
	for i, linea := range lineasDePseudocodigo {
		partes := strings.Split(linea, " ")
		operacion := partes[0]
		parametros := partes[1:]
		instruccion := structs.Instruccion{
			Operacion:  operacion,
			Parametros: parametros,
		}
		pseudocodigoMapeado[i] = instruccion
	}
	/*globals.ProcesosEnMemoriaDeSistema = append(globals.ProcesosEnMemoriaDeSistema, structs.ProcesoMemoria{
		PID:           proceso.PID,
		Instrucciones: pseudocodigoMapeado,
	})*/
	globals.ProcesosEnMemoriaDeSistema[proceso.PID] = structs.ProcesoMemoria{
		PID:           proceso.PID,
		Instrucciones: pseudocodigoMapeado,
	}
	globals.ProcesosEnMemoriaDeUsuario[proceso.PID] = procesoSerializadoMemoriaPrincipal

	slog.Debug(fmt.Sprintf("Proceso cargado en memoria: %d", proceso.PID))
}

func CalcularCantidadDePaginas(tamanio int, config structs.MemoriaConfig) int {
	return int(math.Ceil(float64(tamanio) / float64(config.PageSize)))
}

func MarcosLibresNecesarios(cantidadDePaginas int) []int {

	globals.BitMapMutex.Lock()
	defer globals.BitMapMutex.Unlock()

	libres := 0
	var marcos []int
	for nroMarco, ocupado := range globals.BitMap {
		if !ocupado && libres < cantidadDePaginas {
			libres++
			marcos = append(marcos, nroMarco)
		}

	}
	return marcos
}

func crearTablasDePaginasYMarcos(nivel int, entradasPorNivel int, proceso structs.ProcesoKernelMemoria, marcosLibres *[]int) *structs.TablaNivel {
	if nivel == 0 {
		return nil
	}

	if nivel == 1 {
		tabla := &structs.TablaNivel{
			Entradas: make([]*structs.EntradaTablaNivel, entradasPorNivel),
		}

		for i := 0; i < entradasPorNivel; i++ {
			if len(*marcosLibres) > 0 {
				marcoLibre := getMarcoLibre(marcosLibres)
				globals.BitMap[marcoLibre] = true
				tabla.Entradas[i] = &structs.EntradaTablaNivel{
					Marco:    marcoLibre,
					Presente: true,
				}
				//globals.MetricasXProceso[proceso.PID] = structs.MetricasProceso{}
				globals.MarcosXProceso[proceso.PID] = append(globals.MarcosXProceso[proceso.PID], marcoLibre)
			} else {
				tabla.Entradas[i] = &structs.EntradaTablaNivel{
					Presente: false,
				}

			}
		}
		return tabla
	}

	tabla := &structs.TablaNivel{
		Entradas: make([]*structs.EntradaTablaNivel, entradasPorNivel),
	}

	for i := 0; i < entradasPorNivel; i++ {
		tabla.Entradas[i] = &structs.EntradaTablaNivel{}
		tabla.Entradas[i].SiguienteNivel = crearTablasDePaginasYMarcos(nivel-1, entradasPorNivel, proceso, marcosLibres)
	}

	return tabla
}

func getMarcoLibre(marcosLibres *[]int) int {
	marcoLibre := (*marcosLibres)[0]
	*marcosLibres = (*marcosLibres)[1:]
	return marcoLibre
}

// * ========================================= int to slice de int =========================================* \\

func intToSliceDeInt(numero int) []int {
	if numero < 0 {
		return nil
	}

	slice := []int{}
	for numero > 0 {
		digito := numero % 10
		slice = append(slice, digito)
		numero = numero / 10
	}

	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}

	return slice
}

// * ========================================= Calcular frame =========================================* \\

func CalcularFrame(pid int, entradas []int) int {
	if pid < 0 || len(entradas) == 0 {
		return -1
	}

	_, exists := globals.ProcesosEnMemoriaDeUsuario[pid]
	if !exists {
		return -1
	}

	tablaDeProcesoBuscado := globals.ProcesosEnMemoriaDeUsuario[pid].TablaPaginas
	proceso := globals.MetricasXProceso[pid]

	time.Sleep(time.Duration(configuracionMemoria.MemoriaConfig.MemoryDelay*len(entradas)) * time.Millisecond)

	entrada := tablaDeProcesoBuscado.Entradas[entradas[0]]
	for i := 1; i < configuracionMemoria.MemoriaConfig.NumberOfLevels; i++ {
		proceso.AccesosMemoriaUsuario++
		proceso.Lecturas++
		entrada = entrada.SiguienteNivel.Entradas[entradas[i]]
	}
	globals.MetricasXProceso[pid] = proceso

	if entrada.Presente {
		return entrada.Marco
	}

	return -1
}

// * ========================================= Swapear Proceso ========================================= * \\

func BorrarEstructurasAdministrativas(pid int) {

	globals.ProcesosMutex.Lock()
	defer globals.ProcesosMutex.Unlock()

	proceso, exists := globals.ProcesosEnMemoriaDeUsuario[pid]
	if !exists {
		slog.Debug(fmt.Sprintf("No existe un proceso con PID %d", pid))
		return
	}

	for _, entrada := range proceso.TablaPaginas.Entradas {
		liberarMarco(entrada)
	}

	delete(globals.ProcesosEnMemoriaDeUsuario, pid)
	delete(globals.MarcosXProceso, pid)
}

// * ========================================= Deswapear Proceso ========================================= * \\

func ActualizarTablaDePaginas(proceso structs.ProcesoKernelMemoria, marcosAsignar []int) {

	globals.ProcesosMutex.Lock()
	defer globals.ProcesosMutex.Unlock()

	tablaRaiz := crearTablasDePaginasYMarcos(
		configuracionMemoria.MemoriaConfig.NumberOfLevels,
		configuracionMemoria.MemoriaConfig.EntriesPerPage,
		proceso,
		&marcosAsignar,
	)

	if tablaRaiz == nil {
		fmt.Println("Error: No se pudo crear la tabla de páginas")
		return
	}

	procesoSerializadoMemoriaPrincipal := structs.ProcesoEnMemoriaDeUsuario{
		PID:          proceso.PID,
		TablaPaginas: tablaRaiz,
	}

	globals.ProcesosEnMemoriaDeUsuario[procesoSerializadoMemoriaPrincipal.PID] = procesoSerializadoMemoriaPrincipal
}

// * ========================================= Dumpear Proceso ========================================= * \\

func GenerarDumpMemoria(pid int) ([][]byte, error) {
	if _, existe := globals.ProcesosEnMemoriaDeUsuario[pid]; !existe {
		return nil, fmt.Errorf("proceso %d no encontrado", pid)
	}

	/*
		if swap.SwapManager.ElProcesoEstaEnSwap(pid) {
			contenidoDelProceso, err := swap.SwapManager.LeerDeSwap(pid)
			return contenidoDelProceso, err
		}
	*/

	// marcosDelProceso := globals.MarcosXProceso[pid]
	// time.Sleep(time.Duration(configuracionMemoria.MemoriaConfig.MemoryDelay*configuracionMemoria.MemoriaConfig.NumberOfLevels*len(marcosDelProceso)) * time.Millisecond)

	contenidoMarcosDelProceso, err := ObtenerContenidoEnMarcos(pid)
	return contenidoMarcosDelProceso, err
}

func GuardarDumpEnArchivo(pid int, dump [][]byte) error {

	if _, existe := globals.ProcesosEnMemoriaDeUsuario[pid]; !existe {
		return fmt.Errorf("proceso con PID inexistente")
	}

	var flatDump []byte
	for _, chunk := range dump {
		// el operador ... "desempaqueta" el contenido del slice chunk, es por eso que flatDump no es un [][]byte
		flatDump = append(flatDump, chunk...)
	}

	logger.MemoryDump(pid)

	timestamp := time.Now().Format("2006-01-02_15:04:05")
	filename := fmt.Sprintf("<%d>-<%s>.dmp", pid, timestamp)
	filepath := path.Join(configuracionmemoria.MemoriaConfig.DumpPath, filename)

	err := os.WriteFile(filepath, flatDump, 0644)
	if err != nil {
		return fmt.Errorf("error escribiendo el dump")
	}
	return nil
}

// * ========================================= Leer espacio usuario =========================================* \\

func LeerEspacioUsuario(pid int, direccionFisica int, tamanio int) ([]byte, error) {
	if pid < 0 || direccionFisica < 0 || tamanio <= 0 {
		return nil, fmt.Errorf("PID, direccion fisica y tamanio deben ser mayores a 0")
	}

	tamanioPagina := configuracionMemoria.MemoriaConfig.PageSize
	cantidadDeFrames := len(globals.MemoriaPrincipal) / tamanioPagina

	if direccionFisica >= cantidadDeFrames*tamanioPagina {
		return nil, fmt.Errorf("direccion fisica fuera de los limites de memoria")
	}

	if !framePertenenceAlProceso(pid, direccionFisica/tamanioPagina) {
		return nil, fmt.Errorf("el frame no pertenece al proceso con PID %d", pid)
	}

	// simulo el retardo de acceso a memoria
	time.Sleep(time.Duration(configuracionMemoria.MemoriaConfig.MemoryDelay) * time.Millisecond)

	globals.MemoriaMutex.RLock()
	contenido := make([]byte, tamanio)
	copy(contenido, globals.MemoriaPrincipal[direccionFisica:direccionFisica+tamanio])
	globals.MemoriaMutex.RUnlock()

	ActualizarMetricas(pid, direccionFisica/tamanioPagina, "lectura")

	return contenido, nil
}

// * ========================================= Escribir espacio usuario =========================================* \\

func EscribirEspacioUsuario(pid int, direccionFisica int, datos []byte) error {
	if pid < 0 || direccionFisica < 0 || len(datos) == 0 {
		return fmt.Errorf("PID, direccion fisica y datos deben ser mayores a 0")
	}

	tamanioPagina := configuracionMemoria.MemoriaConfig.PageSize
	cantidadDeFrames := len(globals.MemoriaPrincipal) / tamanioPagina

	if direccionFisica >= cantidadDeFrames*tamanioPagina {
		return fmt.Errorf("direccion fisica fuera de los limites de memoria")
	}

	if !framePertenenceAlProceso(pid, direccionFisica/tamanioPagina) {
		return fmt.Errorf("el frame no pertenece al proceso con PID %d", pid)
	}

	// simulo el retardo de acceso a memoria
	time.Sleep(time.Duration(configuracionMemoria.MemoriaConfig.MemoryDelay) * time.Millisecond)

	globals.MemoriaMutex.Lock()
	copy(globals.MemoriaPrincipal[direccionFisica:direccionFisica+len(datos)], datos)
	globals.MemoriaMutex.Unlock()

	ActualizarMetricas(pid, direccionFisica/tamanioPagina, "escritura")

	return nil
}

// * ========================================= Leer frame completo =========================================* \\

func LeerFrameCompleto(pid int, frame int) ([]byte, error) {
	if pid < 0 || frame < 0 {
		return nil, fmt.Errorf("PID y frame deben ser mayores o iguales a 0")
	}

	tamanioPagina := configuracionMemoria.MemoriaConfig.PageSize
	cantidadDeFrames := len(globals.MemoriaPrincipal) / tamanioPagina

	if !framePertenenceAlProceso(pid, frame) {
		return nil, fmt.Errorf("el frame %d no pertenece al proceso con PID %d", frame, pid)
	}

	if frame < 0 || frame >= cantidadDeFrames {
		return nil, fmt.Errorf("frame fuera de los límites de memoria")
	}

	// simulo el retardo de acceso a memoria
	time.Sleep(time.Duration(configuracionMemoria.MemoriaConfig.MemoryDelay) * time.Millisecond)

	globals.MemoriaMutex.RLock()
	contenidoFrame := make([]byte, tamanioPagina)
	copy(contenidoFrame, globals.MemoriaPrincipal[frame*tamanioPagina:(frame+1)*tamanioPagina])
	globals.MemoriaMutex.RUnlock()

	ActualizarMetricas(pid, frame, "lectura")

	return contenidoFrame, nil
}

// * ========================================= Escribir frame completo =========================================* \\

func EscribirFrameCompleto(pid int, frame int, data []byte) error {
	if pid < 0 || frame < 0 {
		return fmt.Errorf("PID y frame deben ser mayores o iguales a 0")
	}

	tamanioPagina := configuracionMemoria.MemoriaConfig.PageSize
	cantidadDeFrames := len(globals.MemoriaPrincipal) / tamanioPagina

	if !framePertenenceAlProceso(pid, frame) {
		return fmt.Errorf("el frame %d no pertenece al proceso con PID %d", frame, pid)
	}

	if frame < 0 || frame >= cantidadDeFrames {
		return fmt.Errorf("frame fuera de los límites de memoria")
	}

	// simulo el retardo de acceso a memoria
	time.Sleep(time.Duration(configuracionMemoria.MemoriaConfig.MemoryDelay) * time.Millisecond)

	if len(data) > tamanioPagina {
		return fmt.Errorf("los datos a escribir exceden el tamaño de la página")
	}
	globals.MemoriaMutex.Lock()
	copy(globals.MemoriaPrincipal[frame*tamanioPagina:(frame+1)*tamanioPagina], data)
	globals.MemoriaMutex.Unlock()

	ActualizarMetricas(pid, frame, "escritura")

	slog.Debug(fmt.Sprintf("La frame está escrita con esto: %s", string(globals.MemoriaPrincipal[frame*tamanioPagina:(frame+1)*tamanioPagina])))

	return nil
}

func framePertenenceAlProceso(pid int, frame int) bool {

	globals.ProcesosMutex.RLock()
	defer globals.ProcesosMutex.RUnlock()

	slice := globals.MarcosXProceso[pid]
	for pagina := range len(slice) {
		if pagina == frame {
			return true
		}
	}
	return false
}

func ActualizarMetricas(pid int, frame int, tipo string) {

	globals.MetricasMutex.Lock()
	defer globals.MetricasMutex.Unlock()

	proceso, exists := globals.MetricasXProceso[pid]
	if !exists {
		return
	}

	tamanioPagina := configuracionMemoria.MemoriaConfig.PageSize
	direccionFisica := frame * tamanioPagina

	switch tipo {
	case "AccesosTablaPaginas":
		proceso.AccesosTablaPaginas += configuracionMemoria.MemoriaConfig.NumberOfLevels
	case "escritura":
		proceso.Escrituras++
		proceso.AccesosMemoriaUsuario++
		logger.EscrituraLecturaUsuario(pid, "Escritura", direccionFisica, tamanioPagina)
	case "lectura":
		proceso.Lecturas++
		proceso.AccesosMemoriaUsuario++
		logger.EscrituraLecturaUsuario(pid, "Lectura", direccionFisica, tamanioPagina)
	default:
		log.Printf("Tipo de acceso no reconocido: %s", tipo)
		return
	}

	// Guardar los cambios
	globals.MetricasXProceso[pid] = proceso
}

// * ========================================= Liberar espacio de usuario y swap =========================================* \\

func FinalizarProceso(pid int) {
	if pid < 0 {
		log.Printf("PID debe ser mayor o igual a 0")
		return
	}

	globals.MetricasMutex.RLock()
	metricasDelProceso, tieneMetricas := globals.MetricasXProceso[pid]
	globals.MetricasMutex.RUnlock()

	if !tieneMetricas {
		log.Printf("No existen métricas para el proceso con PID %d", pid)
		return
	}

	// log obligatorio
	logger.MetricasProceso(pid, metricasDelProceso.AccesosTablaPaginas, metricasDelProceso.AccesosMemoriaUsuario, metricasDelProceso.SwapIns, metricasDelProceso.SwapOuts, metricasDelProceso.Lecturas, metricasDelProceso.Escrituras)

	BorrarEstructurasAdministrativas(pid)

	globals.ProcesosMutex.Lock()
	delete(globals.ProcesosEnMemoriaDeSistema, pid)
	globals.ProcesosMutex.Unlock()

	globals.MetricasMutex.Lock()
	delete(globals.MetricasXProceso, pid)
	globals.MetricasMutex.Unlock()

	log.Printf("Espacio de usuario del proceso con PID %d borrado correctamente", pid)
}

func liberarMarco(entrada *structs.EntradaTablaNivel) {
	if entrada == nil {
		return
	}

	if entrada.SiguienteNivel != nil {
		for _, subEntrada := range entrada.SiguienteNivel.Entradas {
			liberarMarco(subEntrada)
		}
	} else {
		if entrada.Presente {

			globals.MemoriaMutex.Lock()
			globals.BitMapMutex.Lock()

			inicio := entrada.Marco * configuracionMemoria.MemoriaConfig.PageSize
			fin := inicio + configuracionMemoria.MemoriaConfig.PageSize
			for i := inicio; i < fin; i++ {
				if i < len(globals.MemoriaPrincipal) {
					globals.MemoriaPrincipal[i] = 0
				}
			}
			globals.BitMap[entrada.Marco] = false

			globals.BitMapMutex.Unlock()
			globals.MemoriaMutex.Unlock()
		}
	}
}
