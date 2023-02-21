package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	headerAppName    = "X-App-Name"
	headerAppVersion = "X-App-Version"

	logFormat = "Remote Addr: %s Host Addr: %s Method: %s Path: %s Proto: %s Status: %d Length: %d User Agent: %s Duration: %v"
)

func httpEcho(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(responseStatusCodeFlag)
	w.Write([]byte(responseBodyFlag))
}

func httpHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{status: ok}"))
}

func appendHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(headerAppName, Name)
		w.Header().Add(headerAppVersion, Version)
		next.ServeHTTP(w, r)
	}
}

// Custom ResponseWriter to save information to log
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	length     int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
	}
}

func (lrw *loggingResponseWriter) Header() http.Header {
	return lrw.ResponseWriter.Header()
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.length = len(b)
	return lrw.ResponseWriter.Write(b)
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)

		defer func(start time.Time) {
			reqDuration := time.Now().Sub(start)
			httpRequestLog := fmt.Sprintf(logFormat, r.RemoteAddr, r.Host, r.Method, r.URL.Path, r.Proto, lrw.statusCode, lrw.length, r.UserAgent(), reqDuration)
			if logFullHeader {
				httpRequestLog = fmt.Sprintf("%s Full header: %v", httpRequestLog, r.Header)
			}
			log.Println(httpRequestLog)
		}(time.Now())
	}
}
