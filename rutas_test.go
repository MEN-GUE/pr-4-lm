package main

import "testing"

func cargarRed(t *testing.T) *Grafo {
	t.Helper()
	g, err := LeerMatriz("Matriz_Logistica_del_Valle_40_nodos.csv")
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestLecturaMatriz(t *testing.T) {
	g := cargarRed(t)
	dobles, distintos, unSentido := g.Resumen()
	if len(g.Nombres) != 40 || dobles != 63 || distintos != 14 || unSentido != 18 {
		t.Errorf("resumen inesperado: %d nodos, %d pares dobles, %d distintos, %d de un sentido",
			len(g.Nombres), dobles, distintos, unSentido)
	}
	if _, ok := g.Tiempo("N17", "N01"); ok {
		t.Error("N17 -> N01 no debería existir")
	}
}

func TestRutasActuales(t *testing.T) {
	g := cargarRed(t)
	esperado := map[string]float64{"N08": 59, "N19": 33, "N34": 42, "N40": 93}
	p, err := PoliticaActual(g)
	if err != nil {
		t.Fatal(err)
	}
	for d, r := range p {
		if r.TIda != esperado[d] || r.TRegreso != esperado[d] {
			t.Errorf("%s: ida %g y regreso %g, se esperaba %g", d, r.TIda, r.TRegreso, esperado[d])
		}
	}
}

func TestDijkstra(t *testing.T) {
	g := cargarRed(t)
	casos := []struct {
		desde, hasta string
		tiempo       float64
	}{
		{"N01", "N08", 51}, {"N08", "N01", 59},
		{"N01", "N19", 18}, {"N19", "N01", 24},
		{"N01", "N34", 24}, {"N34", "N01", 34},
		{"N01", "N40", 47}, {"N40", "N01", 61},
	}
	for _, c := range casos {
		camino, tiempo, ok := g.CaminoMasCorto(c.desde, c.hasta)
		if !ok || tiempo != c.tiempo {
			t.Errorf("%s -> %s: %g, se esperaba %g", c.desde, c.hasta, tiempo, c.tiempo)
			continue
		}
		if suma, err := g.TiempoRuta(camino); err != nil || suma != tiempo {
			t.Errorf("%s -> %s: el camino %v no suma %g", c.desde, c.hasta, camino, tiempo)
		}
	}
}

func TestIdaOptimaNoSeInvierte(t *testing.T) {
	g := cargarRed(t)
	for _, d := range destinos {
		ida, _, _ := g.CaminoMasCorto(bodega, d)
		if _, err := g.TiempoRuta(Invertir(ida)); err == nil {
			t.Errorf("la ida a %s sí se puede recorrer al revés", d)
		}
	}
}

func TestRutaReversible(t *testing.T) {
	g := cargarRed(t)
	esperado := map[string][2]float64{"N08": {57, 59}, "N19": {24, 24}, "N34": {34, 34}, "N40": {54, 63}}
	for d, r := range PoliticaReversible(g) {
		if r.TIda != esperado[d][0] || r.TRegreso != esperado[d][1] {
			t.Errorf("%s: ida %g y regreso %g, se esperaba %v", d, r.TIda, r.TRegreso, esperado[d])
		}
	}
}

func TestDesvioPorCierre(t *testing.T) {
	g := cargarRed(t)
	g.W[g.Indice("N17")][g.Indice("N25")] = -1
	g.W[g.Indice("N25")][g.Indice("N17")] = -1
	p, err := PoliticaActual(g)
	if err != nil {
		t.Fatal(err)
	}
	if r := p["N34"]; r.TIda != 49 || r.TRegreso != 61 {
		t.Errorf("N34 con desvío: ida %g y regreso %g, se esperaba 49 y 61", r.TIda, r.TRegreso)
	}
}
