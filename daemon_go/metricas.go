package main

import (
	"time"
)

// Variables globales para calcular el % de CPU
var cpuAnterior = make(map[int]int64) // Guarda PID -> (Utime + Stime)
var tiempoAnterior = time.Now()

// Función para convertir páginas a Megabytes
func calcularMB(paginas int) float64 {
	// 1 página en Linux suele ser de 4096 bytes (4KB)
	return float64(paginas) * 4096.0 / (1024.0 * 1024.0) // Simplificado: float64(paginas) / 256.0
}

// Función para calcular el % de CPU consumido
func calcularPorcentajeCPU(pid int, utime int64, stime int64) float64 {
	cpuActual := utime + stime
	tiempoActual := time.Now()

	usoAnterior, existe := cpuAnterior[pid]
	deltaTiempo := tiempoActual.Sub(tiempoAnterior).Seconds()

	porcentaje := 0.0
	if existe && deltaTiempo > 0 {
		deltaCPU := float64(cpuActual - usoAnterior)
		// En el kernel moderno, utime/stime vienen en nanosegundos (1e9 ns = 1s)
		porcentaje = (deltaCPU / 1000000000.0) / deltaTiempo * 100.0
	}

	// Guardamos el estado actual para el siguiente ciclo
	cpuAnterior[pid] = cpuActual
	return porcentaje
}

func actualizarTiempoCiclo() {
	tiempoAnterior = time.Now()
}
