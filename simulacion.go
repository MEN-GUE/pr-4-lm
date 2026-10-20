package main

import (
	"container/heap"
	"math"
	"math/rand"
)

const limitePuntual = 75.0

type Pedido struct {
	Llegada float64 // minuto en que se recibe la solicitud
	Destino string
}

// GenerarPedidos genera las solicitudes de una jornada. El tiempo entre
// llegadas es exponencial con media 60/porHora minutos (transformada inversa)
// y el destino se elige al azar, con la misma probabilidad para los cuatro.
func GenerarPedidos(porHora, duracion float64, semilla int64) []Pedido {
	if porHora <= 0 {
		return nil
	}
	r := rand.New(rand.NewSource(semilla))
	media := 60 / porHora
	var pedidos []Pedido
	t := 0.0
	for {
		t += -media * math.Log(1-r.Float64())
		if t >= duracion {
			return pedidos
		}
		pedidos = append(pedidos, Pedido{Llegada: t, Destino: destinos[r.Intn(len(destinos))]})
	}
}

type Resultado struct {
	Pedidos         int
	Puntuales       int
	EsperaProm      float64 // minutos en cola antes de salir
	RespuestaProm   float64 // minutos desde la solicitud hasta la entrega
	Utilizacion     float64 // porcentaje del tiempo con los vehículos ocupados
	TiempoAdicional float64 // minutos después del cierre hasta la última entrega
	ColaMax         int
}

func (r Resultado) PorcPuntual() float64 {
	if r.Pedidos == 0 {
		return 0
	}
	return 100 * float64(r.Puntuales) / float64(r.Pedidos)
}

const (
	llegada = iota
	entrega
	regreso
)

type evento struct {
	t        float64
	tipo     int
	pedido   int
	vehiculo int
	orden    int // desempata eventos que ocurren en el mismo instante
}

// agenda es la lista de eventos pendientes, ordenada por tiempo (container/heap).
type agenda []evento

func (a agenda) Len() int { return len(a) }
func (a agenda) Less(i, j int) bool {
	if a[i].t != a[j].t {
		return a[i].t < a[j].t
	}
	return a[i].orden < a[j].orden
}
func (a agenda) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a *agenda) Push(x any)   { *a = append(*a, x.(evento)) }
func (a *agenda) Pop() any {
	viejo := *a
	e := viejo[len(viejo)-1]
	*a = viejo[:len(viejo)-1]
	return e
}

// Simular atiende los pedidos con la política de rutas indicada. Los eventos
// se procesan en orden de tiempo hasta que no queda ninguno, así que la
// simulación sigue después del cierre hasta entregar todo y volver a la bodega.
func Simular(pedidos []Pedido, pol Politica, vehiculos int, duracion float64) Resultado {
	res := Resultado{Pedidos: len(pedidos)}
	if len(pedidos) == 0 || vehiculos <= 0 {
		return res
	}

	var ag agenda
	orden := 0
	programar := func(t float64, tipo, p, v int) {
		heap.Push(&ag, evento{t: t, tipo: tipo, pedido: p, vehiculo: v, orden: orden})
		orden++
	}
	for i, p := range pedidos {
		programar(p.Llegada, llegada, i, -1)
	}

	// libres funciona como cola: el primero es el que lleva más tiempo esperando.
	libres := make([]int, vehiculos)
	for v := range libres {
		libres[v] = v
	}
	var cola []int
	salida := make([]float64, len(pedidos))
	entregado := make([]float64, len(pedidos))
	ocupado, ultimaEntrega, ultimoRegreso := 0.0, 0.0, 0.0

	despachar := func(t float64, p, v int) {
		r := pol[pedidos[p].Destino]
		salida[p] = t
		ocupado += r.TIda + r.TRegreso
		programar(t+r.TIda, entrega, p, v)
	}

	for ag.Len() > 0 {
		e := heap.Pop(&ag).(evento)
		switch e.tipo {
		case llegada:
			if len(libres) > 0 {
				despachar(e.t, e.pedido, libres[0])
				libres = libres[1:]
			} else {
				cola = append(cola, e.pedido)
				if len(cola) > res.ColaMax {
					res.ColaMax = len(cola)
				}
			}
		case entrega:
			entregado[e.pedido] = e.t
			ultimaEntrega = e.t
			programar(e.t+pol[pedidos[e.pedido].Destino].TRegreso, regreso, e.pedido, e.vehiculo)
		case regreso:
			ultimoRegreso = e.t
			if len(cola) > 0 {
				despachar(e.t, cola[0], e.vehiculo)
				cola = cola[1:]
			} else {
				libres = append(libres, e.vehiculo)
			}
		}
	}

	sumaEspera, sumaRespuesta := 0.0, 0.0
	for i, p := range pedidos {
		sumaEspera += salida[i] - p.Llegada
		respuesta := entregado[i] - p.Llegada
		sumaRespuesta += respuesta
		if respuesta <= limitePuntual+1e-9 { // tolerancia por redondeo
			res.Puntuales++
		}
	}
	n := float64(len(pedidos))
	res.EsperaProm = sumaEspera / n
	res.RespuestaProm = sumaRespuesta / n
	res.Utilizacion = 100 * ocupado / (float64(vehiculos) * math.Max(duracion, ultimoRegreso))
	res.TiempoAdicional = math.Max(0, ultimaEntrega-duracion)
	return res
}
