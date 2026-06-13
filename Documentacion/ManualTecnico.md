UNIVERSIDAD DE SAN CARLOS DE GUATEMALA

FACULTAD DE INGENIERÍA

MANEJO E IMPLEMENTACIÓN DE ARCHIVOS

CATEDRÁTICO: ING. SERGIO ARNALDO MENDEZ AGUILAR

TUTOR ACADÉMICO: CAMILO ERNESTO SINCAL SIPAC

======================================================================================

<div style="text-align: center;">
    <h1 style="color: white; text-decoration: underline; text-decoration-color: white;"> MANUAL TECNICO Sonda de Kernel en C y Daemon en Go para Telemetría de contenedores
</div>

=======================================================================================

<p align="right">
OTTONIEL FABRICIO VÁSQUEZ PINEDA
<br>
CARNÉ: 202307820
<br>
SECCIÓN B
</p>

<P align="center">GUATEMALA 12 DE JUNIO DEL 2026</P>

---

## 1. Descripción General

El proyecto consiste en un **sistema integral de monitorización, análisis y gestión autónoma de contenedores Docker** en entornos Linux. El sistema está compuesto por cuatro componentes principales que trabajan de forma coordinada:

| Componente | Lenguaje/Tecnología | Rol |
|---|---|---|
| Módulo de Kernel | C | Sensor de bajo nivel — captura métricas del sistema |
| Daemon | Go | Cerebro del sistema — análisis, decisiones y orquestación |
| Cronjob | Bash/Shell | Generación periódica de contenedores de prueba |
| Dashboard | Grafana + Valkey | Visualización en tiempo real de métricas |

---

## 2. Arquitectura del Sistema

```
┌─────────────────────────────────────────────────────────┐
│                  DAEMON GESTIONADOR (Go)                │
│                  Ejecución paso a paso                  │
└────┬───────────────────┬──────────────────┬─────────────┘
     │ Paso 1            │ Paso 2           │ Paso 3
     ▼                   ▼                  ▼
┌──────────┐     ┌──────────────┐   ┌──────────────────────┐
│  Grafana │     │   Cronjob    │   │  Script de carga de  │
│(contenedor│    │  (cada 2 min)│   │  módulo de kernel    │
│Docker)   │     └──────┬───────┘   └──────────┬───────────┘
└────┬─────┘            │                       │
     │                  ▼                       ▼
     │          ┌───────────────┐      ┌────────────────────┐
     │          │  Contenedores │      │  Módulo de Kernel  │
     │          │  Docker (x5   │      │  (C) — escribe en  │
     │          │  cada 2 min)  │      │  /proc/continfo_   │
     │          └───────────────┘      │  pr1_so1_#CARNET   │
     │                                 └────────────────────┘
     │ Paso 4 (loop cada 20-60s)
     ◄─────────────── Lee /proc ──────────────────────────────
     │
     ▼
┌──────────┐     Escribe logs/métricas
│  Valkey  │◄────────────────────────────────────────────────
│(base de  │
│  datos)  │────► Grafana lee y visualiza
└──────────┘
```

### Flujo de ejecución

1. El Daemon inicia y crea el contenedor de Grafana.
2. El Daemon registra el cronjob que generará contenedores cada 2 minutos.
3. El Daemon ejecuta el script que carga el módulo de kernel en el sistema.
4. **Loop principal** (cada 20–60 segundos): lee `/proc`, deserializa los datos, toma decisiones sobre qué contenedores eliminar, y escribe los logs en Valkey.
5. Al finalizar, el Daemon elimina el cronjob para evitar sobrecarga.

### 2.1 Perfiles de Carga (Generador.sh)
El script bash inyecta aleatoriamente contenedores basándose en tres perfiles:
* **Alto Consumo (RAM):** Utiliza la imagen `roldyoran/go-client`.
* **Alto Consumo (CPU):** Utiliza la imagen `alpine` ejecutando un bucle matemático infinito: `while true; do echo '2^20' | bc > /dev/null; sleep 2; done`.
* **Bajo Consumo:** Utiliza la imagen `alpine` en estado de inactividad: `sleep 240`.

