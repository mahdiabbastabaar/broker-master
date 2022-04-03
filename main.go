package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"therealbroker/api/api"
	"therealbroker/api/api/proto"
)

// Main requirements:
// 1. All tests should be passed
// 2. Your logs should be accessible in Graylog
// 3. Basic prometheus metrics ( latency, throughput, etc. ) should be implemented
// 	  for every base functionality ( publish, subscribe etc. )

var (
	reg = prometheus.NewRegistry()

	SuccessfulRPCCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "server_successful_calls_total",
		Help: "Total number of successful RPCs handled on the server.",
	}, []string{"success"},
	)

	FailedRPCCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "server_failed_calls_total",
		Help: "Total number of failed RPCs handled on the server.",
	}, []string{"failed"},
	)

	EachCallDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "call_duration",
		Help:       "latency of each call.",
		Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.005, 0.99: 0.001},
	}, []string{"duration"})

	ActiveSubscribers = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "active_subscribers_total",
		Help: "Total number of subscribers",
	}, []string{"subscribers"})
)

func init() {
	err := reg.Register(SuccessfulRPCCalls)
	if err != nil {
		return
	}
	err = reg.Register(FailedRPCCalls)
	if err != nil {
		return
	}
	err = reg.Register(EachCallDuration)
	if err != nil {
		return
	}
	err = reg.Register(ActiveSubscribers)
	if err != nil {
		return
	}
}

func main() {

	api.SuccessfulRPCCalls = SuccessfulRPCCalls
	api.FailedRPCCalls = FailedRPCCalls
	api.EachCallDuration = EachCallDuration
	api.ActiveSubscribers = ActiveSubscribers

	fmt.Println("================= Server ===================")

	ls, err := net.Listen("tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	httpServer := &http.Server{Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}), Addr: "localhost:9091"}

	go func() {
		if err = httpServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	gs := grpc.NewServer()
	proto.RegisterBrokerServer(gs, api.NewModule())

	if err = gs.Serve(ls); err != nil {
		log.Fatal(err)
	}

}
