# Formato del estado del torneo

El estado completo vive en **un único documento JSON** (`suizo.json` por defecto,
override con `$SUIZO_FILE`). Este documento es la fuente de verdad: el CLI y la
futura web UI son vistas sobre él.

## Invariantes

- El archivo siempre es JSON válido, indentado con 2 espacios y termina en `\n`.
- La escritura es atómica: primero a un archivo temporal hermano, luego `rename`.
  Un archivo truncado jamás reemplaza al anterior.
- Un archivo inexistente equivale a un torneo vacío (no hay paso de "init").
- Los ids de jugador (`p1`, `p2`, …) son estables y nunca se reutilizan.
- Cada `Match.Result` es `""` (pendiente), `"1-0"`, `"0-1"` o `"0.5-0.5"`.

## Ejemplo

```json
{
  "name": "Suizo tournament",
  "players": [
    { "id": "p1", "name": "Ana García" },
    { "id": "p2", "name": "Beto Ruiz" },
    { "id": "p3", "name": "Carla Pérez" }
  ],
  "rounds": [
    {
      "number": 1,
      "matches": [
        { "white": "p1", "black": "p2", "result": "1-0" },
        { "white": "p3", "isBye": true }
      ]
    }
  ]
}
```

## Campos

### Tournament (raíz)

| Campo     | Tipo      | Significado                                  |
| --------- | --------- | -------------------------------------------- |
| `name`    | `string`  | nombre del torneo                            |
| `players` | `Player[]`| inscriptos, en orden de inscripción          |
| `rounds`  | `Round[]` | rondas jugadas o en curso, en orden          |

### Player

| Campo   | Tipo     | Significado                            |
| ------- | -------- | -------------------------------------- |
| `id`    | `string` | identificador estable (`p<N>`)         |
| `name`  | `string` | nombre visible; no vacío               |

### Round

| Campo     | Tipo       | Significado                                |
| --------- | ---------- | ------------------------------------------ |
| `number`  | `int`      | 1-based, creciente sin huecos              |
| `matches` | `Match[]`  | los emparejamientos de la ronda            |

### Match

| Campo    | Tipo      | Significado                                              |
| -------- | --------- | -------------------------------------------------------- |
| `white`  | `string`  | id del jugador con blancos (o con bye, si `isBye`)       |
| `black`  | `string`  | id del rival; omitido en byes                            |
| `isBye`  | `bool`    | `true` en el punto libre de la ronda (omitido si false)  |
| `result` | `string`  | `""` pendiente, `"1-0"`, `"0-1"`, `"0.5-0.5"`            |

## Tie-breakers (desempates)

La tabla (`suizo standings`) ordena por: puntos (desc), Buchholz (desc),
Buchholz Cut 1 (desc), encuentro directo (desc) y, como último recurso
determinista, id numérico ascendente (`p2` antes que `p10`).

### Buchholz

Suma de los puntos totales de cada rival contra el que el jugador fue
emparejado (victoria 1, tablas 0.5, derrota 0). Los emparejamientos sin
resultado reportado sí cuentan como rivales (con sus puntos actuales).

**Byes:** un bye no aporta rival — el punto libre no existe como contrario,
así que no contribuye nada al Buchholz de nadie (ni propio ni ajeno).

### Buchholz Cut 1

Buchholz menos el puntaje del rival más débil. Si el jugador tiene menos de
dos rivales, vale 0 (no se puede descartar nada con sentido).

### Encuentro directo

Suma de los puntos que el jugador obtuvo en partidas contra jugadores
empatados con él en puntos (el grupo de empate completo, no solo el rival
inmediato en la tabla). Partidas no reportadas y byes no suman. Si el jugador
nunca cruzó a nadie de su grupo, vale 0 — no se extrapola por subgrupos
(elección documentada: simple, determinista y suficiente para el alcance MVP;
los sistemas de subgrupos tipo FIDE quedan fuera).

## Migraciones

Ninguna todavía. Cuando el formato cambie por primera vez, se agrega acá la
regla de compatibilidad y una nota de versión.
