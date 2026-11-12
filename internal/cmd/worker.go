package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/repurpose-ai/re-purpose-be/config"
	"github.com/repurpose-ai/re-purpose-be/internal/job"
	"github.com/repurpose-ai/re-purpose-be/internal/llm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runWorker = &cobra.Command{
	Use:   "worker",
	Short: "run background worker",
	Long:  "Start the asynq worker for asynchronous repurposing jobs.",
	Run:   worker,
}

func init() {
	RootCmd.AddCommand(runWorker)
}

func worker(_ *cobra.Command, _ []string) {
	options, err := asynq.ParseRedisURI(config.RedisWorkerHost())
	if err != nil {
		logrus.WithError(err).Fatal("failed to parse redis worker url")
	}

	server := asynq.NewServer(options, asynq.Config{
		Concurrency: config.WorkerConcurrency(),
	})

	mux := asynq.NewServeMux()
	job.NewProcessor(llm.NewFromConfig()).Register(mux)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logrus.Warn("worker shutdown signal received")
		server.Shutdown()
	}()

	logrus.WithFields(logrus.Fields{
		"redis":       config.RedisWorkerHost(),
		"concurrency": config.WorkerConcurrency(),
		"namespace":   config.WorkerNamespace(),
	}).Info("worker listening ⚙️")

	if err := server.Run(mux); err != nil {
		logrus.WithError(err).Fatal("worker failed")
	}
}
