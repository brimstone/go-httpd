package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/brimstone/go-httpd"
)

func hiHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

func main() {

	h, err := httpd.New(httpd.Port(8000))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	h.HandleFunc("/", hiHandler)
	h.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("/tmp"))))
	err = h.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
