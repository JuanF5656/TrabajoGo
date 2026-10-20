// Servidor web (bonus de interfaz gráfica). Es solo la capa de vista: toda la
// lógica vive en el paquete hospital, que no sabe que esta interfaz existe.
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
)

//go:embed static
var staticFiles embed.FS

func main() {
	addr := flag.String("addr", ":8080", "dirección donde escucha el servidor")
	flag.Parse()

	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	s, err := newServer()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Hospital web en http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, s.routes(static)))
}
