#!/bin/bash
# Despliegue aleatorio de 5 contenedores para el Proyecto 1 [cite: 9, 59]

echo "Iniciando despliegue de contenedores de prueba..."

for i in {1..5}
do
    # Generar un número aleatorio entre 1 y 3
    TIPO=$(( ( RANDOM % 3 ) + 1 ))

    case $TIPO in
        1)
            echo "Desplegando: Alto Consumo RAM (go-client)"
            docker run -d roldyoran/go-client 
            ;;
        2)
            echo "Desplegando: Alto Consumo CPU (alpine con cálculo)"
            docker run -d alpine sh -c "while true; do echo '2^20' | bc > /dev/null; sleep 2; done" 
            ;;
        3)
            echo "Desplegando: Bajo Consumo (alpine sleep)"
            docker run -d alpine sleep 240 
            ;;
    esac
done

echo "Despliegue finalizado."