package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	printLogs    bool
	responseBody string
	listen       string
	SEP          = ","
)

func init() {
	flag.BoolVar(&printLogs, "print_logs", false, "Print logs")
	flag.StringVar(&responseBody, "response_body", "Hello, World!", "Response body")
	flag.StringVar(&listen, "listen", ":8081,:8082", "Ports on which root pattern is going to be listening (Comma separated values)")
}

func httpEcho(w http.ResponseWriter, r *http.Request) {
	if printLogs {
		log.Println(r.Header)
	}

	w.Write([]byte(responseBody))
}

func healthEcho(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{status: ok}"))
}

func listenAndServe(addr string, handler http.Handler, closeCtx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	server := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Printf("[INFO] Server listening on address '%s'", server.Addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("[ERROR] Error starting server. Err: %s", err)
		} else {
			log.Printf("[INFO] Server '%s' gracefully shutdown", server.Addr)
		}
	}()

	<-closeCtx.Done()

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

    server.Shutdown(timeoutCtx)
}

func parseListenAddresses(listen string) ([]string, error) {
	if listen == "" {
		return nil, errors.New("[ERROR] Invalid list of addresses provided")
	}

	addressesList := strings.Split(listen, SEP)

	return addressesList, nil
}

func main() {
	// Parse and validate flags
	flag.Parse()

	if len(flag.Args()) > 0 {
		log.Fatalln("[ERROR] Invalid number of flags provided, see `-help` flag")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", httpEcho)
	mux.HandleFunc("/health", healthEcho)

	addressesList, err := parseListenAddresses(listen)
	if err != nil {
		log.Fatalf("%s", err)
	}

	wg := &sync.WaitGroup{}
	closeCtx, closeServer := context.WithCancel(context.Background())
	for _, addr := range addressesList {
		go listenAndServe(addr, mux, closeCtx, wg)
		wg.Add(1)
	}

	exitChan := make(chan os.Signal, 0)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)

	<-exitChan
	closeServer()

	wg.Wait()
}
