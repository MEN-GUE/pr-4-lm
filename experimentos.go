package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Escenario struct {
	Nombre  string
	Clave   string // prefijo de las columnas en los CSV
	PorHora float64
}

var escenarios = []Escenario{
	{"Normal", "normal", 4},
	{"Semana Santa", "santa", 6},
	{"Navidad", "navidad", 10},
}

var nombresMetricas = []string{
	"Pedidos recibidos",
	"Entregas puntuales",
	"Entregas puntuales (%)",
	"Espera promedio (min)",
	"Solicitud a entrega (min)",
	"Utilización de vehículos (%)",
	"Tiempo adicional (min)",
}

// Nombres cortos de las mismas métricas para las columnas de los CSV.
var clavesMetricas = []string{"pedidos", "puntuales", "pct", "espera", "respuesta", "util", "adicional"}

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

// Exportar corre los experimentos del informe sobre la red g y guarda los
// resultados en archivos CSV dentro de carpeta:
//   - rutas.csv: rutas por destino de las tres políticas;
//   - replicas.csv y resumen.csv: escenarios con 8 vehículos y 8 horas;
//   - flota.csv: puntualidad y espera con 4 a 20 vehículos;
//   - demanda.csv: puntualidad y espera con 1 a 12 pedidos por hora.
func Exportar(g *Grafo, carpeta string, reps int) error {
	actual, err := PoliticaActual(g)
	if err != nil {
		return err
	}
	alg, err := PoliticaAlgoritmo(g)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(carpeta, 0o755); err != nil {
		return err
	}
	const vehiculos, duracion = 8, 480.0

	rutas := [][]string{{"destino", "ruta", "ida", "t_ida", "regreso", "t_regreso", "total"}}
	politicas := []struct {
		nombre string
		pol    Politica
	}{{"actual", actual}, {"algoritmo", alg}, {"reversible", PoliticaReversible(g)}}
	for _, d := range destinos {
		for _, p := range politicas {
			if r, ok := p.pol[d]; ok {
				rutas = append(rutas, []string{d, p.nombre, strings.Join(r.Ida, " "), num(r.TIda),
					strings.Join(r.Regreso, " "), num(r.TRegreso), num(r.Total())})
			}
		}
	}

	replicas := [][]string{{"escenario", "semilla", "politica", "pedidos", "puntuales", "pct",
		"espera", "respuesta", "util", "adicional", "cola_max"}}
	encabezado := []string{"idx", "escenario", "porhora", "carga_act", "carga_alg"}
	for _, c := range clavesMetricas {
		encabezado = append(encabezado, c+"_act", c+"_act_de", c+"_alg", c+"_alg_de")
	}
	resumen := [][]string{encabezado}
	for k, e := range escenarios {
		ra, rb := CorrerParejas(actual, alg, e.PorHora, vehiculos, duracion, 1, reps)
		for i := range ra {
			replicas = append(replicas, filaReplica(e.Nombre, i+1, "actual", ra[i]),
				filaReplica(e.Nombre, i+1, "algoritmo", rb[i]))
		}
		fila := []string{strconv.Itoa(k), e.Nombre, num(e.PorHora),
			num(carga(actual, e.PorHora, vehiculos)), num(carga(alg, e.PorHora, vehiculos))}
		for _, f := range Resumir(ra, rb) {
			fila = append(fila, num(f.MediaAct), num(f.DEAct), num(f.MediaAlg), num(f.DEAlg))
		}
		resumen = append(resumen, fila)

		fmt.Printf("%-12s puntuales %5.1f%% -> %5.1f%%   espera %6.1f -> %6.1f min\n",
			e.Nombre, promedio(ra, Resultado.PorcPuntual), promedio(rb, Resultado.PorcPuntual),
			promedio(ra, espera), promedio(rb, espera))
	}

	flota := [][]string{{"vehiculos"}}
	for _, e := range escenarios {
		flota[0] = append(flota[0], e.Clave+"_pct_act", e.Clave+"_pct_alg", e.Clave+"_espera_act", e.Clave+"_espera_alg")
	}
	for v := 4; v <= 20; v++ {
		fila := []string{strconv.Itoa(v)}
		for _, e := range escenarios {
			ra, rb := CorrerParejas(actual, alg, e.PorHora, v, duracion, 1, reps)
			fila = append(fila, num(promedio(ra, Resultado.PorcPuntual)), num(promedio(rb, Resultado.PorcPuntual)),
				num(promedio(ra, espera)), num(promedio(rb, espera)))
		}
		flota = append(flota, fila)
	}

	demanda := [][]string{{"porhora", "pct_act", "pct_alg", "espera_act", "espera_alg", "adicional_act", "adicional_alg"}}
	for porHora := 1; porHora <= 12; porHora++ {
		ra, rb := CorrerParejas(actual, alg, float64(porHora), vehiculos, duracion, 1, reps)
		demanda = append(demanda, []string{strconv.Itoa(porHora),
			num(promedio(ra, Resultado.PorcPuntual)), num(promedio(rb, Resultado.PorcPuntual)),
			num(promedio(ra, espera)), num(promedio(rb, espera)),
			num(promedio(ra, adicional)), num(promedio(rb, adicional))})
	}

	archivos := map[string][][]string{"rutas.csv": rutas, "replicas.csv": replicas,
		"resumen.csv": resumen, "flota.csv": flota, "demanda.csv": demanda}
	for nombre, filas := range archivos {
		if err := escribirCSV(filepath.Join(carpeta, nombre), filas); err != nil {
			return err
		}
	}
	return nil
}

func espera(r Resultado) float64    { return r.EsperaProm }
func adicional(r Resultado) float64 { return r.TiempoAdicional }

func promedio(rs []Resultado, f func(Resultado) float64) float64 {
	var xs []float64
	for _, r := range rs {
		xs = append(xs, f(r))
	}
	m, _ := mediaDE(xs)
	return m
}

func filaReplica(escenario string, semilla int, politica string, r Resultado) []string {
	return []string{escenario, strconv.Itoa(semilla), politica, strconv.Itoa(r.Pedidos), strconv.Itoa(r.Puntuales),
		num(r.PorcPuntual()), num(r.EsperaProm), num(r.RespuestaProm), num(r.Utilizacion),
		num(r.TiempoAdicional), strconv.Itoa(r.ColaMax)}
}

func num(x float64) string { return strconv.FormatFloat(x, 'f', 2, 64) }

func escribirCSV(archivo string, filas [][]string) error {
	f, err := os.Create(archivo)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	w.WriteAll(filas)
	if err := w.Error(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
