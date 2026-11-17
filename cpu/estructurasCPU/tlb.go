package estructurasCPU

import (
	"container/list"
	"log/slog"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type TLBE struct {
	Page  int
	Frame int
}

type TLB struct {
	Entries     map[int]*list.Element
	LRUList     *list.List
	FIFOQueue   []int
	Size        int
	Algorithm   string
	PageToFrame map[int]int
}

func NewTLB(size int, algorithm string) *TLB {
	return &TLB{
		Entries:     make(map[int]*list.Element),
		LRUList:     list.New(),
		FIFOQueue:   make([]int, 0),
		Size:        size,
		Algorithm:   algorithm,
		PageToFrame: make(map[int]int),
	}
}

func (tlb *TLB) AddEntry(page, frame int) {
	if tlb.Size == 0 {
		return
	}

	logger.ObtenerMarcoDePagina(globals.PCBenEjecucion.PID, page, frame)

	if len(tlb.Entries) >= tlb.Size {
		tlb.EvictEntry()
	}

	entry := &TLBE{Page: page, Frame: frame}
	element := tlb.LRUList.PushFront(entry)
	tlb.Entries[page] = element
	tlb.FIFOQueue = append(tlb.FIFOQueue, page)
	tlb.PageToFrame[page] = frame
}

func (tlb *TLB) EvictEntry() {
	var pageToRemove int
	if tlb.Algorithm == "LRU" {
		element := tlb.LRUList.Back()
		entry := element.Value.(*TLBE)
		pageToRemove = entry.Page
		tlb.LRUList.Remove(element)
	} else if tlb.Algorithm == "FIFO" {
		pageToRemove = tlb.FIFOQueue[0]
		tlb.FIFOQueue = tlb.FIFOQueue[1:]
	}

	delete(tlb.Entries, pageToRemove)
	delete(tlb.PageToFrame, pageToRemove)
}

func (tlb *TLB) Lookup(page int) (int, bool) {
	if tlb.Size == 0 {
		return -1, false
	}
	frame, exists := tlb.PageToFrame[page]
	if exists {
		if tlb.Algorithm == "LRU" {
			element := tlb.Entries[page]
			tlb.LRUList.MoveToFront(element)
		}
		return frame, true
	}
	return -1, false
}

func (tlb *TLB) Clear() {
	tlb.Entries = make(map[int]*list.Element)
	tlb.LRUList.Init()
	tlb.FIFOQueue = make([]int, 0)
	tlb.PageToFrame = make(map[int]int)
}


func (tlb *TLB) PrintLRUList() {
	slog.Debug("Estado actual de LRUList:")
	for e := tlb.LRUList.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*TLBE)
		slog.Debug("Página: %d, Marco: %d", entry.Page, entry.Frame)
	}
}