package main

const bodega = "N01"

var destinos = []string{"N08", "N19", "N34", "N40"}

// Rutas de las hojas de despacho. El regreso es la misma ruta al revés.
var rutasActuales = map[string][]string{
	"N08": {"N01", "N02", "N03", "N04", "N05", "N06", "N07", "N08"},
	"N19": {"N01", "N02", "N03", "N11", "N19"},
	"N34": {"N01", "N09", "N17", "N25", "N33", "N34"},
	"N40": {"N01", "N02", "N03", "N04", "N05", "N06", "N07", "N08", "N16", "N24", "N32", "N40"},
}

func esDestino(nombre string) bool {
	for _, d := range destinos {
		if d == nombre {
			return true
		}
	}
	return false
}
