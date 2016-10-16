package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hihi"))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Request started\n"))

	span := opentracing.StartSpan("/home")
	defer span.Finish()

	asyncReq, _ := http.NewRequest("GET", "http://localhost:8080/async", nil)
	err := span.Tracer().Inject(
		span.Context(),
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(asyncReq.Header),
	)

	if err != nil {
		log.Fatalf("%s: Couldn't inject headers (%v)", r.URL.Path, err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		if _, err := http.DefaultClient.Do(asyncReq); err != nil {
			log.Printf("%s: Async call failed (%v)", r.URL.Path, err)
		}
	}()

	time.Sleep(10 * time.Millisecond)

	syncReq, _ := http.NewRequest("GET", "http://localhost:8080", nil)
	err = span.Tracer().Inject(
		span.Context(),
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(syncReq.Header),
	)

	if err != nil {
		log.Fatalf("%s: Couldn't inject headers (%v)", r.URL.Path, err)
	}

	if _, err = http.DefaultClient.Do(syncReq); err != nil {
		log.Printf("%s: Synchronous call failed (%v)", r.URL.Path, err)
		return
	}

	w.Write([]byte("Request done!\n"))
}

func serviceHandler(w http.ResponseWriter, r *http.Request) {
	opName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	var sp opentracing.Span
	spCtx, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(r.Header),
	)
	if err == nil {
		sp = opentracing.StartSpan(opName, opentracing.ChildOf(spCtx))
	} else {
		sp = opentracing.StartSpan(opName)
	}
	defer sp.Finish()

	time.Sleep(50 * time.Millisecond)

	dbReq, _ := http.NewRequest("GET", "http://localhost:8080/db", nil)
	err = sp.Tracer().Inject(
		sp.Context(),
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(dbReq.Header),
	)
	if err != nil {
		log.Fatalf("%s: Couldn't inject headers (%v)", r.URL.Path, err)
	}

	if _, err := http.DefaultClient.Do(dbReq); err != nil {
		sp.LogEventWithPayload("db request err", err)
	}
}

func dbHandler(w http.ResponseWriter, r *http.Request) {
	var sp opentracing.Span

	spanCtx, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(r.Header),
	)

	if err != nil {
		log.Printf("%s: Could not join trace (%v)", r.URL.Path, err)
		return
	}

	if err == nil {
		sp = opentracing.StartSpan("GET /db", opentracing.ChildOf(spanCtx))
	} else {
		sp = opentracing.StartSpan("GET /db")
	}

	defer sp.Finish()
	time.Sleep(25 * time.Millisecond)
}
