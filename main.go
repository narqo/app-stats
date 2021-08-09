package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		addr     string
		hostname string
	)
	flag.StringVar(&addr, "http-addr", "127.0.0.1:8080", "address to listen on")
	flag.StringVar(&hostname, "hostname", "", "host name")

	flag.Parse()

	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		io.WriteString(rw, hostname)
	})
	mux.HandleFunc("/ping", func(rw http.ResponseWriter, r *http.Request) {
		io.WriteString(rw, "ok")
	})

	handler := RequestLog(os.Stderr, mux)

	log.Printf("listening addr=%s\n", addr)
	return http.ListenAndServe(addr, handler)
}

func RequestLog(out io.Writer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC()

		w := &responseWriter{rw, http.StatusOK}
		next.ServeHTTP(w, r)

		rtime := time.Since(start)
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}

		remoteAddr, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			remoteAddr = r.RemoteAddr
		}

		fmt.Fprintf(
			out,
			"ts=%s method=%s host=%s uri=%s proto=%s status=%v ip=%s ua=%s ref=%s dur=%s\n",
			start.Format(time.RFC3339),
			r.Method,
			host,
			r.RequestURI,
			r.Proto,
			w.status,
			remoteAddr,
			strconv.Quote(r.UserAgent()),
			strconv.Quote(r.Referer()),
			rtime,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode
}
