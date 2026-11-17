package servidormemoria

import (
	"net/http"

	configuracionMemoria "github.com/sisoputnfrba/tp-golang/memoria/configuracionmemoria"
	"github.com/sisoputnfrba/tp-golang/memoria/utils"
	"github.com/sisoputnfrba/tp-golang/utils/httpUtils"
)

func IniciarServidor() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /kernel/handshake", HandlerHandshakeKernel)
	mux.HandleFunc("POST /kernel/consultar-espacio", HandlerConEspacioSuficiente)   
	mux.HandleFunc("POST /kernel/cargar-pseudocodigo", HandlerCargarPseudocodigo)                  //!DONE
	mux.HandleFunc("POST /kernel/suspender-proceso", HandlerSuspenderProceso)                      //!DONE
	mux.HandleFunc("POST /kernel/des-suspender-proceso", HandlerDesSuspenderProceso)               //!DONE
	mux.HandleFunc("POST /kernel/finalizar-proceso", HandlerFinalizarProceso)                      //!DONE
	mux.HandleFunc("POST /kernel/dump-memory", HandlerDumpMemory)                                  //!DOING
	
	mux.HandleFunc("GET /cpu/handshake", HandlerHandshakeCPU)									   //!DONE
	mux.HandleFunc("GET /cpu/solicitar-instruccion", utils.EnviarInstruccion)                      //!DONE
	mux.HandleFunc("GET /cpu/acceder-tabla-de-paginas", HandlerAccederTablaDePaginas)              //!DONE
	mux.HandleFunc("POST /cpu/leer-espacio-usuario", HandlerLecturaEspacioUsuario)  			   //!DONE
	mux.HandleFunc("POST /cpu/escribir-espacio-usuario", HandlerEscrituraEspacioUsuario)           //!DONE
	mux.HandleFunc("GET /cpu/pagina-completa", HandlerLeerPaginaCompleta)                     	   //!DONE
	mux.HandleFunc("POST /cpu/pagina-completa", HandlerActualizarPaginaCompleta)

	httpUtils.StartServer(configuracionMemoria.MemoriaConfig.PortMemory, mux)
}
