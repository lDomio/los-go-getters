package main

// Tamaño configurable desde el config
var MemoriaUsuario []byte

func InicializarMemoriaUsuario(tam int) {
	MemoriaUsuario = make([]byte, tam)
}

var BitmapFrames []bool // true = ocupado, false = libre

func InicializarBitmapFrames(cantFrames int) {
	BitmapFrames = make([]bool, cantFrames)
}

func ReservarFrames(cant int) []int {
	var frames []int
	for i := 0; i < len(BitmapFrames) && len(frames) < cant; i++ {
		if !BitmapFrames[i] {
			BitmapFrames[i] = true
			frames = append(frames, i)
		}
	}
	return frames // Devuelve los índices de los frames reservados
}

// Una entrada de tabla de páginas
type EntradaTabla struct {
	Presente   bool   // ¿Está en memoria?
	Modificado bool   // ¿Está modificada?
	Frame      int    // Número de frame si está presente
	Siguiente  *Tabla // Apunta a la tabla del siguiente nivel (si no es el último nivel)

	// Puedes agregar bits de uso, modificado, etc. si lo pide el TP
}

// Una tabla de páginas de cualquier nivel
type Tabla struct {
	Entradas []*EntradaTabla
}

// Por proceso, la raíz de la tabla de páginas
type TablaDePaginasProceso struct {
	PID  int
	Raiz *Tabla
}

func NuevaTabla(nEntradas int) *Tabla {
	tabla := &Tabla{
		Entradas: make([]*EntradaTabla, nEntradas),
	}
	for i := 0; i < nEntradas; i++ {
		tabla.Entradas[i] = &EntradaTabla{}
	}
	return tabla
}

/*
// Al crear el proceso
procesoTabla := &TablaDePaginasProceso{
    PID:  pid,
    Raiz: NuevaTabla(config.CantEntradasPorNivel),
}

*/
