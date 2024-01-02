package log

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// logrusLogger is logrus (github.com/sirupsen/logrus)
type logrusLogger struct {
	*logrus.Entry
	entryForTraceLevel *logrus.Entry
}

// With returns new logger with custom field
func (l *logrusLogger) With(key string, value interface{}) Logger {
	v := convertBoolValue(value)
	return &logrusLogger{l.WithField(key, v), l.entryForTraceLevel.WithField(key, v)}
}

// WithError returns a new logger with error data set
func (l *logrusLogger) WithError(err error) Logger {
	return &logrusLogger{l.Entry.WithError(err), l.entryForTraceLevel.WithError(err)}
}

func newLogrus(debug bool) *logrusLogger {
	log := logrus.New()

	log.SetReportCaller(true)
	log.SetFormatter(parseFormatter(nil, true, true))

	level := logrus.InfoLevel
	if debug {
		level = logrus.DebugLevel
	}

	log.SetLevel(level)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	logger := log.WithFields(map[string]interface{}{
		"hostname": hostname,
	})

	loggerForTraceLevel := logger.WithFields(map[string]interface{}{
		"TraceLevel": "true",
	})

	return &logrusLogger{logger, loggerForTraceLevel}
}

func parseFormatter(formatter logrus.Formatter, disableColors, disableQuote bool) logrus.Formatter {
	if formatter != nil {
		// override disableColors if it is TextFormatter
		if txtFormatter, ok := formatter.(*logrus.TextFormatter); ok {
			txtFormatter.DisableColors = disableColors
		}
		return formatter
	}
	return &logrus.TextFormatter{
		DisableColors: disableColors,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			funcname := "none"
			if funcnames := strings.Split(f.Function, "."); len(funcnames) > 0 {
				funcname = funcnames[len(funcnames)-1]
			}
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", funcname), filename
		},
		DisableQuote: disableQuote,
	}
}

func (l *logrusLogger) Trace(args ...interface{}) {
	l.entryForTraceLevel.Trace(args...)
}

func (l *logrusLogger) Tracef(format string, args ...interface{}) {
	l.entryForTraceLevel.Tracef(format, args...)
}

// convert value to integer if type is (*bool) or bool.
// (default: original value).
func convertBoolValue(value interface{}) interface{} {
	switch valueType := value.(type) {
	case *bool:
		if valueType != nil {
			return boolToInt(*valueType)
		}
	case bool:
		return boolToInt(valueType)
	}
	return value
}

func boolToInt(val bool) int {
	if val {
		return 1
	}
	return 0
}
