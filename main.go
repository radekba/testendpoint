package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	opsProcessed = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "test_status",
			Help: "passed or failed test",
		},
		[]string{"service"},
	)
	testExecuted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "test_executions",
		Help: "Number of test executions",
	})
	rpcDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "request_durations_seconds",
			Help:       "Request latency distributions.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"service"},
	)
)
var addr, sleep string

func main() {
	sleepSeconds, _ := time.ParseDuration(sleep)
	go statsServer()
	for {
		testExecuted.Inc()
		status, duration := checkSite(addr)
		rpcDurations.WithLabelValues(addr).Observe(duration)
		opsProcessed.WithLabelValues(addr).Set(float64(status))
		time.Sleep(sleepSeconds)
	}
}

func init() {
	flag.StringVar(&addr, "address", "https://www.google.com", "address")
	flag.StringVar(&sleep, "sleep", "10s", "sleep between checks, format: 1s, 1m...")
	flag.Parse()
	prometheus.MustRegister(rpcDurations)
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}

func checkSite(addr string) (int, float64) {
	start := time.Now()
	resp, err := http.Get(addr)
	duration := time.Since(start)
	ms := float64(duration.Seconds() * 1000)
	if err != nil {
		log.Print(err)
		return 0, ms
	}
	log.Print(fmt.Sprintf("Server: %s, code: %v, duration: %v", resp.Request.URL, resp.StatusCode, ms))
	if resp.StatusCode == 200 {
		return 1, ms
	}
	return 0, ms
}

func statsServer() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
