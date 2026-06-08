package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var valkeyClient *redis.Client

// 1. Inicia la conexión con Valkey
func conectarValkey() {
	valkeyClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // El puerto que vimos expuesto en tu docker ps
		Password: "",               // Sin contraseña por defecto
		DB:       0,                // Base de datos por defecto
	})

	_, err := valkeyClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("   [Advertencia] No se pudo conectar a Valkey: %v\n", err)
	} else {
		fmt.Println("   [Éxito] Conectado a la base de datos Valkey.")
	}
}

// 2. Incrementa en 1 el contador cada vez que matamos un contenedor
func registrarEliminacionValkey() {
	if valkeyClient != nil {
		valkeyClient.Incr(ctx, "sistema:contenedores_eliminados")
	}
}

// 3. Guarda las métricas generales de RAM (enviadas por tu módulo en C)
func guardarRAMSistemaValkey(total string, libre string, enUso string) {
	if valkeyClient != nil {
		valkeyClient.Set(ctx, "sistema:ram:total", total, 0)
		valkeyClient.Set(ctx, "sistema:ram:libre", libre, 0)
		valkeyClient.Set(ctx, "sistema:ram:en_uso", enUso, 0)
	}
}

// Actualiza el ranking histórico de RAM (Guarda el pico más alto del contenedor)
func registrarConsumoRAMValkey(id string, pid int, ramMB float64) {
	if valkeyClient != nil {
		// Formateamos el miembro para que muestre PID e ID como pide tu enunciado
		miembro := fmt.Sprintf("PID: %d | ID: %s", pid, id[:8])

		// Con NX no actualiza si ya existe, pero queremos guardar el PICO MÁS ALTO histórico.
		// Podemos usar ZAdd y Valkey se encargará de ordenarlo.
		valkeyClient.ZAdd(ctx, "sistema:ranking:ram", redis.Z{
			Score:  ramMB,
			Member: miembro,
		})
	}
}

// Actualiza el ranking histórico de CPU (Guarda el pico más alto del contenedor)
func registrarConsumoCPUValkey(id string, pid int, cpuPercent float64) {
	if valkeyClient != nil {
		miembro := fmt.Sprintf("PID: %d | ID: %s", pid, id[:8])

		valkeyClient.ZAdd(ctx, "sistema:ranking:cpu", redis.Z{
			Score:  cpuPercent,
			Member: miembro,
		})
	}
}
