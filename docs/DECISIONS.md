# Decisiones

Registro informal de decisiones de diseño, en orden. Cada una: contexto,
decisión, y por qué. (Formato ligero a propósito — si alguna decisión alguna vez
necesita proceso formal, se migra a ADR.)

## 1. Go como lenguaje

**Contexto:** gestor de torneos amateur que va a correr en la notebook de un
organizador de club, quizás sin experiencia técnica, quizás offline.

**Decisión:** Go (1.27.x, estable vigente).

**Por qué:**
- Un solo binario estático, sin runtime: `go build` y copiar. Nada que instalar,
  ni Node, ni Python, ni versiones.
- `net/http` y `encoding/json` en stdlib cubren todo el roadmap (CLI + mini web
  UI) sin framework.
- Compilación y tests rápidos; iteración cómoda para contribuidores ocasionales.

## 2. Estado en un único archivo JSON

**Contexto:** hay que persistir jugadores, rondas y resultados. Opciones
reales: SQLite, archivos por torneo, JSON único, "una base en la nube".

**Decisión:** un único documento JSON (`suizo.json`, override con `$SUIZO_FILE`),
escritura atómica (temp + rename).

**Por qué:**
- Un torneo de club = 8 a 40 jugadores. Una base de datos es operar un servicio
  para gestionar una hoja de cálculo.
- El archivo es inspeccionable con `cat`, copiable como backup, y un diff dice
  exactamente qué cambió.
- Escritura atómica: si se corta la luz a mitad de ronda, el archivo viejo sigue
  íntegro. El peor caso de corrupción es un write parcial nunca renombrado.
- Salir de esto más adelante (si algún día hiciera falta) es un cambio aislado
  detrás de `store.go`.

## 3. Sistema suizo como formato del torneo

**Contexto:** el público son torneos amateur (futbolito del club, abierto de
ajedrez de barrio). No hay calendario de ida y vuelta ni llaves de eliminación.

**Decisión:** emparejamiento suizo: rondas fijas, jugadores con puntajes
similares se cruzan entre sí, sin eliminados, nadie repite rival.

**Por qué:**
- Nadie queda afuera después de perder: en un torneo amateur eliminar gente en
  la ronda 1 significa que un tercio de los inscriptos se va a casa a los 20
  minutos.
- Escala: el mismo algoritmo sirve para 8 o 80 jugadores.
- Los tie-breakers (Buchholz, etc.) son parte natural del formato y quedan como
  feature incremental, no como rediseño.

## 4. CLI primero, web después

**Contexto:** ¿empezar por la web UI o por el CLI?

**Decisión:** la lógica vive en paquete único consumible por ambos; el CLI es la
primera cara del producto y la web (issue #4) es una vista sobre el mismo estado.

**Por qué:**
- El organizador necesita registrar resultados en el momento, rápido y con
  teclado; el CLI cubre eso desde el día uno.
- Separar lógica de presentación fuerza que el estado JSON sea la única fuente
  de verdad — lo que también hace trivial testing.
