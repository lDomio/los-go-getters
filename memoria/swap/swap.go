// Package swap es un paquete que maneja el swap
package swap

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/memoria/configuracionmemoria"
	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

const SWAPSIZE int = 1024 * 1024 // 1 MB de swap

type Swap struct {
	procesosEnSwap map[int]*ProcesoEnSwap
	bitmapSwap     []bool
	file           *os.File // Archivo que representa el swap
	mutex          sync.Mutex
	tamanioPagina  int
}

type ProcesoEnSwap struct {
	PID              int
	PosicionesEnSwap []int // Posiciones del bitmap donde están las páginas en el archivo
	CantidadPaginas  int
}

var SwapManager *Swap

func InicializarSwap() {
	path := configuracionmemoria.MemoriaConfig.SwapfilePath
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		slog.Error("Error al abrir el archivo de swap: %v", err)
		return
	}

	// Crear tamaño fijo
	err = file.Truncate(int64(SWAPSIZE))
	if err != nil {
		slog.Error("Error al truncar el archivo de swap: %v", err)
		return
	}

	cantidadPaginas := SWAPSIZE / configuracionmemoria.MemoriaConfig.PageSize
	SwapManager = &Swap{
		procesosEnSwap: make(map[int]*ProcesoEnSwap),
		bitmapSwap:     make([]bool, cantidadPaginas), // Un bit por página
		file:           file,
		mutex:          sync.Mutex{},
		tamanioPagina:  configuracionmemoria.MemoriaConfig.PageSize,
	}

}

// SwapOut: escribir páginas del proceso en el archivo swap
func (swap *Swap) SwapOut(pid int, contenidoEnMarcos [][]byte) error { // contenidoEnMarcos es un slice de slices de bytes [nro_pag][datos], cada slice representa un marco de memoria
	swap.mutex.Lock()
	defer swap.mutex.Unlock()

	time.Sleep(time.Duration(configuracionmemoria.MemoriaConfig.SwapDelay) * time.Millisecond)

	if _, exists := swap.procesosEnSwap[pid]; exists {
		slog.Error("El proceso %d ya está en swap", pid)
	}

	globals.MetricasMutex.Lock()
	metricasDelProceso := globals.MetricasXProceso[pid]
	metricasDelProceso.SwapOuts++
	globals.MetricasXProceso[pid] = metricasDelProceso
	globals.MetricasMutex.Unlock()

	posicionesLibres := swap.encontrarPaginasLibres(len(contenidoEnMarcos))
	if len(posicionesLibres) < len(contenidoEnMarcos) {
		return fmt.Errorf("no hay suficiente espacio en swap para el proceso %d", pid)
	}

	// Escribir cada página en el archivo
	procesoEnSwap := &ProcesoEnSwap{
		PID:              pid,
		PosicionesEnSwap: make([]int, len(contenidoEnMarcos)),
		CantidadPaginas:  len(contenidoEnMarcos),
	}

	// Recorrer las páginas y escribirlas en el archivo
	for i, contenidoPagina := range contenidoEnMarcos {
		posicionEnArchivo := int64(posicionesLibres[i] * swap.tamanioPagina)

		// Escribir la página en el archivo
		_, err := swap.file.WriteAt(contenidoPagina, posicionEnArchivo)
		if err != nil {
			return fmt.Errorf("error escribiendo página %d del proceso %d: %v", i, pid, err)
		}

		// Marcar la posición como ocupada en el bitmap
		swap.bitmapSwap[posicionesLibres[i]] = true

		//
		procesoEnSwap.PosicionesEnSwap[i] = posicionesLibres[i]
	}

	swap.procesosEnSwap[pid] = procesoEnSwap
	slog.Info(fmt.Sprintf("Proceso %d enviado a swap con %d páginas", pid, len(contenidoEnMarcos)))
	slog.Debug(fmt.Sprintf("Proceso %d movido al swap exitosamente", pid))
	return nil
}

