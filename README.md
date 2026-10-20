# Práctica 4: Logística del Valle

CC2017 Modelación y Simulación, Universidad del Valle de Guatemala.
Juan Fernando Menéndez Guerra (17444).

Aplicación para comparar las rutas actuales de reparto con las que calcula
Dijkstra y simular la atención de pedidos con la flota disponible.

## Requisitos

- Go 1.22 o más reciente (solo se usa la biblioteca estándar).
- Graphviz, para ver el grafo en la página (`neato`). Es opcional.

## Uso

```
go run .
```

Luego abrir <http://localhost:8080>. Desde la página se puede:

- elegir el escenario (normal, Semana Santa, Navidad o personalizado) y
  cambiar los pedidos por hora, los vehículos, las horas de recepción, la
  semilla y las repeticiones;
- simular las dos políticas de rutas con los mismos pedidos;
- cambiar el tiempo de un tramo o cerrarlo, y ver cómo se recalculan las rutas.

Opciones: `-puerto` (8080 por defecto) y `-matriz` (ruta del CSV).

Pruebas: `go test ./...`
