package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	matriz := flag.String("matriz", "Matriz_Logistica_del_Valle_40_nodos.csv", "archivo CSV con la matriz de tiempos")
	puerto := flag.Int("puerto", 8080, "puerto de la aplicación web")
	exportar := flag.String("exportar", "", "carpeta donde guardar los resultados de los experimentos")
	reps := flag.Int("repeticiones", 10, "repeticiones por escenario al exportar")
	dot := flag.String("dot", "", "carpeta donde guardar el grafo en formato DOT")
	cerrar := flag.String("cerrar", "", "tramos cerrados al exportar, por ejemplo N17-N25,N25-N17")
	flag.Parse()

	g, err := LeerMatriz(*matriz)
	if err != nil {
		log.Fatal(err)
	}

	if *exportar == "" && *dot == "" {
		if *cerrar != "" {
			log.Fatal("-cerrar solo se usa junto con -exportar o -dot; en la página los tramos se cierran desde el formulario")
		}
		mux := http.NewServeMux()
		NuevaApp(g).Registrar(mux)
		dir := fmt.Sprintf("localhost:%d", *puerto)
		fmt.Printf("Aplicación disponible en http://%s\n", dir)
		log.Fatal(http.ListenAndServe(dir, mux))
	}

	if *cerrar != "" {
		for _, t := range strings.Split(*cerrar, ",") {
			desde, hasta, _ := strings.Cut(strings.TrimSpace(t), "-")
			i, j := g.Indice(desde), g.Indice(hasta)
			if i < 0 || j < 0 || g.W[i][j] <= 0 {
				log.Fatalf("el tramo %q no existe", t)
			}
			g.W[i][j] = -1
		}
	}
	if *dot != "" {
		if err := exportarDOT(g, *dot); err != nil {
			log.Fatal(err)
		}
	}
	if *exportar != "" {
		if err := Exportar(g, *exportar, *reps); err != nil {
			log.Fatal(err)
		}
	}
}

// exportarDOT guarda el grafo solo (grafo.dot) y con las rutas del algoritmo
// resaltadas (grafo_rutas.dot). Se dibujan con: neato -n2 -Tpdf grafo.dot -o grafo.pdf
func exportarDOT(g *Grafo, carpeta string) error {
	if err := os.MkdirAll(carpeta, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(carpeta, "grafo.dot"), []byte(g.DOT(nil)), 0o644); err != nil {
		return err
	}
	alg, err := PoliticaAlgoritmo(g)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(carpeta, "grafo_rutas.dot"), []byte(g.DOT(resaltarRutas(g, alg))), 0o644)
}
