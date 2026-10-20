package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"
)

//go:embed plantilla.html
var textoPlantilla string

var plantilla = template.Must(template.New("pagina").Funcs(template.FuncMap{
	"ruta": func(r []string) string { return strings.Join(r, " → ") },
	"num":  func(x float64) string { return strconv.FormatFloat(x, 'f', 1, 64) },
}).Parse(textoPlantilla))

type Parametros struct {
	Escenario    string
	PorHora      float64
	Vehiculos    int
	Horas        float64
	Semilla      int64
	Repeticiones int
}

// App guarda el estado de la aplicación: la red original, la red con los
// cambios hechos desde la página y el último resultado simulado.
type App struct {
	mu       sync.Mutex
	original *Grafo
	red      *Grafo
	params   Parametros
	simulado Parametros
	resumen  []FilaResumen
	carga    [2]float64
	mensajes []string
	error    string
}

func NuevaApp(g *Grafo) *App {
	return &App{
		original: g,
		red:      g.Copia(),
		params:   Parametros{Escenario: "Normal", PorHora: 4, Vehiculos: 8, Horas: 8, Semilla: 1, Repeticiones: 10},
	}
}

func (a *App) Registrar(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", a.inicio)
	mux.HandleFunc("POST /simular", a.simular)
	mux.HandleFunc("POST /tramo", a.tramo)
	mux.HandleFunc("POST /restaurar", a.restaurar)
	mux.HandleFunc("GET /grafo.svg", a.grafo)
}

type datosPagina struct {
	P, Simulado                   Parametros
	Escenarios                    []Escenario
	Nombres, Destinos             []string
	Actual, Algoritmo, Reversible Politica
	Desvios                       map[string]bool
	Resumen                       []FilaResumen
	Carga                         [2]float64
	Mensajes, Cambios             []string
	Error                         string
	HayGraphviz                   bool
}

func (a *App) inicio(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	d := datosPagina{
		P:          a.params,
		Simulado:   a.simulado,
		Escenarios: escenarios,
		Nombres:    a.red.Nombres,
		Destinos:   destinos,
		Reversible: PoliticaReversible(a.red),
		Resumen:    a.resumen,
		Carga:      a.carga,
		Mensajes:   a.mensajes,
		Cambios:    a.red.Cambios(a.original),
		Error:      a.error,
	}
	var err error
	if d.Actual, err = PoliticaActual(a.red); err != nil {
		d.Error = err.Error()
	}
	if d.Algoritmo, err = PoliticaAlgoritmo(a.red); err != nil {
		d.Error = err.Error()
	}
	d.Desvios = map[string]bool{}
	for _, dest := range destinos {
		ruta := d.Actual[dest]
		d.Desvios[dest] = ruta.Ida != nil && (!slices.Equal(ruta.Ida, rutasActuales[dest]) ||
			!slices.Equal(ruta.Regreso, Invertir(rutasActuales[dest])))
	}
	a.mensajes, a.error = nil, ""
	a.mu.Unlock()

	_, err = exec.LookPath("neato")
	d.HayGraphviz = err == nil

	var b bytes.Buffer
	if err := plantilla.Execute(&b, d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(b.Bytes())
}

func (a *App) simular(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	a.mensajes = nil
	a.error = a.correrSimulacion(r)
	a.mu.Unlock()
	http.Redirect(w, r, "/#resultados", http.StatusSeeOther)
}

func (a *App) correrSimulacion(r *http.Request) string {
	p, err := leerParametros(r)
	if err != nil {
		return err.Error()
	}
	a.params = p
	actual, err1 := PoliticaActual(a.red)
	alg, err2 := PoliticaAlgoritmo(a.red)
	if err1 != nil || err2 != nil {
		return "No se puede simular: algún destino quedó sin ruta."
	}
	ra, rb := CorrerParejas(actual, alg, p.PorHora, p.Vehiculos, p.Horas*60, p.Semilla, p.Repeticiones)
	a.resumen, a.simulado = Resumir(ra, rb), p
	a.carga = [2]float64{carga(actual, p.PorHora, p.Vehiculos), carga(alg, p.PorHora, p.Vehiculos)}
	return ""
}

// carga es la fracción de la flota que haría falta en promedio para atender
// la demanda: pedidos por minuto por duración media de un viaje redondo.
func carga(pol Politica, porHora float64, vehiculos int) float64 {
	ciclo := 0.0
	for _, d := range destinos {
		ciclo += pol[d].Total() / float64(len(destinos))
	}
	return porHora / 60 * ciclo / float64(vehiculos)
}

func leerParametros(r *http.Request) (Parametros, error) {
	p := Parametros{Escenario: r.FormValue("escenario")}
	for _, e := range escenarios {
		if e.Nombre == p.Escenario {
			p.PorHora = e.PorHora
		}
	}
	if p.PorHora == 0 {
		p.Escenario = "Personalizado"
		v, err := strconv.ParseFloat(r.FormValue("porhora"), 64)
		if err != nil || v <= 0 || v > 200 {
			return p, fmt.Errorf("el promedio de pedidos por hora debe ser mayor que 0 y como máximo 200")
		}
		p.PorHora = v
	}
	var err error
	if p.Vehiculos, err = entero(r.FormValue("vehiculos"), 1, 200); err != nil {
		return p, fmt.Errorf("el número de vehículos debe ser un entero entre 1 y 200")
	}
	if p.Horas, err = strconv.ParseFloat(r.FormValue("horas"), 64); err != nil || p.Horas <= 0 || p.Horas > 24 {
		return p, fmt.Errorf("la duración de la recepción debe estar entre 0 y 24 horas")
	}
	if p.Semilla, err = strconv.ParseInt(r.FormValue("semilla"), 10, 64); err != nil {
		return p, fmt.Errorf("la semilla debe ser un número entero")
	}
	if p.Repeticiones, err = entero(r.FormValue("repeticiones"), 1, 100); err != nil {
		return p, fmt.Errorf("las repeticiones deben ser un entero entre 1 y 100")
	}
	return p, nil
}

func entero(s string, min, max int) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil || v < min || v > max {
		return 0, fmt.Errorf("fuera de rango")
	}
	return v, nil
}

