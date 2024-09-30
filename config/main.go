package config

import (
	"io/ioutil"

	"github.com/akerl/timber/v2/log"
	"github.com/ghodss/yaml"
)

var logger = log.NewLogger("wh31e.config")

// DefaultConfigPath defines the default location to load config from
const DefaultConfigPath = "wh31e.conf"

// Config defines the available configuration options
type Config struct {
	SyslogHost  string         `json:"syslog_host"`
	SyslogPort  int            `json:"syslog_port"`
	Port        int            `json:"port"`
	SensorNames map[int]string `json:"sensor_names"`
}

// LoadConfig creates a config from a file path, using the default if none is provided
func LoadConfig(customPath string) (Config, error) {
	var c Config

	path := customPath
	if path == "" {
		path = DefaultConfigPath
	}
	logger.InfoMsgf("loading config from path: %s", path)

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(contents, &c)
	logger.DebugMsgf("loaded config: %v+", c)
	return c, err
}
