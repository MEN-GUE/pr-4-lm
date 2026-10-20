package main

import "fmt"

const bodega = "N01"

var destinos = []string{"N08", "N19", "N34", "N40"}

// Rutas de las hojas de despacho. El regreso es la misma ruta al revés.
var rutasActuales = map[string][]string{
	"N08": {"N01", "N02", "N03", "N04", "N05", "N06", "N07", "N08"},
	"N19": {"N01", "N02", "N03", "N11", "N19"},
	"N34": {"N01", "N09", "N17", "N25", "N33", "N34"},
	"N40": {"N01", "N02", "N03", "N04", "N05", "N06", "N07", "N08", "N16", "N24", "N32", "N40"},
}

type Ruta struct {
	Ida, Regreso   []string
	TIda, TRegreso float64
}

func (r Ruta) Total() float64 { return r.TIda + r.TRegreso }

// Politica indica qué ruta de ida y de regreso se usa para cada destino.
type Politica map[string]Ruta

func nuevaRuta(g *Grafo, ida, regreso []string) Ruta {
	ti, _ := g.TiempoRuta(ida)
	tr, _ := g.TiempoRuta(regreso)
	return Ruta{Ida: ida, Regreso: regreso, TIda: ti, TRegreso: tr}
}

// PoliticaActual usa las rutas de las hojas. Si algún tramo ya no existe, el
// conductor toma el desvío más corto entre los extremos de ese tramo y luego
// sigue con su ruta de siempre.
func PoliticaActual(g *Grafo) (Politica, error) {
	p := Politica{}
	for _, d := range destinos {
		ida, err := g.conDesvios(rutasActuales[d])
		if err != nil {
			return nil, err
		}
		regreso, err := g.conDesvios(Invertir(rutasActuales[d]))
		if err != nil {
			return nil, err
		}
		p[d] = nuevaRuta(g, ida, regreso)
	}
	return p, nil
}

func (g *Grafo) conDesvios(ruta []string) ([]string, error) {
	nueva := []string{ruta[0]}
	for k := 0; k+1 < len(ruta); k++ {
		if _, ok := g.Tiempo(ruta[k], ruta[k+1]); ok {
			nueva = append(nueva, ruta[k+1])
			continue
		}
		desvio, _, ok := g.CaminoMasCorto(ruta[k], ruta[k+1])
		if !ok {
			return nil, fmt.Errorf("no hay forma de ir de %s a %s", ruta[k], ruta[k+1])
		}
		nueva = append(nueva, desvio[1:]...)
	}
	return nueva, nil
}

// PoliticaAlgoritmo calcula por separado la ida (desde la bodega) y el
// regreso (hacia la bodega) con Dijkstra.
func PoliticaAlgoritmo(g *Grafo) (Politica, error) {
	p := Politica{}
	for _, d := range destinos {
		ida, _, ok1 := g.CaminoMasCorto(bodega, d)
		regreso, _, ok2 := g.CaminoMasCorto(d, bodega)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("no hay ruta entre %s y %s", bodega, d)
		}
		p[d] = nuevaRuta(g, ida, regreso)
	}
	return p, nil
}

// PoliticaReversible solo se usa para comparar rutas; si por algún cierre no
// existe un camino de doble sentido, ese destino queda sin ruta.
func PoliticaReversible(g *Grafo) Politica {
	p := Politica{}
	for _, d := range destinos {
		if ida, ok := g.RutaReversible(bodega, d); ok {
			p[d] = nuevaRuta(g, ida, Invertir(ida))
		}
	}
	return p
}

func esDestino(nombre string) bool {
	for _, d := range destinos {
		if d == nombre {
			return true
		}
	}
	return false
}
