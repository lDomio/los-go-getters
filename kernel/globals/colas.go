package globals

import (
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var (
	// Cola de procesos
	ColaNew     *structs.ListaPCBSegura
	ColaReady   *structs.ListaPCBSegura
	ColaExec    *structs.ListaPCBSegura
	ColaBlocked *structs.ListaPCBSegura
	ColaExit    *structs.ListaPCBSegura
	//TODO preguntar en el soporte si estaria bueno implementar un canal para las colas

	// Cola de procesos suspendidos
	ColaSuspendedReady   *structs.ListaPCBSegura
	ColaSuspendedBlocked *structs.ListaPCBSegura
)

func init() {
	ColaNew = &structs.ListaPCBSegura{}
	ColaReady = &structs.ListaPCBSegura{}
	ColaBlocked = &structs.ListaPCBSegura{}
	ColaExit = &structs.ListaPCBSegura{}
	ColaExec = &structs.ListaPCBSegura{}

	ColaSuspendedReady = &structs.ListaPCBSegura{}
	ColaSuspendedBlocked = &structs.ListaPCBSegura{}
}
