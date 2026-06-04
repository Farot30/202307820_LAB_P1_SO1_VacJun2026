package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	fmt.Println("=== Iniciando Daemon Gestor de Contenedores ===")
	rutaProc := "/proc/continfo_pr1_so1_202307820"

	datos, err := os.ReadFile(rutaProc)
	if err != nil {
		fmt.Printf("Error: No se pudo leer /proc: %v\n", err)
		return
	}

	lineas := strings.Split(string(datos), "\n")
	var listaProcesos []Proceso

	for _, linea := range lineas {
		if strings.HasPrefix(linea, "PID:") {
			var p Proceso
			_, err := fmt.Sscanf(linea, "PID: %d | Nombre: %s | VSZ: %d | RSS: %d | Utime: %d | Stime: %d",
				&p.PID, &p.Nombre, &p.VSZ, &p.RSS, &p.Utime, &p.Stime)

			if err == nil {
				listaProcesos = append(listaProcesos, p)
			}
		}
	}

	// --- ORDENAMIENTO POR RAM (RSS) ---
	sort.Slice(listaProcesos, func(i, j int) bool {
		return listaProcesos[i].RSS > listaProcesos[j].RSS
	})

	fmt.Println("\n--- TOP 5 PROCESOS POR USO DE RAM (RSS) ---")
	for i := 0; i < 5 && i < len(listaProcesos); i++ {
		fmt.Printf("%d. %s (PID: %d) - RSS: %d páginas\n", i+1, listaProcesos[i].Nombre, listaProcesos[i].PID, listaProcesos[i].RSS)
	}

	// --- ORDENAMIENTO POR CPU (Utime + Stime) ---
	sort.Slice(listaProcesos, func(i, j int) bool {
		cpuI := listaProcesos[i].Utime + listaProcesos[i].Stime
		cpuJ := listaProcesos[j].Utime + listaProcesos[j].Stime
		return cpuI > cpuJ
	})

	fmt.Println("\n--- TOP 5 PROCESOS POR USO DE CPU (Ticks) ---")
	for i := 0; i < 5 && i < len(listaProcesos); i++ {
		cpuTotal := listaProcesos[i].Utime + listaProcesos[i].Stime
		fmt.Printf("%d. %s (PID: %d) - CPU Ticks: %d\n", i+1, listaProcesos[i].Nombre, listaProcesos[i].PID, cpuTotal)
	}
}
