package listener

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/akerl/metrics/metrics"
	"github.com/akerl/metrics/server"
	"github.com/akerl/wh31e/config"

	"github.com/akerl/timber/v2/log"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

var logger = log.NewLogger("wh31e.listener")

type message struct {
	TimeStr      string  `json:"time"`
	IDInt        int     `json:"id"`
	Battery      int     `json:"battery_ok"`
	TemperatureC float64 `json:"temperature_C"`
	Humidity     int     `json:"humidity"`
}

// Listener defines the syslog engine
type Listener struct {
	SensorNames   map[int]string
	SyslogHost    string
	SyslogPort    int
	Cache         *server.Cache
	SensorMetrics map[int]metrics.MetricSet
	channel       syslog.LogPartsChannel
}

// NewListener creates a new syslog engine from the given config
func NewListener(conf config.Config, cache *server.Cache) *Listener {
	t := time.Unix(0, 0)

	preload := map[int]metrics.MetricSet{}
	for k, v := range conf.SensorNames {
		tags := map[string]string{
			"name": v,
			"id":   fmt.Sprintf("%d", k),
		}
		preload[k] = metrics.MetricSet{
			metrics.Metric{
				Name:  "wh31e_last_updated",
				Type:  "gauge",
				Tags:  tags,
				Value: fmt.Sprintf("%d", t.Unix()),
			},
		}
	}
	return &Listener{
		SensorNames:   conf.SensorNames,
		SyslogHost:    conf.SyslogHost,
		SyslogPort:    conf.SyslogPort,
		SensorMetrics: preload,
		Cache:         cache,
	}
}

// RunAsync launches the syslog engine in the background
func (l *Listener) RunAsync() error {
	l.channel = make(syslog.LogPartsChannel)
	if err := l.launchSyslogServer(); err != nil {
		return err
	}

	go l.loop()
	return nil
}

func (l *Listener) launchSyslogServer() error {
	server := syslog.NewServer()
	server.SetFormat(syslog.RFC5424)

	handler := syslog.NewChannelHandler(l.channel)
	server.SetHandler(handler)

	host := fmt.Sprintf("%s:%d", l.SyslogHost, l.SyslogPort)
	logger.InfoMsgf("launching syslog listener on %s", host)
	server.ListenUDP(host)

	return server.Boot()
}

func (l *Listener) loop() {
	for log := range l.channel {
		logger.DebugMsgf("received syslog event: %v+", log)
		err := l.logEvent(log)
		if err != nil {
			panic(err)
		}
	}
}

func (l *Listener) logEvent(log format.LogParts) error {
	data, ok := log["message"].(string)
	if !ok {
		return fmt.Errorf("failed to cast message to string")
	}

	var m message
	err := json.Unmarshal([]byte(data), &m)
	if err != nil {
		return err
	}

	logger.InfoMsgf("logging event for %s", m.Name(l.SensorNames))
	l.SensorMetrics[m.IDInt], err = m.Metrics(l.SensorNames)
	if err != nil {
		return err
	}

	mergedMetricSet := metrics.MetricSet{}
	for _, v := range l.SensorMetrics {
		mergedMetricSet = append(mergedMetricSet, v...)
	}
	l.Cache.MetricSet = mergedMetricSet
	return nil
}

func (m message) IDStr() string {
	return fmt.Sprintf("%d", m.IDInt)
}

func (m message) Name(sensorNames map[int]string) string {
	name := sensorNames[m.IDInt]
	if name == "" {
		name = fmt.Sprintf("%d", m.IDInt)
	}
	return name
}

func (m message) Tags(sensorNames map[int]string) map[string]string {
	return map[string]string{
		"name": m.Name(sensorNames),
		"id":   m.IDStr(),
	}
}

func (m message) Metrics(sensorNames map[int]string) (metrics.MetricSet, error) {
	t, err := time.Parse("2006-01-02 15:04:05", m.TimeStr)
	if err != nil {
		return metrics.MetricSet{}, err
	}

	if m.Humidity == 0 {
		logger.InfoMsgf("recieved null message: %v+", m)
		return metrics.MetricSet{}, nil
	}

	tags := m.Tags(sensorNames)

	return metrics.MetricSet{
		metrics.Metric{
			Name:  "wh31e_last_updated",
			Type:  "gauge",
			Tags:  tags,
			Value: fmt.Sprintf("%d", t.Unix()),
		},
		metrics.Metric{
			Name:  "wh31e_humidity",
			Type:  "gauge",
			Tags:  tags,
			Value: fmt.Sprintf("%d", m.Humidity),
		},
		metrics.Metric{
			Name:  "wh31e_battery",
			Type:  "gauge",
			Tags:  tags,
			Value: fmt.Sprintf("%d", m.Battery),
		},
		metrics.Metric{
			Name:  "wh31e_temperature_c",
			Type:  "gauge",
			Tags:  tags,
			Value: fmt.Sprintf("%g", m.TemperatureC),
		},
		metrics.Metric{
			Name:  "wh31e_temperature_f",
			Type:  "gauge",
			Tags:  tags,
			Value: fmt.Sprintf("%g", m.TemperatureC*1.8+32),
		},
	}, nil
}
