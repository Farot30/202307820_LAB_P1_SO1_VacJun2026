#include <linux/sysinfo.h>
#include <linux/module.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/mm.h>
#include <linux/sched/signal.h>
#include <linux/sched.h>

#define PROC_NAME "continfo_pr1_so1_202307820" 

MODULE_LICENSE("GPL");
MODULE_AUTHOR("Farot");
MODULE_DESCRIPTION("Sonda de Kernel para Telemetria de Contenedores");
MODULE_VERSION("1.0");

// Esta función se ejecuta cada vez que alguien hace 'cat' al archivo en /proc
static int escribir_archivo(struct seq_file *m, void *v) {
    struct sysinfo i;
    long total_ram_mb, free_ram_mb, used_ram_mb;
    struct task_struct *task;

    // Variables para la memoria de cada proceso
    unsigned long vsz = 0;
    unsigned long rss = 0;
    // Variables para el CPU de cada proceso
    u64 utime, stime;

    si_meminfo(&i);

    total_ram_mb = (i.totalram * i.mem_unit) / (1024 * 1024);
    free_ram_mb = (i.freeram * i.mem_unit) / (1024 * 1024);
    used_ram_mb = total_ram_mb - free_ram_mb;

    seq_printf(m, "¡Módulo de kernel de Farot cargado correctamente!\n");
    seq_printf(m, "Total de memoria RAM: %ld MB\n", total_ram_mb);
    seq_printf(m, "Memoria RAM libre: %ld MB\n", free_ram_mb);
    seq_printf(m, "Memoria RAM en uso: %ld MB\n\n", used_ram_mb);

    seq_printf(m, "--- Lista de Procesos ---\n");

    for_each_process(task) {
        vsz = 0;
        rss = 0;

        // Extraer VSZ y RSS (se debe bloquear/desbloquear para evitar condiciones de carrera)
        if (task->mm) {
            vsz = task->mm->total_vm; // Memoria Virtual (en páginas)
            rss = get_mm_rss(task->mm); // Memoria Física residente (en páginas)
        }

        // Extraer tiempo de CPU (User time y System time en nanosegundos)
        utime = task->utime;
        stime = task->stime;

        // Formateamos la salida. El Daemon en Go tendrá que convertir las páginas a MB y los tiempos a %
        seq_printf(m, "PID: %d | Nombre: %s | VSZ: %lu | RSS: %lu | Utime: %llu | Stime: %llu\n", 
                   task->pid, task->comm, vsz, rss, utime, stime);
    }

    return 0;
}

// Función que se llama al abrir el archivo
static int al_abrir(struct inode *inode, struct file *file) {
    return single_open(file, escribir_archivo, NULL);
}

// Estructura de operaciones para kernels modernos (5.6+)
static const struct proc_ops operaciones_proc = {
    .proc_open = al_abrir,
    .proc_read = seq_read,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};

// Función de inicialización (se ejecuta al hacer insmod)
static int __init modulo_init(void) {
    proc_create(PROC_NAME, 0, NULL, &operaciones_proc);
    printk(KERN_INFO "Modulo %s cargado exitosamente en el kernel.\n", PROC_NAME);
    return 0;
}

// Función de limpieza (se ejecuta al hacer rmmod)
static void __exit modulo_exit(void) {
    remove_proc_entry(PROC_NAME, NULL);
    printk(KERN_INFO "Modulo %s descargado del kernel.\n", PROC_NAME);
}

module_init(modulo_init);
module_exit(modulo_exit);