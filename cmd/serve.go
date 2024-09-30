package cmd

import (
	"github.com/akerl/wh31e/config"
	"github.com/akerl/wh31e/listener"
	"github.com/akerl/wh31e/register"
	"github.com/akerl/wh31e/server"

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

	reg := register.NewRegister(conf)

	l := listener.NewListener(conf, reg)
	s := server.NewServer(conf, reg)

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
