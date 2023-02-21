package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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
	listenFlag             string
	responseStatusCodeFlag int
	responseBodyFlag       string
	logFullHeader          bool
	VersionFlag            bool

	Name      string
	Version   string
	GitCommit string
)

const (
	SEP = ","
)

func init() {
	flag.StringVar(&listenFlag, "listen", ":8081,", "Ports on an HTTP echo server is going to be listening (Comma separated values)")
	flag.IntVar(&responseStatusCodeFlag, "response_status", 200, "Response status code to be sent")
	flag.StringVar(&responseBodyFlag, "response_body", "Hello, World!", "Response body to be sent")
	flag.BoolVar(&logFullHeader, "log_headers", false, "Log full header")

	flag.BoolVar(&VersionFlag, "version", false, "Display version information at start")
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
			os.Exit(1)
		}
	}()

	<-closeCtx.Done()

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(timeoutCtx)
	if err != nil {
		log.Printf("[ERROR] Error shutting down server. Err: %s", err)
	} else {
		log.Printf("[INFO] Server '%s' gracefully shutdown", server.Addr)
	}
}

func parseListenAddresses(listen string) ([]string, error) {
	var addressesList []string

	for _, addr := range strings.Split(listen, SEP) {
		trimmedAddr := strings.TrimSpace(addr)
		if trimmedAddr == "" {
			continue
		}
		addressesList = append(addressesList, trimmedAddr)
	}

	if len(addressesList) == 0 {
		return nil, errors.New("[ERROR] Invalid list of addresses provided")
	}

	return addressesList, nil
}

func main() {
	// Parse and validate flags
	flag.Parse()

	if len(flag.Args()) > 0 {
		log.Fatalln("[ERROR] Invalid number of flags provided, see `-help` flag")
	}

	// Return version details if flag is set
	if VersionFlag {
		fmt.Printf("%s - version: %s (SHA: %s)", Name, Version, GitCommit)
		os.Exit(0)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", appendHeaders(logRequest(httpEcho)))
	mux.HandleFunc("/health", appendHeaders(httpHealth))

	addressesList, err := parseListenAddresses(listenFlag)
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
