package main

import (
	"fmt"
	"net/http"
)

// Index welcomes
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}
