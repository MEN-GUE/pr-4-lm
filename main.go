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

	fmt.Println("Rutas actuales:")
	for _, d := range destinos {
		ruta := rutasActuales[d]
		fmt.Printf("%s: %s\n", d, strings.Join(ruta, " "))
		fmt.Printf("    ida:     %s\n", describir(g.TiempoRuta(ruta)))
		fmt.Printf("    regreso: %s\n", describir(g.TiempoRuta(Invertir(ruta))))
	}

	if err := os.WriteFile("informe/figuras/grafo.dot", []byte(g.DOT(nil)), 0o644); err != nil {
		log.Fatal(err)
	}
}

func describir(t float64, err error) string {
	if err != nil {
		return "no válida (" + err.Error() + ")"
	}
	return fmt.Sprintf("%g min", t)
}
