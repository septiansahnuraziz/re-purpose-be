package cmd

import (
	"os"

	"github.com/repurpose-ai/re-purpose-be/config"
	"github.com/repurpose-ai/re-purpose-be/internal/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "repurpose-be",
	Short: "RePurpose.ai backend service",
	Long:  "RePurpose.ai backend service: HTTP API, background worker, and database migrations.",
}

func init() {
	config.LoadConfig()
	logger.SetupLogger(logger.Options{
		Level:       config.LogLevel(),
		Format:      config.LogFormat(),
		Environment: config.EnvironmentMode(),
		Service:     config.AppSlugName(),
		Version:     config.AppVersion(),
		Build:       config.AppBuild(),
	})
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
