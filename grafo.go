package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// W[i][j] es el tiempo en minutos para ir de i a j. Vale -1 si no hay
// conexión directa y 0 en la diagonal.
type Grafo struct {
	Nombres []string
	W       [][]float64
}

func LeerMatriz(archivo string) (*Grafo, error) {
	f, err := os.Open(archivo)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	filas, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(filas) < 2 {
		return nil, fmt.Errorf("la matriz no tiene datos")
	}

	nombres := filas[0][1:]
	n := len(nombres)
	if len(filas)-1 != n {
		return nil, fmt.Errorf("la matriz no es cuadrada: %d filas y %d columnas", len(filas)-1, n)
	}

	g := &Grafo{Nombres: nombres, W: make([][]float64, n)}
	for i, fila := range filas[1:] {
		if strings.TrimSpace(fila[0]) != nombres[i] {
			return nil, fmt.Errorf("la fila %s no corresponde a la columna %s", fila[0], nombres[i])
		}
		g.W[i] = make([]float64, n)
		for j, celda := range fila[1:] {
			celda = strings.TrimSpace(celda)
			// Una celda sin número (vacía o con un guion) indica que no hay conexión.
			if !strings.ContainsAny(celda, "0123456789") {
				g.W[i][j] = -1
				continue
			}
			v, err := strconv.ParseFloat(celda, 64)
			if err != nil {
				return nil, fmt.Errorf("valor %q no válido en %s -> %s", celda, nombres[i], nombres[j])
			}
			if (i == j && v != 0) || (i != j && v <= 0) {
				return nil, fmt.Errorf("tiempo %v no válido en %s -> %s", v, nombres[i], nombres[j])
			}
			g.W[i][j] = v
		}
	}
	return g, nil
}

func (g *Grafo) Indice(nombre string) int {
	for i, n := range g.Nombres {
		if n == nombre {
			return i
		}
	}
	return -1
}

func (g *Grafo) Tiempo(desde, hasta string) (float64, bool) {
	i, j := g.Indice(desde), g.Indice(hasta)
	if i < 0 || j < 0 || g.W[i][j] <= 0 {
		return 0, false
	}
	return g.W[i][j], true
}

func (g *Grafo) TiempoRuta(ruta []string) (float64, error) {
	total := 0.0
	for k := 0; k+1 < len(ruta); k++ {
		t, ok := g.Tiempo(ruta[k], ruta[k+1])
		if !ok {
			return 0, fmt.Errorf("no hay conexión directa de %s a %s", ruta[k], ruta[k+1])
		}
		total += t
	}
	return total, nil
}

func Invertir(ruta []string) []string {
	r := make([]string, len(ruta))
	for i, n := range ruta {
		r[len(ruta)-1-i] = n
	}
	return r
}

func (g *Grafo) Copia() *Grafo {
	c := &Grafo{Nombres: g.Nombres, W: make([][]float64, len(g.W))}
	for i := range g.W {
		c.W[i] = append([]float64(nil), g.W[i]...)
	}
	return c
}

// Cambios lista los tramos cuyo tiempo es distinto al de la red original.
func (g *Grafo) Cambios(original *Grafo) []string {
	var cambios []string
	for i := range g.W {
		for j := range g.W[i] {
			if g.W[i][j] == original.W[i][j] {
				continue
			}
			nuevo := "cerrado"
			if g.W[i][j] > 0 {
				nuevo = fmt.Sprintf("%g min", g.W[i][j])
			}
			cambios = append(cambios, fmt.Sprintf("De %s a %s: %s (antes %g min)",
				g.Nombres[i], g.Nombres[j], nuevo, original.W[i][j]))
		}
	}
	return cambios
}

