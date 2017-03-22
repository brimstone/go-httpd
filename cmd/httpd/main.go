package main

import (
	"net/http"

	"github.com/brimstone/go-httpd"
)

func hiHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

func main() {

	h := httpd.New()

	h.HandleFunc("/", hiHandler)
	h.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("/tmp"))))
	h.ListenAndServe(":8080")
}
