package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIConfig struct {
	IPAddress string
	port      int
}

var templates map[string]*template.Template

func api() {
	loadTemplates()

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/"))))
	r.HandleFunc("/info", info)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8000", r))
}

func info(w http.ResponseWriter, r *http.Request) {
	if err := templates["template.html"].Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func loadTemplates() {
	// var baseTemplate = "templates/layout/_base.html"
	templates = make(map[string]*template.Template)
	templates["template.html"] = template.Must(template.ParseFiles("static/template.html"))
}
