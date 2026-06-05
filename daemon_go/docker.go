package main

import (
	"fmt"
	"os/exec"
	"sort"
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
		cmdInspect := exec.Command("docker", "inspect", "-f", "{{.State.Pid}}|{{.Config.Image}}|{{join .Config.Cmd \" \"}}", id)
		salidaInspect, errInspect := cmdInspect.Output()
		if errInspect != nil {
			continue
		}

		datos := strings.SplitN(strings.TrimSpace(string(salidaInspect)), "|", 3)
		if len(datos) == 3 {
			pid, _ := strconv.Atoi(datos[0])
			imagen := datos[1]
			comando := datos[2]

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

func gestionarContenedores(monitoreados []ContenedorMonitoreado) {
	var altos []ContenedorMonitoreado
	var bajos []ContenedorMonitoreado

	for _, m := range monitoreados {
		if m.Tipo == "Alto Consumo" {
			altos = append(altos, m)
		} else if m.Tipo == "Bajo Consumo" {
			bajos = append(bajos, m)
		}
	}

	sort.Slice(altos, func(i, j int) bool { return altos[i].RSS > altos[j].RSS })
	sort.Slice(bajos, func(i, j int) bool { return bajos[i].RSS > bajos[j].RSS })

	if len(altos) > 2 {
		fmt.Printf("\n[ALERTA] Exceso de Alto Consumo detectado: %d/2. Eliminando excedentes...\n", len(altos))
		for i := 2; i < len(altos); i++ {
			matarContenedor(altos[i])
		}
	}

	if len(bajos) > 3 {
		fmt.Printf("\n[ALERTA] Exceso de Bajo Consumo detectado: %d/3. Eliminando excedentes...\n", len(bajos))
		for i := 3; i < len(bajos); i++ {
			matarContenedor(bajos[i])
		}
	}
}

func matarContenedor(c ContenedorMonitoreado) {
	fmt.Printf(" -> Matando contenedor %s (PID: %d) - Tipo: %s - RAM: %d pág\n", c.IDContenedor[:8], c.PID, c.Tipo, c.RSS)
	cmd := exec.Command("docker", "rm", "-f", c.IDContenedor)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("    [Error] No se pudo eliminar %s: %v\n", c.IDContenedor[:8], err)
	} else {
		fmt.Printf("    [Éxito] Contenedor %s eliminado correctamente.\n", c.IDContenedor[:8])
	}
}
