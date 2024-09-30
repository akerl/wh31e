package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/akerl/wh31e/config"
	"github.com/akerl/wh31e/metrics"
	"github.com/akerl/wh31e/register"

	"github.com/akerl/timber/v2/log"
)

var logger = log.NewLogger("wh31e.server")

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
	logger.InfoMsgf("binding metrics server to %s", bindStr)
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", s.handleMetrics)
	return http.ListenAndServe(bindStr, mux)
}

func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	mf := metrics.MetricFile{}
	for _, v := range s.Register.Latest {
		mf.Metrics = append(mf.Metrics, v.Metrics()...)
	}
	if !mf.Validate() {
		logger.InfoMsg("invalid metrics file requested")
		http.Error(w, "invalid metrics file", http.StatusInternalServerError)
	} else {
		logger.InfoMsg("successful metrics request")
		io.WriteString(w, mf.String())
	}
}
