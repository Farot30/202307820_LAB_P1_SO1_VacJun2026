package main

type Proceso struct {
	PID    int
	Nombre string
	VSZ    int
	RSS    int
	Utime  int64
	Stime  int64
}
