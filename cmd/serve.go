package cmd

import (
	"github.com/akerl/wh31e/config"
	"github.com/akerl/wh31e/listener"

	"github.com/akerl/metrics/server"
	"github.com/spf13/cobra"
)

func serveRunner(_ *cobra.Command, args []string) error {
	var configPath string
	if len(args) > 0 {
		configPath = args[0]
	}

	conf, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	cache := server.Cache{}

	l := listener.NewListener(conf, &cache)
	s := server.NewServer(conf.Port, &cache)

	err = l.RunAsync()
	if err != nil {
		return err
	}
	return s.Run()
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run web server to serve metrics",
	RunE:  serveRunner,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