// Cuenta los pares de nodos conectados en ambos sentidos (y cuántos de ellos
// tienen un tiempo distinto en cada sentido) y los tramos de un solo sentido.
func (g *Grafo) Resumen() (dobles, distintos, unSentido int) {
	n := len(g.Nombres)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			ida, vuelta := g.W[i][j] > 0, g.W[j][i] > 0
			if ida && vuelta {
				dobles++
				if g.W[i][j] != g.W[j][i] {
					distintos++
				}
			} else if ida || vuelta {
				unSentido++
			}
		}
	}
	return
}

// DOT genera el grafo para Graphviz (neato -n2). Los nodos se acomodan en una
// cuadrícula de 8 columnas según su numeración; la posición no es geográfica.
// resaltar asigna un color a algunos tramos (i, j), por ejemplo una ruta.
func (g *Grafo) DOT(resaltar map[[2]int]string) string {
	const columnas, sep = 8, 90
	n := len(g.Nombres)
	filas := (n + columnas - 1) / columnas

	pos := func(i int) (int, int) {
		return (i % columnas) * sep, (filas - 1 - i/columnas) * sep
	}

	var b strings.Builder
	b.WriteString("digraph red {\n")
	b.WriteString("  graph [splines=true, esep=\"+14\", outputorder=edgesfirst, pad=0.2]\n")
	b.WriteString("  node [shape=circle, fixedsize=true, width=0.5, style=filled, fillcolor=white, fontname=Helvetica, fontsize=12]\n")
	b.WriteString("  edge [fontname=Helvetica, fontsize=12, arrowsize=0.7]\n")

	for i, nombre := range g.Nombres {
		x, y := pos(i)
		estilo := ""
		if nombre == bodega {
			estilo = ", shape=box, fillcolor=gray75"
		} else if esDestino(nombre) {
			estilo = ", fillcolor=gray88, penwidth=2"
		}
		fmt.Fprintf(&b, "  %s [pos=\"%d,%d\"%s]\n", nombre, x, y, estilo)
	}

	usados := map[[2]int]bool{}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			ida, vuelta := g.W[i][j] > 0, g.W[j][i] > 0
			if !ida && !vuelta {
				continue
			}
			a, c := g.Nombres[i], g.Nombres[j]
			var estilo, texto string
			switch {
			case ida && vuelta && g.W[i][j] == g.W[j][i]:
				estilo, texto = "dir=none", fmt.Sprintf("%g", g.W[i][j])
			case ida && vuelta:
				estilo, texto = "dir=none, color=blue, fontcolor=blue", fmt.Sprintf("%g/%g", g.W[i][j], g.W[j][i])
			case ida:
				estilo, texto = "color=red, fontcolor=red", fmt.Sprintf("%g", g.W[i][j])
			default:
				a, c = c, a
				estilo, texto = "color=red, fontcolor=red", fmt.Sprintf("%g", g.W[j][i])
			}

			// La etiqueta va a la mitad del tramo, excepto si el tramo pasa por
			// encima de otro nodo o si cruza otra diagonal en el mismo punto
			// medio; ahí se acerca al origen para que no se encime.
			xi, yi := pos(i)
			xj, yj := pos(j)
			medio := [2]int{xi + xj, yi + yj}
			largo := (xi-xj)*(xi-xj)+(yi-yj)*(yi-yj) > 2*sep*sep
			attr := fmt.Sprintf("%s, label=\"%s\"", estilo, texto)
			if largo || usados[medio] {
				attr = fmt.Sprintf("%s, taillabel=\"%s\", labelangle=0, labeldistance=2.5", estilo, texto)
			}
			usados[medio] = true

			color := resaltar[[2]int{i, j}]
			if otro := resaltar[[2]int{j, i}]; otro != "" {
				if color != "" && color != otro {
					color += ":" + otro
				} else {
					color = otro
				}
			}
			if color != "" {
				attr += fmt.Sprintf(", color=\"%s\", penwidth=3", color)
			}
			fmt.Fprintf(&b, "  %s -> %s [%s]\n", a, c, attr)
		}
	}
	b.WriteString("}\n")
	return b.String()
}
