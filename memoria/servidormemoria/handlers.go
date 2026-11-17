// Package servidormemoria provides HTTP server functionality for the memoria module.
// It handles requests from kernel and CPU components.
package servidormemoria

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	configuracionMemoria "github.com/sisoputnfrba/tp-golang/memoria/configuracionmemoria"
	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	"github.com/sisoputnfrba/tp-golang/memoria/swap"
	"github.com/sisoputnfrba/tp-golang/memoria/utils"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//* ============================= HANDLERS PARA KERNEL ============================= //

func HandlerHandshakeKernel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensaje":"Conección a memoria aceptada"}`))
}

func HandlerConEspacioSuficiente(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var procesoEnviadoPorKernel structs.ProcesoKernelMemoria
	if err := json.NewDecoder(r.Body).Decode(&procesoEnviadoPorKernel); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	cantidadDePaginas := utils.CalcularCantidadDePaginas(procesoEnviadoPorKernel.Size, *configuracionMemoria.MemoriaConfig)
	marcosLibres := utils.MarcosLibresNecesarios(cantidadDePaginas)
	var respuestaMemoria structs.RespuestaEspacioMemoria

	if len(marcosLibres) < cantidadDePaginas {
		respuestaMemoria.OK = false
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respuestaMemoria)
		return
	}

	respuestaMemoria.OK = true

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuestaMemoria)

}

func HandlerCargarPseudocodigo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var procesoEnviadoPorKernel structs.ProcesoKernelMemoria
	if err := json.NewDecoder(r.Body).Decode(&procesoEnviadoPorKernel); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	cantidadDePaginas := utils.CalcularCantidadDePaginas(procesoEnviadoPorKernel.Size, *configuracionMemoria.MemoriaConfig)
	marcosLibres := utils.MarcosLibresNecesarios(cantidadDePaginas)
	var respuestaMemoria structs.RespuestaEspacioMemoria

	if len(marcosLibres) < cantidadDePaginas {
		respuestaMemoria.OK = false
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respuestaMemoria)
		return
	}

	utils.CargarProceso(procesoEnviadoPorKernel, cantidadDePaginas, &marcosLibres)

	// log obligatorio
	logger.CreacionProceso(procesoEnviadoPorKernel.PID, procesoEnviadoPorKernel.Size)

	respuestaMemoria.OK = true

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuestaMemoria)
}

func HandlerSuspenderProceso(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var procesoSuspender structs.ProcesoKernelMemoria
	if err := json.NewDecoder(r.Body).Decode(&procesoSuspender); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	slog.Debug(fmt.Sprintf("Recibida solicitud de suspender proceso: %d", procesoSuspender.PID))

	contenidoEnMarcos, err := utils.ObtenerContenidoEnMarcos(procesoSuspender.PID)
	if err != nil {
		http.Error(w, "Error obteniendo contenido en marcos", http.StatusInternalServerError)
		return
	}
	swap.SwapManager.SwapOut(procesoSuspender.PID, contenidoEnMarcos)

	utils.BorrarEstructurasAdministrativas(procesoSuspender.PID)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Proceso suspendido"))
}

