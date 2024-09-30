package metrics

import (
	"fmt"
	"regexp"
	"strings"
)

// Metric defines a single metric point
type Metric struct {
	Name  string            `json:"name"`
	Type  string            `json:"type"`
	Tags  map[string]string `json:"tags"`
	Value string            `json:"value"`
}

// MetricFile defines a set of Metrics
type MetricFile []Metric

var textRegex = regexp.MustCompile(`^[\w\-/]+$`)
var valueRegex = regexp.MustCompile(`^\d+(.\+)?$`)

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
		return false
	}
	if !textRegex.MatchString(m.Type) {
		return false
	}
	if !valueRegex.MatchString(m.Value) {
		return false
	}
	for k, v := range m.Tags {
		if !textRegex.MatchString(k) {
			return false
		}
		if !textRegex.MatchString(v) {
			return false
		}
	}
	return true
}

// String formats the set of Metrics into Prometheus text format
func (mf *MetricFile) String() string {
	var sb strings.Builder
	for _, x := range *mf {
		sb.WriteString(x.String())
	}
	return sb.String()
}

// Validate confirms that the Metrics fields are set with valid values
func (mf *MetricFile) Validate() bool {
	for _, x := range *mf {
		if !x.Validate() {
			return false
		}
	}
	return true
}
