package main

import (
	"log"

	configuracionMemoria "github.com/sisoputnfrba/tp-golang/memoria/configuracionmemoria"
	servidorMemoria "github.com/sisoputnfrba/tp-golang/memoria/servidormemoria"
	"github.com/sisoputnfrba/tp-golang/memoria/utils"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

func main() {
	configuracionMemoria.MemoriaConfig = config.IniciarConfiguracion[structs.MemoriaConfig]("configuracionmemoria/MEMORYConfig.json")
	logger.ConfigurarLogger("memoria.log", configuracionMemoria.MemoriaConfig.LogLevel)

	if configuracionMemoria.MemoriaConfig == nil {
		log.Fatalf("No se pudo cargar la configuracion")
	}

	utils.InicializarEstructuras(configuracionMemoria.MemoriaConfig)
	servidorMemoria.IniciarServidor()
}
