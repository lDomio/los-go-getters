package globals

import (
	"github.com/sisoputnfrba/tp-golang/utils/structs"

)

// "time"

type CPUConfig struct {
	Port_cpu    		int    `json:"port_cpu"`
	Ip_cpu      		string `json:"ip_cpu"`
	Ip_memory   		string `json:"ip_memory"`
	Port_memory 		int    `json:"port_memory"`
	Ip_kernel   		string `json:"ip_kernel"`
	Port_kernel 		int    `json:"port_kernel"`
	Tlb_entries 		int    `json:"tlb_entries"`
	Tlb_replacement 	string `json:"tlb_replacement"`
	Cache_entries 		int    `json:"cache_entries"`
	Cache_replacement 	string `json:"cache_replacement"`
	Cache_delay 		int    `json:"cache_delay"`
	Log_level   		string `json:"log_level"`
}

var CpuConfig *CPUConfig

type Estado string

const (
	NEW     Estado = "NEW"
	READY   Estado = "READY"
	EXEC    Estado = "EXEC"
	BLOCKED Estado = "BLOCKED"
	EXIT    Estado = "EXIT"

	// SUSPENDED_READY Estado = "SUSPENDED_READY"
	// SUSPENDED_BLOCKED Estado = "SUSPENDED_BLOCKED"
)

var PCBenEjecucion structs.ProcesoEnEjecucion

var CheckInterrupt bool = false
var CicloFinalizado bool = false

// Falta agregar el resto al json y a la estructura

var DatosMemoria struct{
	TamanioPagina    	int `json:"tamanio_pagina"`
	Cant_Entradas_Tabla int `json:"Cant_Entradas_Tabla"` 
	NumeroNiveles    	int `json:"numero_niveles"` 
}

