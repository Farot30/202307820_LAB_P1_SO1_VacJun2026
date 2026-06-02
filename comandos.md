## Fase A: Gestión del Módulo de Kernel (C)
Estos comandos los ejecutas desde la terminal, estando ubicado dentro de la carpeta kernel_module/ donde están tu continfo.c y tu Makefile:

<br>
<br>
<br>

### make
Lee las instrucciones de tu Makefile, compila tu código fuente en C y genera el archivo objeto del kernel (continfo.ko).

### sudo insmod continfo.ko
Inserta (carga) tu módulo compilado directamente en el núcleo de Linux. A partir de este momento, el módulo está vivo y ejecutándose en tu sistema.

### cat /proc/continfo_pr1_so1_202307820
Lee e imprime en la terminal el contenido del archivo virtual que tu módulo creó. Esto es lo que activa la función escribir_archivo que programaste para recolectar la RAM y la estructura de procesos.

### sudo dmesg | tail -n 10
Muestra las últimas 10 líneas de los registros (logs) internos del kernel. Es vital para depurar y ver los mensajes que envías desde el código C usando la función printk().

### sudo rmmod continfo
Descarga (detiene y elimina) tu módulo del kernel. Esto ejecuta tu función modulo_exit, la cual limpia la memoria y borra el archivo de /proc. Regla de oro: Siempre ejecuta esto antes de volver a hacer make.

<br>
<br>

## Simulación de Carga: Script y Docker (Cronjob)
Estos comandos los utilizas para gestionar los contenedores de prueba y la automatización:

### chmod +x generador.sh
Le otorga permisos de ejecución a tu script en bash. Sin esto, Linux tratará el archivo solo como un documento de texto y no te dejará correrlo.

### ./generador.sh
Ejecuta tu script de forma manual e inmediata. Útil para hacer pruebas rápidas sin tener que esperar a que el cronjob se active.

### crontab -e
Abre la tabla de tareas automatizadas (cron) de tu usuario en un editor de texto (como Nano). Aquí es donde pegas la instrucción */2 * * * * ... para que el sistema ejecute tu script automáticamente cada 2 minutos en segundo plano.

### docker ps
Lista todos los contenedores que están actualmente en ejecución en tu sistema. Te permite verificar si el cronjob hizo su trabajo y si tienes la mezcla correcta de contenedores de alto y bajo consumo.

### docker rm -f $(docker ps -aq)
Comando de emergencia/limpieza. Fuerza la detención y eliminación de absolutamente todos los contenedores que existan en tu máquina. Muy útil cuando el cronjob ya generó demasiados contenedores de prueba y quieres limpiar el entorno para empezar de cero.