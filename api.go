package main

import (
	"gorilla/mux"
	"html/template"
	"log"
	"net/http"
)

type APIConfig struct {
	IPAddress string
	port      int
}

func api() {
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/info", info)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8000", r))
}

func info(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("Gorilla!\n"))
	// w.Write([]byte(*configFilePath))
	t := template.New("some template")
	t, _ = t.ParseFiles("./template.html")
	t.Execute(w, t)
}
