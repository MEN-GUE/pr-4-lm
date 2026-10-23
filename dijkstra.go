package main

import "math"

// Dijkstra sobre la matriz de adyacencia. En cada paso se toma el nodo no
// visitado con menor distancia, O(V^2), que para 40 nodos es suficiente.
// Si hay empate se queda el nodo de menor índice.
func (g *Grafo) Dijkstra(origen int) (dist []float64, previo []int) {
	n := len(g.Nombres)
	dist = make([]float64, n)
	previo = make([]int, n)
	visitado := make([]bool, n)
	for i := range dist {
		dist[i] = math.Inf(1)
		previo[i] = -1
	}
	dist[origen] = 0

	for {
		u := -1
		for i := 0; i < n; i++ {
			if !visitado[i] && !math.IsInf(dist[i], 1) && (u == -1 || dist[i] < dist[u]) {
				u = i
			}
		}
		if u == -1 {
			break
		}
		visitado[u] = true
		for v := 0; v < n; v++ {
			if g.W[u][v] > 0 && dist[u]+g.W[u][v] < dist[v] {
				dist[v] = dist[u] + g.W[u][v]
				previo[v] = u
			}
		}
	}
	return dist, previo
}

func (g *Grafo) CaminoMasCorto(desde, hasta string) ([]string, float64, bool) {
	o, d := g.Indice(desde), g.Indice(hasta)
	if o < 0 || d < 0 {
		return nil, 0, false
	}
	dist, previo := g.Dijkstra(o)
	if math.IsInf(dist[d], 1) {
		return nil, 0, false
	}
	var camino []string
	for v := d; v != -1; v = previo[v] {
		camino = append(camino, g.Nombres[v])
	}
	return Invertir(camino), dist[d], true
}

// RutaReversible busca el mejor camino que solo usa calles de doble sentido,
// de modo que se pueda volver por el mismo camino. Cada tramo cuesta la suma
// de sus dos sentidos, así se minimiza directamente el viaje de ida y vuelta.
func (g *Grafo) RutaReversible(desde, hasta string) ([]string, bool) {
	n := len(g.Nombres)
	aux := &Grafo{Nombres: g.Nombres, W: make([][]float64, n)}
	for i := range aux.W {
		aux.W[i] = make([]float64, n)
		for j := range aux.W[i] {
			aux.W[i][j] = -1
			if g.W[i][j] > 0 && g.W[j][i] > 0 {
				aux.W[i][j] = g.W[i][j] + g.W[j][i]
			}
		}
	}
	camino, _, ok := aux.CaminoMasCorto(desde, hasta)
	return camino, ok
}
