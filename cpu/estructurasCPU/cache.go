package estructurasCPU

import (
	"fmt"
	"time"
	"log/slog"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/request"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type PageCacheEntry struct {
	Page       int
	Data       []byte
	Referenced bool
	Modified   bool
}

type PageCache struct {
	Entries   map[int]*PageCacheEntry
	Size      int
	Algorithm string
	ClockHand int
	Pages     []int
	Delay     int
	TLB       *TLB
	MMU       *MMU
}

func NewPageCache(size int, algorithm string) *PageCache {
	return &PageCache{
		Entries:   make(map[int]*PageCacheEntry),
		Size:      size,
		Algorithm: algorithm,
		ClockHand: 0,
		Pages:     make([]int, 0),
		Delay:     globals.CpuConfig.Cache_delay,
	}
}

func (pc *PageCache) SetTLB(tlb *TLB) {
	pc.TLB = tlb
}

func (pc *PageCache) SetMMU(mmu *MMU) {
	pc.MMU = mmu
}

func (pc *PageCache) AddPage(page int, data []byte) {
	if pc.Size == 0 {
		return
	}

	if len(pc.Entries) >= pc.Size {
		pc.EvictPage()
	}

	entry := &PageCacheEntry{Page: page, Data: data, Referenced: true, Modified: false}
	pc.Entries[page] = entry
	pc.Pages = append(pc.Pages, page)
	logger.PaginaIngresoCache(globals.PCBenEjecucion.PID, page)

	time.Sleep(time.Duration(pc.Delay) * time.Millisecond)

}

func (pc *PageCache) EvictPage() {

	time.Sleep(time.Duration(pc.Delay) * time.Millisecond)

	if pc.Algorithm == "CLOCK-M" {
		// Primera pasada: buscar (0,0)
		for i := 0; i < len(pc.Pages); i++ {
			idx := (pc.ClockHand + i) % len(pc.Pages)
			page := pc.Pages[idx]
			entry := pc.Entries[page]
			if !entry.Referenced && !entry.Modified {

				pc.BuscarFrameParaActualizar(entry.Page, entry.Data)

				delete(pc.Entries, entry.Page)
				pc.Pages = append(pc.Pages[:idx], pc.Pages[idx+1:]...)
				pc.ClockHand = idx % len(pc.Pages)
				return
			}
		}
		// Segunda pasada: buscar (0,1) y poner Referenced en 0
		for i := 0; i < len(pc.Pages); i++ {
			idx := (pc.ClockHand + i) % len(pc.Pages)
			page := pc.Pages[idx]
			entry := pc.Entries[page]
			if !entry.Referenced && entry.Modified {

				pc.BuscarFrameParaActualizar(entry.Page, entry.Data)

				delete(pc.Entries, entry.Page)
				pc.Pages = append(pc.Pages[:idx], pc.Pages[idx+1:]...)
				pc.ClockHand = idx % len(pc.Pages)
				return
			}
			entry.Referenced = false
		}
		// Tercera pasada: buscar (0,0) de nuevo
		for i := 0; i < len(pc.Pages); i++ {
			idx := (pc.ClockHand + i) % len(pc.Pages)
			page := pc.Pages[idx]
			entry := pc.Entries[page]
			if !entry.Referenced && !entry.Modified {

				pc.BuscarFrameParaActualizar(entry.Page, entry.Data)

				delete(pc.Entries, entry.Page)
				pc.Pages = append(pc.Pages[:idx], pc.Pages[idx+1:]...)
				pc.ClockHand = idx % len(pc.Pages)
				return
			}
		}
		// Cuarta pasada: buscar (0,1) de nuevo
		for i := 0; i < len(pc.Pages); i++ {
			idx := (pc.ClockHand + i) % len(pc.Pages)
			page := pc.Pages[idx]
			entry := pc.Entries[page]
			if !entry.Referenced && entry.Modified {

				pc.BuscarFrameParaActualizar(entry.Page, entry.Data)

				delete(pc.Entries, entry.Page)
				pc.Pages = append(pc.Pages[:idx], pc.Pages[idx+1:]...)
				pc.ClockHand = idx % len(pc.Pages)
				return
			}
		}
	} else if pc.Algorithm == "CLOCK" {
		for {
			if len(pc.Pages) == 0 {
				break
			}
			page := pc.Pages[pc.ClockHand]
			entry := pc.Entries[page]

			if !entry.Referenced {

				pc.BuscarFrameParaActualizar(entry.Page, entry.Data)

				delete(pc.Entries, entry.Page)
				pc.Pages = append(pc.Pages[:pc.ClockHand], pc.Pages[pc.ClockHand+1:]...)
				if pc.ClockHand >= len(pc.Pages) {
					pc.ClockHand = 0
				}
				break
			}

			entry.Referenced = false
			pc.ClockHand = (pc.ClockHand + 1) % len(pc.Pages)
		}
	}
}

