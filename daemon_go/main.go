package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== Iniciando Daemon Gestor de Contenedores ===")

	// 1. Definimos la ruta exacta del archivo que tu módulo en C está generando
	// Usamos tu carnet exacto según lo que probamos en la Fase A
	rutaProc := "/proc/continfo_pr1_so1_202307820"

	// 2. Leemos el archivo completo de un solo golpe
	// os.ReadFile nos devuelve los bytes del archivo y un error (si lo hubo)
	datos, err := os.ReadFile(rutaProc)
	if err != nil {
		fmt.Printf("Error fatal: No se pudo leer el archivo del kernel. ¿Está cargado el módulo?\nDetalle: %v\n", err)
		return
	}

	// 3. Si no hay error, convertimos los bytes a texto (string) y los imprimimos
	fmt.Println("¡Lectura exitosa! Esto es lo que Go está viendo desde el Kernel:")
	fmt.Println("--------------------------------------------------")
	fmt.Println(string(datos))
	fmt.Println("--------------------------------------------------")
}
