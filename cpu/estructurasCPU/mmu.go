package estructurasCPU

import (
	"math"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type MMU struct {
	TLB         *TLB
	PageCache   *PageCache
	PageSize    int
	TableEntries int
	Levels      int
}

func NewMMU(tlb *TLB, pageCache *PageCache, pageSize int, tableEntries int, levels int) *MMU {
	return &MMU{
		TLB:         tlb,
		PageCache:   pageCache,
		PageSize:    pageSize,
		TableEntries: tableEntries,
		Levels:      levels,
	}
}


func (mmu *MMU) ObtenerPaginaYOffsetDeDireccionLogica(logicalAddress int) (int, int) {
	pageNumber := int(math.Floor(float64(logicalAddress) / float64(mmu.PageSize)))
	offset := logicalAddress % mmu.PageSize
	return pageNumber, offset
}

func (mmu *MMU) CalculateFrameFromMultiLevelTables(logicalAddress int) []int {

	page, _ := mmu.ObtenerPaginaYOffsetDeDireccionLogica(logicalAddress)
	entradas := make([]int, mmu.Levels)
	for x := 0; x < mmu.Levels; x++ {

		divisor := int(math.Pow(float64(mmu.TableEntries), float64(mmu.Levels-x-1))) //TODO revisar si la cuenta es con el -1 o no
		entradaNivelX := (page/divisor) % mmu.TableEntries 
		entradas[x] = entradaNivelX
	}

	return entradas
}

func (mmu *MMU) LeerPaginaEnCache(logicalAddress int, size int) ([]byte, bool) {
	// Buscar en la caché de páginas
	page, offset := mmu.ObtenerPaginaYOffsetDeDireccionLogica(logicalAddress)
	data, hit := mmu.PageCache.Lookup(page, offset, size)
	if hit {
		logger.LecturaMemoria(globals.PCBenEjecucion.PID, 0, string(data))
		return data, true
	}

	return data, false
}

func (mmu *MMU) EstaEnLaTLB(logicalAddress int) (int, bool) {
	pageNumber, _ := mmu.ObtenerPaginaYOffsetDeDireccionLogica(logicalAddress)
	frame, hit := mmu.TLB.Lookup(pageNumber)
	if hit {
		logger.TlbHit(globals.PCBenEjecucion.PID, pageNumber)
		return frame, true
	}
	if (mmu.TLB.Size <=0){
		return -1, false
	}

	logger.TlbMiss(globals.PCBenEjecucion.PID, pageNumber)
	return -1, false
}

func (mmu *MMU) EscribirPaginaEnCache(logicalAddress int, data []byte) (bool) {
	page, offset := mmu.ObtenerPaginaYOffsetDeDireccionLogica(logicalAddress)
	return mmu.PageCache.UpdatePage(page, offset, data)
}


func (mmu *MMU) AgregarPaginaCacheyTLB(logicalAddress int, frame int, data []byte) {
	page, _ := mmu.ObtenerPaginaYOffsetDeDireccionLogica(logicalAddress)
	mmu.TLB.AddEntry(page, frame)
	mmu.PageCache.AddPage(page, data)
}

func (mmu *MMU) AgregarPaginaCache(logicalAddress int, data []byte) {
	page, _ := mmu.ObtenerPaginaYOffsetDeDireccionLogica(logicalAddress)
	mmu.PageCache.AddPage(page, data)
}

func (mmu *MMU) AgregarPaginaTLB(logicalAddress int, frame int) {
	page, _ := mmu.ObtenerPaginaYOffsetDeDireccionLogica(logicalAddress)
	mmu.TLB.AddEntry(page, frame)
}

func (mmu *MMU) TraducirDireccionLogica(frame int, logicalAddress int) (int) {
	_, offset := mmu.ObtenerPaginaYOffsetDeDireccionLogica(logicalAddress)
	return frame * mmu.PageSize + offset
}


func (mmu *MMU) Clear() {
	mmu.PageCache.Clear()
	mmu.TLB.Clear()
}