func (pc *PageCache) Lookup(page int, offset int, size int) ([]byte, bool) {

	if pc.Size == 0 {
		return nil, false
	}
	entry, exists := pc.Entries[page]
	if !exists {
		logger.PaginaFaltanteCache(globals.PCBenEjecucion.PID, page)
		return nil, false
	}
	entry.Referenced = true

	if offset < 0 || offset >= len(entry.Data) || size < 0 || offset+size > len(entry.Data) {
		return nil, false
	}

	logger.PaginaEnCache(globals.PCBenEjecucion.PID, page)

	time.Sleep(time.Duration(pc.Delay) * time.Millisecond)

	return entry.Data[offset : offset+size], true
}

func (pc *PageCache) UpdatePage(page int, offset int, data []byte) bool {
	entry, exists := pc.Entries[page]
	if(pc.Size <= 0){
		return false
	}

	if !exists {
		logger.PaginaFaltanteCache(globals.PCBenEjecucion.PID, page)
		return false
	}
	if offset < 0 || offset+len(data) > len(entry.Data) {
		return false
	}
	logger.PaginaEnCache(globals.PCBenEjecucion.PID, page)

	time.Sleep(time.Duration(pc.Delay) * time.Millisecond)
	copy(entry.Data[offset:], data)

	logger.EscrituraMemoria(globals.PCBenEjecucion.PID, 0, string(data))
	entry.Modified = true
	return true
}

func (pc *PageCache) Clear() {
	for page, entry := range pc.Entries {
		if entry.Modified {

			pc.BuscarFrameParaActualizar(entry.Page, entry.Data)

			time.Sleep(time.Duration(pc.Delay) * time.Millisecond)
			
		}
		slog.Debug("Clearing page %d from cache\n", page)
	}
	pc.Entries = make(map[int]*PageCacheEntry)
	pc.Pages = make([]int, 0)
	pc.ClockHand = 0
}

func (pc *PageCache) BuscarFrameParaActualizar(page int, data []byte) {
	TLBFrame, exists := pc.MMU.EstaEnLaTLB(page * pc.MMU.PageSize)
	if exists {
		request.ActualizarPaginaCompleta(TLBFrame, data)
		logger.PaginaDesalojoCache(globals.PCBenEjecucion.PID, page, TLBFrame)
	} else {
		entradasMultiNivel := pc.MMU.CalculateFrameFromMultiLevelTables(page * pc.MMU.PageSize)
		frame, err := request.ObtenerFramePorEntradasMultinivel(entradasMultiNivel)
		if err != nil {
			fmt.Printf("Error al obtener frame por entradas multinivel: %v\n", err)
			return
		}
		pc.TLB.AddEntry(page, frame)
		request.ActualizarPaginaCompleta(frame, data)
		logger.PaginaDesalojoCache(globals.PCBenEjecucion.PID, page, frame)
	}

	

}