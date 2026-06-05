package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func obtenerContenedores() []ContenedorDocker {
	var lista []ContenedorDocker
	cmd := exec.Command("docker", "ps", "-q")
	salida, err := cmd.Output()
	if err != nil {
		return lista
	}

	ids := strings.Split(strings.TrimSpace(string(salida)), "\n")
	if len(ids) == 0 || ids[0] == "" {
		return lista
	}

	for _, id := range ids {
		// Extraemos PID, Imagen y Comando
		cmdInspect := exec.Command("docker", "inspect", "-f", "{{.State.Pid}}|{{.Config.Image}}|{{join .Config.Cmd \" \"}}", id)
		salidaInspect, errInspect := cmdInspect.Output()
		if errInspect != nil {
			continue
		}

		// CORRECCIÓN: Usamos SplitN limitado a 3 partes para que los '|' internos del comando no rompan el parseo
		datos := strings.SplitN(strings.TrimSpace(string(salidaInspect)), "|", 3)
		if len(datos) == 3 {
			pid, _ := strconv.Atoi(datos[0])
			imagen := datos[1]
			comando := datos[2]

			// Clasificación estricta
			tipo := "Desconocido"
			if strings.Contains(imagen, "go-client") || strings.Contains(comando, "bc") {
				tipo = "Alto Consumo"
			} else if strings.Contains(comando, "sleep") {
				tipo = "Bajo Consumo"
			}

			lista = append(lista, ContenedorDocker{ID: id, Imagen: imagen, PID: pid, Tipo: tipo})
		}
	}
	return lista
}

func main() {
	fmt.Println("=== Iniciando Daemon Gestor de Contenedores ===")

	// 1. Obtenemos los contenedores de Docker
	contenedoresActivos := obtenerContenedores()

	// 2. Leemos el archivo del Kernel
	rutaProc := "/proc/continfo_pr1_so1_202307820"
	datos, err := os.ReadFile(rutaProc)
	if err != nil {
		fmt.Printf("Error al leer /proc: %v\n", err)
		return
	}

	// 3. EL CRUCE DE DATOS (El Match)
	var monitoreados []ContenedorMonitoreado
	lineas := strings.Split(string(datos), "\n")

	for _, linea := range lineas {
		if strings.HasPrefix(linea, "PID:") {
			var p Proceso
			_, err := fmt.Sscanf(linea, "PID: %d | Nombre: %s | VSZ: %d | RSS: %d | Utime: %d | Stime: %d",
				&p.PID, &p.Nombre, &p.VSZ, &p.RSS, &p.Utime, &p.Stime)

			if err == nil {
				// Buscamos si este PID del Kernel pertenece a un contenedor
				for _, cont := range contenedoresActivos {
					if p.PID == cont.PID {
						monitoreados = append(monitoreados, ContenedorMonitoreado{
							IDContenedor: cont.ID,
							Imagen:       cont.Imagen,
							Tipo:         cont.Tipo,
							PID:          p.PID,
							Nombre:       p.Nombre,
							VSZ:          p.VSZ,
							RSS:          p.RSS,
							Utime:        p.Utime,
							Stime:        p.Stime,
						})
					}
				}
			}
		}
	}

	// 4. Imprimimos el resultado final limpio
	fmt.Println("\n--- CONTENEDORES MONITOREADOS DESDE EL KERNEL ---")
	for _, m := range monitoreados {
		fmt.Printf("ID: %s | Tipo: %-15s | PID: %d | RSS: %d pág | CPU Ticks: %d\n",
			m.IDContenedor[:8], m.Tipo, m.PID, m.RSS, (m.Utime + m.Stime))
	}
}
