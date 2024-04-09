package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"handler", "method", "status"},
	)
)

type server struct {
	redis  redis.UniversalClient
	logger *zap.Logger
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("unable to initialize logger")
	}

	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{os.Getenv("REDIS_ADDR")},
		Password: "",
		DB:       0,
	})

	srv := &server{
		redis:  rdb,
		logger: logger,
	}

	router := httprouter.New()

	// Wrap the handler with promhttp.InstrumentHandlerDuration
	router.GET("/", prometheusHandler(srv.indexHandler))
	router.GET("/health", srv.healthCheckHandler)
	router.Handler("GET", "/metrics", promhttp.Handler())

	logger.Info("server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// prometheusHandler wraps the given handler with Prometheus instrumentation.
func prometheusHandler(h http.HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h(w, r) // Call the original handler first

		// Extract relevant labels from the request and increment the counter
		labels := []string{"indexHandler", r.Method, fmt.Sprintf("%d", http.StatusOK)}
		httpRequestsTotal.WithLabelValues(labels...).Inc()
	}
}


func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {

	var v string
	var err error
	if v, err = s.redis.Get(context.Background(), "updated_time").Result(); err != nil {
		s.logger.Info("updated_time not found, setting it")
		v = time.Now().Format("2006-01-02 03:04:05")
		s.redis.Set(context.Background(), "updated_time", v, 5*time.Second)
	} else {
		s.logger.Info("got updated_time")
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "hello world: updated_time=%s\n", v)

	httpRequestsTotal.WithLabelValues("indexHandler", r.Method, fmt.Sprintf("%d", http.StatusOK)).Inc()
}

func (s *server) healthCheckHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