---

## 3. Requisitos del Sistema

### Software

| Herramienta | Versión recomendada | Obligatorio |
|---|---|---|
| Linux (kernel ≥ 5.x) | Ubuntu 22.04 LTS o superior | Sí |
| GCC + Make | Sistema de compilación estándar | Sí |
| Linux Kernel Headers | Matching con el kernel en uso | Sí |
| Go | 1.21 o superior | Sí |
| Docker | 24.x o superior | Sí |
| Docker Compose | v2.x | Sí |
| Grafana | Imagen oficial Docker | Sí |
| Valkey | Imagen oficial Docker | Sí |

### Hardware mínimo recomendado

- CPU: 2 núcleos
- RAM: 4 GB
- Almacenamiento: 20 GB libres

### Permisos requeridos

- Acceso `root` o `sudo` para cargar módulos de kernel (`insmod`/`modprobe`)
- Acceso al socket de Docker (`/var/run/docker.sock`) desde el Daemon en Go

---

## 4. Módulo de Kernel en C

### 4.1 Descripción

El módulo de kernel es un **LKM (Loadable Kernel Module)** desarrollado en C que actúa como sensor de bajo nivel. Accede directamente a las estructuras internas del kernel Linux para recopilar métricas de procesos y memoria, exponiéndolas a través del sistema de archivos virtual `/proc`.

### 4.2 Métricas capturadas

- Total de memoria RAM
- Memoria RAM libre
- Memoria RAM en uso

**Por cada proceso de contenedor y del sistema:**

| Campo | Descripción |
|---|---|
| `PID` | Identificador del proceso |
| `Nombre` | Nombre del proceso |
| `Cmdline` | Línea de comando ejecutada o ID del contenedor |
| `VSZ` | Virtual Memory Size (KB) |
| `RSS` | Resident Set Size — memoria física (KB) |
| `%MEM` | Porcentaje de memoria utilizada |
| `%CPU` | Porcentaje de CPU utilizado |


### 4.3 Archivos en /proc

El módulo crea el siguiente archivo:

```
/proc/continfo_pr1_so1_202307820
```


---

## 5. Subsistema de Usuario (Daemon en Go)
El Daemon está diseñado bajo principios de concurrencia estructurada y manejo asíncrono de eventos de sistema. Su modularidad se divide de la siguiente manera:

### Módulos del Daemon:
- main.go: Punto de anclaje que gestiona la secuencia de arranque del entorno, invoca la compilación del módulo vía exec.Command("make") e implementa el bucle de control iterativo.

- docker.go (Orquestación Autónoma): Consume la API de Docker mediante comandos CLI estructurados (docker inspect, docker run, docker rm). Implementa el algoritmo de ordenamiento y filtrado por peso de consumo.

- cron.go (Planificación Operativa): Modifica el archivo crontab del sistema operativo de forma programática:

~~~Go
comando := fmt.Sprintf("(crontab -l 2>/dev/null | grep -v '%s'; echo '*/2 * * * * %s') | crontab -", rutaScript, rutaScript)
~~~
- metricas.go (Capa Matemática y Algorítmica): Resuelve las conversiones analíticas de métricas del sistema.

- valkey.go (Capa de Datos): Abstrae la comunicación de red TCP con la base de datos usando un pool de conexiones optimizado de go-redis/v9.

---

## 6. Modelo de Datos y Almacenamiento (Valkey DB)
La persistencia de telemetría utiliza estructuras de datos avanzadas en memoria dentro de Valkey para optimizar los tiempos de respuesta ante las consultas masivas de Grafana.

