package register

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/akerl/timber/v2/log"
	"github.com/akerl/wh31e/config"
	"github.com/akerl/wh31e/metrics"

	"gopkg.in/mcuadros/go-syslog.v2/format"
)

var logger = log.NewLogger("wh31e.register")

type message struct {
	TimeStr      string  `json:"time"`
	IDInt        int     `json:"id"`
	Battery      int     `json:"battery_ok"`
	TemperatureC float64 `json:"temperature_C"`
	Humidity     int     `json:"humidity"`
}

type event struct {
	Time         time.Time
	ID           string
	Name         string
	Battery      int
	TemperatureC float64
	TemperatureF float64
	Humidity     int
}

// Register defines the shared event tracking state
type Register struct {
	SensorNames map[int]string
	Latest      map[string]event
}

func (e *event) Tags() map[string]string {
	return map[string]string{
		"name": e.Name,
		"id":   e.ID,
	}
}

// Metrics returns Metric-formatted entries for this event
func (e *event) Metrics() []metrics.Metric {
	t := e.Tags()
	mf := []metrics.Metric{
		metrics.Metric{
			Name:  "wh31e_last_updated",
			Type:  "gauge",
			Tags:  t,
			Value: fmt.Sprintf("%d", e.Time.Unix()),
		},
	}
	if e.Humidity != 0 {
		mf = append(
			mf,
			metrics.Metric{
				Name:  "wh31e_humidity",
				Type:  "gauge",
				Tags:  t,
				Value: fmt.Sprintf("%d", e.Humidity),
			},
			metrics.Metric{
				Name:  "wh31e_battery",
				Type:  "gauge",
				Tags:  t,
				Value: fmt.Sprintf("%d", e.Battery),
			},
			metrics.Metric{
				Name:  "wh31e_temperature_c",
				Type:  "gauge",
				Tags:  t,
				Value: fmt.Sprintf("%g", e.TemperatureC),
			},
			metrics.Metric{
				Name:  "wh31e_temperature_f",
				Type:  "gauge",
				Tags:  t,
				Value: fmt.Sprintf("%g", e.TemperatureF),
			},
		)
	}
	return mf
}

// NewRegister creates a new Register object from the provided config
func NewRegister(conf config.Config) *Register {
	r := Register{
		SensorNames: conf.SensorNames,
		Latest:      map[string]event{},
	}
	for k, v := range r.SensorNames {
		r.Latest[v] = event{
			Time: time.Unix(0, 0),
			ID:   fmt.Sprintf("%d", k),
			Name: v,
		}
	}
	return &r
}

func (r *Register) parseSensorName(c int) string {
	name := r.SensorNames[c]
	if name == "" {
		name = fmt.Sprintf("%d", c)
	}
	return name
}

// LogEvent creates a new Event object from a syslog message and adds it to the Register
func (r *Register) LogEvent(log format.LogParts) error {
	data, ok := log["message"].(string)
	if !ok {
		return fmt.Errorf("failed to cast message to string")
	}

	var m message
	err := json.Unmarshal([]byte(data), &m)

	t, err := time.Parse("2006-01-02 15:04:05", m.TimeStr)
	if err != nil {
		return err
	}

	e := event{
		Time:         t,
		ID:           fmt.Sprintf("%d", m.IDInt),
		Name:         r.parseSensorName(m.IDInt),
		Battery:      m.Battery,
		TemperatureC: m.TemperatureC,
		TemperatureF: m.TemperatureC*1.8 + 32,
		Humidity:     m.Humidity,
	}
	logger.InfoMsgf("logging event for %s", e.Name)

	r.Latest[e.Name] = e
	return nil
}
