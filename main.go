package main

import (
	"fmt"
	"log"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

func main() {
	var port = 8800

	collector, err := zipkin.NewHTTPCollector("http://localhost:9411/api/v1/spans")
	if err != nil {
		log.Fatalf("unable to create Zipkin HTTP collector: %+v", err)
	}
	recoder := zipkin.NewRecorder(collector, true, "127.0.0.1:0", "testggg")
	tracer, err := zipkin.NewTracer(recoder,
		zipkin.ClientServerSameSpan(true),
		zipkin.TraceID128Bit(true))
	opentracing.InitGlobalTracer(tracer)

	httpServerAddr := fmt.Sprintf(":%d", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/home", middleware(homeHandler))
	mux.HandleFunc("/async", middleware(serviceHandler))
	mux.HandleFunc("/service", serviceHandler)
	mux.HandleFunc("/db", dbHandler)
	fmt.Printf("http server is up and listening on port %d\n", port)
	log.Fatal(http.ListenAndServe(httpServerAddr, mux))
}

func middleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Access-Control-Allow-Origin", "*")
		h.Add("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE")
		h.Add("Access-Control-Allow-Headers", "Accept-Language, Content-Type, x-b3-traceid, x-b3-spanid, x-b3-sampled")
		handler(w, r)
	}
}
