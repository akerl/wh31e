package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/akerl/wh31e/config"
	"github.com/akerl/wh31e/metrics"
	"github.com/akerl/wh31e/register"
)

// Server defines a Prometheus-compatible metrics engine
type Server struct {
	SensorNames map[int]string
	Port        int
	Register    *register.Register
}

// NewServer creates a new Server object
func NewServer(conf config.Config, reg *register.Register) *Server {
	return &Server{
		SensorNames: conf.SensorNames,
		Port:        conf.Port,
		Register:    reg,
	}
}

// Run starts the Server object in the foreground
func (s *Server) Run() error {
	bindStr := fmt.Sprintf(":%d", s.Port)
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", s.handleMetrics)
	return http.ListenAndServe(bindStr, mux)
}

func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	mf := metrics.MetricFile{}
	for _, v := range s.Register.Latest {
		mf = append(mf, v.Metrics()...)
	}
	mf = append(mf, s.Register.CounterMetrics()...)
	if !mf.Validate() {
		http.Error(w, "invalid metrics file", http.StatusInternalServerError)
	} else {
		io.WriteString(w, mf.String())
	}
}
