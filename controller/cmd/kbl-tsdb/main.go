package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

func main() {
	addr := flag.String("addr", ":9090", "HTTP listen address")
	dataDir := flag.String("data-dir", "/var/kbl/tsdb", "TSDB data directory on node-local storage")
	flag.Parse()

	if err := os.MkdirAll(*dataDir, 0o755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	engine, err := store.OpenTSDBEngine(*dataDir)
	if err != nil {
		log.Fatalf("open tsdb engine: %v", err)
	}

	handler := store.NewTSDBHandler(engine)
	log.Printf("kbl-tsdb listening on %s data-dir=%s", *addr, *dataDir)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}
