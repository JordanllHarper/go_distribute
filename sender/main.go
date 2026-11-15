package main

import (
	"flag"
	"log"
	"net/http"

	shared "github.com/JordanllHarper/distributed_file_backup/shared"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "address to set the server to")
	dir := flag.String("dir", "", "where to send files to")
	flag.Parse()

	mux := http.NewServeMux()

	fs := shared.NewFileStorage(*dir)
	d := newDependencies(fs)
	addEndpointsV1(mux, d)
	srv := addLogging(mux)

	log.Println("Listening on:", *addr)

	log.Fatalln(http.ListenAndServe(*addr, srv))
}
