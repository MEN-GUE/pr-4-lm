package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	matriz := flag.String("matriz", "Matriz_Logistica_del_Valle_40_nodos.csv", "archivo CSV con la matriz de tiempos")
	puerto := flag.Int("puerto", 8080, "puerto de la aplicación web")
	flag.Parse()

	g, err := LeerMatriz(*matriz)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	NuevaApp(g).Registrar(mux)
	dir := fmt.Sprintf("localhost:%d", *puerto)
	fmt.Printf("Aplicación disponible en http://%s\n", dir)
	log.Fatal(http.ListenAndServe(dir, mux))
}
