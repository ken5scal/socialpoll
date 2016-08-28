package main

import (
	"flag"
	"net/http"
	"log"
)

func main() {
	var addr = flag.String("addr", ":8081", "Web Address")
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("/", http.StripPrefix("/", http.FileServer(http.Dir("public"))))
	log.Println("Web site address: ", *addr)
	http.ListenAndServe(*addr, mux)
}
