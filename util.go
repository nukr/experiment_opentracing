package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/traceapp"
)

func startAppdashServer(appdashPort int) string {
	store := appdash.NewMemoryStore()
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})

	if err != nil {
		log.Fatal(err)
	}
	collectorPort := l.Addr().(*net.TCPAddr).Port
	// collectorApp := fmt.Sprintf(":%d", collectorPort)

	cs := appdash.NewServer(l, appdash.NewLocalCollector(store))
	go cs.Start()

	appdashURLStr := fmt.Sprintf("http://localhost:%d", appdashPort)
	appdashURL, err := url.Parse(appdashURLStr)

	if err != nil {
		log.Fatalf("Error parsing %s: %s", appdashURLStr, err)
	}

	fmt.Printf("To see your traces, go to %s/trace\n", appdashURL)

	tapp, err := traceapp.New(nil, appdashURL)
	if err != nil {
		log.Fatal(err)
	}

	tapp.Store = store
	tapp.Queryer = store

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", appdashPort), tapp))
	}()
	return fmt.Sprintf(":%d", collectorPort)
}