func (a *App) tramo(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	a.mensajes = nil
	a.error = a.cambiarTramo(r)
	a.mu.Unlock()
	http.Redirect(w, r, "/#red", http.StatusSeeOther)
}

// cambiarTramo modifica o cierra un tramo y vuelve a calcular las rutas. Solo
// se aceptan tramos que existen en la matriz original, así un tramo cerrado
// se puede reabrir dándole un tiempo.
func (a *App) cambiarTramo(r *http.Request) string {
	desde, hasta := r.FormValue("desde"), r.FormValue("hasta")
	cerrar := r.FormValue("cerrar") != ""
	tiempo := 0.0
	if !cerrar {
		v, err := strconv.ParseFloat(r.FormValue("tiempo"), 64)
		if err != nil || v <= 0 {
			return "Indique un tiempo mayor que cero o marque la opción de cerrar el tramo."
		}
		tiempo = v
	}

	nueva := a.red.Copia()
	var hecho []string
	aplicar := func(o, d string) bool {
		i, j := nueva.Indice(o), nueva.Indice(d)
		if i < 0 || j < 0 || a.original.W[i][j] <= 0 {
			return false
		}
		if cerrar {
			nueva.W[i][j] = -1
			hecho = append(hecho, fmt.Sprintf("Se cerró el tramo %s → %s.", o, d))
		} else {
			nueva.W[i][j] = tiempo
			hecho = append(hecho, fmt.Sprintf("El tramo %s → %s ahora tarda %g min.", o, d, tiempo))
		}
		return true
	}
	if !aplicar(desde, hasta) {
		return fmt.Sprintf("No existe una conexión directa de %s a %s en la matriz.", desde, hasta)
	}
	if r.FormValue("ambos") != "" && !aplicar(hasta, desde) {
		hecho = append(hecho, fmt.Sprintf("No existe el tramo %s → %s (calle de un solo sentido).", hasta, desde))
	}

	antesAct, _ := PoliticaActual(a.red)
	antesAlg, _ := PoliticaAlgoritmo(a.red)
	despuesAct, err1 := PoliticaActual(nueva)
	despuesAlg, err2 := PoliticaAlgoritmo(nueva)
	if err1 != nil || err2 != nil {
		return "El cambio deja algún destino sin conexión con la bodega, por eso no se aplicó."
	}

	a.red = nueva
	a.resumen = nil
	cambios := append(diferencias("Algoritmo", antesAlg, despuesAlg), diferencias("Ruta actual", antesAct, despuesAct)...)
	if len(cambios) == 0 {
		cambios = []string{"Ninguna ruta cambió."}
	}
	a.mensajes = append(append(hecho, "Rutas recalculadas con Dijkstra sobre la red modificada:"), cambios...)
	return ""
}

func diferencias(nombre string, antes, despues Politica) []string {
	var msgs []string
	for _, d := range destinos {
		x, y := antes[d], despues[d]
		if !slices.Equal(x.Ida, y.Ida) || x.TIda != y.TIda {
			msgs = append(msgs, fmt.Sprintf("%s, ida a %s: %s, %g min (antes %g min).",
				nombre, d, strings.Join(y.Ida, " "), y.TIda, x.TIda))
		}
		if !slices.Equal(x.Regreso, y.Regreso) || x.TRegreso != y.TRegreso {
			msgs = append(msgs, fmt.Sprintf("%s, regreso desde %s: %s, %g min (antes %g min).",
				nombre, d, strings.Join(y.Regreso, " "), y.TRegreso, x.TRegreso))
		}
	}
	return msgs
}

func (a *App) restaurar(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	a.red = a.original.Copia()
	a.resumen = nil
	a.mensajes, a.error = []string{"Se restauró la red original."}, ""
	a.mu.Unlock()
	http.Redirect(w, r, "/#red", http.StatusSeeOther)
}

// grafo dibuja la red actual con Graphviz y resalta las rutas de una política:
// verde para la ida y naranja para el regreso.
func (a *App) grafo(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	red := a.red.Copia()
	a.mu.Unlock()

	var pol Politica
	if r.URL.Query().Get("rutas") == "actual" {
		pol, _ = PoliticaActual(red)
	} else {
		pol, _ = PoliticaAlgoritmo(red)
	}

	cmd := exec.Command("neato", "-n2", "-Tsvg")
	cmd.Stdin = strings.NewReader(red.DOT(resaltarRutas(red, pol)))
	svg, err := cmd.Output()
	if err != nil {
		http.Error(w, "no se pudo generar el grafo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(svg)
}
