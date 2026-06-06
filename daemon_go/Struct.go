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

type ContenedorMonitoreado struct {
	IDContenedor string  `json:"id"`
	Imagen       string  `json:"imagen"`
	Tipo         string  `json:"tipo"`
	PID          int     `json:"pid"`
	Nombre       string  `json:"nombre"`
	VSZ_MB       float64 `json:"vsz_mb"`
	RSS_MB       float64 `json:"rss_mb"`
	CPU_Percent  float64 `json:"cpu"`
	Utime        int64   `json:"-"`
	Stime        int64   `json:"-"`
}
