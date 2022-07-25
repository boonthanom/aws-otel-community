package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/aws-otel-commnunity/sample-apps/go-sample-app/collection"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric/global"
)

// This sample application is in conformance with the ADOT SampleApp requirements spec.
func main() {
	ctx := context.Background()

	// Creates and configures random based metrics based on a configuration file (config.yaml).
	mp := global.MeterProvider()
	cfg := collection.GetConfiguration()
	rmc := collection.NewRandomMetricCollector(mp)
	rmc.RegisterMetricsClient(ctx, *cfg)

	// Starts request based metric and registers necessary callbacks
	rqmc := collection.NewRequestBasedMetricCollector(ctx, *cfg, mp)
	rqmc.StartTotalRequestCallback()

	shutdown, err := collection.StartClient(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer shutdown(ctx)

	// Creates a router, client and web server with several endpoints
	r := mux.NewRouter()
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	r.Use(otelmux.Middleware("Go-Sampleapp-Server"))

	r.HandleFunc("/aws-sdk-call", rqmc.AwsSdkCall)

	r.HandleFunc("/outgoing-http-call", func(w http.ResponseWriter, r *http.Request) {
		rqmc.OutgoingHttpCall(w, r, client)
	})

	r.HandleFunc("/outgoing-sampleapp", func(w http.ResponseWriter, r *http.Request) {
		rqmc.OutgoingSampleApp(w, r, client)
	})
	http.Handle("/", r)

	srv := &http.Server{
		Addr: net.JoinHostPort(cfg.Host, cfg.Port),
	}
	log.Fatal(srv.ListenAndServe())

}
