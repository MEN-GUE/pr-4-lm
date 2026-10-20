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

## Informe

El informe está en `informe/informe.pdf`, con su fuente en `informe/informe.tex`.
Se compila con pdfLaTeX (por ejemplo en Overleaf) junto con las carpetas
`informe/figuras` e `informe/datos`.

## Datos y figuras del informe

Los experimentos usan 8 vehículos, 8 horas de recepción y las semillas 1 a 10
(se cambia con `-repeticiones`). Los resultados quedan en CSV:

```
go run . -exportar informe/datos
go run . -cerrar N17-N25,N25-N17 -exportar informe/datos/incidente
```

El grafo se exporta en formato DOT y se dibuja con Graphviz:

```
go run . -dot informe/figuras
neato -n2 -Tpdf informe/figuras/grafo.dot -o informe/figuras/grafo.pdf
neato -n2 -Tpdf informe/figuras/grafo_rutas.dot -o informe/figuras/grafo_rutas.pdf
```

Pruebas: `go test ./...`