func HandlerDesSuspenderProceso(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var procesoDesSuspender structs.ProcesoKernelMemoria
	if err := json.NewDecoder(r.Body).Decode(&procesoDesSuspender); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	slog.Debug(fmt.Sprintf("Recibida solicitud de des-suspender proceso: %d", procesoDesSuspender.PID))

	contenidoDelProcesoAEscribir, err := swap.SwapManager.SwapIn(procesoDesSuspender.PID)
	if err != nil {
		slog.Error("No se pudo recuperar el proceso de swap")
		return
	}
	cantidadDePaginas := len(contenidoDelProcesoAEscribir)
	marcosLibres := utils.MarcosLibresNecesarios(cantidadDePaginas)

	utils.ActualizarTablaDePaginas(procesoDesSuspender, marcosLibres)

	globals.ProcesosMutex.Lock()
	globals.MarcosXProceso[procesoDesSuspender.PID] = marcosLibres
	globals.ProcesosMutex.Unlock()

	globals.BitMapMutex.Lock()
	for _, marco := range marcosLibres {
		globals.BitMap[marco] = true
	}
	globals.BitMapMutex.Unlock()

	for i := range cantidadDePaginas {
		marcoAEscribir := marcosLibres[i]
		contenidoAEscirbir := contenidoDelProcesoAEscribir[i]
		utils.EscribirFrameCompleto(procesoDesSuspender.PID, marcoAEscribir, contenidoAEscirbir)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Proceso des-suspendido"))
}

func HandlerFinalizarProceso(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var procesoFinalizar structs.ProcesoKernelMemoria

	if err := json.NewDecoder(r.Body).Decode(&procesoFinalizar); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	// borra el proceso de la lista de procesos en memoria
	utils.FinalizarProceso(procesoFinalizar.PID)

	slog.Debug(fmt.Sprintf("Recibida solicitud de finalizar proceso: %d", procesoFinalizar.PID))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Proceso finalizado"))
}

func HandlerDumpMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		panic("method not allowed")
	}

	var request structs.ProcesoKernelMemoria
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		panic("error decoding JSON")
	}

	slog.Debug(fmt.Sprintf("Recibida solicitud de dump de memoria para proceso: %d", request.PID))

	// Generar el dump de memoria
	dump, err := utils.GenerarDumpMemoria(request.PID)
	if err != nil {
		http.Error(w, "Error generando dump de memoria", http.StatusInternalServerError)
		panic("error generando dump de memoria")
	}

	var respuestaMemoria structs.RespuestaEspacioMemoria
	// Guardar dump en archivo según configuración
	err = utils.GuardarDumpEnArchivo(request.PID, dump)
	if err != nil {
		respuestaMemoria.OK = false
		log.Printf("Error guardando dump en archivo: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respuestaMemoria)
	}

	respuestaMemoria.OK = true
	// Enviar el dump como respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuestaMemoria)
}

// *============================= HANDLERS PARA CPU ============================= //

func HandlerHandshakeCPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	handshake := structs.HandshakeMemoria{
		TamanioPagina:     configuracionMemoria.MemoriaConfig.PageSize,
		CantEntradasTabla: configuracionMemoria.MemoriaConfig.EntriesPerPage,
		NumeroNiveles:     configuracionMemoria.MemoriaConfig.NumberOfLevels,
	}

	response, err := json.Marshal(handshake)
	if err != nil {
		http.Error(w, "Error creando respuesta JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// ========================================= Calcular el frame ========================================= \\

func HandlerAccederTablaDePaginas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request structs.EnvioEntradasMultinivel
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	frame := utils.CalcularFrame(request.PID, request.Entradas)
	if frame == -1 {
		http.Error(w, "Error al calcular frame", http.StatusInternalServerError)
		return
	}

	response := structs.RespuestaEnvioEntradasMultinivel{
		Frame: frame,
	}

	slog.Debug(fmt.Sprintf("Frame calculado: %d para PID: %d", frame, request.PID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error al codificar la respuesta", http.StatusInternalServerError)
		return
	}
}

// ========================================= Leer espacio de usuario ========================================= \\

func HandlerLecturaEspacioUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request structs.LeerMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	contenidoFrame, err := utils.LeerEspacioUsuario(request.PID, request.PhysicalAddress, request.Size)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(contenidoFrame)
}

// ========================================= Escribir espacio de usuario ========================================= \\

func HandlerEscrituraEspacioUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request structs.EscribirMemoriaRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	erro := utils.EscribirEspacioUsuario(request.PID, request.PhysicalAddress, request.Datos)
	if erro != nil {
		http.Error(w, "Error al escribir en el espacio de usuario", http.StatusInternalServerError)
		return
	}

	tamanioDatos := len(request.Datos)
	logger.EscrituraLecturaUsuario(request.PID, "Escritura", request.PhysicalAddress, tamanioDatos)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

// ========================================= Leer pagina completa de espacio de usuario ========================================= \\

func HandlerLeerPaginaCompleta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request structs.SolicitudFrameCompleta
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	contenidoFrame, err := utils.LeerFrameCompleto(request.PID, request.Frame)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tamanioPagina := configuracionMemoria.MemoriaConfig.PageSize
	direccionFisica := request.Frame * tamanioPagina

	logger.EscrituraLecturaUsuario(request.PID, "Lectura", direccionFisica, configuracionMemoria.MemoriaConfig.PageSize)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(contenidoFrame)
}

// ========================================= Actualizar pagina completa de espacio de usuario ========================================= \\

func HandlerActualizarPaginaCompleta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request structs.ActualizarFrameCompletaRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	err := utils.EscribirFrameCompleto(request.PID, request.Frame, request.Datos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tamanioPagina := configuracionMemoria.MemoriaConfig.PageSize
	direccionFisica := request.Frame * tamanioPagina

	logger.EscrituraLecturaUsuario(request.PID, "Escritura", direccionFisica, configuracionMemoria.MemoriaConfig.PageSize)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}
