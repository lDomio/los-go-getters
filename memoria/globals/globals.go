// Package globals contiene variables globales utilizadas para la gestión de memoria en el sistema.
package globals

import (
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// * ========================================= RECURSOS COMPARTIDOS =========================================* \\

var MemoriaPrincipal []byte

var ProcesosEnMemoriaDeSistema map[int]structs.ProcesoMemoria //[pid]DatoDelProceso (instrucciones)

var CantidadDeMarcos int

var BitMap []bool

var ProcesosEnMemoriaDeUsuario map[int]structs.ProcesoEnMemoriaDeUsuario // [pid]DatoDelProceso ()

var MetricasXProceso map[int]structs.MetricasProceso // [pid]metricas

var MarcosXProceso map[int][]int

// * ========================================= MUTEX PARA SINCRONIZACIÓN =========================================* \\

var (
	// Proteger la memoria principal
	MemoriaMutex sync.RWMutex

	// Proteger el bitmap de marcos
	BitMapMutex sync.Mutex

	// Proteger las estructuras de procesos
	ProcesosMutex sync.RWMutex

	// Proteger las métricas
	MetricasMutex sync.RWMutex

	// Proteger la asignación de marcos
	MarcosAsignacionMutex sync.Mutex
)
