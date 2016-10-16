package main

import (
	"fmt"
	"log"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/appdash"
	appdashot "sourcegraph.com/sourcegraph/appdash/opentracing"
)

func main() {
	var port = 8080
	var appdashPort = 8700

	addr := startAppdashServer(appdashPort)
	tracer := appdashot.NewTracer(appdash.NewRemoteCollector(addr))
	opentracing.InitGlobalTracer(tracer)

	httpServerAddr := fmt.Sprintf(":%d", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/home", homeHandler)
	mux.HandleFunc("/async", serviceHandler)
	mux.HandleFunc("/service", serviceHandler)
	mux.HandleFunc("/db", dbHandler)
	fmt.Printf("http server is up and listening on port %d", port)
	log.Fatal(http.ListenAndServe(httpServerAddr, mux))
}
