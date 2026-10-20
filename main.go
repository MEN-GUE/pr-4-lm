package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	g, err := LeerMatriz("Matriz_Logistica_del_Valle_40_nodos.csv")
	if err != nil {
		log.Fatal(err)
	}

	dobles, distintos, unSentido := g.Resumen()
	fmt.Printf("%d nodos y %d tramos dirigidos\n", len(g.Nombres), 2*dobles+unSentido)
	fmt.Printf("%d pares de doble sentido (%d con tiempos distintos) y %d tramos de un solo sentido\n\n",
		dobles, distintos, unSentido)

	actual, err := PoliticaActual(g)
	if err != nil {
		log.Fatal(err)
	}
	algoritmo, err := PoliticaAlgoritmo(g)
	if err != nil {
		log.Fatal(err)
	}
	reversible := PoliticaReversible(g)

	for _, d := range destinos {
		fmt.Println("Destino", d)
		mostrarRuta("actual", actual[d])
		mostrarRuta("Dijkstra", algoritmo[d])
		mostrarRuta("reversible", reversible[d])
		if _, err := g.TiempoRuta(Invertir(algoritmo[d].Ida)); err != nil {
			fmt.Printf("  la ida de Dijkstra no sirve para regresar: %v\n", err)
		}
		fmt.Println()
	}

	if err := os.WriteFile("informe/figuras/grafo.dot", []byte(g.DOT(nil)), 0o644); err != nil {
		log.Fatal(err)
	}
}

func mostrarRuta(nombre string, r Ruta) {
	fmt.Printf("  %-10s ida %3g min: %s\n", nombre, r.TIda, strings.Join(r.Ida, " "))
	fmt.Printf("  %-10s reg %3g min: %s  (total %g)\n", "", r.TRegreso, strings.Join(r.Regreso, " "), r.Total())
}
