package metrics

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/akerl/timber/v2/log"
)

var logger = log.NewLogger("wh31e.metrics")

// Metric defines a single metric point
type Metric struct {
	Name  string            `json:"name"`
	Type  string            `json:"type"`
	Tags  map[string]string `json:"tags"`
	Value string            `json:"value"`
}

// MetricFile defines a set of Metrics
type MetricFile struct {
	Metrics []Metric
}

var textRegex = regexp.MustCompile(`^[\w\-/]+$`)
var valueRegex = regexp.MustCompile(`^-?\d+(.\d+)?$`)

// String formats the Metric into Prometheus text format
func (m *Metric) String() string {
	return fmt.Sprintf(
		"# TYPE %s %s\n%s%s %s\n\n",
		m.Name,
		m.Type,
		m.Name,
		m.TagString(),
		m.Value,
	)
}

// TagString formats the Tags on a Metric into Prometheus text format
func (m *Metric) TagString() string {
	if len(m.Tags) == 0 {
		return ""
	}
	tagStrings := []string{}
	for k, v := range m.Tags {
		tagStrings = append(tagStrings, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(tagStrings, ","))
}

// Validate confirms that the Metric fields are set with valid values
func (m *Metric) Validate() bool {
	if !textRegex.MatchString(m.Name) {
		logger.DebugMsgf("invalid metric name: %s", m.Name)
		return false
	}
	if !textRegex.MatchString(m.Type) {
		logger.DebugMsgf("invalid metric type: %s (%s)", m.Type, m.Name)
		return false
	}
	if !valueRegex.MatchString(m.Value) {
		logger.DebugMsgf("invalid metric value: %s (%s)", m.Value, m.Name)
		return false
	}
	for k, v := range m.Tags {
		if !textRegex.MatchString(k) {
			logger.DebugMsgf("invalid metric tag key: %s (%s)", k, m.Name)
			return false
		}
		if !textRegex.MatchString(v) {
			logger.DebugMsgf("invalid metric tag value: %s (%s)", v, m.Name)
			return false
		}
	}
	return true
}

// String formats the set of Metrics into Prometheus text format
func (mf *MetricFile) String() string {
	var sb strings.Builder
	for _, x := range mf.Metrics {
		sb.WriteString(x.String())
	}
	return sb.String()
}

// Validate confirms that the Metrics fields are set with valid values
func (mf *MetricFile) Validate() bool {
	for _, x := range mf.Metrics {
		if !x.Validate() {
			return false
		}
	}
	return true
}