### Mapeo de Estructuras y Claves de Base de Datos:
1. Variables de Estado Instantáneo (Strings):

    - ```sistema:ram:total:``` Capacidad física de la RAM (ej. "7932 MB").

    - ```sistema:ram:libre:``` Espacio disponible actual.

    - ```sistema:ram:en_uso:``` Consumo actual.

    - ```sistema:contenedores_eliminados:``` Contador acumulativo de tipo entero, modificado atómicamente mediante comandos INCR.

2. Series Temporales Históricas (Streams):

    - sistema:historico:ram: Estructura de tipo append-only optimizada para series de tiempo. Almacena las fluctuaciones globales de memoria. Se le aplica una política de truncado automático (MaxLen: 100, Approx: true) para evitar el desbordamiento de memoria de la base de datos de telemetría.

3. Estructuras de Ranking y Clasificación (Sorted Sets / ZSETS):

    - ```sistema:ranking:ram y sistema:ranking:cpu:``` Registran los récords históricos de consumo. Se ordenan de forma nativa por su puntaje numérico (Score representando MB o % CPU). Los miembros están parametrizados bajo el formato PID: [pid] | ID: [id_corto] para evitar colisiones de nombres y mantener registros históricos únicos de contenedores ya destruidos.

## 7. Fundamentos Matemáticos de las Métricas

### 7.1 Conversión de Páginas de Memoria a Megabytes (MB)
El kernel de Linux reporta la utilización de memoria (`vsz` y `rss`) en páginas físicas. Sabiendo que una página estándar en Linux x86_64 es de 4096 Bytes (4 KB), la conversión en Go se realiza así:
* **Fórmula:** `MB = (páginas * 4096) / (1024 * 1024)`

### 7.2 Cálculo del Porcentaje de CPU
Los valores del kernel `utime` y `stime` representan el tiempo de CPU en nanosegundos. Para calcular el porcentaje de uso en tiempo real, el Daemon utiliza un cálculo diferencial entre ciclos:
1. Se suma `utime + stime` del ciclo actual.
2. Se resta el valor del ciclo anterior para obtener el `Delta CPU`.
3. Se divide entre el `Delta Tiempo` (segundos transcurridos) y se multiplica por 100.

## 8. Algoritmo de Estabilización y Balanceo
El Daemon actúa como un controlador de lazo cerrado que asegura la restricción obligatoria: **exactamente 3 contenedores de bajo consumo y 2 de alto consumo en todo momento**.

1. **Clasificación:** El sistema lee la salida de Docker y clasifica los procesos por su imagen y comandos ejecutados (`go-client` o `bc` = Alto; `sleep` = Bajo).
2. **Ordenamiento de purga:** Si hay un exceso de contenedores (por la inyección del cronjob), la lista se ordena de forma ascendente según la memoria física (`RSS_MB`). 
3. **Ejecución:** El algoritmo preserva los procesos más ligeros (los primeros 2 o 3 del arreglo) y envía una señal de destrucción (`docker rm -f`) a los contenedores sobrantes más pesados. Si hay un déficit, el Daemon ejecuta un `docker run` inmediato para cubrir la cuota.


## 9. Mecanismo de Cierre Seguro (Graceful Shutdown)
Para garantizar la estabilidad del servidor anfitrión, el Daemon implementa un manejador de señales POSIX (`SIGINT`, `SIGTERM`). Al interceptar un `Ctrl+C`, se ejecuta un protocolo de limpieza atómica:
1. **Desinstalación del Cronjob:** Se remueve la regla de `crontab` para detener la generación en segundo plano.
2. **Purga de Contenedores:** Se iteran y destruyen todos los contenedores de prueba (go-client y alpine).
3. **Apagado de Infraestructura:** Se detienen los servicios de Grafana y Valkey mediante `docker compose down`.
4. **Descarga del Kernel:** Se ejecuta el comando `rmmod continfo_pr1_so1_202307820` para liberar el espacio de memoria del núcleo de Linux de manera segura.