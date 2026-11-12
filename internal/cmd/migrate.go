package cmd

import (
	"strconv"

	"github.com/repurpose-ai/re-purpose-be/internal/database"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate database",
	Long:  "Apply database migrations from internal/database/migrations.",
	Run:   processMigration,
}

func init() {
	migrateCmd.PersistentFlags().Int("step", 0, "maximum migration steps (0 = all)")
	migrateCmd.PersistentFlags().String("direction", "up", "migration direction: up | down")
	RootCmd.AddCommand(migrateCmd)
}

func processMigration(cmd *cobra.Command, _ []string) {
	direction := cmd.Flag("direction").Value.String()

	step, err := strconv.Atoi(cmd.Flag("step").Value.String())
	if err != nil {
		logrus.WithError(err).Fatal("failed to parse migration step")
	}

	db, err := database.InitializePostgresConnection()
	if err != nil {
		logrus.WithError(err).Fatal("failed to connect to PostgreSQL")
	}
	defer closeDatabase(db)

	sqlDB, err := db.DB()
	if err != nil {
		logrus.WithError(err).Fatal("failed to get sql db")
	}

	migrationDirection := migrate.Up
	if direction == "down" {
		migrationDirection = migrate.Down
	}

	source := &migrate.FileMigrationSource{Dir: "internal/database/migrations/"}
	migrate.SetTable("migrations")

	applied, err := migrate.ExecMax(sqlDB, "postgres", source, migrationDirection, step)
	if err != nil {
		logrus.WithError(err).Fatal("failed to migrate database")
	}

	logrus.WithFields(logrus.Fields{
		"applied":   applied,
		"direction": direction,
	}).Info("database migration applied ✅")
}
