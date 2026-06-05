package main

import (
	"fmt"
	"os/exec"
)

func instalarCronjob(rutaScript string) {
	fmt.Println("\n[INFO] Configurando Cronjob para ejecutar la carga cada 2 minutos...")
	comando := fmt.Sprintf("(crontab -l 2>/dev/null | grep -v '%s'; echo '*/2 * * * * %s') | crontab -", rutaScript, rutaScript)
	cmd := exec.Command("bash", "-c", comando)
	if err := cmd.Run(); err != nil {
		fmt.Printf("   [Error] No se pudo instalar el cronjob: %v\n", err)
	} else {
		fmt.Println("   [Éxito] Cronjob instalado correctamente.")
	}
}

func eliminarCronjob(rutaScript string) {
	fmt.Println("\n[INFO] Limpiando el sistema: Eliminando Cronjob...")
	comando := fmt.Sprintf("crontab -l 2>/dev/null | grep -v '%s' | crontab -", rutaScript)
	cmd := exec.Command("bash", "-c", comando)
	if err := cmd.Run(); err != nil {
		fmt.Printf("   [Error] No se pudo eliminar el cronjob: %v\n", err)
	} else {
		fmt.Println("   [Éxito] Cronjob eliminado correctamente.")
	}
}
