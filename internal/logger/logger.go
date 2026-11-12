package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Log formats supported by SetupLogger.
const (
	FormatText = "text"
	FormatJSON = "json"
)

// Options configures the global logrus logger.
type Options struct {
	Level       string
	Format      string
	Environment string
	Service     string
	Version     string
	Build       string
}

// SetupLogger configures logrus with a readable formatter (development) or
// JSON output (production), caller information, and context-aware hooks.
func SetupLogger(options Options) {
	format := strings.ToLower(strings.TrimSpace(options.Format))
	if format == "" {
		if isProduction(options.Environment) {
			format = FormatJSON
		} else {
			format = FormatText
		}
	}

	logrus.SetFormatter(newFormatter(format, options.Environment))
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(parseLevel(options.Level))
	logrus.SetReportCaller(true)

	// Keep SetupLogger idempotent when called more than once (tests, embedders).
	logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	logrus.AddHook(&defaultFieldsHook{fields: logrus.Fields{
		"service":     options.Service,
		"version":     options.Version,
		"build":       options.Build,
		"environment": options.Environment,
	}})
	logrus.AddHook(&contextHook{})
}

func newFormatter(format, environment string) logrus.Formatter {
	if format == FormatJSON {
		return &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		}
	}

	return &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   isProduction(environment),
	}
}

func parseLevel(level string) logrus.Level {
	parsed, err := logrus.ParseLevel(strings.ToLower(strings.TrimSpace(level)))
	if err != nil {
		return logrus.InfoLevel
	}
	return parsed
}

func isProduction(environment string) bool {
	return strings.EqualFold(environment, "prod") || strings.EqualFold(environment, "production")
}
