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

	// 1. ORDENAMIENTO CORRECTO (De MENOR a MAYOR consumo)
	// Los índices 0 y 1 serán los más ligeros (los que queremos conservar)
	sort.Slice(altos, func(i, j int) bool { return altos[i].RSS_MB < altos[j].RSS_MB })
	sort.Slice(bajos, func(i, j int) bool { return bajos[i].RSS_MB < bajos[j].RSS_MB })

	// --- GESTIÓN DE ALTO CONSUMO (Debe ser EXACTAMENTE 2) ---
	if len(altos) > 2 {
		fmt.Printf("\n[ALERTA] Exceso de Alto Consumo: %d/2. Eliminando los más pesados...\n", len(altos))
		for i := 2; i < len(altos); i++ {
			matarContenedor(altos[i])
		}
	} else if len(altos) < 2 {
		faltantes := 2 - len(altos)
		fmt.Printf("\n[INFO] Faltan contenedores de Alto Consumo: %d. Levantando nuevos...\n", faltantes)
		for i := 0; i < faltantes; i++ {
			// Levantar contenedor de alto consumo (go-client o el de CPU)
			exec.Command("docker", "run", "-d", "roldyoran/go-client").Run()
		}
	}

	// --- GESTIÓN DE BAJO CONSUMO (Debe ser EXACTAMENTE 3) ---
	if len(bajos) > 3 {
		fmt.Printf("\n[ALERTA] Exceso de Bajo Consumo: %d/3. Eliminando los más pesados...\n", len(bajos))
		for i := 3; i < len(bajos); i++ {
			matarContenedor(bajos[i])
		}
	} else if len(bajos) < 3 {
		faltantes := 3 - len(bajos)
		fmt.Printf("\n[INFO] Faltan contenedores de Bajo Consumo: %d. Levantando nuevos...\n", faltantes)
		for i := 0; i < faltantes; i++ {
			// Levantar contenedor de bajo consumo
			exec.Command("docker", "run", "-d", "alpine", "sleep", "240").Run()
		}
	}
}

func matarContenedor(c ContenedorMonitoreado) {
	fmt.Printf(" -> Matando contenedor %s (PID: %d) - Tipo: %s - RAM: %.2f MB\n", c.IDContenedor[:8], c.PID, c.Tipo, c.RSS_MB)
	cmd := exec.Command("docker", "rm", "-f", c.IDContenedor)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("    [Error] No se pudo eliminar %s: %v\n", c.IDContenedor[:8], err)
	} else {
		fmt.Printf("    [Éxito] Contenedor %s eliminado correctamente.\n", c.IDContenedor[:8])
		registrarEliminacionValkey()
	}
}

// limpiarTodoAlSalir detiene y elimina tanto los contenedores de prueba como Valkey y Grafana
func limpiarTodoAlSalir(rutaDocker string) {
	fmt.Println("[INFO] Buscando contenedores de prueba activos para eliminar...")

	// Reutilizamos tu función para listar los contenedores que están corriendo
	contenedores := obtenerContenedores()

	if len(contenedores) == 0 {
		fmt.Println("   [INFO] No se encontraron contenedores de prueba activos.")
	} else {
		for _, c := range contenedores {
			fmt.Printf("   -> Eliminando contenedor de prueba huérfano: %s (%s) | PID: %d\n", c.ID[:8], c.Tipo, c.PID)
			// 'rm -f' detiene y elimina el contenedor inmediatamente
			exec.Command("docker", "rm", "-f", c.ID).Run()
		}
		fmt.Println("   [Éxito] Todos los contenedores de prueba fueron removidos.")
	}

	// Apagamos Valkey y Grafana de forma segura para no dejar procesos Docker vivos
	fmt.Println("[INFO] Apagando servicios base (Valkey y Grafana) con Docker Compose...")
	cmdDockerDown := exec.Command("bash", "-c", fmt.Sprintf("cd %s && docker compose down", rutaDocker))
	if err := cmdDockerDown.Run(); err != nil {
		fmt.Printf("   [Advertencia] No se pudo ejecutar docker compose down: %v\n", err)
	} else {
		fmt.Println("   [Éxito] Infraestructura de Docker Compose apagada limpiamente.")
	}
}
