package register

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/akerl/wh31e/config"
	"github.com/akerl/wh31e/metrics"

	"gopkg.in/mcuadros/go-syslog.v2/format"
)

type message struct {
	TimeStr      string  `json:"time"`
	IDInt        int     `json:"id"`
	ChannelInt   int     `json:"channel"`
	Battery      int     `json:"battery_ok"`
	TemperatureC float64 `json:"temperature_C"`
	Humidity     int     `json:"humidity"`
}

type event struct {
	Time         time.Time
	ID           string
	Name         string
	Channel      string
	Battery      int
	TemperatureC float64
	TemperatureF float64
	Humidity     int
}

type counter struct {
	Time time.Time
	Name string
}

// Register defines the shared event tracking state
type Register struct {
	SensorNames map[int]string
	Latest      map[string]event
	Counters    []counter
}

func (e *event) Counter() counter {
	return counter{
		Time: e.Time,
		Name: e.Name,
	}
}

func (e *event) Tags() map[string]string {
	return map[string]string{
		"name":    e.Name,
		"id":      e.ID,
		"channel": e.Channel,
	}
}

// Metrics returns Metric-formatted entries for this event
func (e *event) Metrics() metrics.MetricFile {
	t := e.Tags()
	return metrics.MetricFile{
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
	}
}

// NewRegister creates a new Register object from the provided config
func NewRegister(conf config.Config) *Register {
	return &Register{
		SensorNames: conf.SensorNames,
	}
}

func (r *Register) parseChannelName(c int) string {
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
		Name:         r.parseChannelName(m.IDInt),
		Channel:      fmt.Sprintf("%d", m.ChannelInt),
		Battery:      m.Battery,
		TemperatureC: m.TemperatureC,
		TemperatureF: m.TemperatureC*1.8 + 32,
		Humidity:     m.Humidity,
	}

	r.prune()

	r.Latest[e.Name] = e
	r.Counters = append(r.Counters, e.Counter())
	return nil
}

// CounterMetrics provides Metrics objects for the active counters
func (r *Register) CounterMetrics() metrics.MetricFile {
	counts := map[string]int{}
	for _, v := range r.SensorNames {
		counts[v] = 0
	}
	for _, v := range r.Counters {
		counts[v.Name]++
	}
	mf := metrics.MetricFile{}
	for k, v := range counts {
		m := metrics.Metric{
			Name:  "wh31e_events_last_hour",
			Type:  "gauge",
			Tags:  map[string]string{"name": k},
			Value: fmt.Sprintf("%d", v),
		}
		mf = append(mf, m)
	}
	return mf
}

func (r *Register) prune() {
	var breakPoint int
	checkPoint := time.Now().Add(time.Hour * -1)
	stalePoint := time.Now().Add(time.Minute * -5)
	seen := map[string]struct{}{}
	for index, val := range r.Counters {
		if breakPoint == 0 && val.Time.After(checkPoint) {
			breakPoint = index
		}
		if breakPoint != 0 && val.Time.After(stalePoint) {
			seen[val.Name] = struct{}{}
		}
	}
	r.Counters = r.Counters[0:]
	for k := range r.Latest {
		if _, ok := seen[k]; !ok {
			delete(r.Latest, k)
		}
	}
}
