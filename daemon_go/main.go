package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("=== Iniciando Daemon Gestor de Contenedores ===")
	rutaProc := "/proc/continfo_pr1_so1_202307820"

	fmt.Println("Buscando contenedores activos...")
	contenedores := obtenerContenedores()
	for _, c := range contenedores {
		fmt.Printf("Contenedor ID: %s | Imagen: %s | PID: %d\n", c.ID, c.Imagen, c.PID)
	}

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

func obtenerContenedores() []ContenedorDocker {
	var listaContenedores []ContenedorDocker

	// Ejecutamos 'docker ps -q' para obtener solo los IDs de los contenedores
	cmd := exec.Command("docker", "ps", "-q")
	salida, err := cmd.Output()
	if err != nil {
		fmt.Println("Error al ejecutar docker ps:", err)
		return listaContenedores
	}

	ids := strings.Split(strings.TrimSpace(string(salida)), "\n")
	if len(ids) == 0 || ids[0] == "" {
		return listaContenedores // No hay contenedores corriendo
	}

	// Por cada ID, extraemos su PID y su Imagen usando 'docker inspect'
	for _, id := range ids {
		// Formato: {{.State.Pid}}|{{.Config.Image}}
		cmdInspect := exec.Command("docker", "inspect", "-f", "{{.State.Pid}}|{{.Config.Image}}", id)
		salidaInspect, errInspect := cmdInspect.Output()
		if errInspect != nil {
			continue
		}

		datos := strings.Split(strings.TrimSpace(string(salidaInspect)), "|")
		if len(datos) == 2 {
			pid, _ := strconv.Atoi(datos[0])
			imagen := datos[1]
			tipo := "Desconocido"

			// Clasificamos según la tabla del enunciado
			if strings.Contains(imagen, "go-client") || strings.Contains(imagen, "alpine") {
				// Alpine puede ser ambos, pero el go-client es de alto consumo de RAM
				tipo = "Monitoreado"
			}

			listaContenedores = append(listaContenedores, ContenedorDocker{
				ID:     id,
				Imagen: imagen,
				PID:    pid,
				Tipo:   tipo,
			})
		}
	}

	return listaContenedores
}
