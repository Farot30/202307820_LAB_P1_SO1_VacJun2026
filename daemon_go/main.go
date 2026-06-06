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

	// Centralizamos las rutas absolutas del proyecto
	rutaScript := "/home/ubuntu/Documentos/202307820_LAB_P1_SO1_VacJun2026/docker/generador.sh"
	rutaDocker := "/home/ubuntu/Documentos/202307820_LAB_P1_SO1_VacJun2026/docker"
	rutaKernel := "/home/ubuntu/Documentos/202307820_LAB_P1_SO1_VacJun2026/kernel_module"

	// Inicializamos el entorno (Docker Compose, Kernel y Carga Inicial Inmediata)
	inicializarEntorno(rutaDocker, rutaKernel, rutaScript)

	// Conectar a la base de datos Valkey al arrancar
	conectarValkey()

	// Configuración del apagado seguro (Capturar Ctrl+C)
	canalSalida := make(chan os.Signal, 1)
	signal.Notify(canalSalida, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-canalSalida
		fmt.Println("\n\n[ALERTA] Señal de detención recibida. Iniciando limpieza total...")

		// 1. Eliminar Cronjob para que no intente crear más contenedores
		eliminarCronjob(rutaScript)

		// 2. Eliminar contenedores de prueba huérfanos y apagar Valkey/Grafana safely
		limpiarTodoAlSalir(rutaDocker)

		// 3. Limpieza del Módulo del Kernel
		fmt.Println("[INFO] Limpiando el módulo del kernel...")
		cmdRmmod := exec.Command("bash", "-c", "rmmod continfo")
		if err := cmdRmmod.Run(); err != nil {
			fmt.Printf("   [Advertencia] No se pudo remover el módulo: %v\n", err)
		} else {
			fmt.Println("   [Éxito] Módulo removido con sudo rmmod.")
		}

		cmdClean := exec.Command("bash", "-c", fmt.Sprintf("cd %s && make clean", rutaKernel))
		if err := cmdClean.Run(); err != nil {
			fmt.Printf("   [Advertencia] No se pudo ejecutar make clean: %v\n", err)
		} else {
			fmt.Println("   [Éxito] Archivos compilados limpiados con make clean.")
		}

		fmt.Println("=== Daemon Detenido de forma segura y sistema limpio ===")
		os.Exit(0)
	}()

	// Instalar el Cronjob para mantener el flujo automatizado cada 2 minutos
	instalarCronjob(rutaScript)

	// Loop Principal de Monitoreo
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
							// MEJORA: Si es un contenedor de infraestructura (Valkey/Grafana), lo ignoramos
							if cont.Tipo == "Desconocido" {
								continue
							}

							// Convertimos métricas usando tu archivo de calculos (metricas.go)
							vszMB := calcularMB(p.VSZ)
							rssMB := calcularMB(p.RSS)
							cpuPct := calcularPorcentajeCPU(p.PID, p.Utime, p.Stime)

							monitoreados = append(monitoreados, ContenedorMonitoreado{
								IDContenedor: cont.ID,
								Imagen:       cont.Imagen,
								Tipo:         cont.Tipo,
								PID:          p.PID,
								Nombre:       p.Nombre,
								VSZ_MB:       vszMB,
								RSS_MB:       rssMB,
								CPU_Percent:  cpuPct,
								Utime:        p.Utime,
								Stime:        p.Stime,
							})
						}
					}
				}
			}
		}

		// Persistir métricas filtradas en la Base de Datos
		guardarRAMSistemaValkey(ramTotal, ramLibre, ramUso)
		guardarTopContenedoresValkey(monitoreados)

		// Imprimir estado limpio en Consola
		fmt.Println("--- ESTADO ACTUAL DE CONTENEDORES ---")
		if len(monitoreados) == 0 {
			fmt.Println("   [INFO] No hay contenedores de prueba activos en este ciclo.")
		} else {
			for _, m := range monitoreados {
				fmt.Printf("ID: %s | Tipo: %-15s | PID: %d | RSS: %.2f MB | CPU: %.2f%%\n",
					m.IDContenedor[:8], m.Tipo, m.PID, m.RSS_MB, m.CPU_Percent)
			}
		}

		// Ejecutar la lógica de toma de decisiones autónoma (Mantener 3 bajos y 2 altos)
		gestionarContenedores(monitoreados)

		// Actualizar el reloj para calcular de forma precisa el uso de CPU en la siguiente vuelta
		actualizarTiempoCiclo()

		fmt.Println("Esperando 20 segundos para el próximo ciclo...")
		time.Sleep(20 * time.Second)
	}
}

func inicializarEntorno(rutaDocker string, rutaKernel string, rutaScript string) {
	fmt.Println("\n[INFO] Inicializando servicios base...")

	// 1. Levantar Docker Compose (Valkey y Grafana)
	cmdDocker := exec.Command("bash", "-c", fmt.Sprintf("cd %s && docker compose up -d", rutaDocker))
	if err := cmdDocker.Run(); err != nil {
		fmt.Printf("   [Advertencia] Verifica Docker Compose: %v\n", err)
	} else {
		fmt.Println("   [Éxito] Contenedores de Grafana y Valkey inicializados.")
	}

	time.Sleep(2 * time.Second)

	// 2. Compilar Módulo de Kernel
	cmdMake := exec.Command("bash", "-c", fmt.Sprintf("cd %s && make all", rutaKernel))
	if err := cmdMake.Run(); err != nil {
		fmt.Printf("   [Advertencia] Verifica el make all: %v\n", err)
	} else {
		fmt.Println("   [Éxito] Módulo del kernel compilado correctamente.")
	}

	// 3. Inyectar Sonda en el Kernel
	cmdKernel := exec.Command("bash", "-c", fmt.Sprintf("dmesg -c > /dev/null && insmod %s/continfo.ko", rutaKernel))
	if err := cmdKernel.Run(); err != nil {
		fmt.Printf("   [Advertencia] Módulo de kernel ya estaba cargado o requiere permisos: %v\n", err)
	} else {
		fmt.Println("   [Éxito] Sonda de kernel inyectada en el sistema.")
	}

	// 4. MEJORA CRÍTICA: Forzar carga inicial de contenedores inmediatamente en el Segundo 0
	fmt.Println("[INFO] Lanzando carga inicial de contenedores de prueba...")
	cmdCargaInicial := exec.Command("bash", rutaScript)
	if err := cmdCargaInicial.Run(); err != nil {
		fmt.Printf("   [Advertencia] No se pudo ejecutar el script generador al inicio: %v\n", err)
	} else {
		fmt.Println("   [Éxito] Primera tanda de contenedores desplegada correctamente.")
	}
}
