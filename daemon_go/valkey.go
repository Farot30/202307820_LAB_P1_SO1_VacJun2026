package main

import (
	"context"
	"encoding/json"
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

// 4. Guarda el Top de contenedores en formato JSON para Grafana
func guardarTopContenedoresValkey(monitoreados []ContenedorMonitoreado) {
	if valkeyClient != nil {
		// En Grafana será más fácil leer un JSON con la lista completa
		jsonData, err := json.Marshal(monitoreados)
		if err == nil {
			valkeyClient.Set(ctx, "contenedores:monitoreados", string(jsonData), 0)
		}
	}
}
