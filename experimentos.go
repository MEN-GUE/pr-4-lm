package main

import "math"

type Escenario struct {
	Nombre  string
	PorHora float64
}

var escenarios = []Escenario{{"Normal", 4}, {"Semana Santa", 6}, {"Navidad", 10}}

var nombresMetricas = []string{
	"Pedidos recibidos",
	"Entregas puntuales",
	"Entregas puntuales (%)",
	"Espera promedio (min)",
	"Solicitud a entrega (min)",
	"Utilización de vehículos (%)",
	"Tiempo adicional (min)",
}

func (r Resultado) Valores() []float64 {
	return []float64{float64(r.Pedidos), float64(r.Puntuales), r.PorcPuntual(),
		r.EsperaProm, r.RespuestaProm, r.Utilizacion, r.TiempoAdicional}
}

// CorrerParejas hace n repeticiones. En cada una se generan los pedidos con
// la semilla semilla+k y esa misma lista se simula con las dos políticas.
func CorrerParejas(actual, alg Politica, porHora float64, vehiculos int, duracion float64, semilla int64, n int) ([]Resultado, []Resultado) {
	var ra, rb []Resultado
	for k := 0; k < n; k++ {
		pedidos := GenerarPedidos(porHora, duracion, semilla+int64(k))
		ra = append(ra, Simular(pedidos, actual, vehiculos, duracion))
		rb = append(rb, Simular(pedidos, alg, vehiculos, duracion))
	}
	return ra, rb
}

func mediaDE(xs []float64) (media, de float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	for _, x := range xs {
		media += x
	}
	media /= float64(len(xs))
	if len(xs) < 2 {
		return media, 0
	}
	for _, x := range xs {
		de += (x - media) * (x - media)
	}
	return media, math.Sqrt(de / float64(len(xs)-1))
}

type FilaResumen struct {
	Metrica         string
	MediaAct, DEAct float64
	MediaAlg, DEAlg float64
	Diferencia      float64 // algoritmo menos actual
}

func Resumir(ra, rb []Resultado) []FilaResumen {
	filas := make([]FilaResumen, len(nombresMetricas))
	for k, nombre := range nombresMetricas {
		var xa, xb []float64
		for i := range ra {
			xa = append(xa, ra[i].Valores()[k])
			xb = append(xb, rb[i].Valores()[k])
		}
		f := FilaResumen{Metrica: nombre}
		f.MediaAct, f.DEAct = mediaDE(xa)
		f.MediaAlg, f.DEAlg = mediaDE(xb)
		f.Diferencia = f.MediaAlg - f.MediaAct
		filas[k] = f
	}
	return filas
}
