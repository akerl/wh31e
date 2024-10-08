package listener

import (
	"fmt"

	"github.com/akerl/wh31e/config"
	"github.com/akerl/wh31e/register"

	"github.com/akerl/timber/v2/log"
	"gopkg.in/mcuadros/go-syslog.v2"
)

var logger = log.NewLogger("wh31e.listener")

// Listener defines the syslog engine
type Listener struct {
	SyslogHost string
	SyslogPort int
	Register   *register.Register
	channel    syslog.LogPartsChannel
}

// NewListener creates a new syslog engine from the given config
func NewListener(conf config.Config, reg *register.Register) *Listener {
	return &Listener{
		SyslogHost: conf.SyslogHost,
		SyslogPort: conf.SyslogPort,
		Register:   reg,
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
		err := l.Register.LogEvent(log)
		if err != nil {
			panic(err)
		}
	}
}
