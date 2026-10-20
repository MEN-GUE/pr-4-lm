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

	pedidos := GenerarPedidos(4, 480, 1)
	fmt.Printf("Prueba: %d pedidos en 8 horas (4 por hora), 8 vehículos, semilla 1\n", len(pedidos))
	mostrarResultado("actual", Simular(pedidos, actual, 8, 480))
	mostrarResultado("Dijkstra", Simular(pedidos, algoritmo, 8, 480))

	if err := os.WriteFile("informe/figuras/grafo.dot", []byte(g.DOT(nil)), 0o644); err != nil {
		log.Fatal(err)
	}
}

func mostrarResultado(nombre string, r Resultado) {
	fmt.Printf("  %-9s puntuales %d (%.1f%%), espera %.1f min, solicitud a entrega %.1f min, utilización %.1f%%, adicional %.1f min\n",
		nombre, r.Puntuales, r.PorcPuntual(), r.EsperaProm, r.RespuestaProm, r.Utilizacion, r.TiempoAdicional)
}

func mostrarRuta(nombre string, r Ruta) {
	fmt.Printf("  %-10s ida %3g min: %s\n", nombre, r.TIda, strings.Join(r.Ida, " "))
	fmt.Printf("  %-10s reg %3g min: %s  (total %g)\n", "", r.TRegreso, strings.Join(r.Regreso, " "), r.Total())
}
