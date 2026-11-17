package globals

type IOConfig struct {
	Ip_kernel   string `json:"ip_kernel"`
	Port_kernel int    `json:"port_kernel"`
	Port_io     int    `json:"port_io"`
	Ip_io       string `json:"ip_io"`
	Log_level   string `json:"log_level"`
	Nombre      string // por el momento hasta saber como pasar el nombre por argumento del main
}

var IoConfig *IOConfig

var PIDEnEjecucion int // PID del proceso que está ejecutando la IO
