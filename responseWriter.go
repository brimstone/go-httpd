package httpd

import "net/http"

type responseWriter struct {
	Writer http.ResponseWriter
	Status int
	Size   int
}

func (r *responseWriter) Header() http.Header {
	return r.Writer.Header()
}

func (r *responseWriter) Write(b []byte) (int, error) {
	if r.Status == 0 {
		r.Status = 200
	}
	r.Size += len(b)
	return r.Writer.Write(b)
}

func (r *responseWriter) WriteHeader(i int) {
	r.Status = i
	r.Writer.WriteHeader(i)
}
