package main

import (
	"math"
	"testing"
)

func TestSinEsperaConMuchosVehiculos(t *testing.T) {
	g := cargarRed(t)
	actual, err := PoliticaActual(g)
	if err != nil {
		t.Fatal(err)
	}
	pedidos := GenerarPedidos(6, 480, 1)
	r := Simular(pedidos, actual, 200, 480)
	if r.EsperaProm != 0 {
		t.Errorf("espera promedio %g, se esperaba 0", r.EsperaProm)
	}
	// La ida actual a N40 tarda 93 min: esos pedidos nunca son puntuales.
	esperado := 0
	for _, p := range pedidos {
		if p.Destino != "N40" {
			esperado++
		}
	}
	if r.Puntuales != esperado {
		t.Errorf("%d puntuales, se esperaban %d", r.Puntuales, esperado)
	}
}

// Con cola FIFO cada pedido sale con el primer vehículo que se desocupa, así
// que la espera también se puede calcular directamente, pedido por pedido.
func TestComparacionConCalculoDirecto(t *testing.T) {
	g := cargarRed(t)
	pol, err := PoliticaAlgoritmo(g)
	if err != nil {
		t.Fatal(err)
	}
	for _, porHora := range []float64{4, 6, 10} {
		pedidos := GenerarPedidos(porHora, 480, 7)
		libre := make([]float64, 8)
		suma, puntuales := 0.0, 0
		for _, p := range pedidos {
			v := 0
			for k := range libre {
				if libre[k] < libre[v] {
					v = k
				}
			}
			salida := math.Max(p.Llegada, libre[v])
			r := pol[p.Destino]
			libre[v] = salida + r.TIda + r.TRegreso
			suma += salida - p.Llegada
			if salida+r.TIda-p.Llegada <= limitePuntual {
				puntuales++
			}
		}
		res := Simular(pedidos, pol, 8, 480)
		if math.Abs(res.EsperaProm-suma/float64(len(pedidos))) > 1e-9 || res.Puntuales != puntuales {
			t.Errorf("%g pedidos/h: simulación (%g, %d) y cálculo directo (%g, %d) no coinciden",
				porHora, res.EsperaProm, res.Puntuales, suma/float64(len(pedidos)), puntuales)
		}
	}
}
