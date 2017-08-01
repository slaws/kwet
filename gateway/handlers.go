package main

import (
	"html/template"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Index welcomes
func Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Errorf("Unable to parse index template : %s", err)
	}

	// t.ExecuteTemplate(w, "templates/index.html", "foo")
	t.Execute(w, "bar")

}

// // ServeStatic serves static files
// func ServeStatic(w http.ResponseWriter, r *http.Request) {
// 	http.FileServer(http.Dir("public"))
// }
