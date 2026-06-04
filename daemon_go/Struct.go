package main

type Proceso struct {
	PID    int
	Nombre string
	VSZ    int
	RSS    int
	Utime  int64
	Stime  int64
}

// Estructura para almacenar los datos de un contenedor vivo
type ContenedorDocker struct {
	ID     string
	Imagen string
	PID    int
	Tipo   string
}
