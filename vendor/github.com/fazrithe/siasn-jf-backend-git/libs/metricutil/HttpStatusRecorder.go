package metricutil

import "net/http"

// HttpStatusRecorder implements the default http.ResponseWriter interface, used to record the returned HTTP status code.
type HttpStatusRecorder struct {
	http.ResponseWriter
	Status       int
	BytesWritten int
}

func NewHttpStatusRecorder(writer http.ResponseWriter) *HttpStatusRecorder {
	return &HttpStatusRecorder{ResponseWriter: writer, Status: http.StatusOK}
}

// WriteHeader overrides the default WriteHeader behavior to also write access logs.
func (rec *HttpStatusRecorder) WriteHeader(code int) {
	rec.Status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *HttpStatusRecorder) Write(p []byte) (n int, err error) {
	n, err = rec.ResponseWriter.Write(p)
	rec.BytesWritten += n
	return
}
