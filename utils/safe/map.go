package safe

import (
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// TODO Meter listas

// ----------------------( MAP CPU )----------------------\\
type MapCPUSeguro struct {
	mx sync.Mutex
	m  map[int]*structs.CPU
}

func InicializarMapCPUSeguro() MapCPUSeguro {
	return MapCPUSeguro{
		m: make(map[int]*structs.CPU),
	}
}

func (sMap *MapCPUSeguro) Set(key int, value *structs.CPU) {
	sMap.mx.Lock()
	sMap.m[key] = value
	sMap.mx.Unlock()
}

func (sMap *MapCPUSeguro) Get(key int) *structs.CPU {
	sMap.mx.Lock()
	value := sMap.m[key]
	sMap.mx.Unlock()
	return value
}

func (sMap *MapCPUSeguro) GetLibre() (*structs.CPU, bool) {
	sMap.mx.Lock()
	defer sMap.mx.Unlock()

	for _, cpu := range sMap.m {
		if cpu.Libre {
			return cpu, true
		}
	}
	return &structs.CPU{}, false
}

// ----------------------( MAP IO )----------------------\\

type MapIOConectadaSeguro struct {
	mx sync.Mutex
	m  map[string]*structs.IOConectada
}

func InicializarMapIOConectadaSeguro() *MapIOConectadaSeguro {
	return &MapIOConectadaSeguro{
		m: make(map[string]*structs.IOConectada),
	}
}

// SetInstancia agrega una nueva instancia de IO al mapa seguro, o agrega la instancia a una IO existente.
func (sMap *MapIOConectadaSeguro) SetInstancia(key string, value *structs.InstanciaIO) {
	sMap.mx.Lock()
	defer sMap.mx.Unlock()

	if io, exists := sMap.m[key]; exists {
		io.AgregarInstancia(value)
	} else {
		sMap.m[key] = &structs.IOConectada{
			Nombre:          key,
			InstanciasIO:    []*structs.InstanciaIO{},
			ColaPCBsWaiting: []*structs.PCBWaitingIO{},
		}
		sMap.m[key].AgregarInstancia(value)
	}
}

// Get devuelve la IO conectada asociada a la clave proporcionada, o nil si no existe.
func (sMap *MapIOConectadaSeguro) Get(key string) *structs.IOConectada {
	sMap.mx.Lock()
	defer sMap.mx.Unlock()

	if io, exists := sMap.m[key]; exists {
		return io
	}
	return nil
}

// ----------------------( MAP SEGURO )----------------------\\

type MapSeguro[K comparable, V any] struct {
	mx sync.Mutex
	m  map[K]V
}

func InicializarMapSeguro[K comparable, V any]() *MapSeguro[K, V] {
	return &MapSeguro[K, V]{
		m: make(map[K]V),
	}
}

func (sMap *MapSeguro[K, V]) Set(key K, value V) {
	sMap.mx.Lock()
	sMap.m[key] = value
	sMap.mx.Unlock()
}

func (sMap *MapSeguro[K, V]) Get(key K) V {
	sMap.mx.Lock()
	value := sMap.m[key]
	sMap.mx.Unlock()
	return value
}

func (sMap *MapSeguro[K, V]) Remove(key K) {
	sMap.mx.Lock()
	delete(sMap.m, key)
	sMap.mx.Unlock()
}

func (sMap *MapSeguro[K, V]) GetAndRemove(key K) (V, bool) {
	sMap.mx.Lock()
	defer sMap.mx.Unlock()

	value, exists := sMap.m[key]
	if exists {
		delete(sMap.m, key)
		return value, true
	}
	var zeroValue V // Valor cero para el tipo V
	return zeroValue, false
}

func (sMap *MapSeguro[K, V]) GetAll() map[K]V {
	sMap.mx.Lock()
	defer sMap.mx.Unlock()

	// Crear una copia del mapa para evitar problemas de concurrencia
	copia := make(map[K]V, len(sMap.m))
	for k, v := range sMap.m {
		copia[k] = v
	}
	return copia
}

// ----------------------( MAP IO )----------------------\\