// Encuentra páginas libres en el bitmap para asignar al proceso
func (swap *Swap) encontrarPaginasLibres(cantidadNecesaria int) []int {
	var posicionesLibres []int
	for i, ocupado := range swap.bitmapSwap {
		if !ocupado {
			posicionesLibres = append(posicionesLibres, i)
			if len(posicionesLibres) == cantidadNecesaria {
				break
			}
		}
	}
	return posicionesLibres
}

// SwapIn: leer páginas del proceso desde el archivo swap
func (swap *Swap) SwapIn(pid int) ([][]byte, error) {
	swap.mutex.Lock()
	defer swap.mutex.Unlock()

	time.Sleep(time.Duration(configuracionmemoria.MemoriaConfig.SwapDelay) * time.Millisecond)
	metricasDelProceso := globals.MetricasXProceso[pid]
	metricasDelProceso.SwapIns++
	globals.MetricasXProceso[pid] = metricasDelProceso

	// Verificar si el proceso está en swap
	procesoEnSwap, exists := swap.procesosEnSwap[pid]
	if !exists {
		return nil, fmt.Errorf("el proceso %d no está en swap", pid)
	}

	// Leer cada página del archivo
	contenidoMarcos := make([][]byte, procesoEnSwap.CantidadPaginas)
	for i, posicion := range procesoEnSwap.PosicionesEnSwap {
		posicionEnArchivo := int64(posicion * swap.tamanioPagina)

		// Leer la página del archivo
		contenidoPagina := make([]byte, swap.tamanioPagina)
		_, err := swap.file.ReadAt(contenidoPagina, posicionEnArchivo)
		if err != nil {
			return nil, fmt.Errorf("error leyendo página %d del proceso %d: %v", i, pid, err)
		}

		contenidoMarcos[i] = contenidoPagina

		// Liberar la posición en el bitmap
		swap.bitmapSwap[posicion] = false
	}

	// Eliminar el proceso del swap
	delete(swap.procesosEnSwap, pid)
	slog.Info("Proceso %d recuperado del swap con %d páginas", pid, procesoEnSwap.CantidadPaginas)
	log.Printf("Proceso %d movido de swap a memoria exitosamente", pid)
	return contenidoMarcos, nil
}

func (swap *Swap) LeerDeSwap(pid int) ([][]byte, error) {
	swap.mutex.Lock()
	defer swap.mutex.Unlock()

	procesoEnSwap, exists := swap.procesosEnSwap[pid]
	if !exists {
		return nil, fmt.Errorf("el proceso %d no está en swap", pid)
	}

	// Leer cada página del archivo
	contenidoMarcos := make([][]byte, procesoEnSwap.CantidadPaginas)
	for i, posicion := range procesoEnSwap.PosicionesEnSwap {
		posicionEnArchivo := int64(posicion * swap.tamanioPagina)

		// Leer la página del archivo
		contenidoPagina := make([]byte, swap.tamanioPagina)
		_, err := swap.file.ReadAt(contenidoPagina, posicionEnArchivo)
		if err != nil {
			return nil, fmt.Errorf("error leyendo página %d del proceso %d: %v", i, pid, err)
		}

		contenidoMarcos[i] = contenidoPagina
	}
	return contenidoMarcos, nil
}

// Liberar todas las páginas de un proceso (si el proceso termina sin volver a memoria)
func (swap *Swap) LiberarProceso(pid int) error {
	swap.mutex.Lock()
	defer swap.mutex.Unlock()

	procesoEnSwap, exists := swap.procesosEnSwap[pid]
	if !exists {
		return fmt.Errorf("el proceso %d no está en swap", pid)
	}

	// Liberar todas las posiciones en el bitmap
	for _, posicion := range procesoEnSwap.PosicionesEnSwap {
		swap.bitmapSwap[posicion] = false
	}

	delete(swap.procesosEnSwap, pid)
	slog.Info("Proceso %d liberado del swap", pid)
	return nil
}

func (swap *Swap) ElProcesoEstaEnSwap(pid int) bool {
	swap.mutex.Lock()
	defer swap.mutex.Unlock()

	_, exists := swap.procesosEnSwap[pid]
	return exists
}
