package httpd

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"
)

// Httpd main struct
type Httpd struct {
	server *http.Server
	mux    *http.ServeMux
}

// New creates a new Httpd instance
func New() *Httpd {
	h := &Httpd{}
	h.mux = http.NewServeMux()

	h.server = &http.Server{
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Register pprof handlers
	h.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	h.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	h.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	h.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	h.mux.HandleFunc("/debug/pprof/", pprof.Index)

	return h
}

func (h *Httpd) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// inspiration from https://github.com/lestrrat/go-apache-logformat
	newWriter := &responseWriter{
		Writer: w,
	}
	h.mux.ServeHTTP(newWriter, r)

	// Log results
	userAgent := r.Header["User-Agent"]
	referer := r.Header["Referer"]
	userID := "-"
	fmt.Printf("%s - %s [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" \"-\"\n",
		r.RemoteAddr,
		userID,
		time.Now().Format("2/Jan/2006:15:04:05 -0700"),
		r.Method,
		r.RequestURI,
		r.Proto,
		newWriter.Status,
		newWriter.Size,
		referer[0],
		userAgent[0],
	)
	//spew.Dump(r)
}

// HandleFunc adds a route to the map of handlers
func (h *Httpd) HandleFunc(
	newroute string,
	handler func(http.ResponseWriter, *http.Request),
) {
	h.mux.HandleFunc(newroute, handler)
}

func (h *Httpd) Handle(route string, handler http.Handler) {
	h.mux.Handle(route, handler)
}

// ListenAndServe starts Httpd listening on the defined address
func (h *Httpd) ListenAndServe(address string) error {
	h.server.Addr = address
	fmt.Println("Serving on", address)
	return h.server.ListenAndServe()
}
