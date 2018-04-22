package httpd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"time"
)

// Httpd main struct
type Httpd struct {
	Address string
	server  *http.Server
	mux     *http.ServeMux
}

// New creates a new Httpd instance
func New(options ...func(*Httpd) error) (*Httpd, error) {
	h := &Httpd{}
	h.mux = http.NewServeMux()

	h.server = &http.Server{
		Handler:        h,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Register pprof handlers
	h.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	h.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	h.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	h.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	h.mux.HandleFunc("/debug/pprof/", pprof.Index)

	for _, option := range options {
		err := option(h)
		if err != nil {
			return nil, err
		}
	}
	return h, nil
}

func Port(port int) func(*Httpd) error {
	return func(h *Httpd) error {
		h.Address = ":" + strconv.Itoa(port)
		return nil
	}
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
	if len(referer) == 0 {
		referer = []string{""}
	}
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
func (h *Httpd) ListenAndServe(address ...string) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go h.HandleSignal(quit)

	h.server.Addr = h.Address
	if len(address) > 0 {
		h.server.Addr = address[0]
	}
	fmt.Println("Serving on", h.server.Addr)
	err := h.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (h *Httpd) HandleSignal(q chan os.Signal) {
	<-q
	log.Println("Server is shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	h.server.SetKeepAlivesEnabled(false)
	if err := h.server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
}
