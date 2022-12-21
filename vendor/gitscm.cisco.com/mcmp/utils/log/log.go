/*
Package log creates and configures a common logger for use across MCMP services.
*/
package log

import (
	"context"
	stdlog "log"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"gitscm.cisco.com/mcmp/utils/ctxutil"
	"gitscm.cisco.com/mcmp/utils/env"
)

var (
	initialize sync.Once
	_logger    logrus.FieldLogger
)

// RFC3339NanoFixed is time.RFC3339Nano with nanoseconds padded using zeros to
// ensure the formatted time is always the same number of characters.
const RFC3339NanoFixed = "2006-01-02T15:04:05.000000000Z07:00"

// GetLogger retrieves the current logger or creates a new one if needed.
func GetLogger(hooks ...logrus.Hook) logrus.FieldLogger {
	checkLogger(hooks...)

	return _logger
}

// SetLevel sets the level at which log messages are published/written.
func SetLevel(level string) {
	checkLogger()

	// If there's no explicit logging level specified, set the level to INFO
	if level == "" {
		level = "info"
	}

	loglevel, err := logrus.ParseLevel(level)
	if err == nil {
		// set default logger and the custom logger levels
		logrus.SetLevel(loglevel)
		_logger.(*logrus.Entry).Logger.SetLevel(loglevel)
	}
}

// SetService adds the Service field to each log message if name provided.
func SetService(name string) {
	checkLogger()

	if name != "" {
		_logger = _logger.WithField("Service", name)
	}
}

// Logger configures a logger instance using the provided context.
func Logger(ctx context.Context) logrus.FieldLogger {
	checkLogger()

	l := _logger

	if ctx == nil {
		return l
	}

	if reqID := ctxutil.RequestID(ctx); reqID.String() != "" {
		l = l.WithField("CorrelationId", reqID.String())
	}

	if principal := ctxutil.Principal(ctx); principal != "" {
		l = l.WithField("Actor", principal)
	}

	if clientID := ctxutil.ClientID(ctx); clientID != "" {
		l = l.WithField("CliendID", clientID)
	}

	if tenantID := ctxutil.TenantID(ctx); tenantID.String() != "" {
		l = l.WithField("TenantID", tenantID.String())
	}

	if sgID := ctxutil.ServiceGroupID(ctx); sgID.String() != "" {
		l = l.WithField("ServiceGroupID", sgID.String())
	}

	if acctID := ctxutil.AccountID(ctx); acctID.String() != "" {
		l = l.WithField("AccountID", acctID)
	}

	return l
}

func checkLogger(hooks ...logrus.Hook) {
	if _logger == nil {
		newLogger(hooks...)
	}
}

func newLogger(hooks ...logrus.Hook) {
	l := logrus.New()
	// configure the default logger to include timestamps and quote empty fields to make visually
	// seeing an empty Field easier. These configurations will not impact or influence the
	// configuration of the logstash hook below.
	l.Formatter = &logrus.TextFormatter{
		TimestampFormat:  RFC3339NanoFixed,
		FullTimestamp:    true,
		QuoteEmptyFields: true,
	}

	// Add each of the supplied hooks to the logger
	for _, h := range hooks {
		// add the hook to the logger
		l.Hooks.Add(h)
	}

	_logger = l.WithFields(captureFields())

	// configure standard logger to use the configured logger to write log entries
	initialize.Do(func() {
		// disable any flags that result in a prefix to the message
		// otherwise there will be duplicate timestamps, etc.
		stdlog.SetFlags(0)
		// use WriterLevel to define at what level library code messages should
		// be included into the logs. Most of the time the messages should be silent
		// unless additional diagnostics are required.
		stdlog.SetOutput(_logger.(*logrus.Entry).WriterLevel(logrus.DebugLevel))
	})

	_logger.Debug("Logger initialized")
}

func captureFields() logrus.Fields {
	return logrus.Fields{
		"container":    os.Getenv("HOSTNAME"),
		"NodeName":     viper.GetString(env.NodeName),
		"PodNamespace": viper.GetString(env.PodNS),
		"PodName":      viper.GetString(env.PodName),
		"PodIP":        viper.GetString(env.PodIP),
		"Region":       viper.GetString(env.SvcRegion),
	}
}
