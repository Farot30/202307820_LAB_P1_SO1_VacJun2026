package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	fmt.Println("=== Iniciando Daemon Gestor de Contenedores ===")

	rutaScript := "/home/ubuntu/Documentos/202307820_LAB_P1_SO1_VacJun2026/docker/generador.sh"

	// Inicializamos el entorno (Docker Compose y Kernel)
	inicializarEntorno()

	// 1. Conectar a la base de datos al arrancar
	conectarValkey()

	// Configuración del apagado seguro (Capturar Ctrl+C)
	canalSalida := make(chan os.Signal, 1)
	signal.Notify(canalSalida, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-canalSalida
		fmt.Println("\n\n[ALERTA] Señal de detención recibida.")

		// 1. Eliminar Cronjob
		eliminarCronjob(rutaScript)

		// 2. Limpieza del Kernel (Las instrucciones que pediste)
		fmt.Println("[INFO] Limpiando el módulo del kernel...")
		rutaKernel := "/home/ubuntu/Documentos/202307820_LAB_P1_SO1_VacJun2026/kernel_module"

		// OJO: El nombre del módulo en 'rmmod' va SIN la extensión .ko
		cmdRmmod := exec.Command("bash", "-c", "rmmod continfo")
		if err := cmdRmmod.Run(); err != nil {
			fmt.Printf("   [Advertencia] No se pudo remover el módulo (quizá ya no estaba): %v\n", err)
		} else {
			fmt.Println("   [Éxito] Módulo removido con sudo rmmod.")
		}

		cmdClean := exec.Command("bash", "-c", fmt.Sprintf("cd %s && make clean", rutaKernel))
		if err := cmdClean.Run(); err != nil {
			fmt.Printf("   [Advertencia] No se pudo ejecutar make clean: %v\n", err)
		} else {
			fmt.Println("   [Éxito] Archivos compilados limpiados con make clean.")
		}

		fmt.Println("=== Daemon Detenido de forma segura ===")
		os.Exit(0)
	}()

	// Instalar el Cronjob
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
		var ramTotal, ramLibre, ramUso string

		lineas := strings.Split(string(datos), "\n")

		for _, linea := range lineas {
			if strings.HasPrefix(linea, "Total de memoria RAM:") {
				fmt.Sscanf(linea, "Total de memoria RAM: %s MB", &ramTotal)
			} else if strings.HasPrefix(linea, "Memoria RAM libre:") {
				fmt.Sscanf(linea, "Memoria RAM libre: %s MB", &ramLibre)
			} else if strings.HasPrefix(linea, "Memoria RAM en uso:") {
				fmt.Sscanf(linea, "Memoria RAM en uso: %s MB", &ramUso)
			}

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

		guardarRAMSistemaValkey(ramTotal, ramLibre, ramUso)
		guardarTopContenedoresValkey(monitoreados)

		fmt.Println("--- ESTADO ACTUAL DE CONTENEDORES ---")
		for _, m := range monitoreados {
			fmt.Printf("ID: %s | Tipo: %-15s | PID: %d | RSS: %d pág\n", m.IDContenedor[:8], m.Tipo, m.PID, m.RSS)
		}

		gestionarContenedores(monitoreados)

		fmt.Println("Esperando 20 segundos para el próximo ciclo...")
		time.Sleep(20 * time.Second)
	}
}

// Mueve la función de inicialización aquí abajo para mantener el orden
func inicializarEntorno() {
	fmt.Println("\n[INFO] Inicializando servicios base...")

	rutaDocker := "/home/ubuntu/Documentos/202307820_LAB_P1_SO1_VacJun2026/docker"
	rutaKernel := "/home/ubuntu/Documentos/202307820_LAB_P1_SO1_VacJun2026/kernel_module"

	// 1. Levantar Docker Compose (Entramos a la carpeta y ejecutamos)
	cmdDocker := exec.Command("bash", "-c", fmt.Sprintf("cd %s && docker compose up -d", rutaDocker))
	if err := cmdDocker.Run(); err != nil {
		fmt.Printf("   [Advertencia] Verifica Docker Compose: %v\n", err)
	} else {
		fmt.Println("   [Éxito] Contenedores de Grafana y Valkey inicializados.")
	}

	// Pausa de 2 segundos para darle tiempo a Valkey de despertar antes de que Go intente conectarse
	time.Sleep(2 * time.Second)

	// 2. Ejecutar el Make All (Entramos a la carpeta del kernel y compilamos)
	cmdMake := exec.Command("bash", "-c", fmt.Sprintf("cd %s && make all", rutaKernel))
	if err := cmdMake.Run(); err != nil {
		fmt.Printf("   [Advertencia] Verifica el make all: %v\n", err)
	} else {
		fmt.Println("   [Éxito] Módulo del kernel compilado correctamente.")
	}

	// 3. Ejecutar el script para cargar el módulo del kernel
	cmdKernel := exec.Command("bash", "-c", fmt.Sprintf("dmesg -c > /dev/null && insmod %s/continfo.ko", rutaKernel))
	if err := cmdKernel.Run(); err != nil {
		fmt.Printf("   [Advertencia] Módulo de kernel ya estaba cargado o requiere permisos: %v\n", err)
	} else {
		fmt.Println("   [Éxito] Sonda de kernel inyectada en el sistema.")
	}
}
