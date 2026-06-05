package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	fmt.Println("=== Iniciando Daemon Gestor de Contenedores ===")

	rutaScript := "/home/ubuntu/Documentos/202307820_LAB_P1_SO1_VacJun2026/docker/generador.sh"

	// Configurar el apagado seguro (Capturar Ctrl+C)
	canalSalida := make(chan os.Signal, 1)
	signal.Notify(canalSalida, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-canalSalida
		fmt.Println("\n\n[ALERTA] Señal de detención recibida.")
		eliminarCronjob(rutaScript)
		fmt.Println("=== Daemon Detenido de forma segura ===")
		os.Exit(0)
	}()

	// Instalar el Cronjob al iniciar
	instalarCronjob(rutaScript)

	// Loop Principal
	for {
		fmt.Println("\n==================================================")
		fmt.Println("Realizando escaneo del sistema...")

		contenedoresActivos := obtenerContenedores()
		rutaProc := "/proc/continfo_pr1_so1_202307820"
		datos, err := os.ReadFile(rutaProc)
		if err != nil {
			fmt.Printf("Error al leer /proc: %v\n", err)
			time.Sleep(20 * time.Second)
			continue
		}

		var monitoreados []ContenedorMonitoreado
		lineas := strings.Split(string(datos), "\n")

		for _, linea := range lineas {
			if strings.HasPrefix(linea, "PID:") {
				var p Proceso
				_, err := fmt.Sscanf(linea, "PID: %d | Nombre: %s | VSZ: %d | RSS: %d | Utime: %d | Stime: %d",
					&p.PID, &p.Nombre, &p.VSZ, &p.RSS, &p.Utime, &p.Stime)

				if err == nil {
					for _, cont := range contenedoresActivos {
						if p.PID == cont.PID {
							monitoreados = append(monitoreados, ContenedorMonitoreado{
								IDContenedor: cont.ID, Imagen: cont.Imagen, Tipo: cont.Tipo,
								PID: p.PID, Nombre: p.Nombre, VSZ: p.VSZ, RSS: p.RSS, Utime: p.Utime, Stime: p.Stime,
							})
						}
					}
				}
			}
		}

		fmt.Println("--- ESTADO ACTUAL DE CONTENEDORES ---")
		for _, m := range monitoreados {
			fmt.Printf("ID: %s | Tipo: %-15s | PID: %d | RSS: %d pág\n", m.IDContenedor[:8], m.Tipo, m.PID, m.RSS)
		}

		gestionarContenedores(monitoreados)

		fmt.Println("Esperando 20 segundos para el próximo ciclo...")
		time.Sleep(20 * time.Second)
	}
}
